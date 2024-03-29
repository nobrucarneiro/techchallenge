package api

import (
	"g37-lanchonete/internal/controllers/_api"

	"github.com/gin-gonic/gin"
)

type ApiParams struct {
	CustomerController _api.CustomeController
	ProductController  _api.ProductController
	OrderController    _api.OrderController
}

func NewApi(params ApiParams) *gin.Engine {
	router := gin.Default()
	v1 := router.Group("/v1")
	{
		v1.GET("/customers", params.CustomerController.GetCustomers)
		v1.POST("/customers", params.CustomerController.SaveCustomer)

		v1.GET("/products", params.ProductController.GetProducts)
		v1.POST("/products", params.ProductController.CreateProducts)
		v1.PUT("/products/:id", params.ProductController.UpdateProduct)
		v1.DELETE("/products/:id", params.ProductController.DeleteProduct)

		v1.GET("/orders", params.OrderController.GetAllOrders)
		v1.POST("/orders", params.OrderController.CreateOrder)
		v1.GET("/orders/:id/status", params.OrderController.GetOrderStatus)
		v1.PUT("/orders/:id/status", params.OrderController.UpdateOrderStatus)
		v1.PUT("/orders/:id/payment", params.OrderController.HandleOrderPayment)
	}

	return router
}
