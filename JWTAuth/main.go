package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/SIMCO-TECH/simco/JWTAuth/database"
	"github.com/SIMCO-TECH/simco/JWTAuth/models"
	routes "github.com/SIMCO-TECH/simco/JWTAuth/routes"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ---------- Rate limiting (per client IP) ----------

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type ipRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     rate.Limit
	burst    int
}

func newIPRateLimiter(ctx context.Context, r rate.Limit, burst int) *ipRateLimiter {
	l := &ipRateLimiter{visitors: make(map[string]*visitor), rate: r, burst: burst}
	go l.cleanup(ctx)
	return l
}

func (l *ipRateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	v, ok := l.visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(l.rate, l.burst)}
		l.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter.Allow()
}

// cleanup removes idle IPs so the map can't grow forever.
func (l *ipRateLimiter) cleanup(ctx context.Context) {
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			l.mu.Lock()
			for ip, v := range l.visitors {
				if time.Since(v.lastSeen) > 10*time.Minute {
					delete(l.visitors, ip)
				}
			}
			l.mu.Unlock()
		}
	}
}

// ---------- Middleware ----------

// rateLimit applies a strict limit to login/signup (brute-force protection)
// and a looser one to everything else.
func rateLimit(general, auth *ipRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		limiter := general
		switch c.Request.URL.Path {
		case "/users/login", "/users/signup":
			limiter = auth
		}
		if !limiter.allow(c.ClientIP()) {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests, try again later"})
			return
		}
		c.Next()
	}
}

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		h.Set("Cache-Control", "no-store")
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		}
		c.Next()
	}
}

// limitBody rejects request bodies larger than max bytes.
func limitBody(max int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, max)
		c.Next()
	}
}

// ---------- Startup checks ----------

func mustHaveStrongSecret() {
	s := os.Getenv("SECRET_KEY")
	if len(s) < 32 || s == "your_secret_key_here" {
		log.Fatal("SECRET_KEY must be a random string of at least 32 characters")
	}
}

func main() {
	database.ConnectDB() // also loads .env
	mustHaveStrongSecret()

	if err := database.DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Release mode unless GIN_MODE is set explicitly (no debug route dump, less info leakage)
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	router := gin.New()
	router.HandleMethodNotAllowed = true

	// Don't trust X-Forwarded-For from arbitrary clients (stops IP spoofing
	// that would bypass the rate limiter). If you run behind nginx/Caddy on
	// the same machine, use []string{"127.0.0.1"} instead of nil.
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatal("Failed to set trusted proxies:", err)
	}

	router.Use(
		gin.Logger(),
		gin.Recovery(), // a panic in a handler becomes a 500, not a crash
		securityHeaders(),
		limitBody(1<<20), // 1 MB
		rateLimit(
			newIPRateLimiter(ctx, 10, 20),                        // general: 10 req/s, burst 20
			newIPRateLimiter(ctx, rate.Every(12*time.Second), 5), // login/signup: ~5 per minute
		),
	)

	// Public routes (registered BEFORE UserRoutes, which adds the auth middleware)
	routes.AuthRoutes(router)
	router.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	router.GET("/api-1", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "Hello from API-1"}) })
	router.GET("/api-2", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "Hello from API-2"}) })

	// Protected routes
	routes.UserRoutes(router)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})

	// http.Server with timeouts protects against slow-connection (Slowloris) attacks
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Server failed:", err)
		}
	}()
	log.Println("Server listening on :" + port)

	// Graceful shutdown on Ctrl+C: finish in-flight requests first
	<-ctx.Done()
	log.Println("Shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Println("Forced shutdown:", err)
	}
}
