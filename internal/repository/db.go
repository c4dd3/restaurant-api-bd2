package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// NewDB opens a PostgreSQL connection using credentials from environment variables
// and verifies the connection with a ping. Falls back to sensible local defaults
// if any variable is unset.
func NewDB() (*sql.DB, error) {
	// Read each connection parameter from the environment, using a local default if absent.
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "restaurant_db")

	// Build the PostgreSQL connection string (DSN).
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// sql.Open only validates the arguments — it does not actually connect yet.
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Ping establishes a real connection and confirms the DB is reachable.
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("cannot connect to DB: %w", err)
	}

	// Tune the connection pool: allow up to 25 open connections and keep 10 idle.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)

	log.Println("Database connected successfully")
	return db, nil
}

// RunMigrations creates all required tables if they do not already exist.
// It is safe to run on every startup — CREATE TABLE IF NOT EXISTS is idempotent.
func RunMigrations(db *sql.DB) error {
	queries := []string{
		// Enable the uuid-ossp extension so uuid_generate_v4() is available for primary keys.
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,

		// users — stores registered accounts. Email must be unique across the system.
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'client',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);`,

		// restaurants — each restaurant is owned by a user (admin_id).
		// Deleting a user cascades and removes their restaurants.
		`CREATE TABLE IF NOT EXISTS restaurants (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			address TEXT NOT NULL,
			phone VARCHAR(50) NOT NULL,
			description TEXT,
			admin_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			capacity INT NOT NULL DEFAULT 50,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);`,

		// menus — each menu belongs to a restaurant.
		// Deleting a restaurant cascades and removes its menus.
		`CREATE TABLE IF NOT EXISTS menus (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			restaurant_id UUID NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);`,

		// menu_items — individual dishes within a menu.
		// Deleting a menu cascades and removes all its items.
		`CREATE TABLE IF NOT EXISTS menu_items (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			menu_id UUID NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			price NUMERIC(10,2) NOT NULL,
			available BOOLEAN DEFAULT TRUE
		);`,

		// reservations — a table booking made by a user at a restaurant.
		// Deleting the user or restaurant cascades and removes their reservations.
		`CREATE TABLE IF NOT EXISTS reservations (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			restaurant_id UUID NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			date TIMESTAMPTZ NOT NULL,
			party_size INT NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			notes TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);`,

		// orders — a food order placed by a user at a restaurant.
		// reservation_id is optional (NULL for standalone/takeaway orders).
		// If the linked reservation is deleted, reservation_id is set to NULL (SET NULL).
		`CREATE TABLE IF NOT EXISTS orders (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			restaurant_id UUID NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
			reservation_id UUID REFERENCES reservations(id) ON DELETE SET NULL,
			total NUMERIC(10,2) NOT NULL DEFAULT 0,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			pickup BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);`,

		// order_items — individual lines within an order.
		// Price is stored here as a snapshot so it is unaffected by future menu price changes.
		`CREATE TABLE IF NOT EXISTS order_items (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
			menu_item_id UUID NOT NULL REFERENCES menu_items(id),
			quantity INT NOT NULL,
			price NUMERIC(10,2) NOT NULL
		);`,
	}

	// Execute each statement in order. Stop and return the error if any statement fails.
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("migration error: %w", err)
		}
	}

	log.Println("Migrations applied successfully")
	return nil
}

// getEnv reads an environment variable by key and returns its value.
// If the variable is not set or is empty, it returns the provided fallback value.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
