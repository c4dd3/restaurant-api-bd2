package repository

import (
	"database/sql"
	"errors"
	"restaurant-api/internal/models"
)

// UserRepository handles all database operations for user accounts.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a UserRepository with the given database connection.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user and writes the generated ID and timestamps back into the struct.
// The password stored in user.Password must already be hashed before calling this method.
func (r *UserRepository) Create(user *models.User) error {
	query := `INSERT INTO users (id, name, email, password, role, created_at, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, user.Name, user.Email, user.Password, user.Role).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// FindByEmail looks up a user by their email address.
// Used during login to retrieve the account before verifying the password.
// Returns (nil, nil) if no user with that email exists.
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, name, email, password, role, created_at, updated_at FROM users WHERE email = $1`
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // User not found — return nil without an error.
	}
	return user, err
}

// FindByID looks up a user by their unique ID.
// Used by the /me endpoint to fetch the logged-in user's full profile.
// Returns (nil, nil) if no user with that ID exists.
func (r *UserRepository) FindByID(id string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, name, email, password, role, created_at, updated_at FROM users WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // User not found — return nil without an error.
	}
	return user, err
}

// Update modifies a user's name and/or email.
// COALESCE(NULLIF($1,''), name) means: keep the existing value if an empty string is sent,
// so callers can update just one field without affecting the other.
// The password is intentionally excluded — password changes are not supported here.
// Returns (nil, nil) if no user with that ID exists.
func (r *UserRepository) Update(id string, req *models.UpdateUserRequest) (*models.User, error) {
	query := `UPDATE users SET name = COALESCE(NULLIF($1,''), name), email = COALESCE(NULLIF($2,''), email),
		updated_at = NOW() WHERE id = $3
		RETURNING id, name, email, role, created_at, updated_at`
	user := &models.User{}
	err := r.db.QueryRow(query, req.Name, req.Email, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // User not found — return nil without an error.
	}
	return user, err
}

// Delete removes a user account by ID.
// Deleting a user also cascades and removes their restaurants and reservations
// via the ON DELETE CASCADE constraints defined in the schema.
func (r *UserRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id = $1`, id)
	return err
}
