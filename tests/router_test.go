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

// Auth routes must NOT exist in the API service — they live in the auth-service container.
func TestRouterSetup_AuthRegister_NotFound(t *testing.T) {
	userRepo := new(MockUserRepo)
	restRepo := new(MockRestaurantRepo)
	menuRepo := new(MockMenuRepo)
	resRepo := new(MockReservationRepo)
	orderRepo := new(MockOrderRepo)
	jwtSvc := auth.NewJWTService()

	r := router.Setup(userRepo, restRepo, menuRepo, resRepo, orderRepo, jwtSvc)

	req, _ := http.NewRequest(http.MethodPost, "/auth/register", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRouterSetup_AuthLogin_NotFound(t *testing.T) {
	userRepo := new(MockUserRepo)
	restRepo := new(MockRestaurantRepo)
	menuRepo := new(MockMenuRepo)
	resRepo := new(MockReservationRepo)
	orderRepo := new(MockOrderRepo)
	jwtSvc := auth.NewJWTService()

	r := router.Setup(userRepo, restRepo, menuRepo, resRepo, orderRepo, jwtSvc)

	req, _ := http.NewRequest(http.MethodPost, "/auth/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// Protected routes without a token must return 401
func TestRouterSetup_ProtectedRoutes_Unauthorized(t *testing.T) {
	userRepo := new(MockUserRepo)
	restRepo := new(MockRestaurantRepo)
	menuRepo := new(MockMenuRepo)
	resRepo := new(MockReservationRepo)
	orderRepo := new(MockOrderRepo)
	jwtSvc := auth.NewJWTService()

	r := router.Setup(userRepo, restRepo, menuRepo, resRepo, orderRepo, jwtSvc)

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/users/me"},
		{http.MethodGet, "/restaurants"},
		{http.MethodPost, "/restaurants"},
		{http.MethodGet, "/menus/some-id"},
		{http.MethodPost, "/reservations"},
		{http.MethodPost, "/orders"},
	}

	for _, route := range routes {
		req, _ := http.NewRequest(route.method, route.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code, "expected 401 for %s %s", route.method, route.path)
	}
}
