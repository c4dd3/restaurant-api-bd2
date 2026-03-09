package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"restaurant-api/internal/middleware"
	"restaurant-api/internal/models"
	"restaurant-api/internal/repository"
)

type ReservationHandler struct {
	repo           *repository.ReservationRepository
	restaurantRepo *repository.RestaurantRepository
}

func NewReservationHandler(repo *repository.ReservationRepository, restaurantRepo *repository.RestaurantRepository) *ReservationHandler {
	return &ReservationHandler{repo: repo, restaurantRepo: restaurantRepo}
}

// CreateReservation godoc
// @Summary      Create a reservation
// @Tags         reservations
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body models.CreateReservationRequest true "Reservation data"
// @Success      201  {object}  models.Reservation
// @Router       /reservations [post]
func (h *ReservationHandler) Create(c *gin.Context) {
	var req models.CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check restaurant exists
	rest, err := h.restaurantRepo.FindByID(req.RestaurantID)
	if err != nil || rest == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "restaurant not found"})
		return
	}

	// Check capacity
	available, err := h.repo.CheckAvailability(req.RestaurantID, req.PartySize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error checking availability"})
		return
	}
	if available < 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "insufficient capacity for requested party size"})
		return
	}

	claims := middleware.ExtractClaims(c)
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

// CancelReservation godoc
// @Summary      Cancel a reservation
// @Tags         reservations
// @Security     BearerAuth
// @Param        id   path  string  true  "Reservation ID"
// @Success      200  {object}  gin.H
// @Router       /reservations/{id} [delete]
func (h *ReservationHandler) Cancel(c *gin.Context) {
	id := c.Param("id")
	reservation, err := h.repo.FindByID(id)
	if err != nil || reservation == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "reservation not found"})
		return
	}

	claims := middleware.ExtractClaims(c)
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

// OrderHandler

type OrderHandler struct {
	orderRepo    *repository.OrderRepository
	menuRepo     *repository.MenuRepository
	restaurantRepo *repository.RestaurantRepository
}

func NewOrderHandler(orderRepo *repository.OrderRepository, menuRepo *repository.MenuRepository, restaurantRepo *repository.RestaurantRepository) *OrderHandler {
	return &OrderHandler{orderRepo: orderRepo, menuRepo: menuRepo, restaurantRepo: restaurantRepo}
}

// CreateOrder godoc
// @Summary      Place an order
// @Tags         orders
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body models.CreateOrderRequest true "Order data"
// @Success      201  {object}  models.Order
// @Router       /orders [post]
func (h *OrderHandler) Create(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check restaurant exists
	rest, err := h.restaurantRepo.FindByID(req.RestaurantID)
	if err != nil || rest == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "restaurant not found"})
		return
	}

	claims := middleware.ExtractClaims(c)
	order := &models.Order{
		UserID:        claims.UserID,
		RestaurantID:  req.RestaurantID,
		ReservationID: req.ReservationID,
		Status:        models.StatusPending,
		Pickup:        req.Pickup,
	}

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
		lineTotal := menuItem.Price * float64(itemReq.Quantity)
		total += lineTotal
		order.Items = append(order.Items, models.OrderItem{
			MenuItemID: itemReq.MenuItemID,
			Quantity:   itemReq.Quantity,
			Price:      menuItem.Price,
		})
	}
	order.Total = total

	if err := h.orderRepo.Create(order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating order"})
		return
	}

	c.JSON(http.StatusCreated, order)
}

// GetOrder godoc
// @Summary      Get order details
// @Tags         orders
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  string  true  "Order ID"
// @Success      200  {object}  models.Order
// @Router       /orders/{id} [get]
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

	claims := middleware.ExtractClaims(c)
	if claims.Role != models.RoleAdmin && order.UserID != claims.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot access another user's order"})
		return
	}

	c.JSON(http.StatusOK, order)
}
