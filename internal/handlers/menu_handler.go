package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"restaurant-api/internal/models"
)

type MenuHandler struct {
	menuRepo       MenuRepository
	restaurantRepo RestaurantRepository
}

func NewMenuHandler(menuRepo MenuRepository, restaurantRepo RestaurantRepository) *MenuHandler {
	return &MenuHandler{menuRepo: menuRepo, restaurantRepo: restaurantRepo}
}

func (h *MenuHandler) Create(c *gin.Context) {
	var req models.CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rest, err := h.restaurantRepo.FindByID(req.RestaurantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if rest == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "restaurant not found"})
		return
	}

	menu := &models.Menu{
		RestaurantID: req.RestaurantID,
		Name:         req.Name,
		Description:  req.Description,
	}
	for _, item := range req.Items {
		menu.Items = append(menu.Items, models.MenuItem{
			Name:        item.Name,
			Description: item.Description,
			Price:       item.Price,
			Available:   item.Available,
		})
	}

	if err := h.menuRepo.Create(menu); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating menu"})
		return
	}

	c.JSON(http.StatusCreated, menu)
}

func (h *MenuHandler) Get(c *gin.Context) {
	menu, err := h.menuRepo.FindByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if menu == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "menu not found"})
		return
	}
	c.JSON(http.StatusOK, menu)
}

func (h *MenuHandler) Update(c *gin.Context) {
	var req models.UpdateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	menu, err := h.menuRepo.Update(c.Param("id"), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error updating menu"})
		return
	}
	if menu == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "menu not found"})
		return
	}
	c.JSON(http.StatusOK, menu)
}

func (h *MenuHandler) Delete(c *gin.Context) {
	if err := h.menuRepo.Delete(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting menu"})
		return
	}
	c.Status(http.StatusNoContent)
}
