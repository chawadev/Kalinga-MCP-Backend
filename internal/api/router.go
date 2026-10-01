package api

import (
	"github.com/chawadev/kalinga-backend/internal/auth"
	"github.com/chawadev/kalinga-backend/internal/handlers"
	"github.com/gin-gonic/gin"
)

type Router struct {
	authHandler *handlers.AuthHandler
}

func NewRouter(authService *auth.Service) *Router {
	return &Router{
		authHandler: handlers.NewAuthHandler(authService),
	}
}

func (r *Router) SetupRoutes(router *gin.Engine) {
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

	// Auth routes
	router.POST("/api/auth/register", r.authHandler.Register)
	router.POST("/api/auth/login", r.authHandler.Login)
	router.GET("/api/auth/user", r.authHandler.GetUser)
}
