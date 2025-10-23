package main

import (
	"log"
	"os"

	"spotifybackend/internal/config"
	"spotifybackend/internal/handlers"
	"spotifybackend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// load .env in dev
	_ = godotenv.Load()

	cfg, err := config.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := models.NewGormDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	// auto-migrate (dev)
	if err := models.AutoMigrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	r := gin.Default()

	api := r.Group("/api/v1")
	{
		handlers.RegisterAuthRoutes(api, db, cfg)
		handlers.RegisterUserRoutes(api, db, cfg)
		handlers.RegisterTrackRoutes(api, db, nil, cfg)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
