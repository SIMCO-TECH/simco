package routes

import (
	controller "github.com/SIMCO-TECH/simco/JWTAuth/controllers"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRouter *gin.Engine) {
	incomingRouter.POST("/users/signup", controller.SignUp())
	incomingRouter.POST("/users/login", controller.Login())

}
