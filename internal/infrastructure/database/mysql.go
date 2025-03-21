package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/vizardkill/order-management/config"
)

var DB *sql.DB

func InitMySQL(cfg *config.Config) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Error al conectar con MySQL:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("No se pudo hacer ping a MySQL:", err)
	}

	// Ejecutar migraciones
	RunMigrations(db)

	log.Println("Conexión a MySQL exitosa")
	DB = db
}
