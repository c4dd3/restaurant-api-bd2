package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"restaurant-api/internal/handlers"
	"restaurant-api/internal/models"
)

func TestMenuCreate_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewMenuHandler(menuRepo, restRepo)

	r := gin.New()
	r.POST("/menus", h.Create)

	req, _ := http.NewRequest(http.MethodPost, "/menus", bytes.NewBufferString(`{"restaurant_id":`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMenuGet_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewMenuHandler(menuRepo, restRepo)

	menu := &models.Menu{
		ID:           "menu-1",
		RestaurantID: "rest-1",
		Name:         "Lunch",
		Description:  "desc",
	}
	menuRepo.On("FindByID", "menu-1").Return(menu, nil)

	r := gin.New()
	r.GET("/menus/:id", h.Get)

	req, _ := http.NewRequest(http.MethodGet, "/menus/menu-1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Lunch")
	menuRepo.AssertExpectations(t)
}

func TestMenuUpdate_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewMenuHandler(menuRepo, restRepo)

	menuRepo.On("Update", "menu-404", mock.AnythingOfType("*models.UpdateMenuRequest")).Return(nil, nil)

	r := gin.New()
	r.PUT("/menus/:id", h.Update)

	req, _ := http.NewRequest(http.MethodPut, "/menus/menu-404",
		bytes.NewBufferString(`{"name":"Updated"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "menu not found")
	menuRepo.AssertExpectations(t)
}
