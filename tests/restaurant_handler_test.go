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

func TestRestaurantList_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockRestaurantRepo)
	h := handlers.NewRestaurantHandler(repo)

	restaurants := []models.Restaurant{
		{ID: "r1", Name: "Rest 1", Capacity: 50},
		{ID: "r2", Name: "Rest 2", Capacity: 80},
	}
	repo.On("FindAll").Return(restaurants, nil)

	r := gin.New()
	r.GET("/restaurants", h.List)

	req, _ := http.NewRequest(http.MethodGet, "/restaurants", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Rest 1")
	assert.Contains(t, w.Body.String(), "Rest 2")
	repo.AssertExpectations(t)
}

func TestRestaurantCreate_Unauthorized_NoClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockRestaurantRepo)
	h := handlers.NewRestaurantHandler(repo)

	r := gin.New()
	r.POST("/restaurants", h.Create)

	req, _ := http.NewRequest(http.MethodPost, "/restaurants",
		bytes.NewBufferString(`{"name":"My Rest","address":"123 St","phone":"2222","capacity":60}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRestaurantCreate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockRestaurantRepo)
	h := handlers.NewRestaurantHandler(repo)

	repo.On("Create", mock.AnythingOfType("*models.Restaurant")).Return(nil)

	r := gin.New()
	r.POST("/restaurants", func(c *gin.Context) {
		c.Set("claims", &models.Claims{UserID: "admin-1", Role: models.RoleAdmin})
		h.Create(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/restaurants",
		bytes.NewBufferString(`{"name":"My Rest","address":"123 St","phone":"2222","description":"nice","capacity":60}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "My Rest")
	repo.AssertExpectations(t)
}
