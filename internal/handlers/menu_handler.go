package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"restaurant-api/internal/models"
)

// MenuHandler handles HTTP requests related to menus and their items.
// It needs both the menu repo (to read/write menus) and the restaurant repo
// (to verify the restaurant exists before creating a menu for it).
type MenuHandler struct {
	menuRepo       MenuRepository
	restaurantRepo RestaurantRepository
}

// NewMenuHandler creates a MenuHandler with the given repositories.
func NewMenuHandler(menuRepo MenuRepository, restaurantRepo RestaurantRepository) *MenuHandler {
	return &MenuHandler{menuRepo: menuRepo, restaurantRepo: restaurantRepo}
}

// Create adds a new menu (with its items) to a restaurant.
// Steps: validate the request body → confirm the restaurant exists → build the menu
// model with its items → save it to the database.
func (h *MenuHandler) Create(c *gin.Context) {
	// Parse and validate the JSON request body.
	var req models.CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Make sure the target restaurant exists before attaching a menu to it.
	rest, err := h.restaurantRepo.FindByID(req.RestaurantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if rest == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "restaurant not found"})
		return
	}

	// Build the menu model and convert each requested item into a MenuItem.
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

	// Persist the menu (and its items) to the database.
	if err := h.menuRepo.Create(menu); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating menu"})
		return
	}

	c.JSON(http.StatusCreated, menu)
}

// Get returns a single menu by the ID provided in the URL parameter.
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

// Update modifies the fields of an existing menu identified by the URL parameter ID.
func (h *MenuHandler) Update(c *gin.Context) {
	// Parse and validate the JSON request body.
	var req models.UpdateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Apply the update; the repository returns nil if no menu with that ID exists.
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

// Delete removes a menu by the ID provided in the URL parameter.
// Returns 204 No Content on success (no body needed).
func (h *MenuHandler) Delete(c *gin.Context) {
	if err := h.menuRepo.Delete(c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting menu"})
		return
	}
	c.Status(http.StatusNoContent)
}
