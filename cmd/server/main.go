package main

import (
	"log"

	"github.com/chawadev/kalinga-backend/internal/ai"
	"github.com/chawadev/kalinga-backend/internal/auth"
	"github.com/chawadev/kalinga-backend/internal/config"
	"github.com/chawadev/kalinga-backend/internal/database"
	"github.com/chawadev/kalinga-backend/internal/handlers"
	"github.com/chawadev/kalinga-backend/internal/mcp"
	"github.com/chawadev/kalinga-backend/internal/users"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Disconnect()

	// Create database indexes (ignore if already exists)
	if err := database.CreateIndexes(database.Database); err != nil {
		log.Println("Warning: Failed to create indexes:", err)
	}

	// Initialize repositories
	userRepo := users.NewRepository(database.Database)

	// Initialize services
	authService := auth.NewService(userRepo)

	// Initialize AI service
	aiService, err := ai.NewService(cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize AI service: %v", err)
		aiService = nil
	}
	if aiService != nil {
		defer aiService.Close()
	}

	// Initialize MCP server
	mcpServer := mcp.NewServer()

	// Initialize handlers
	chatHandler := handlers.NewChatHandler(aiService, mcpServer)

	// Setup Gin router
	router := gin.Default()

	// Setup routes
	setupRoutes(router, authService, mcpServer, chatHandler)

	log.Printf("Kalinga server listening on http://localhost:%s", cfg.Port)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}

func setupRoutes(router *gin.Engine, authService *auth.Service, mcpServer *mcp.Server, chatHandler *handlers.ChatHandler) {
	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "kalinga-mcp-backend",
			"version": "0.1.0",
		})
	})

	// MCP endpoint
	router.POST("/mcp", mcpServer.HandleRequest)

	// Chat endpoint
	router.POST("/api/chat", chatHandler.Chat)

	// Auth routes (for user authentication)
	router.POST("/api/auth/register", registerHandler(authService))
	router.POST("/api/auth/login", loginHandler(authService))
	router.GET("/api/auth/user", getUserHandler(authService))
}

// Auth handlers
func registerHandler(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req users.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		response, err := authService.Register(c.Request.Context(), &req)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, response)
	}
}

func loginHandler(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req users.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		response, err := authService.Login(c.Request.Context(), &req)
		if err != nil {
			c.JSON(401, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, response)
	}
}

func getUserHandler(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		user, err := authService.GetUser(c.Request.Context(), userID)
		if err != nil {
			c.JSON(404, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, user)
	}
}
