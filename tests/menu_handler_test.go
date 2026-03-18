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

func TestMenuCreate_RestaurantNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewMenuHandler(menuRepo, restRepo)

	restRepo.On("FindByID", "rest-404").Return(nil, nil)

	r := gin.New()
	r.POST("/menus", h.Create)

	req, _ := http.NewRequest(http.MethodPost, "/menus",
		bytes.NewBufferString(`{
			"restaurant_id":"rest-404",
			"name":"Dinner",
			"description":"desc",
			"items":[{"name":"Pasta","description":"ok","price":12.5,"available":true}]
		}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "restaurant not found")
	restRepo.AssertExpectations(t)
}

func TestMenuCreate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewMenuHandler(menuRepo, restRepo)

	rest := &models.Restaurant{ID: "rest-1", Name: "Rest 1"}
	restRepo.On("FindByID", "rest-1").Return(rest, nil)
	menuRepo.On("Create", mock.AnythingOfType("*models.Menu")).Return(nil)

	r := gin.New()
	r.POST("/menus", h.Create)

	req, _ := http.NewRequest(http.MethodPost, "/menus",
		bytes.NewBufferString(`{
			"restaurant_id":"rest-1",
			"name":"Dinner",
			"description":"desc",
			"items":[
				{"name":"Pasta","description":"ok","price":12.5,"available":true},
				{"name":"Pizza","description":"ok","price":15.0,"available":true}
			]
		}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Dinner")
	assert.Contains(t, w.Body.String(), "Pasta")
	menuRepo.AssertExpectations(t)
	restRepo.AssertExpectations(t)
}

func TestMenuGet_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewMenuHandler(menuRepo, restRepo)

	menuRepo.On("FindByID", "menu-404").Return(nil, nil)

	r := gin.New()
	r.GET("/menus/:id", h.Get)

	req, _ := http.NewRequest(http.MethodGet, "/menus/menu-404", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "menu not found")
	menuRepo.AssertExpectations(t)
}

func TestMenuUpdate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewMenuHandler(menuRepo, restRepo)

	updated := &models.Menu{
		ID:           "menu-1",
		RestaurantID: "rest-1",
		Name:         "Updated Menu",
		Description:  "new desc",
	}
	menuRepo.On("Update", "menu-1", mock.AnythingOfType("*models.UpdateMenuRequest")).Return(updated, nil)

	r := gin.New()
	r.PUT("/menus/:id", h.Update)

	req, _ := http.NewRequest(http.MethodPut, "/menus/menu-1",
		bytes.NewBufferString(`{"name":"Updated Menu","description":"new desc"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Updated Menu")
	menuRepo.AssertExpectations(t)
}

func TestMenuDelete_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewMenuHandler(menuRepo, restRepo)

	menuRepo.On("Delete", "menu-1").Return(nil)

	r := gin.New()
	r.DELETE("/menus/:id", h.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/menus/menu-1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	menuRepo.AssertExpectations(t)
}
