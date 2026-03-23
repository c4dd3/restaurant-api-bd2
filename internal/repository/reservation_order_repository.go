package repository

import (
	"database/sql"
	"errors"
	"restaurant-api/internal/models"
)

// ReservationRepository handles all database operations for reservations.
type ReservationRepository struct {
	db *sql.DB
}

// NewReservationRepository creates a ReservationRepository with the given database connection.
func NewReservationRepository(db *sql.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

// Create inserts a new reservation and writes the generated ID and timestamp back into the struct.
func (r *ReservationRepository) Create(res *models.Reservation) error {
	query := `INSERT INTO reservations (id, restaurant_id, user_id, date, party_size, status, notes, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, created_at`
	return r.db.QueryRow(query, res.RestaurantID, res.UserID, res.Date, res.PartySize, res.Status, res.Notes).
		Scan(&res.ID, &res.CreatedAt)
}

// FindByID returns the reservation with the given ID.
// Returns (nil, nil) if no reservation with that ID exists.
func (r *ReservationRepository) FindByID(id string) (*models.Reservation, error) {
	res := &models.Reservation{}
	err := r.db.QueryRow(
		`SELECT id, restaurant_id, user_id, date, party_size, status, notes, created_at
		FROM reservations WHERE id = $1`, id,
	).Scan(&res.ID, &res.RestaurantID, &res.UserID, &res.Date, &res.PartySize, &res.Status, &res.Notes, &res.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // Reservation not found — return nil without an error.
	}
	return res, err
}

// Cancel sets the status of a reservation to 'cancelled'.
// It does not delete the row so the booking history is preserved.
func (r *ReservationRepository) Cancel(id string) error {
	_, err := r.db.Exec(`UPDATE reservations SET status = 'cancelled' WHERE id = $1`, id)
	return err
}

// CheckAvailability returns how many seats are still available at the restaurant
// after accounting for the requested party size.
// A negative result means the party is larger than the restaurant's total capacity.
// Note: this is a simple capacity check — it does not account for existing reservations.
func (r *ReservationRepository) CheckAvailability(restaurantID string, partySize int) (int, error) {
	var capacity int
	err := r.db.QueryRow(`SELECT capacity FROM restaurants WHERE id = $1`, restaurantID).Scan(&capacity)
	if err != nil {
		return 0, err
	}
	// Subtract the requested party size from the total capacity.
	// The handler rejects the booking if this value is negative.
	return capacity - partySize, nil
}

// OrderRepository handles all database operations for orders and their items.
type OrderRepository struct {
	db *sql.DB
}

// NewOrderRepository creates an OrderRepository with the given database connection.
func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create inserts a new order and all its items in a single transaction.
// Using a transaction ensures that if any item insertion fails, the whole
// operation is rolled back and no partial order is left in the database.
func (r *OrderRepository) Create(order *models.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	// Rollback is a no-op if Commit has already been called successfully.
	defer tx.Rollback()

	// Insert the order row and read back the generated ID and timestamp.
	err = tx.QueryRow(
		`INSERT INTO orders (id, user_id, restaurant_id, reservation_id, total, status, pickup, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, created_at`,
		order.UserID, order.RestaurantID, order.ReservationID, order.Total, order.Status, order.Pickup,
	).Scan(&order.ID, &order.CreatedAt)
	if err != nil {
		return err
	}

	// Insert each order item, linking it to the newly created order ID.
	// We iterate by index so we can write the generated ID back into the slice element.
	for i := range order.Items {
		item := &order.Items[i]
		err = tx.QueryRow(
			`INSERT INTO order_items (id, order_id, menu_item_id, quantity, price)
			VALUES (uuid_generate_v4(), $1, $2, $3, $4) RETURNING id`,
			order.ID, item.MenuItemID, item.Quantity, item.Price,
		).Scan(&item.ID)
		if err != nil {
			return err
		}
		item.OrderID = order.ID
	}

	return tx.Commit()
}

// FindByID returns the order with the given ID, including all its items.
// Returns (nil, nil) if no order with that ID exists.
func (r *OrderRepository) FindByID(id string) (*models.Order, error) {
	order := &models.Order{}

	// Fetch the order row first.
	err := r.db.QueryRow(
		`SELECT id, user_id, restaurant_id, reservation_id, total, status, pickup, created_at
		FROM orders WHERE id = $1`, id,
	).Scan(&order.ID, &order.UserID, &order.RestaurantID, &order.ReservationID,
		&order.Total, &order.Status, &order.Pickup, &order.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // Order not found — return nil without an error.
	}
	if err != nil {
		return nil, err
	}

	// Fetch all items that belong to this order.
	rows, err := r.db.Query(
		`SELECT id, order_id, menu_item_id, quantity, price FROM order_items WHERE order_id = $1`, order.ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Scan each item row and append it to the order's Items slice.
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.MenuItemID, &item.Quantity, &item.Price); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}
