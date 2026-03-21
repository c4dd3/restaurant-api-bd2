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

type MockRestaurantRepo struct{ mock.Mock }

func (m *MockRestaurantRepo) Create(rest *models.Restaurant) error {
	args := m.Called(rest)
	if rest != nil && rest.ID == "" {
		rest.ID = "rest-1"
	}
	return args.Error(0)
}
func (m *MockRestaurantRepo) FindAll() ([]models.Restaurant, error) {
	args := m.Called()
	return args.Get(0).([]models.Restaurant), args.Error(1)
}
func (m *MockRestaurantRepo) FindByID(id string) (*models.Restaurant, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Restaurant), args.Error(1)
}

type MockMenuRepo struct{ mock.Mock }

func (m *MockMenuRepo) Create(menu *models.Menu) error {
	args := m.Called(menu)
	if menu != nil && menu.ID == "" {
		menu.ID = "menu-1"
	}
	return args.Error(0)
}
func (m *MockMenuRepo) FindByID(id string) (*models.Menu, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Menu), args.Error(1)
}
func (m *MockMenuRepo) Update(id string, req *models.UpdateMenuRequest) (*models.Menu, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Menu), args.Error(1)
}
func (m *MockMenuRepo) Delete(id string) error {
	return m.Called(id).Error(0)
}
func (m *MockMenuRepo) FindItemByID(id string) (*models.MenuItem, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MenuItem), args.Error(1)
}

type MockReservationRepo struct{ mock.Mock }

func (m *MockReservationRepo) Create(res *models.Reservation) error {
	args := m.Called(res)
	if res != nil && res.ID == "" {
		res.ID = "res-1"
	}
	return args.Error(0)
}
func (m *MockReservationRepo) FindByID(id string) (*models.Reservation, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reservation), args.Error(1)
}
func (m *MockReservationRepo) Cancel(id string) error {
	return m.Called(id).Error(0)
}
func (m *MockReservationRepo) CheckAvailability(restaurantID string, partySize int) (int, error) {
	args := m.Called(restaurantID, partySize)
	return args.Int(0), args.Error(1)
}

type MockOrderRepo struct{ mock.Mock }

func (m *MockOrderRepo) Create(order *models.Order) error {
	args := m.Called(order)
	if order != nil && order.ID == "" {
		order.ID = "order-1"
	}
	return args.Error(0)
}
func (m *MockOrderRepo) FindByID(id string) (*models.Order, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func TestReservationCreate_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resRepo := new(MockReservationRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewReservationHandler(resRepo, restRepo)

	r := gin.New()
	r.POST("/reservations", h.Create)

	req, _ := http.NewRequest(http.MethodPost, "/reservations", bytes.NewBufferString(`{"restaurant_id":`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReservationCreate_Unauthorized_NoClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resRepo := new(MockReservationRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewReservationHandler(resRepo, restRepo)

	rest := &models.Restaurant{ID: "rest-1", Name: "Rest 1"}
	restRepo.On("FindByID", "rest-1").Return(rest, nil)
	resRepo.On("CheckAvailability", "rest-1", 4).Return(10, nil)

	r := gin.New()
	r.POST("/reservations", h.Create)

	req, _ := http.NewRequest(http.MethodPost, "/reservations",
		bytes.NewBufferString(`{
			"restaurant_id":"rest-1",
			"date":"2026-03-20T18:00:00Z",
			"party_size":4
		}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	restRepo.AssertExpectations(t)
	resRepo.AssertExpectations(t)
}

func TestReservationCreate_CreateError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resRepo := new(MockReservationRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewReservationHandler(resRepo, restRepo)

	rest := &models.Restaurant{ID: "rest-1", Name: "Rest 1"}
	restRepo.On("FindByID", "rest-1").Return(rest, nil)
	resRepo.On("CheckAvailability", "rest-1", 4).Return(10, nil)
	resRepo.On("Create", mock.AnythingOfType("*models.Reservation")).Return(assert.AnError)

	r := gin.New()
	r.POST("/reservations", func(c *gin.Context) {
		c.Set("claims", &models.Claims{UserID: "user-1", Role: models.RoleClient})
		h.Create(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/reservations",
		bytes.NewBufferString(`{
			"restaurant_id":"rest-1",
			"date":"2026-03-20T18:00:00Z",
			"party_size":4
		}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	resRepo.AssertExpectations(t)
	restRepo.AssertExpectations(t)
}

func TestReservationCancel_Unauthorized_NoClaims(t *testing.T) {
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

	r := gin.New()
	r.DELETE("/reservations/:id", h.Cancel)

	req, _ := http.NewRequest(http.MethodDelete, "/reservations/res-1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	resRepo.AssertExpectations(t)
}

func TestOrderCreate_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderRepo := new(MockOrderRepo)
	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewOrderHandler(orderRepo, menuRepo, restRepo)

	r := gin.New()
	r.POST("/orders", h.Create)

	req, _ := http.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(`{"restaurant_id":`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOrderCreate_Unauthorized_NoClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderRepo := new(MockOrderRepo)
	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewOrderHandler(orderRepo, menuRepo, restRepo)

	rest := &models.Restaurant{ID: "rest-1", Name: "Rest 1"}
	restRepo.On("FindByID", "rest-1").Return(rest, nil)

	r := gin.New()
	r.POST("/orders", h.Create)

	req, _ := http.NewRequest(http.MethodPost, "/orders",
		bytes.NewBufferString(`{
			"restaurant_id":"rest-1",
			"items":[{"menu_item_id":"item-1","quantity":2}],
			"pickup":true
		}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	restRepo.AssertExpectations(t)
}

func TestOrderGet_Unauthorized_NoClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderRepo := new(MockOrderRepo)
	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewOrderHandler(orderRepo, menuRepo, restRepo)

	order := &models.Order{
		ID:           "order-1",
		UserID:       "user-1",
		RestaurantID: "rest-1",
		Status:       models.StatusPending,
		Total:        35,
	}
	orderRepo.On("FindByID", "order-1").Return(order, nil)

	r := gin.New()
	r.GET("/orders/:id", h.Get)

	req, _ := http.NewRequest(http.MethodGet, "/orders/order-1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	orderRepo.AssertExpectations(t)
}
