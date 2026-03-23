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

func TestRegister_CreateError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockUserRepo)
	jwtSvc := auth.NewJWTService()
	h := handlers.NewAuthHandler(repo, jwtSvc)

	repo.On("FindByEmail", "ana@test.com").Return(nil, nil)
	repo.On("Create", mock.AnythingOfType("*models.User")).Return(assert.AnError)

	r := gin.New()
	r.POST("/auth/register", h.Register)

	req, _ := http.NewRequest(http.MethodPost, "/auth/register",
		bytes.NewBufferString(`{"name":"Ana","email":"ana@test.com","password":"secret123","role":"client"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	repo.AssertExpectations(t)
}

func TestMenuGet_DBError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewMenuHandler(menuRepo, restRepo)

	menuRepo.On("FindByID", "menu-1").Return(nil, assert.AnError)

	r := gin.New()
	r.GET("/menus/:id", h.Get)

	req, _ := http.NewRequest(http.MethodGet, "/menus/menu-1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	menuRepo.AssertExpectations(t)
}

func TestReservationCancel_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resRepo := new(MockReservationRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewReservationHandler(resRepo, restRepo)

	reservation := &models.Reservation{
		ID:           "res-1",
		RestaurantID: "rest-1",
		UserID:       "user-1",
		Status:       models.StatusPending,
	}
	resRepo.On("FindByID", "res-1").Return(reservation, nil)
	resRepo.On("Cancel", "res-1").Return(assert.AnError)

	r := gin.New()
	r.DELETE("/reservations/:id", func(c *gin.Context) {
		c.Set("claims", &models.Claims{UserID: "user-1", Role: models.RoleClient})
		h.Cancel(c)
	})

	req, _ := http.NewRequest(http.MethodDelete, "/reservations/res-1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resRepo.AssertExpectations(t)
}

func TestOrderCreate_CreateError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderRepo := new(MockOrderRepo)
	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewOrderHandler(orderRepo, menuRepo, restRepo)

	rest := &models.Restaurant{ID: "rest-1", Name: "Rest 1"}
	item := &models.MenuItem{ID: "item-1", Name: "Pasta", Price: 10, Available: true}

	restRepo.On("FindByID", "rest-1").Return(rest, nil)
	menuRepo.On("FindItemByID", "item-1").Return(item, nil)
	orderRepo.On("Create", mock.AnythingOfType("*models.Order")).Return(assert.AnError)

	r := gin.New()
	r.POST("/orders", func(c *gin.Context) {
		c.Set("claims", &models.Claims{UserID: "user-1", Role: models.RoleClient})
		h.Create(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/orders",
		bytes.NewBufferString(`{"restaurant_id":"rest-1","items":[{"menu_item_id":"item-1","quantity":2}],"pickup":true}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	orderRepo.AssertExpectations(t)
	menuRepo.AssertExpectations(t)
	restRepo.AssertExpectations(t)
}
