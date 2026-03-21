package router

import (
	"github.com/gin-gonic/gin"

	"restaurant-api/internal/auth"
	"restaurant-api/internal/handlers"
	"restaurant-api/internal/middleware"
)

// Setup wires all API routes.
// NOTE: /auth/register and /auth/login are intentionally absent — they are
// owned by the auth-service container (cmd/auth). This service only validates
// JWT tokens that the auth-service issues.
func Setup(
	userRepo handlers.UserRepository,
	restaurantRepo handlers.RestaurantRepository,
	menuRepo handlers.MenuRepository,
	reservationRepo handlers.ReservationRepository,
	orderRepo handlers.OrderRepository,
	jwtSvc *auth.JWTService,
) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "api", "status": "ok"})
	})

	authHandler := handlers.NewAuthHandler(userRepo, jwtSvc)
	userHandler := handlers.NewUserHandler(userRepo)
	restaurantHandler := handlers.NewRestaurantHandler(restaurantRepo)
	menuHandler := handlers.NewMenuHandler(menuRepo, restaurantRepo)
	reservationHandler := handlers.NewReservationHandler(reservationRepo, restaurantRepo)
	orderHandler := handlers.NewOrderHandler(orderRepo, menuRepo, restaurantRepo)

	protected := r.Group("/")
	protected.Use(middleware.Auth(jwtSvc))
	{
		// Users
		protected.GET("/users/me", authHandler.Me)
		protected.PUT("/users/:id", userHandler.Update)
		protected.DELETE("/users/:id", userHandler.Delete)

		// Restaurants
		protected.GET("/restaurants", restaurantHandler.List)
		protected.POST("/restaurants", middleware.AdminOnly(), restaurantHandler.Create)

		// Menus
		protected.POST("/menus", middleware.AdminOnly(), menuHandler.Create)
		protected.GET("/menus/:id", menuHandler.Get)
		protected.PUT("/menus/:id", middleware.AdminOnly(), menuHandler.Update)
		protected.DELETE("/menus/:id", middleware.AdminOnly(), menuHandler.Delete)

		// Reservations
		protected.POST("/reservations", reservationHandler.Create)
		protected.DELETE("/reservations/:id", reservationHandler.Cancel)

		// Orders
		protected.POST("/orders", orderHandler.Create)
		protected.GET("/orders/:id", orderHandler.Get)
	}

	return r
}
