package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/recolab/backend/internal/config"
	"github.com/recolab/backend/internal/gorse"
	"github.com/recolab/backend/internal/handlers"
	"github.com/recolab/backend/internal/store"
)

func main() {
	// Try to load .env from several locations
	for _, envPath := range []string{".env", "../../.env", "../../../.env"} {
		if err := godotenv.Load(envPath); err == nil {
			log.Printf("Loaded env from %s", envPath)
			break
		}
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Determine data directory
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		// Try a few relative paths
		candidates := []string{"data", "../data", "../../data", "../../../data", "/app/data"}
		for _, c := range candidates {
			if info, err := os.Stat(c); err == nil && info.IsDir() {
				dataDir, _ = filepath.Abs(c)
				break
			}
		}
	}
	if dataDir == "" {
		dataDir = "data"
	}
	log.Printf("Using data directory: %s", dataDir)

	// Initialize Gorse client
	gorseClient := gorse.NewClient(cfg)

	// Initialize item store and load metadata
	itemStore := store.NewItemStore()
	if err := itemStore.LoadFromDataDir(dataDir); err != nil {
		log.Printf("Warning: item store load error: %v", err)
	} else {
		log.Printf("Item store loaded: %d items, %d users", itemStore.TotalItems(), itemStore.TotalUsers())
	}

	// Set up Gin
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "tenant_id": cfg.TenantID})
	})

	// Initialize handlers
	newsH := handlers.NewNewsHandler(gorseClient, itemStore, cfg.TenantID)
	productH := handlers.NewProductHandler(gorseClient, itemStore, cfg.TenantID)
	expoH := handlers.NewExpoHandler(gorseClient, itemStore, cfg.TenantID)
	eventH := handlers.NewEventHandler(gorseClient, cfg.TenantID)
	dataH := handlers.NewDataHandler(gorseClient, itemStore, dataDir, cfg.TenantID)

	api := r.Group("/api")
	{
		// Events
		api.POST("/events", eventH.PostEvent)

		demo := api.Group("/demo")
		{
			// Data management
			demo.POST("/load-data", dataH.LoadData)
			demo.GET("/dashboard", dataH.GetDashboard)

			// News
			demo.GET("/trending-news", newsH.GetTrendingNews)
			demo.GET("/personalized-news/:userId", newsH.GetPersonalizedNews)

			// Products
			demo.GET("/products/user/:userId/recommendations", productH.GetProductRecommendations)
			demo.GET("/products/item/:itemId/similar", productH.GetSimilarProducts)

			// Expo
			demo.GET("/expo/buyer/:buyerId/booths", expoH.GetBuyerBoothRecommendations)
			demo.GET("/expo/buyer/:buyerId/suppliers", expoH.GetBuyerSupplierRecommendations)
			demo.GET("/expo/buyer/:buyerId/sessions", expoH.GetRecommendedSessions)
			demo.GET("/expo/seller/:sellerId/buyers", expoH.GetSellerBuyers)
			demo.GET("/expo/trending", expoH.GetTrendingBooths)
		}
	}

	addr := fmt.Sprintf("%s:%d", cfg.BackendHost, cfg.BackendPort)
	log.Printf("RecoLab backend starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
