package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"restaurant-api/internal/handlers"
	"restaurant-api/internal/models"
)

func TestRestaurantCreate_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockRestaurantRepo)
	h := handlers.NewRestaurantHandler(repo)

	r := gin.New()
	r.POST("/restaurants", h.Create)

	req, _ := http.NewRequest(http.MethodPost, "/restaurants", bytes.NewBufferString(`{"name":`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRestaurantList_EmptySlice(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockRestaurantRepo)
	h := handlers.NewRestaurantHandler(repo)

	repo.On("FindAll").Return([]models.Restaurant{}, nil)

	r := gin.New()
	r.GET("/restaurants", h.List)

	req, _ := http.NewRequest(http.MethodGet, "/restaurants", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "[]")
	repo.AssertExpectations(t)
}
