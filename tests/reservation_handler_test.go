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

func TestReservationCreate_RestaurantNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resRepo := new(MockReservationRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewReservationHandler(resRepo, restRepo)

	restRepo.On("FindByID", "rest-404").Return(nil, nil)

	r := gin.New()
	r.POST("/reservations", h.Create)

	req, _ := http.NewRequest(http.MethodPost, "/reservations",
		bytes.NewBufferString(`{
			"restaurant_id":"rest-404",
			"date":"2026-03-20T18:00:00Z",
			"party_size":4,
			"notes":"window"
		}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "restaurant not found")
	restRepo.AssertExpectations(t)
}

func TestReservationCreate_InsufficientCapacity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resRepo := new(MockReservationRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewReservationHandler(resRepo, restRepo)

	rest := &models.Restaurant{ID: "rest-1", Name: "Rest 1"}
	restRepo.On("FindByID", "rest-1").Return(rest, nil)
	resRepo.On("CheckAvailability", "rest-1", 6).Return(-1, nil)

	r := gin.New()
	r.POST("/reservations", func(c *gin.Context) {
		c.Set("claims", &models.Claims{UserID: "user-1", Role: models.RoleClient})
		h.Create(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/reservations",
		bytes.NewBufferString(`{
			"restaurant_id":"rest-1",
			"date":"2026-03-20T18:00:00Z",
			"party_size":6,
			"notes":"window"
		}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "insufficient capacity")
	restRepo.AssertExpectations(t)
	resRepo.AssertExpectations(t)
}

func TestReservationCreate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resRepo := new(MockReservationRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewReservationHandler(resRepo, restRepo)

	rest := &models.Restaurant{ID: "rest-1", Name: "Rest 1"}
	restRepo.On("FindByID", "rest-1").Return(rest, nil)
	resRepo.On("CheckAvailability", "rest-1", 4).Return(10, nil)
	resRepo.On("Create", mock.AnythingOfType("*models.Reservation")).Return(nil)

	r := gin.New()
	r.POST("/reservations", func(c *gin.Context) {
		c.Set("claims", &models.Claims{UserID: "user-1", Role: models.RoleClient})
		h.Create(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/reservations",
		bytes.NewBufferString(`{
			"restaurant_id":"rest-1",
			"date":"2026-03-20T18:00:00Z",
			"party_size":4,
			"notes":"window"
		}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"restaurant_id":"rest-1"`)
	resRepo.AssertExpectations(t)
	restRepo.AssertExpectations(t)
}

func TestReservationCancel_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resRepo := new(MockReservationRepo)
	restRepo := new(MockRestaurantRepo)
	h := handlers.NewReservationHandler(resRepo, restRepo)

	reservation := &models.Reservation{
		ID:           "res-1",
		RestaurantID: "rest-1",
		UserID:       "owner-user",
		Status:       models.StatusPending,
	}
	resRepo.On("FindByID", "res-1").Return(reservation, nil)

	r := gin.New()
	r.DELETE("/reservations/:id", func(c *gin.Context) {
		c.Set("claims", &models.Claims{UserID: "other-user", Role: models.RoleClient})
		h.Cancel(c)
	})

	req, _ := http.NewRequest(http.MethodDelete, "/reservations/res-1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "cannot cancel another user's reservation")
	resRepo.AssertExpectations(t)
}

func TestReservationCancel_Success(t *testing.T) {
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
	resRepo.On("Cancel", "res-1").Return(nil)

	r := gin.New()
	r.DELETE("/reservations/:id", func(c *gin.Context) {
		c.Set("claims", &models.Claims{UserID: "user-1", Role: models.RoleClient})
		h.Cancel(c)
	})

	req, _ := http.NewRequest(http.MethodDelete, "/reservations/res-1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "reservation cancelled")
	resRepo.AssertExpectations(t)
}
