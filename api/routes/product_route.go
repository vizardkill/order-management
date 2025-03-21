package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/vizardkill/order-management/api/handlers"
	"github.com/vizardkill/order-management/internal/app"
	"github.com/vizardkill/order-management/internal/infrastructure/cache"
	"github.com/vizardkill/order-management/internal/infrastructure/database"
	"github.com/vizardkill/order-management/internal/infrastructure/repo"
)

// RegisterProductRoutes configura las rutas relacionadas con productos.
func RegisterProductRoutes(router *gin.Engine, redis *cache.RedisClient) {
	// Inicializar el repositorio de productos
	productRepo := repo.NewProductRepository(database.DB)

	// Inicializar el servicio de productos
	productService := app.NewProductService(productRepo, redis)

	// Inicializar el manejador de productos
	productHandler := handlers.NewProductHandler(productService, redis)

	// Grupo de rutas para productos
	productRoutes := router.Group("/products")
	{
		productRoutes.GET("/", func(c *gin.Context) {
			productHandler.GetProducts(c.Writer, c.Request)
		})

		productRoutes.PUT("/:product_id/stock", func(c *gin.Context) {
			productId := c.Param("product_id")
			productHandler.UpdateProductStock(c.Writer, c.Request, productId)
		})
	}
}
