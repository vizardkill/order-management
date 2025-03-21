package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/vizardkill/order-management/api/routes"
	"github.com/vizardkill/order-management/config"
	"github.com/vizardkill/order-management/internal/infrastructure/cache"
	"github.com/vizardkill/order-management/internal/infrastructure/database"
)

func main() {
	// Cargar configuración
	cfg := config.LoadConfig()

	// Inicializar MySQL y Redis
	database.InitMySQL(cfg)

	// Inicializar el router
	router := gin.Default()

	// Inicializar Redis
	redisClient := cache.NewRedisClient(cfg.RedisHost, cfg.RedisPort, cfg.RedisPass, cfg.RedisDB)

	// Ruta de prueba para verificar el estado de la API
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "API funcionando correctamente"})
	})

	// Registrar las rutas de productos
	routes.RegisterProductRoutes(router, redisClient)

	// Registrar las rutas de ordenes
	routes.RegisterOrders(router, redisClient)

	// Iniciar servidor
	log.Println("Servidor corriendo en el puerto 8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
