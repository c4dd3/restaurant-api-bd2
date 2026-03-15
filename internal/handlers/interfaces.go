// internal/handlers/interfaces.go
package handlers

import "restaurant-api/internal/models"

type UserRepository interface {
	FindByEmail(email string) (*models.User, error)
	FindByID(id string) (*models.User, error)
	Create(user *models.User) error
	Update(id string, req *models.UpdateUserRequest) (*models.User, error)
	Delete(id string) error
}

type RestaurantRepository interface {
	Create(rest *models.Restaurant) error
	FindAll() ([]models.Restaurant, error)
	FindByID(id string) (*models.Restaurant, error)
}

type MenuRepository interface {
	Create(menu *models.Menu) error
	FindByID(id string) (*models.Menu, error)
	Update(id string, req *models.UpdateMenuRequest) (*models.Menu, error)
	Delete(id string) error
	FindItemByID(id string) (*models.MenuItem, error)
}

type ReservationRepository interface {
	Create(res *models.Reservation) error
	FindByID(id string) (*models.Reservation, error)
	Cancel(id string) error
	CheckAvailability(restaurantID string, partySize int) (int, error)
}

type OrderRepository interface {
	Create(order *models.Order) error
	FindByID(id string) (*models.Order, error)
}
