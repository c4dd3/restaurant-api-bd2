package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"restaurant-api/internal/auth"
	"restaurant-api/internal/authrouter"
	"restaurant-api/internal/repository"
)

var (
	loadEnv       = godotenv.Load
	newDB         = repository.NewDB
	runMigrations = repository.RunMigrations
	newJWTService = auth.NewJWTService
	runServer     = func(run func(...string) error, addr string) error { return run(addr) }
	runApp        = run
	exitFunc      = os.Exit
)

func run() error {
	_ = loadEnv()

	db, err := newDB()
	if err != nil {
		return err
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		return err
	}

	userRepo := repository.NewUserRepository(db)
	jwtSvc := newJWTService()
	r := authrouter.Setup(userRepo, jwtSvc)

	port := os.Getenv("AUTH_PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("🔐 Auth service running on port %s", port)
	return runServer(r.Run, ":"+port)
}

func main() {
	if err := runApp(); err != nil {
		log.Printf("auth-service error: %v", err)
		exitFunc(1)
	}
}
