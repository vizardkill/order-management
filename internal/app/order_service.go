package app

import (
	"database/sql"
	"errors"
	"time"

	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/vizardkill/order-management/internal/domain"
	"github.com/vizardkill/order-management/internal/infrastructure/cache"
	"github.com/vizardkill/order-management/internal/infrastructure/repo"
)

type OrderService struct {
	OrderRepo   *repo.OrderRepository
	ProductRepo *repo.ProductRepository
	RedisClient *cache.RedisClient
	Validate    *validator.Validate
}

func NewOrderService(orderRepo *repo.OrderRepository, productRepo *repo.ProductRepository, redisClient *cache.RedisClient) *OrderService {
	return &OrderService{
		OrderRepo:   orderRepo,
		ProductRepo: productRepo,
		RedisClient: redisClient,
		Validate:    validator.New(),
	}
}

// CreateOrder crea una nueva orden y reduce el stock de los productos.
func (s *OrderService) CreateOrder(order domain.CreateOrderService) (domain.Order, error) {
	// Validar datos
	if err := s.Validate.Struct(order); err != nil {
		return domain.Order{}, errors.New("Estructura de datos inválidos: " + err.Error())
	}

	var orderData domain.Order
	var totalAmount float64

	// Obtener los productos de la base de datos
	for _, item := range order.Items {
		product, err := s.ProductRepo.GetProductByID(item.ProductID)
		if err != nil {
			return domain.Order{}, errors.New("Error obteniendo el producto con ID: " + fmt.Sprint(item.ProductID))
		}

		if product.ID == 0 {
			return domain.Order{}, errors.New("Producto no encontrado con ID: " + fmt.Sprint(item.ProductID))
		}

		if product.Stock < item.Quantity {
			return domain.Order{}, errors.New("Stock insuficiente para el producto: " + product.Name)
		}

		// Subtotal
		subtotal := product.Price * float64(item.Quantity)

		// Calcular el total
		totalAmount += subtotal

		// Asignar valores a la estructura de la orden
		orderData.Items = append(orderData.Items, domain.OrderItem{
			ProductID: product.ID,
			Quantity:  item.Quantity,
			Subtotal:  subtotal,
		})
	}

	orderData.CustomerName = order.CustomerName
	orderData.TotalAmount = totalAmount

	// Crear la orden y reducir el stock dentro de una transacción
	createdOrder, err := s.OrderRepo.CreateOrder(orderData, func(tx *sql.Tx) error {
		for _, item := range order.Items {
			// Adquirir un lock para el producto
			lockKey := fmt.Sprintf("lock:product:%d", item.ProductID)
			locked, err := s.RedisClient.AcquireLock(lockKey, 5*time.Second)
			if err != nil || !locked {
				return errors.New("no se pudo adquirir el lock para el producto")
			}
			defer s.RedisClient.ReleaseLock(lockKey)

			// Reducir el stock
			err = s.ProductRepo.ReduceStockWithTransaction(tx, item.ProductID, item.Quantity)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return domain.Order{}, err
	}

	return createdOrder, nil
}

// GetOrder obtiene una orden por id incluyendo sus items.
func (s *OrderService) GetOrderByID(id int) (domain.Order, error) {
	return s.OrderRepo.GetOrderWithItemsByID(id)
}
