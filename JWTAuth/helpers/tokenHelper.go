package helper

import (
	"errors"
	"os"
	"simco/JWTAuth/database"
	"simco/JWTAuth/models"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

type SignedDetails struct {
	Email    string
	Name     string
	UserType string
	UserId   uint64
	jwt.StandardClaims
}

func secretKey() ([]byte, error) {
	key := os.Getenv("SECRET_KEY")
	if key == "" {
		return nil, errors.New("SECRET_KEY environment variable is not set")
	}
	return []byte(key), nil
}

func GenerateAllTokens(email string, name string, userType string, userId uint) (string, string, error) {
	key, err := secretKey()
	if err != nil {
		return "", "", err
	}

	claims := &SignedDetails{
		Email:    email,
		Name:     name,
		UserType: userType,
		UserId:   userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(168 * time.Hour).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
	if err != nil {
		return "", "", err
	}
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(key)
	if err != nil {
		return "", "", err
	}

	return token, refreshToken, nil

}

func ValidateToken(signedToken string) (*SignedDetails, error) {
	key, err := secretKey()
	if err != nil {
		return nil, err
	}

	token, err = jwt.ParseWithClaims(signedToken, &SignedDetails{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return key, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, errors.New("token has expired")
	}
	return claims, nil
}

func UpdateAllTokens(signedToken, signedRefreshToken string, userId uint) error {
	return database.DB.Model(&models.User{}).
		Where("id = ?", userId).
		Updates(map[string]interface{}{
			"token":         signedToken,
			"refresh_token": signedRefreshToken,
		}).Error
}
