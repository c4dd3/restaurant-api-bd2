package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"restaurant-api/internal/auth"
	"restaurant-api/internal/authrouter"
	"restaurant-api/internal/repository"
)

// Dependency injection vars — replaceable in tests to mock infrastructure behavior.
var (
	loadEnv       = godotenv.Load
	newDB         = repository.NewDB
	runMigrations = repository.RunMigrations
	newJWTService = auth.NewJWTService
	runServer     = func(run func(...string) error, addr string) error { return run(addr) }
	runApp        = run
	exitFunc      = os.Exit
)

// run initializes the auth service dependencies and starts its HTTP server.
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

	// Build the user repository and JWT service, then wire them into the auth router.
	userRepo := repository.NewUserRepository(db)
	jwtSvc := newJWTService()
	r := authrouter.Setup(userRepo, jwtSvc)

	// Resolve the listen port from the environment, defaulting to 8081.
	port := os.Getenv("AUTH_PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("🔐 Auth service running on port %s", port)
	return runServer(r.Run, ":"+port)
}

// main is the entry point; it delegates to run and exits with code 1 on failure.
func main() {
	if err := runApp(); err != nil {
		log.Printf("auth-service error: %v", err)
		exitFunc(1)
	}
}
