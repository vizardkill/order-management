package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/vizardkill/order-management/api/handlers"
	"github.com/vizardkill/order-management/internal/app"
	"github.com/vizardkill/order-management/internal/infrastructure/cache"
	"github.com/vizardkill/order-management/internal/infrastructure/database"
	"github.com/vizardkill/order-management/internal/infrastructure/repo"
)

// RegisterOrders configura las rutas relacionadas con ordenes.
func RegisterOrders(router *gin.Engine, redis *cache.RedisClient) {
	// Inicializar el repositorio de ordenes
	orderRepo := repo.NewOrderRepository(database.DB)

	// Inicializar el repositorio de productos
	productRepo := repo.NewProductRepository(database.DB)

	// Inicializar el servicio de ordenes
	orderService := app.NewOrderService(orderRepo, productRepo, redis)

	// Inicializar el manejador de ordenes
	orderHandler := handlers.NewOrderHandler(orderService, redis)

	// Grupo de rutas para ordenes
	orderRoutes := router.Group("/orders")
	{
		orderRoutes.GET("/:order_id", func(c *gin.Context) {
			orderId := c.Param("order_id")
			orderHandler.GetOrder(c.Writer, c.Request, orderId)
		})

		orderRoutes.POST("/", func(c *gin.Context) {
			orderHandler.CreateOrder(c.Writer, c.Request)
		})
	}
}
