package database

import (
	"context"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"hsduc.com/rag/config"
	"hsduc.com/rag/models"
)

var (
	DB    *gorm.DB
	Redis *redis.Client
	Minio *minio.Client
)

func ConnectMySQL() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.App.DBUser,
		config.App.DBPassword,
		config.App.DBHost,
		config.App.DBPort,
		config.App.DBName,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to MySQL:", err)
	}

	log.Println("Connected to MySQL successfully")

	// Migrate models
	err = DB.AutoMigrate(&models.User{}, &models.Conversation{}, &models.Message{}, &models.Document{}, &models.DocumentFile{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
}

func ConnectRedis() {
	Redis = redis.NewClient(&redis.Options{
		Addr:     config.App.RedisAddr,
		Password: config.App.RedisPassword,
		DB:       config.App.RedisDB,
	})

	_, err := Redis.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	log.Println("Connected to Redis successfully")
}

func ConnectMinio() {
	var err error
	Minio, err = minio.New(config.App.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.App.MinioAccessKey, config.App.MinioSecretKey, ""),
		Secure: config.App.MinioUseSSL,
	})
	if err != nil {
		log.Fatal("Failed to connect to MinIO:", err)
	}

	ctx := context.Background()
	exists, err := Minio.BucketExists(ctx, config.App.MinioBucket)
	if err != nil {
		log.Fatal("Failed to check MinIO bucket:", err)
	}
	if !exists {
		if err = Minio.MakeBucket(ctx, config.App.MinioBucket, minio.MakeBucketOptions{}); err != nil {
			log.Fatal("Failed to create MinIO bucket:", err)
		}
		log.Printf("Created MinIO bucket: %s", config.App.MinioBucket)
	}

	log.Println("Connected to MinIO successfully")
}
