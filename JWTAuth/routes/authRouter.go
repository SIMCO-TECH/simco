package routes

import (
	controller "simco/JWTAuth/controllers"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRouter *gin.Engine) {
	incomingRouter.POST("users/signup", controller.Signup())
	incomingRouter.POST("users/login", controller.Login())
	
}