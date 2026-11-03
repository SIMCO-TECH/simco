package routes

import (
	controller "simco/JWTAuth/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.RouterGroup) {
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("/users", controller.GetAllUsers())
	incomingRoutes.GET("/users/:userId", controller.GetUser())

}
