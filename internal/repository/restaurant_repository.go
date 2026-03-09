package repository

import (
	"database/sql"
	"errors"
	"restaurant-api/internal/models"
)

type RestaurantRepository struct {
	db *sql.DB
}

func NewRestaurantRepository(db *sql.DB) *RestaurantRepository {
	return &RestaurantRepository{db: db}
}

func (r *RestaurantRepository) Create(rest *models.Restaurant) error {
	query := `INSERT INTO restaurants (id, name, address, phone, description, admin_id, capacity, created_at, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, rest.Name, rest.Address, rest.Phone, rest.Description, rest.AdminID, rest.Capacity).
		Scan(&rest.ID, &rest.CreatedAt, &rest.UpdatedAt)
}

func (r *RestaurantRepository) FindAll() ([]models.Restaurant, error) {
	rows, err := r.db.Query(`SELECT id, name, address, phone, description, admin_id, capacity, created_at, updated_at FROM restaurants ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (r *RestaurantRepository) FindByID(id string) (*models.Restaurant, error) {
	rest := &models.Restaurant{}
	err := r.db.QueryRow(`SELECT id, name, address, phone, description, admin_id, capacity, created_at, updated_at FROM restaurants WHERE id = $1`, id).
		Scan(&rest.ID, &rest.Name, &rest.Address, &rest.Phone, &rest.Description,
			&rest.AdminID, &rest.Capacity, &rest.CreatedAt, &rest.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return rest, err
}
