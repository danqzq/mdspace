package main

import (
	"log"
	"net/http"

	"github.com/danqzq/mdspace/internal/config"
	"github.com/danqzq/mdspace/internal/handlers"
	"github.com/danqzq/mdspace/internal/router"
	"github.com/danqzq/mdspace/internal/storage"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cfg := config.Load()

	store, err := storage.NewStore(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer store.Close()
	log.Println("Connected to Redis")

	h := handlers.NewHandler(store, cfg.BaseURL)
	r := router.New(h, cfg.StaticDir)

	log.Printf("mdspace server starting on :%s", cfg.Port)
	log.Printf("Base URL: %s", cfg.BaseURL)

	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
