package handlers

import "restaurant-api/internal/models"

// UserRepository defines the database operations the handlers need to work with users.
type UserRepository interface {
	FindByEmail(email string) (*models.User, error)          // Look up a user by their email address.
	FindByID(id string) (*models.User, error)                // Look up a user by their unique ID.
	Create(user *models.User) error                          // Insert a new user into the database.
	Update(id string, req *models.UpdateUserRequest) (*models.User, error) // Update an existing user's fields.
	Delete(id string) error                                  // Remove a user from the database.
}

// RestaurantRepository defines the database operations the handlers need to work with restaurants.
type RestaurantRepository interface {
	Create(rest *models.Restaurant) error            // Insert a new restaurant into the database.
	FindAll() ([]models.Restaurant, error)           // Return every restaurant in the database.
	FindByID(id string) (*models.Restaurant, error)  // Look up a single restaurant by its ID.
}

// MenuRepository defines the database operations the handlers need to work with menus and menu items.
type MenuRepository interface {
	Create(menu *models.Menu) error                                        // Insert a new menu into the database.
	FindByID(id string) (*models.Menu, error)                              // Look up a menu by its ID.
	Update(id string, req *models.UpdateMenuRequest) (*models.Menu, error) // Update an existing menu's fields.
	Delete(id string) error                                                // Remove a menu from the database.
	FindItemByID(id string) (*models.MenuItem, error)                      // Look up a single menu item by its ID.
}

// ReservationRepository defines the database operations the handlers need to work with reservations.
type ReservationRepository interface {
	Create(res *models.Reservation) error                                 // Insert a new reservation into the database.
	FindByID(id string) (*models.Reservation, error)                      // Look up a reservation by its ID.
	Cancel(id string) error                                               // Mark a reservation as cancelled.
	CheckAvailability(restaurantID string, partySize int) (int, error)    // Return the number of available seats for the given restaurant and party size.
}

// OrderRepository defines the database operations the handlers need to work with orders.
type OrderRepository interface {
	Create(order *models.Order) error        // Insert a new order into the database.
	FindByID(id string) (*models.Order, error) // Look up an order by its ID.
}
