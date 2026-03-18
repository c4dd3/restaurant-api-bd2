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

func TestOrderCreate_RestaurantNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderRepo := new(MockOrderRepo)
	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewOrderHandler(orderRepo, menuRepo, restRepo)

	restRepo.On("FindByID", "rest-404").Return(nil, nil)

	r := gin.New()
	r.POST("/orders", h.Create)

	req, _ := http.NewRequest(http.MethodPost, "/orders",
		bytes.NewBufferString(`{
			"restaurant_id":"rest-404",
			"items":[{"menu_item_id":"item-1","quantity":2}],
			"pickup":true
		}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "restaurant not found")
	restRepo.AssertExpectations(t)
}

func TestOrderCreate_MenuItemNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderRepo := new(MockOrderRepo)
	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewOrderHandler(orderRepo, menuRepo, restRepo)

	rest := &models.Restaurant{ID: "rest-1", Name: "Rest 1"}
	restRepo.On("FindByID", "rest-1").Return(rest, nil)
	menuRepo.On("FindItemByID", "item-404").Return(nil, nil)

	r := gin.New()
	r.POST("/orders", func(c *gin.Context) {
		c.Set("claims", &models.Claims{UserID: "user-1", Role: models.RoleClient})
		h.Create(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/orders",
		bytes.NewBufferString(`{
			"restaurant_id":"rest-1",
			"items":[{"menu_item_id":"item-404","quantity":2}],
			"pickup":true
		}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "menu item not found")
	restRepo.AssertExpectations(t)
	menuRepo.AssertExpectations(t)
}

func TestOrderCreate_MenuItemUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderRepo := new(MockOrderRepo)
	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewOrderHandler(orderRepo, menuRepo, restRepo)

	rest := &models.Restaurant{ID: "rest-1", Name: "Rest 1"}
	item := &models.MenuItem{ID: "item-1", Name: "Pasta", Price: 12.5, Available: false}

	restRepo.On("FindByID", "rest-1").Return(rest, nil)
	menuRepo.On("FindItemByID", "item-1").Return(item, nil)

	r := gin.New()
	r.POST("/orders", func(c *gin.Context) {
		c.Set("claims", &models.Claims{UserID: "user-1", Role: models.RoleClient})
		h.Create(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/orders",
		bytes.NewBufferString(`{
			"restaurant_id":"rest-1",
			"items":[{"menu_item_id":"item-1","quantity":2}],
			"pickup":true
		}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "menu item not available")
	restRepo.AssertExpectations(t)
	menuRepo.AssertExpectations(t)
}

func TestOrderCreate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderRepo := new(MockOrderRepo)
	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewOrderHandler(orderRepo, menuRepo, restRepo)

	rest := &models.Restaurant{ID: "rest-1", Name: "Rest 1"}
	item1 := &models.MenuItem{ID: "item-1", Name: "Pasta", Price: 10.0, Available: true}
	item2 := &models.MenuItem{ID: "item-2", Name: "Pizza", Price: 5.0, Available: true}

	restRepo.On("FindByID", "rest-1").Return(rest, nil)
	menuRepo.On("FindItemByID", "item-1").Return(item1, nil)
	menuRepo.On("FindItemByID", "item-2").Return(item2, nil)
	orderRepo.On("Create", mock.AnythingOfType("*models.Order")).Return(nil)

	r := gin.New()
	r.POST("/orders", func(c *gin.Context) {
		c.Set("claims", &models.Claims{UserID: "user-1", Role: models.RoleClient})
		h.Create(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/orders",
		bytes.NewBufferString(`{
			"restaurant_id":"rest-1",
			"items":[
				{"menu_item_id":"item-1","quantity":2},
				{"menu_item_id":"item-2","quantity":3}
			],
			"pickup":true
		}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"total":35`)
	orderRepo.AssertExpectations(t)
	menuRepo.AssertExpectations(t)
	restRepo.AssertExpectations(t)
}

func TestOrderGet_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderRepo := new(MockOrderRepo)
	menuRepo := new(MockMenuRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewOrderHandler(orderRepo, menuRepo, restRepo)

	order := &models.Order{
		ID:           "order-1",
		UserID:       "owner-user",
		RestaurantID: "rest-1",
		Status:       models.StatusPending,
	}
	orderRepo.On("FindByID", "order-1").Return(order, nil)

	r := gin.New()
	r.GET("/orders/:id", func(c *gin.Context) {
		c.Set("claims", &models.Claims{UserID: "other-user", Role: models.RoleClient})
		h.Get(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/orders/order-1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "cannot access another user's order")
	orderRepo.AssertExpectations(t)
}

func TestOrderGet_Success(t *testing.T) {
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
	r.GET("/orders/:id", func(c *gin.Context) {
		c.Set("claims", &models.Claims{UserID: "user-1", Role: models.RoleClient})
		h.Get(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/orders/order-1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"total":35`)
	orderRepo.AssertExpectations(t)
}
