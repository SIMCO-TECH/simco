package routes

import (
	controller "simco/JWTAuth/controllers"
	middleware "simco/JWTAuth/middleware"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("/users", controller.GetAllUsers())
	incomingRoutes.GET("/users/:user_id", controller.GetUser())

}
