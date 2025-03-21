package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/vizardkill/order-management/internal/app"
	"github.com/vizardkill/order-management/internal/infrastructure/cache"
)

type ProductHandler struct {
	ProductService *app.ProductService
	RedisClient    *cache.RedisClient
	Validator      *validator.Validate
}

type PutProductStockRequest struct {
	NewStock int `json:"new_stock" validate:"required,min=0"`
}

func NewProductHandler(productService *app.ProductService, redisClient *cache.RedisClient) *ProductHandler {
	return &ProductHandler{ProductService: productService, RedisClient: redisClient, Validator: validator.New()}
}

// GET /products
func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.ProductService.GetAllProducts()
	if err != nil {
		http.Error(w, "Error obteniendo productos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// PUT /products/{product_id}/stock
func (h *ProductHandler) UpdateProductStock(w http.ResponseWriter, r *http.Request, productId string) {
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
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	// Marcar la solicitud como IN_PROGRESS en Redis
	err = h.RedisClient.SetIdempotencyKey(idempotencyKey, cache.IdempotencyData{
		Status:   "IN_PROGRESS",
		Response: "",
	}, 24*time.Hour)
	if err != nil {
		http.Error(w, "Error configurando la clave de idempotencia", http.StatusInternalServerError)
		return
	}

	var data PutProductStockRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&data); err != nil {
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
	if err := h.Validator.Struct(data); err != nil {
		delErr := h.RedisClient.DeleteIdempotencyKey(idempotencyKey)
		if delErr != nil {
			http.Error(w, "Error eliminando la clave de idempotencia: "+delErr.Error(), http.StatusInternalServerError)
			return
		}

		http.Error(w, "Estructura de datos inválidos: "+err.Error(), http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(productId)
	if err != nil {
		delErr := h.RedisClient.DeleteIdempotencyKey(idempotencyKey)
		if delErr != nil {
			http.Error(w, "Error eliminando la clave de idempotencia: "+delErr.Error(), http.StatusInternalServerError)
			return
		}

		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	if err := h.ProductService.UpdateProductStockByID(id, data.NewStock); err != nil {
		delErr := h.RedisClient.DeleteIdempotencyKey(idempotencyKey)
		if delErr != nil {
			http.Error(w, "Error eliminando la clave de idempotencia: "+delErr.Error(), http.StatusInternalServerError)
			return
		}

		http.Error(w, "Error actualizando el stock del producto", http.StatusInternalServerError)
		return
	}

	// Marcar la solicitud como COMPLETED en Redis
	err = h.RedisClient.SetIdempotencyKey(idempotencyKey, cache.IdempotencyData{
		Status:   "COMPLETED",
		Response: string(productId),
	}, 24*time.Hour)
	if err != nil {
		http.Error(w, "Error almacenando la clave de idempotencia", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
