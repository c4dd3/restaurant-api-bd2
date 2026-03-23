// Package authrouter sets up the HTTP routes for the authentication service.
// This is a separate service that only handles user registration and login.
package authrouter

import (
	"github.com/gin-gonic/gin"

	"restaurant-api/internal/auth"
	"restaurant-api/internal/handlers"
)

// Setup creates and returns the router for the auth service.
// It receives the user repository (to read/write users in the DB) and the JWT service
// (to sign tokens on login). It registers three routes:
//   - GET  /health        → returns a simple status response to confirm the service is up
//   - POST /auth/register → creates a new user account
//   - POST /auth/login    → validates credentials and returns a JWT token
func Setup(userRepo handlers.UserRepository, jwtSvc *auth.JWTService) *gin.Engine {
	r := gin.Default()

	// Simple ping endpoint so external tools can check if this service is running.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "auth", "status": "ok"})
	})

	// Create the auth handler and register the login and register routes.
	// These routes are public — no JWT is required to access them.
	authHandler := handlers.NewAuthHandler(userRepo, jwtSvc)
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)

	return r
}
