// Package authrouter builds the Gin engine used by the auth-service.
// Exposing it as a package allows integration tests to spin up both
// services in-process without needing real network calls between them.
package authrouter

import (
	"github.com/gin-gonic/gin"

	"restaurant-api/internal/auth"
	"restaurant-api/internal/handlers"
)

// Setup creates the auth-service router with only two routes:
//
//	POST /auth/register
//	POST /auth/login
//
// All other routes (including JWT-protected endpoints) live in the main API router.
func Setup(userRepo handlers.UserRepository, jwtSvc *auth.JWTService) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "auth", "status": "ok"})
	})

	authHandler := handlers.NewAuthHandler(userRepo, jwtSvc)
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)

	return r
}
