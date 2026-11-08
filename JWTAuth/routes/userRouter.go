package routes

import (
	controller "github.com/SIMCO-TECH/simco/JWTAuth/controllers"
	middleware "github.com/SIMCO-TECH/simco/JWTAuth/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("/users", controller.GetAllUsers())
	incomingRoutes.GET("/users/:user_id", controller.GetUser())

}
