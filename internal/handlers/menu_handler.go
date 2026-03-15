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
	return &MenuHandler{
		menuRepo:       menuRepo,
		restaurantRepo: restaurantRepo,
	}
}

// CreateMenu godoc
// @Summary Create a new menu
// @Tags menus
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.CreateMenuRequest true "Menu data"
// @Success 201 {object} models.Menu
// @Router /menus [post]
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

// GetMenu godoc
// @Summary Get menu by ID
// @Tags menus
// @Security BearerAuth
// @Produce json
// @Param id path string true "Menu ID"
// @Success 200 {object} models.Menu
// @Router /menus/{id} [get]
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

// UpdateMenu godoc
// @Summary Update a menu
// @Tags menus
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Menu ID"
// @Param body body models.UpdateMenuRequest true "Update data"
// @Success 200 {object} models.Menu
// @Router /menus/{id} [put]
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

// DeleteMenu godoc
// @Summary Delete a menu
// @Tags menus
// @Security BearerAuth
// @Param id path string true "Menu ID"
// @Success 204
// @Router /menus/{id} [delete]
func (h *MenuHandler) Delete(c *gin.Context) {
	if err := h.menuRepo.Delete(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting menu"})
		return
	}

	c.Status(http.StatusNoContent)
}
