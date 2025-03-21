package repo

import (
	"database/sql"
	"errors"

	"github.com/vizardkill/order-management/internal/domain"
)

// ProductRepository maneja las operaciones relacionadas con los productos en la base de datos.
type ProductRepository struct {
	DB *sql.DB
}

// NewProductRepository crea una nueva instancia de ProductRepository.
func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{DB: db}
}

// GetAllProducts obtiene todos los productos de la base de datos.
// Retorna una lista de productos y un error en caso de que ocurra algún problema.
func (r *ProductRepository) GetAllProducts() ([]domain.Product, error) {
	// Consulta SQL para obtener todos los productos
	query := "SELECT id, name, price, stock, created_at, updated_at FROM products"

	// Ejecutar la consulta
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, errors.New("error al ejecutar la consulta para obtener productos: " + err.Error())
	}
	defer rows.Close()

	var products []domain.Product

	// Iterar sobre las filas del resultado
	for rows.Next() {
		var p domain.Product

		// Escanear los datos de la fila en la estructura Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt); err != nil {
			// Loguear el error y continuar con la siguiente fila
			// Esto asegura que un error en una fila no detenga el proceso completo
			continue
		}

		// Agregar el producto
		products = append(products, p)
	}

	// Verificar si ocurrió algún error durante la iteración
	if err := rows.Err(); err != nil {
		return nil, errors.New("error al iterar sobre los resultados: " + err.Error())
	}

	// Retornar la lista de productos
	return products, nil
}

func (r *ProductRepository) GetProductByID(id int) (domain.Product, error) {
	// Consulta SQL para obtener un producto por su ID
	query := "SELECT id, name, price, stock, created_at, updated_at FROM products WHERE id = ?"

	// Ejecutar la consulta
	row := r.DB.QueryRow(query, id)

	// Crear una estructura Product para almacenar el resultado
	var p domain.Product

	// Escanear los datos de la fila en la estructura Product
	if err := row.Scan(&p.ID, &p.Name, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return domain.Product{}, errors.New("error al obtener el producto: " + err.Error())
	}

	// Retornar el producto
	return p, nil
}

// ReduceStock reduce el stock de un producto en la base de datos.
func (r *ProductRepository) ReduceStockWithTransaction(tx *sql.Tx, productID int, quantity int) error {
	query := "UPDATE products SET stock = stock - ? WHERE id = ?"
	result, err := tx.Exec(query, quantity, productID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no se pudo reducir el stock, stock insuficiente o producto no encontrado")
	}

	return nil
}

func (r *ProductRepository) UpdateStock(productID int, quantity int) error {
	query := "UPDATE products SET stock = ? WHERE id = ?"
	result, err := r.DB.Exec(query, quantity, productID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no se pudo actualizar el stock o producto no encontrado")
	}

	return nil
}
