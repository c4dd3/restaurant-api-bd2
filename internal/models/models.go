package models

import "time"

const (
	RoleClient = "client"
	RoleAdmin  = "admin"
)

const (
	StatusPending   = "pending"
	StatusConfirmed = "confirmed"
	StatusCancelled = "cancelled"
)

type User struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Restaurant struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Address     string    `json:"address" db:"address"`
	Phone       string    `json:"phone" db:"phone"`
	Description string    `json:"description" db:"description"`
	AdminID     string    `json:"admin_id" db:"admin_id"`
	Capacity    int       `json:"capacity" db:"capacity"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Menu struct {
	ID           string     `json:"id" db:"id"`
	RestaurantID string     `json:"restaurant_id" db:"restaurant_id"`
	Name         string     `json:"name" db:"name"`
	Description  string     `json:"description" db:"description"`
	Items        []MenuItem `json:"items,omitempty"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

type MenuItem struct {
	ID          string  `json:"id" db:"id"`
	MenuID      string  `json:"menu_id" db:"menu_id"`
	Name        string  `json:"name" db:"name"`
	Description string  `json:"description" db:"description"`
	Price       float64 `json:"price" db:"price"`
	Available   bool    `json:"available" db:"available"`
}

type Reservation struct {
	ID           string    `json:"id" db:"id"`
	RestaurantID string    `json:"restaurant_id" db:"restaurant_id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Date         time.Time `json:"date" db:"date"`
	PartySize    int       `json:"party_size" db:"party_size"`
	Status       string    `json:"status" db:"status"`
	Notes        string    `json:"notes" db:"notes"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type Order struct {
	ID            string      `json:"id" db:"id"`
	UserID        string      `json:"user_id" db:"user_id"`
	RestaurantID  string      `json:"restaurant_id" db:"restaurant_id"`
	ReservationID *string     `json:"reservation_id,omitempty" db:"reservation_id"`
	Items         []OrderItem `json:"items,omitempty"`
	Total         float64     `json:"total" db:"total"`
	Status        string      `json:"status" db:"status"`
	Pickup        bool        `json:"pickup" db:"pickup"`
	CreatedAt     time.Time   `json:"created_at" db:"created_at"`
}

type OrderItem struct {
	ID         string  `json:"id" db:"id"`
	OrderID    string  `json:"order_id" db:"order_id"`
	MenuItemID string  `json:"menu_item_id" db:"menu_item_id"`
	Quantity   int     `json:"quantity" db:"quantity"`
	Price      float64 `json:"price" db:"price"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=client admin"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateRestaurantRequest struct {
	Name        string `json:"name" binding:"required"`
	Address     string `json:"address" binding:"required"`
	Phone       string `json:"phone" binding:"required"`
	Description string `json:"description"`
	Capacity    int    `json:"capacity" binding:"required,min=1"`
}

type CreateMenuRequest struct {
	RestaurantID string           `json:"restaurant_id" binding:"required"`
	Name         string           `json:"name" binding:"required"`
	Description  string           `json:"description"`
	Items        []MenuItemRequest `json:"items"`
}

type MenuItemRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Available   bool    `json:"available"`
}

type UpdateMenuRequest struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Items       []MenuItemRequest `json:"items"`
}

type CreateReservationRequest struct {
	RestaurantID string    `json:"restaurant_id" binding:"required"`
	Date         time.Time `json:"date" binding:"required"`
	PartySize    int       `json:"party_size" binding:"required,min=1"`
	Notes        string    `json:"notes"`
}

type CreateOrderRequest struct {
	RestaurantID  string             `json:"restaurant_id" binding:"required"`
	ReservationID *string            `json:"reservation_id"`
	Items         []OrderItemRequest `json:"items" binding:"required,min=1"`
	Pickup        bool               `json:"pickup"`
}

type OrderItemRequest struct {
	MenuItemID string `json:"menu_item_id" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required,min=1"`
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}
