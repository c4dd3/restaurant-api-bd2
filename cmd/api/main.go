package main

import (
	"log"
	"os"

	"restaurant-api/internal/auth"
	"restaurant-api/internal/repository"
	"restaurant-api/internal/router"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env if present
	_ = godotenv.Load()

	// Connect to DB
	db, err := repository.NewDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations ("Migrations" to create tables and extensions if they don't exist)
	if err := repository.RunMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Build repositories ("Repository pattern" for better separation of concerns)
	userRepo := repository.NewUserRepository(db)
	restaurantRepo := repository.NewRestaurantRepository(db)
	menuRepo := repository.NewMenuRepository(db)
	reservationRepo := repository.NewReservationRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	// JWT service
	jwtSvc := auth.NewJWTService()

	// Setup router
	r := router.Setup(userRepo, restaurantRepo, menuRepo, reservationRepo, orderRepo, jwtSvc)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
