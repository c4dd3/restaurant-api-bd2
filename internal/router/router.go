package router

import (
	"github.com/gin-gonic/gin"
	"restaurant-api/internal/auth"
	"restaurant-api/internal/handlers"
	"restaurant-api/internal/middleware"
	"restaurant-api/internal/repository"
)

func Setup(
	userRepo *repository.UserRepository,
	restaurantRepo *repository.RestaurantRepository,
	menuRepo *repository.MenuRepository,
	reservationRepo *repository.ReservationRepository,
	orderRepo *repository.OrderRepository,
	jwtSvc *auth.JWTService,
) *gin.Engine {
	// r initializes a new Gin engine with default middleware including Logger and Recovery.
	// The returned *gin.Engine can be used to define routes and start an HTTP server.
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Handlers
	authHandler := handlers.NewAuthHandler(userRepo, jwtSvc)
	userHandler := handlers.NewUserHandler(userRepo)
	restaurantHandler := handlers.NewRestaurantHandler(restaurantRepo)
	menuHandler := handlers.NewMenuHandler(menuRepo, restaurantRepo)
	reservationHandler := handlers.NewReservationHandler(reservationRepo, restaurantRepo)
	orderHandler := handlers.NewOrderHandler(orderRepo, menuRepo, restaurantRepo)

	// Public routes
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)

	// Protected routes
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
