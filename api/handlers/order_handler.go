package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/vizardkill/order-management/internal/app"
	"github.com/vizardkill/order-management/internal/domain"
	"github.com/vizardkill/order-management/internal/infrastructure/cache"
)

type OrderHandler struct {
	OrderService *app.OrderService
	Validator    *validator.Validate
	RedisClient  *cache.RedisClient
}

type CreateOrderItemRequest struct {
	ProductID int `json:"product_id" validate:"required"`
	Quantity  int `json:"quantity" validate:"required,min=1"`
}

type CreateOrderRequest struct {
	CustomerName string                   `json:"customer_name" validate:"required"`
	Items        []CreateOrderItemRequest `json:"items" validate:"required,dive,required"`
}

func NewOrderHandler(orderService *app.OrderService, redisClient *cache.RedisClient) *OrderHandler {
	return &OrderHandler{
		OrderService: orderService,
		Validator:    validator.New(),
		RedisClient:  redisClient,
	}
}

// POST /orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Obtener la clave de idempotencia del encabezado
	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		http.Error(w, "Idempotency-Key es requerido", http.StatusBadRequest)
		return
	}

	// Verificar si la clave ya existe en Redis
	idempotencyData, err := h.RedisClient.GetIdempotencyKey(idempotencyKey)
	if err != nil && err.Error() != "redis: nil" {
		http.Error(w, "Error obteniendo la clave de idempotencia", http.StatusInternalServerError)
		return
	}

	if err == nil {
		if idempotencyData.Status == "IN_PROGRESS" {
			http.Error(w, "Solicitud en progreso", http.StatusConflict)
			return
		}
		if idempotencyData.Status == "COMPLETED" {
			// Si ya está completada, devolver la respuesta almacenada
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(idempotencyData.Response))
			return
		}
	}

	// Marcar la solicitud como IN_PROGRESS en Redis
	err = h.RedisClient.SetIdempotencyKey(idempotencyKey, cache.IdempotencyData{
		Status:   "IN_PROGRESS",
		Response: "",
	}, 24*time.Hour)
	if err != nil {
		delErr := h.RedisClient.DeleteIdempotencyKey(idempotencyKey)
		if delErr != nil {
			http.Error(w, "Error eliminando la clave de idempotencia: "+delErr.Error(), http.StatusInternalServerError)
			return
		}

		http.Error(w, "Error configurando la clave de idempotencia", http.StatusInternalServerError)
		return
	}

	// Leer el cuerpo de la solicitud
	var order CreateOrderRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&order); err != nil {
		if delErr := h.RedisClient.DeleteIdempotencyKey(idempotencyKey); delErr != nil {
			http.Error(w, "Error eliminando la clave de idempotencia: "+delErr.Error(), http.StatusInternalServerError)
			return
		}

		if _, ok := err.(*json.UnmarshalTypeError); ok {
			http.Error(w, "Error en el formato de los datos enviados: "+err.Error(), http.StatusBadRequest)
			return
		}

		if strings.HasPrefix(err.Error(), "json: unknown field") {
			http.Error(w, "Se enviaron campos no esperados en el cuerpo de la solicitud: "+err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, "Error decodificando la solicitud: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validar los datos
	if err := h.Validator.Struct(order); err != nil {
		delErr := h.RedisClient.DeleteIdempotencyKey(idempotencyKey)
		if delErr != nil {
			http.Error(w, "Error eliminando la clave de idempotencia: "+delErr.Error(), http.StatusInternalServerError)
			return
		}

		http.Error(w, "Estructura de datos inválidos: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Crear la estructura de la orden para el servicio
	domainOrder := domain.CreateOrderService{
		CustomerName: order.CustomerName,
		Items:        make([]domain.CreateOrderItemService, len(order.Items)),
	}

	for i, item := range order.Items {
		domainOrder.Items[i] = domain.CreateOrderItemService{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	// Crear la orden
	createdOrder, err := h.OrderService.CreateOrder(domainOrder)
	if err != nil {
		delErr := h.RedisClient.DeleteIdempotencyKey(idempotencyKey)
		if delErr != nil {
			http.Error(w, "Error eliminando la clave de idempotencia: "+delErr.Error(), http.StatusInternalServerError)
			return
		}

		http.Error(w, "Error creando la orden: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Serializar la respuesta
	response, err := json.Marshal(createdOrder)
	if err != nil {
		delErr := h.RedisClient.DeleteIdempotencyKey(idempotencyKey)
		if delErr != nil {
			http.Error(w, "Error eliminando la clave de idempotencia: "+delErr.Error(), http.StatusInternalServerError)
			return
		}

		http.Error(w, "Error procesando la respuesta", http.StatusInternalServerError)
		return
	}

	// Marcar la solicitud como COMPLETED en Redis
	err = h.RedisClient.SetIdempotencyKey(idempotencyKey, cache.IdempotencyData{
		Status:   "COMPLETED",
		Response: string(response),
	}, 24*time.Hour)
	if err != nil {
		http.Error(w, "Error almacenando la clave de idempotencia", http.StatusInternalServerError)
		return
	}

	// Devolver la respuesta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

// GET /orders/{order_id}
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request, orderId string) {
	id, err := strconv.Atoi(orderId)
	if err != nil {
		http.Error(w, "ID de orden inválido", http.StatusBadRequest)
		return
	}

	order, err := h.OrderService.GetOrderByID(id)
	if err != nil {
		http.Error(w, "Error obteniendo la orden", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
