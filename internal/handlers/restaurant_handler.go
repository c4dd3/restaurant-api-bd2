package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"restaurant-api/internal/middleware"
	"restaurant-api/internal/models"
)

type RestaurantHandler struct {
	repo RestaurantRepository
}

func NewRestaurantHandler(repo RestaurantRepository) *RestaurantHandler {
	return &RestaurantHandler{repo: repo}
}

func (h *RestaurantHandler) Create(c *gin.Context) {
	var req models.CreateRestaurantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims := middleware.ExtractClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	rest := &models.Restaurant{
		Name:        req.Name,
		Address:     req.Address,
		Phone:       req.Phone,
		Description: req.Description,
		Capacity:    req.Capacity,
		AdminID:     claims.UserID,
	}

	if err := h.repo.Create(rest); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating restaurant"})
		return
	}

	c.JSON(http.StatusCreated, rest)
}

func (h *RestaurantHandler) List(c *gin.Context) {
	restaurants, err := h.repo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching restaurants"})
		return
	}

	if restaurants == nil {
		restaurants = []models.Restaurant{}
	}

	c.JSON(http.StatusOK, restaurants)
}
