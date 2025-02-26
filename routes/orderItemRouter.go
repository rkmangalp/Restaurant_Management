package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/rkmangalp/Restaurant_Management/controllers"
)

func OrderItemRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/ordersItems", controller.GetOrderItems())
	incomingRoutes.GET("/orderItemss/:orderItem_id", controller.GetOrderItem())
	incomingRoutes.GET("/orderItems-order/:order_id", controller.GetOrderItemsByOrder())
	incomingRoutes.POST("orderItems", controller.CreateOrderItem())
	incomingRoutes.PATCH("/orderItems/:orderItem_id", controller.UpdateOrderItem())

}
