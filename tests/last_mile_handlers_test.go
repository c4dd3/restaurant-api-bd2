package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"restaurant-api/internal/auth"
	"restaurant-api/internal/handlers"
	"restaurant-api/internal/models"
)

func TestMenuHandler_Update_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewMenuHandler(menuRepo, restRepo)

	menuRepo.On("Update", "menu-1", mock.AnythingOfType("*models.UpdateMenuRequest")).Return(nil, assert.AnError)

	r := gin.New()
	r.PUT("/menus/:id", h.Update)

	req, _ := http.NewRequest(http.MethodPut, "/menus/menu-1",
		bytes.NewBufferString(`{"name":"Updated"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	menuRepo.AssertExpectations(t)
}

func TestReservationCreate_CheckAvailabilityError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resRepo := new(MockReservationRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewReservationHandler(resRepo, restRepo)

	rest := &models.Restaurant{ID: "rest-1", Name: "Rest 1"}
	restRepo.On("FindByID", "rest-1").Return(rest, nil)
	resRepo.On("CheckAvailability", "rest-1", 4).Return(0, assert.AnError)

	r := gin.New()
	r.POST("/reservations", func(c *gin.Context) {
		c.Set("claims", &models.Claims{UserID: "user-1", Role: models.RoleClient})
		h.Create(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/reservations",
		bytes.NewBufferString(`{"restaurant_id":"rest-1","date":"2026-03-20T18:00:00Z","party_size":4}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	restRepo.AssertExpectations(t)
	resRepo.AssertExpectations(t)
}

func TestOrderGet_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderRepo := new(MockOrderRepo)
	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewOrderHandler(orderRepo, menuRepo, restRepo)

	orderRepo.On("FindByID", "order-1").Return(nil, assert.AnError)

	r := gin.New()
	r.GET("/orders/:id", h.Get)

	req, _ := http.NewRequest(http.MethodGet, "/orders/order-1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	orderRepo.AssertExpectations(t)
}

func TestLogin_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockUserRepo)
	jwtSvc := auth.NewJWTService()
	h := handlers.NewAuthHandler(repo, jwtSvc)

	r := gin.New()
	r.POST("/auth/login", h.Login)

	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
