package router

import (
	"github.com/gin-gonic/gin"

	"restaurant-api/internal/auth"
	"restaurant-api/internal/handlers"
	"restaurant-api/internal/middleware"
)

// Setup creates the main API router and registers all routes.
//
// Every route in this router requires a valid JWT token in the
// Authorization header — there are no public routes here.
// Registration and login live in the separate auth service (cmd/auth).
//
// Routes marked with middleware.AdminOnly() can only be called by
// users whose JWT token contains role = "admin".
func Setup(
	userRepo handlers.UserRepository,
	restaurantRepo handlers.RestaurantRepository,
	menuRepo handlers.MenuRepository,
	reservationRepo handlers.ReservationRepository,
	orderRepo handlers.OrderRepository,
	jwtSvc *auth.JWTService,
) *gin.Engine {
	r := gin.Default()

	// Public health check — no token required, used to verify the service is running.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"service": "api", "status": "ok"})
	})

	// Create one handler per resource, injecting the repositories each one needs.
	authHandler := handlers.NewAuthHandler(userRepo, jwtSvc)
	userHandler := handlers.NewUserHandler(userRepo)
	restaurantHandler := handlers.NewRestaurantHandler(restaurantRepo)
	menuHandler := handlers.NewMenuHandler(menuRepo, restaurantRepo)
	reservationHandler := handlers.NewReservationHandler(reservationRepo, restaurantRepo)
	orderHandler := handlers.NewOrderHandler(orderRepo, menuRepo, restaurantRepo)

	// All routes below this line require a valid JWT token.
	// middleware.Auth(jwtSvc) reads the token from the Authorization header,
	// validates it, and stores the user's claims in the request context.
	// If the token is missing or invalid the request is rejected with 401.
	protected := r.Group("/")
	protected.Use(middleware.Auth(jwtSvc))
	{
		// User routes — any authenticated user can view or edit their own account.
		protected.GET("/users/me", authHandler.Me)        // Get the profile of the logged-in user.
		protected.PUT("/users/:id", userHandler.Update)   // Update a user account (own account or admin).
		protected.DELETE("/users/:id", userHandler.Delete) // Delete a user account (own account or admin).

		// Restaurant routes — listing is open to all authenticated users; creating requires admin.
		protected.GET("/restaurants", restaurantHandler.List)
		protected.POST("/restaurants", middleware.AdminOnly(), restaurantHandler.Create) // Admin only.

		// Menu routes — reading is open to all authenticated users; write operations require admin.
		protected.POST("/menus", middleware.AdminOnly(), menuHandler.Create)          // Admin only.
		protected.GET("/menus/:id", menuHandler.Get)
		protected.PUT("/menus/:id", middleware.AdminOnly(), menuHandler.Update)       // Admin only.
		protected.DELETE("/menus/:id", middleware.AdminOnly(), menuHandler.Delete)    // Admin only.

		// Reservation routes — any authenticated user can create or cancel their own reservation.
		protected.POST("/reservations", reservationHandler.Create)
		protected.DELETE("/reservations/:id", reservationHandler.Cancel)

		// Order routes — any authenticated user can place or view their own orders.
		protected.POST("/orders", orderHandler.Create)
		protected.GET("/orders/:id", orderHandler.Get)
	}

	return r
}
