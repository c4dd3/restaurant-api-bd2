package repository

import (
	"database/sql"
	"errors"
	"restaurant-api/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	query := `INSERT INTO users (id, name, email, password, role, created_at, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, user.Name, user.Email, user.Password, user.Role).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// FindByEmail retrieves a user from the database by their email address
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	// Initialize an empty User struct to hold the query results
	user := &models.User{}
	// Define the SQL query to select all user fields where email matches
	query := `SELECT id, name, email, password, role, created_at, updated_at FROM users WHERE email = $1`
	// Execute the query with the provided email and scan results into the user struct
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	// Check if no rows were found and return nil instead of an error
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	// Return the user and any other error that occurred
	return user, err
}

func (r *UserRepository) FindByID(id string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, name, email, password, role, created_at, updated_at FROM users WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return user, err
}

func (r *UserRepository) Update(id string, req *models.UpdateUserRequest) (*models.User, error) {
	query := `UPDATE users SET name = COALESCE(NULLIF($1,''), name), email = COALESCE(NULLIF($2,''), email),
		updated_at = NOW() WHERE id = $3
		RETURNING id, name, email, role, created_at, updated_at`
	user := &models.User{}
	err := r.db.QueryRow(query, req.Name, req.Email, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return user, err
}

func (r *UserRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id = $1`, id)
	return err
}
