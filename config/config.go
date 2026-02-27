package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port          string
	DBUser        string
	DBPassword    string
	DBHost        string
	DBPort        string
	DBName        string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	JWTSecretKey  string
}

var App *Config

func LoadConfig() {
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	App = &Config{
		Port:          getEnv("PORT", "8080"),
		DBUser:        getEnv("DB_USER", "rag_user"),
		DBPassword:    getEnv("DB_PASSWORD", "rag_password"),
		DBHost:        getEnv("DB_HOST", "127.0.0.1"),
		DBPort:        getEnv("DB_PORT", "3306"),
		DBName:        getEnv("DB_NAME", "rag_db"),
		RedisAddr:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,
		JWTSecretKey:  getEnv("JWT_SECRET_KEY", "default_secret_key"),
	}

	log.Println("Configuration loaded successfully")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
