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

func TestRegister_DatabaseErrorOnFindByEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockUserRepo)
	jwtSvc := auth.NewJWTService()
	h := handlers.NewAuthHandler(repo, jwtSvc)

	repo.On("FindByEmail", "ana@test.com").Return(nil, assert.AnError)

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

func TestMenuDelete_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewMenuHandler(menuRepo, restRepo)

	menuRepo.On("Delete", "menu-1").Return(assert.AnError)

	r := gin.New()
	r.DELETE("/menus/:id", h.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/menus/menu-1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	menuRepo.AssertExpectations(t)
}

func TestReservationCancel_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resRepo := new(MockReservationRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewReservationHandler(resRepo, restRepo)

	resRepo.On("FindByID", "res-404").Return(nil, nil)

	r := gin.New()
	r.DELETE("/reservations/:id", h.Cancel)

	req, _ := http.NewRequest(http.MethodDelete, "/reservations/res-404", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	resRepo.AssertExpectations(t)
}

func TestOrderGet_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderRepo := new(MockOrderRepo)
	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewOrderHandler(orderRepo, menuRepo, restRepo)

	orderRepo.On("FindByID", "order-404").Return(nil, nil)

	r := gin.New()
	r.GET("/orders/:id", h.Get)

	req, _ := http.NewRequest(http.MethodGet, "/orders/order-404", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	orderRepo.AssertExpectations(t)
}

func TestUserUpdate_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockUserRepo)
	h := handlers.NewUserHandler(repo)

	repo.On("Update", "user-1", mock.AnythingOfType("*models.UpdateUserRequest")).Return(nil, nil)

	r := gin.New()
	r.PUT("/users/:id", func(c *gin.Context) {
		c.Set("claims", &models.Claims{UserID: "user-1", Role: models.RoleClient})
		h.Update(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/users/user-1", bytes.NewBufferString(`{"name":"X"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	repo.AssertExpectations(t)
}
