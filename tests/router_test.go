package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"restaurant-api/internal/auth"
	"restaurant-api/internal/router"
)

func TestRouterSetup_HealthRoute(t *testing.T) {
	userRepo := new(MockUserRepo)
	restRepo := new(MockRestaurantRepo)
	menuRepo := new(MockMenuRepo)
	resRepo := new(MockReservationRepo)
	orderRepo := new(MockOrderRepo)
	jwtSvc := auth.NewJWTService()

	r := router.Setup(userRepo, restRepo, menuRepo, resRepo, orderRepo, jwtSvc)

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}
