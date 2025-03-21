package database

import (
	"database/sql"
	"fmt"
	"log"
)

// RunMigrations ejecuta la creación de la base de datos y las tablas
func RunMigrations(db *sql.DB) {
	queries := []string{
		`CREATE DATABASE IF NOT EXISTS order_management;`,
		`USE order_management;`,
		`CREATE TABLE IF NOT EXISTS products (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			price DECIMAL(10,2) NOT NULL,
			stock INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS orders (
			id INT AUTO_INCREMENT PRIMARY KEY,
			customer_name VARCHAR(255) NOT NULL,
			total_amount DECIMAL(10,2) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS order_items (
			id INT AUTO_INCREMENT PRIMARY KEY,
			order_id INT NOT NULL,
			product_id INT NOT NULL,
			quantity INT NOT NULL,
			subtotal DECIMAL(10,2) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
			FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
		);`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			log.Fatalf("Error ejecutando la migración: %v", err)
		}
	}

	fmt.Println("Migraciones ejecutadas correctamente ✅")
	
    // Insertar datos de prueba en la tabla products si no existen
    insertTestData(db)
}

// insertTestData inserta datos de prueba en la tabla products si está vacía
func insertTestData(db *sql.DB) {
	// Verificar si la tabla products ya tiene datos
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		log.Fatalf("Error verificando datos en la tabla products: %v", err)
	}

	// Si no hay datos, insertar productos de prueba
	if count == 0 {
		fmt.Println("Insertando datos de prueba en la tabla products...")

		queries := []string{
			`INSERT INTO products (name, price, stock) VALUES ('Producto A', 10.50, 100);`,
			`INSERT INTO products (name, price, stock) VALUES ('Producto B', 20.00, 50);`,
			`INSERT INTO products (name, price, stock) VALUES ('Producto C', 15.75, 200);`,
		}

		for _, query := range queries {
			_, err := db.Exec(query)
			if err != nil {
				log.Fatalf("Error insertando datos de prueba: %v", err)
			}
		}

		fmt.Println("Datos de prueba insertados correctamente ✅")
	} else {
		fmt.Println("La tabla products ya contiene datos, no se insertaron datos de prueba.")
	}
}
