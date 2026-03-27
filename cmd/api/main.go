package main

// This file contains the main function, which is the entry point of the application. It initializes all dependencies and starts the HTTP server.
// This is the main api compared to the auth service, which has its own main.go in cmd/auth/main.go. The two services are separate and communicate over HTTP, so they have distinct entry points and dependency setups.

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"restaurant-api/internal/auth"
	"restaurant-api/internal/repository"
	"restaurant-api/internal/router"
)

// Dependency injection vars — replaceable in tests to mock infrastructure behavior.
var (
	loadEnv       = godotenv.Load
	newDB         = repository.NewDB
	runMigrations = repository.RunMigrations
	newJWTService = auth.NewJWTService
	setupRouter   = router.Setup
	runServer     = func(run func(...string) error, addr string) error { return run(addr) }
	runApp        = run
	exitFunc      = os.Exit
)

// run initializes all application dependencies and starts the HTTP server.
func run() error {
	// Load environment variables from .env file; ignore error if file is absent.
	_ = loadEnv()

	// Open the database connection.
	db, err := newDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Apply pending schema migrations before accepting traffic.
	if err := runMigrations(db); err != nil {
		return err
	}

	// Build repository instances, each wrapping the shared DB connection.
	userRepo := repository.NewUserRepository(db)
	restaurantRepo := repository.NewRestaurantRepository(db)
	menuRepo := repository.NewMenuRepository(db)
	reservationRepo := repository.NewReservationRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	// Create the JWT service used for token signing and verification.
	jwtSvc := newJWTService()

	// Wire all repositories and services into the HTTP router.
	r := setupRouter(userRepo, restaurantRepo, menuRepo, reservationRepo, orderRepo, jwtSvc)

	// Resolve the listen port from the environment, defaulting to 8080.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	return runServer(r.Run, ":"+port)
}

// main is the entry point; it delegates to run and exits with code 1 on failure.
func main() {
	if err := runApp(); err != nil {
		log.Printf("startup/server error: %v", err)
		exitFunc(1)
	}
}
