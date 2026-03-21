package repository

import (
	"database/sql"
	"errors"
	"restaurant-api/internal/models"
)

type ReservationRepository struct {
	db *sql.DB
}

func NewReservationRepository(db *sql.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

func (r *ReservationRepository) Create(res *models.Reservation) error {
	query := `INSERT INTO reservations (id, restaurant_id, user_id, date, party_size, status, notes, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, created_at`
	return r.db.QueryRow(query, res.RestaurantID, res.UserID, res.Date, res.PartySize, res.Status, res.Notes).
		Scan(&res.ID, &res.CreatedAt)
}

func (r *ReservationRepository) FindByID(id string) (*models.Reservation, error) {
	res := &models.Reservation{}
	err := r.db.QueryRow(
		`SELECT id, restaurant_id, user_id, date, party_size, status, notes, created_at
		FROM reservations WHERE id = $1`, id,
	).Scan(&res.ID, &res.RestaurantID, &res.UserID, &res.Date, &res.PartySize, &res.Status, &res.Notes, &res.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return res, err
}

func (r *ReservationRepository) Cancel(id string) error {
	_, err := r.db.Exec(`UPDATE reservations SET status = 'cancelled' WHERE id = $1`, id)
	return err
}

func (r *ReservationRepository) CheckAvailability(restaurantID string, partySize int) (int, error) {
	var capacity int
	err := r.db.QueryRow(`SELECT capacity FROM restaurants WHERE id = $1`, restaurantID).Scan(&capacity)
	if err != nil {
		return 0, err
	}
	return capacity - partySize, nil
}

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(order *models.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRow(
		`INSERT INTO orders (id, user_id, restaurant_id, reservation_id, total, status, pickup, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, created_at`,
		order.UserID, order.RestaurantID, order.ReservationID, order.Total, order.Status, order.Pickup,
	).Scan(&order.ID, &order.CreatedAt)
	if err != nil {
		return err
	}

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

func (r *OrderRepository) FindByID(id string) (*models.Order, error) {
	order := &models.Order{}
	err := r.db.QueryRow(
		`SELECT id, user_id, restaurant_id, reservation_id, total, status, pickup, created_at
		FROM orders WHERE id = $1`, id,
	).Scan(&order.ID, &order.UserID, &order.RestaurantID, &order.ReservationID,
		&order.Total, &order.Status, &order.Pickup, &order.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(
		`SELECT id, order_id, menu_item_id, quantity, price FROM order_items WHERE order_id = $1`, order.ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.MenuItemID, &item.Quantity, &item.Price); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}
