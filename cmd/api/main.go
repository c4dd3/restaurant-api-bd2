package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"restaurant-api/internal/auth"
	"restaurant-api/internal/repository"
	"restaurant-api/internal/router"
)

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
	restaurantRepo := repository.NewRestaurantRepository(db)
	menuRepo := repository.NewMenuRepository(db)
	reservationRepo := repository.NewReservationRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	jwtSvc := newJWTService()

	r := setupRouter(userRepo, restaurantRepo, menuRepo, reservationRepo, orderRepo, jwtSvc)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server running on port %s", port)
	return runServer(r.Run, ":"+port)
}

func main() {
	if err := runApp(); err != nil {
		log.Printf("startup/server error: %v", err)
		exitFunc(1)
	}
}
