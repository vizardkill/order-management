package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	RedisHost  string
	RedisPort  string
	RedisPass  string
	RedisDB    int
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No se pudo cargar el archivo .env, usando valores por defecto")
	}

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "order_management"),
		RedisHost:  getEnv("REDIS_HOST", "localhost"),
		RedisPort:  getEnv("REDIS_PORT", "6379"),
		RedisDB: func() int {
			val, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
			if err != nil {
				log.Printf("Error converting REDIS_DB to int: %v. Using default value 0.", err)
				return 0
			}
			return val
		}(),
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
