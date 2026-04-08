package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port              string
	DBUser            string
	DBPassword        string
	DBHost            string
	DBPort            string
	DBName            string
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	JWTSecretKey      string
	OpenAIApiKey      string
	OpenAIBaseURL     string
	FRONTEND_BASE_URL string
	MinioEndpoint     string
	MinioAccessKey    string
	MinioSecretKey    string
	MinioBucket       string
	MinioUseSSL       bool
}

var App *Config

func LoadConfig() {
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	minioUseSSL, _ := strconv.ParseBool(getEnv("MINIO_USE_SSL", "false"))

	App = &Config{
		Port:              getEnv("PORT", "8080"),
		DBUser:            getEnv("DB_USER", "rag_user"),
		DBPassword:        getEnv("DB_PASSWORD", "rag_password"),
		DBHost:            getEnv("DB_HOST", "127.0.0.1"),
		DBPort:            getEnv("DB_PORT", "3306"),
		DBName:            getEnv("DB_NAME", "rag_db"),
		RedisAddr:         getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		RedisDB:           redisDB,
		JWTSecretKey:      getEnv("JWT_SECRET_KEY", "default_secret_key"),
		OpenAIApiKey:      getEnv("OPENAI_API_KEY", ""),
		OpenAIBaseURL:     getEnv("OPENAI_BASE_URL", ""),
		FRONTEND_BASE_URL: getEnv("FRONTEND_BASE_URL", "http://localhost:3000"),
		MinioEndpoint:     getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey:    getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey:    getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:       getEnv("MINIO_BUCKET", "documents"),
		MinioUseSSL:       minioUseSSL,
	}

	log.Println("Configuration loaded successfully")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
