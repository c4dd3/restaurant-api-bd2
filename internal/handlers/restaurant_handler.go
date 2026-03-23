package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"restaurant-api/internal/middleware"
	"restaurant-api/internal/models"
)

// RestaurantHandler handles HTTP requests for creating and listing restaurants.
type RestaurantHandler struct {
	repo RestaurantRepository
}

// NewRestaurantHandler creates a RestaurantHandler with the given repository.
func NewRestaurantHandler(repo RestaurantRepository) *RestaurantHandler {
	return &RestaurantHandler{repo: repo}
}

// Create registers a new restaurant.
// Steps: validate the request body → read the admin ID from JWT → build the
// restaurant model → save it to the database.
func (h *RestaurantHandler) Create(c *gin.Context) {
	// Parse and validate the JSON request body.
	var req models.CreateRestaurantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the caller's identity from the JWT claims injected by the auth middleware.
	claims := middleware.ExtractClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Build the restaurant model, assigning the logged-in user as its admin.
	rest := &models.Restaurant{
		Name:        req.Name,
		Address:     req.Address,
		Phone:       req.Phone,
		Description: req.Description,
		Capacity:    req.Capacity,
		AdminID:     claims.UserID,
	}

	// Persist the new restaurant to the database.
	if err := h.repo.Create(rest); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating restaurant"})
		return
	}

	c.JSON(http.StatusCreated, rest)
}

// List returns all restaurants in the database.
// If no restaurants exist yet, it returns an empty array instead of null
// so the client always receives a valid JSON array.
func (h *RestaurantHandler) List(c *gin.Context) {
	restaurants, err := h.repo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching restaurants"})
		return
	}

	// Return an empty slice instead of null to keep the response consistent for clients.
	if restaurants == nil {
		restaurants = []models.Restaurant{}
	}

	c.JSON(http.StatusOK, restaurants)
}
