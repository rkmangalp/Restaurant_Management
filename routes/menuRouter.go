package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/rkmangalp/Restaurant_Management/controllers"
)

func MenuRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/menus", controller.GetMenus())
	incomingRoutes.GET("/menus/:menu_id", controller.GetMenuByID())
	incomingRoutes.POST("menus", controller.CreateMenu())
	incomingRoutes.PATCH("/menus", controller.UpdateMenu())
	incomingRoutes.DELETE("/menus/:menu_id", controller.DeleteMenu())
}
