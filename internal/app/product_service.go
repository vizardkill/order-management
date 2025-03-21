package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/vizardkill/order-management/internal/domain"
	"github.com/vizardkill/order-management/internal/infrastructure/cache"
	"github.com/vizardkill/order-management/internal/infrastructure/repo"
)

type ProductService struct {
	ProductRepo *repo.ProductRepository
	RedisClient *cache.RedisClient
}

func NewProductService(productRepo *repo.ProductRepository, redisClient *cache.RedisClient) *ProductService {
	return &ProductService{ProductRepo: productRepo, RedisClient: redisClient}
}

func (s *ProductService) GetAllProducts() ([]domain.Product, error) {
	return s.ProductRepo.GetAllProducts()
}

func (s *ProductService) UpdateProductStockByID(productID int, newStock int) error {
	// Adquirir un lock para el producto
	lockKey := fmt.Sprintf("lock:product:%d", productID)
	locked, err := s.RedisClient.AcquireLock(lockKey, 5*time.Second)
	if err != nil || !locked {
		return errors.New("no se pudo adquirir el lock para el producto")
	}
	defer s.RedisClient.ReleaseLock(lockKey)

	// Actualizar el stock
	err = s.ProductRepo.UpdateStock(productID, newStock)
	if err != nil {
		return err
	}

	return nil
}
