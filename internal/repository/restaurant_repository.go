package repository

import (
	"database/sql"
	"errors"
	"restaurant-api/internal/models"
)

// RestaurantRepository handles all database operations for restaurants.
type RestaurantRepository struct {
	db *sql.DB
}

// NewRestaurantRepository creates a RestaurantRepository with the given database connection.
func NewRestaurantRepository(db *sql.DB) *RestaurantRepository {
	return &RestaurantRepository{db: db}
}

// Create inserts a new restaurant and writes the generated ID and timestamps back into the struct.
func (r *RestaurantRepository) Create(rest *models.Restaurant) error {
	query := `INSERT INTO restaurants (id, name, address, phone, description, admin_id, capacity, created_at, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, rest.Name, rest.Address, rest.Phone, rest.Description, rest.AdminID, rest.Capacity).
		Scan(&rest.ID, &rest.CreatedAt, &rest.UpdatedAt)
}

// FindAll returns every restaurant in the database, ordered from newest to oldest.
// Returns nil (not an empty slice) if there are no restaurants — the handler converts
// this to an empty slice before sending the response.
func (r *RestaurantRepository) FindAll() ([]models.Restaurant, error) {
	rows, err := r.db.Query(`SELECT id, name, address, phone, description, admin_id, capacity, created_at, updated_at FROM restaurants ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Scan each row into a Restaurant and collect the results.
	var restaurants []models.Restaurant
	for rows.Next() {
		var rest models.Restaurant
		if err := rows.Scan(&rest.ID, &rest.Name, &rest.Address, &rest.Phone, &rest.Description,
			&rest.AdminID, &rest.Capacity, &rest.CreatedAt, &rest.UpdatedAt); err != nil {
			return nil, err
		}
		restaurants = append(restaurants, rest)
	}
	return restaurants, nil
}

// FindByID returns the restaurant with the given ID.
// Returns (nil, nil) if no restaurant with that ID exists.
func (r *RestaurantRepository) FindByID(id string) (*models.Restaurant, error) {
	rest := &models.Restaurant{}
	err := r.db.QueryRow(`SELECT id, name, address, phone, description, admin_id, capacity, created_at, updated_at FROM restaurants WHERE id = $1`, id).
		Scan(&rest.ID, &rest.Name, &rest.Address, &rest.Phone, &rest.Description,
			&rest.AdminID, &rest.Capacity, &rest.CreatedAt, &rest.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // Restaurant not found — return nil without an error.
	}
	return rest, err
}
