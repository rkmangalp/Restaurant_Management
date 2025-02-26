package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/rkmangalp/Restaurant_Management/controllers"
)

func TableRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/tables", controller.GetTabels())
	incomingRoutes.GET("/tables/:table_id", controller.GetTabel())
	incomingRoutes.POST("tables", controller.CreateTable())
	incomingRoutes.PATCH("/tables/:table_id", controller.UpdateTable())
}
