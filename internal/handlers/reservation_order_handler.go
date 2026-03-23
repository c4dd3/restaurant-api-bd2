package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"restaurant-api/internal/middleware"
	"restaurant-api/internal/models"
)

// ReservationHandler handles HTTP requests for creating and cancelling reservations.
// It needs the restaurant repo to verify the restaurant exists before booking a table.
type ReservationHandler struct {
	repo           ReservationRepository
	restaurantRepo RestaurantRepository
}

// NewReservationHandler creates a ReservationHandler with the given repositories.
func NewReservationHandler(repo ReservationRepository, restaurantRepo RestaurantRepository) *ReservationHandler {
	return &ReservationHandler{repo: repo, restaurantRepo: restaurantRepo}
}

// Create books a new reservation at a restaurant for the logged-in user.
// Steps: validate request body → confirm restaurant exists → check the restaurant
// has enough seats for the party → read user ID from JWT → save the reservation.
func (h *ReservationHandler) Create(c *gin.Context) {
	// Parse and validate the JSON request body.
	var req models.CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Confirm the restaurant exists before trying to book a table there.
	rest, err := h.restaurantRepo.FindByID(req.RestaurantID)
	if err != nil || rest == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "restaurant not found"})
		return
	}

	// Check that the restaurant has enough free seats for the requested party size.
	// A negative result means there is not enough capacity.
	available, err := h.repo.CheckAvailability(req.RestaurantID, req.PartySize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error checking availability"})
		return
	}
	if available < 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "insufficient capacity for requested party size"})
		return
	}

	// Get the user ID from the JWT claims injected by the auth middleware.
	claims := middleware.ExtractClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Build and save the reservation, linking it to the logged-in user.
	reservation := &models.Reservation{
		RestaurantID: req.RestaurantID,
		UserID:       claims.UserID,
		Date:         req.Date,
		PartySize:    req.PartySize,
		Status:       models.StatusPending,
		Notes:        req.Notes,
	}

	if err := h.repo.Create(reservation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating reservation"})
		return
	}

	c.JSON(http.StatusCreated, reservation)
}

// Cancel marks a reservation as cancelled.
// Only the user who made the reservation or an admin can cancel it.
func (h *ReservationHandler) Cancel(c *gin.Context) {
	id := c.Param("id")

	// Fetch the reservation to verify it exists and to check ownership below.
	reservation, err := h.repo.FindByID(id)
	if err != nil || reservation == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "reservation not found"})
		return
	}

	// Get the caller's identity from the JWT claims.
	claims := middleware.ExtractClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Prevent a regular user from cancelling someone else's reservation.
	// Admins can cancel any reservation.
	if claims.Role != models.RoleAdmin && reservation.UserID != claims.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot cancel another user's reservation"})
		return
	}

	if err := h.repo.Cancel(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error cancelling reservation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "reservation cancelled"})
}

// OrderHandler handles HTTP requests for creating and retrieving orders.
// It needs three repositories: orders (to save/read orders), menus (to validate and
// price each item), and restaurants (to confirm the target restaurant exists).
type OrderHandler struct {
	orderRepo      OrderRepository
	menuRepo       MenuRepository
	restaurantRepo RestaurantRepository
}

// NewOrderHandler creates an OrderHandler with the given repositories.
func NewOrderHandler(orderRepo OrderRepository, menuRepo MenuRepository, restaurantRepo RestaurantRepository) *OrderHandler {
	return &OrderHandler{orderRepo: orderRepo, menuRepo: menuRepo, restaurantRepo: restaurantRepo}
}

// Create places a new order at a restaurant for the logged-in user.
// Steps: validate request body → confirm restaurant exists → read user ID from JWT
// → validate each ordered item (exists and is available) → calculate total price → save the order.
func (h *OrderHandler) Create(c *gin.Context) {
	// Parse and validate the JSON request body.
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Confirm the restaurant exists before placing an order there.
	rest, err := h.restaurantRepo.FindByID(req.RestaurantID)
	if err != nil || rest == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "restaurant not found"})
		return
	}

	// Get the user ID from the JWT claims injected by the auth middleware.
	claims := middleware.ExtractClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Start building the order linked to the logged-in user.
	order := &models.Order{
		UserID:        claims.UserID,
		RestaurantID:  req.RestaurantID,
		ReservationID: req.ReservationID,
		Status:        models.StatusPending,
		Pickup:        req.Pickup,
	}

	// Loop over each requested item: verify it exists, check it is available,
	// accumulate the total price, and add it to the order.
	var total float64
	for _, itemReq := range req.Items {
		menuItem, err := h.menuRepo.FindItemByID(itemReq.MenuItemID)
		if err != nil || menuItem == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "menu item not found: " + itemReq.MenuItemID})
			return
		}
		if !menuItem.Available {
			c.JSON(http.StatusBadRequest, gin.H{"error": "menu item not available: " + menuItem.Name})
			return
		}
		total += menuItem.Price * float64(itemReq.Quantity)
		order.Items = append(order.Items, models.OrderItem{
			MenuItemID: itemReq.MenuItemID,
			Quantity:   itemReq.Quantity,
			Price:      menuItem.Price,
		})
	}
	order.Total = total

	// Persist the completed order to the database.
	if err := h.orderRepo.Create(order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating order"})
		return
	}

	c.JSON(http.StatusCreated, order)
}

// Get returns a single order by the ID provided in the URL parameter.
// Only the user who placed the order or an admin can view it.
func (h *OrderHandler) Get(c *gin.Context) {
	order, err := h.orderRepo.FindByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	// Get the caller's identity from the JWT claims.
	claims := middleware.ExtractClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Prevent a regular user from viewing someone else's order.
	// Admins can access any order.
	if claims.Role != models.RoleAdmin && order.UserID != claims.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot access another user's order"})
		return
	}

	c.JSON(http.StatusOK, order)
}
