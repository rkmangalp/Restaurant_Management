package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/rkmangalp/Restaurant_Management/controllers"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/user", controllers.GetUsers())
	incomingRoutes.GET("/users/:user_id", controllers.GetUser())
	incomingRoutes.GET("/users/signup", controllers.SignUp())
	incomingRoutes.GET("/users/login", controllers.Login())
}
