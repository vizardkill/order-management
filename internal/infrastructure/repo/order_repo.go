package repo

import (
	"database/sql"

	"github.com/vizardkill/order-management/internal/domain"
)

type OrderRepository struct {
	DB *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{DB: db}
}

func (r *OrderRepository) CreateOrder(order domain.Order, reduceStockFunc func(tx *sql.Tx) error) (domain.Order, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return domain.Order{}, err
	}

	query := "INSERT INTO orders (customer_name, total_amount) VALUES (?, ?)"

	result, err := tx.Exec(query, order.CustomerName, order.TotalAmount)
	if err != nil {
		tx.Rollback()
		return domain.Order{}, err
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return domain.Order{}, err
	}

	// Insertar los items de la orden
	for _, item := range order.Items {
		_, err = tx.Exec("INSERT INTO order_items (order_id, product_id, quantity, subtotal) VALUES (?, ?, ?, ?)",
			orderID, item.ProductID, item.Quantity, item.Subtotal)
		if err != nil {
			tx.Rollback()
			return domain.Order{}, err
		}
	}

	// Reducir el stock de los productos
	if err := reduceStockFunc(tx); err != nil {
		tx.Rollback()
		return domain.Order{}, err
	}

	err = tx.Commit()
	if err != nil {
		return domain.Order{}, err
	}

	// Obtener la orden creada con los items
	return r.GetOrderWithItemsByID(int(orderID))
}

// Obtener una orden por su ID y sus items
func (r *OrderRepository) GetOrderWithItemsByID(id int) (domain.Order, error) {
	query := `
        SELECT 
            o.id AS order_id, 
            o.customer_name, 
            o.total_amount, 
            o.created_at, 
            o.updated_at, 
            oi.product_id, 
            oi.quantity, 
            oi.subtotal
        FROM 
            orders o
        LEFT JOIN 
            order_items oi 
        ON 
            o.id = oi.order_id
        WHERE 
            o.id = ?
    `

	rows, err := r.DB.Query(query, id)
	if err != nil {
		return domain.Order{}, err
	}
	defer rows.Close()

	var order domain.Order
	order.Items = []domain.OrderItem{}

	for rows.Next() {
		var item domain.OrderItem
		var orderID int

		// Escanear los datos de la fila
		err := rows.Scan(
			&orderID,
			&order.CustomerName,
			&order.TotalAmount,
			&order.CreatedAt,
			&order.UpdatedAt,
			&item.ProductID,
			&item.Quantity,
			&item.Subtotal,
		)
		if err != nil {
			return domain.Order{}, err
		}

		// Agregar el item a la lista si no es nulo
		if item.ProductID != 0 {
			order.Items = append(order.Items, item)
		}

		// Asignar el ID de la orden (solo una vez)
		order.ID = orderID
	}

	// Verificar si no se encontró la orden
	if order.ID == 0 {
		return domain.Order{}, sql.ErrNoRows
	}

	return order, nil
}
