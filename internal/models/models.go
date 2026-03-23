package models

import "time"

// User role constants — used to control access throughout the API.
const (
	RoleClient = "client" // Regular user who can make reservations and place orders.
	RoleAdmin  = "admin"  // Administrator who can manage restaurants, menus, and all users.
)

// Status constants shared by reservations and orders.
const (
	StatusPending   = "pending"   // Newly created, waiting for confirmation.
	StatusConfirmed = "confirmed" // Confirmed by the restaurant or system.
	StatusCancelled = "cancelled" // Cancelled by the user or an admin.
)

// User represents a registered account in the system.
// The Password field is excluded from JSON responses (json:"-") so it is never sent to clients.
type User struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"` // Stored as a bcrypt hash, never exposed in responses.
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Restaurant represents a restaurant registered in the system.
// AdminID links the restaurant to the user who created and manages it.
type Restaurant struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Address     string    `json:"address" db:"address"`
	Phone       string    `json:"phone" db:"phone"`
	Description string    `json:"description" db:"description"`
	AdminID     string    `json:"admin_id" db:"admin_id"` // ID of the user who owns this restaurant.
	Capacity    int       `json:"capacity" db:"capacity"` // Total number of seats available.
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Menu represents a menu belonging to a restaurant.
// Items is omitted from JSON when empty (omitempty) to keep responses clean.
type Menu struct {
	ID           string     `json:"id" db:"id"`
	RestaurantID string     `json:"restaurant_id" db:"restaurant_id"`
	Name         string     `json:"name" db:"name"`
	Description  string     `json:"description" db:"description"`
	Items        []MenuItem `json:"items,omitempty"` // The dishes listed in this menu.
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// MenuItem represents a single dish within a menu.
// Available indicates whether the item can currently be ordered.
type MenuItem struct {
	ID          string  `json:"id" db:"id"`
	MenuID      string  `json:"menu_id" db:"menu_id"`
	Name        string  `json:"name" db:"name"`
	Description string  `json:"description" db:"description"`
	Price       float64 `json:"price" db:"price"`
	Available   bool    `json:"available" db:"available"` // False means the item is temporarily unavailable.
}

// Reservation represents a table booking made by a user at a restaurant.
type Reservation struct {
	ID           string    `json:"id" db:"id"`
	RestaurantID string    `json:"restaurant_id" db:"restaurant_id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Date         time.Time `json:"date" db:"date"`
	PartySize    int       `json:"party_size" db:"party_size"` // Number of people in the booking.
	Status       string    `json:"status" db:"status"`         // One of: pending, confirmed, cancelled.
	Notes        string    `json:"notes" db:"notes"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Order represents a food order placed by a user at a restaurant.
// ReservationID is optional — an order can exist without a reservation (e.g. takeaway).
// Pickup indicates whether the order is for collection rather than table service.
type Order struct {
	ID            string      `json:"id" db:"id"`
	UserID        string      `json:"user_id" db:"user_id"`
	RestaurantID  string      `json:"restaurant_id" db:"restaurant_id"`
	ReservationID *string     `json:"reservation_id,omitempty" db:"reservation_id"` // nil if not linked to a reservation.
	Items         []OrderItem `json:"items,omitempty"`
	Total         float64     `json:"total" db:"total"`   // Sum of (price × quantity) for all items.
	Status        string      `json:"status" db:"status"` // One of: pending, confirmed, cancelled.
	Pickup        bool        `json:"pickup" db:"pickup"` // True if the customer will collect the order.
	CreatedAt     time.Time   `json:"created_at" db:"created_at"`
}

// OrderItem represents a single line in an order (one menu item and its quantity).
// Price is stored at the time of the order so it is not affected by future price changes.
type OrderItem struct {
	ID         string  `json:"id" db:"id"`
	OrderID    string  `json:"order_id" db:"order_id"`
	MenuItemID string  `json:"menu_item_id" db:"menu_item_id"`
	Quantity   int     `json:"quantity" db:"quantity"`
	Price      float64 `json:"price" db:"price"` // Snapshot of the item price at the time of ordering.
}

// RegisterRequest is the body expected by POST /auth/register.
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=client admin"`
}

// LoginRequest is the body expected by POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse is returned after a successful login or registration.
// It contains the JWT token and the full user profile.
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// UpdateUserRequest is the body expected by PATCH /users/:id.
// All fields are optional — only provided fields will be updated.
type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CreateRestaurantRequest is the body expected by POST /restaurants.
type CreateRestaurantRequest struct {
	Name        string `json:"name" binding:"required"`
	Address     string `json:"address" binding:"required"`
	Phone       string `json:"phone" binding:"required"`
	Description string `json:"description"`
	Capacity    int    `json:"capacity" binding:"required,min=1"`
}

// CreateMenuRequest is the body expected by POST /menus.
// Items is optional — a menu can be created without dishes and updated later.
type CreateMenuRequest struct {
	RestaurantID string            `json:"restaurant_id" binding:"required"`
	Name         string            `json:"name" binding:"required"`
	Description  string            `json:"description"`
	Items        []MenuItemRequest `json:"items"`
}

// MenuItemRequest describes a single dish when creating or updating a menu.
type MenuItemRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"` // Must be greater than zero.
	Available   bool    `json:"available"`
}

// UpdateMenuRequest is the body expected by PATCH /menus/:id.
// All fields are optional — only provided fields will be updated.
type UpdateMenuRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Items       []MenuItemRequest `json:"items"`
}

// CreateReservationRequest is the body expected by POST /reservations.
type CreateReservationRequest struct {
	RestaurantID string    `json:"restaurant_id" binding:"required"`
	Date         time.Time `json:"date" binding:"required"`
	PartySize    int       `json:"party_size" binding:"required,min=1"`
	Notes        string    `json:"notes"`
}

// CreateOrderRequest is the body expected by POST /orders.
// ReservationID is optional — orders can be standalone (e.g. takeaway).
// Pickup indicates whether the customer will collect the order themselves.
type CreateOrderRequest struct {
	RestaurantID  string             `json:"restaurant_id" binding:"required"`
	ReservationID *string            `json:"reservation_id"`
	Items         []OrderItemRequest `json:"items" binding:"required,min=1"` // At least one item is required.
	Pickup        bool               `json:"pickup"`
}

// OrderItemRequest describes a single item when placing an order.
type OrderItemRequest struct {
	MenuItemID string `json:"menu_item_id" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required,min=1"` // Must order at least 1.
}

// Claims holds the user information extracted from a validated JWT token.
// This is what the auth middleware stores in the Gin context for handlers to read.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}
