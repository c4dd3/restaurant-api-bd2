//go:build integration
// +build integration

package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"restaurant-api/internal/auth"
	"restaurant-api/internal/repository"
	"restaurant-api/internal/router"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDB     *sql.DB
	testEngine http.Handler
	jwtSvc     *auth.JWTService
)

func getTestDB() (*sql.DB, error) {
	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_DB_PORT", "5432")
	user := getEnvOrDefault("TEST_DB_USER", "postgres")
	password := getEnvOrDefault("TEST_DB_PASSWORD", "postgres")
	dbname := getEnvOrDefault("TEST_DB_NAME", "restaurant_test")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	return sql.Open("postgres", dsn)
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func setupIntegration(t *testing.T) {
	t.Helper()
	db, err := getTestDB()
	if err != nil {
		t.Skip("no test database available:", err)
	}
	if err := db.Ping(); err != nil {
		t.Skip("cannot reach test database:", err)
	}

	if err := repository.RunMigrations(db); err != nil {
		t.Fatal("migration failed:", err)
	}

	testDB = db
	jwtSvc = auth.NewJWTService()

	userRepo := repository.NewUserRepository(db)
	restaurantRepo := repository.NewRestaurantRepository(db)
	menuRepo := repository.NewMenuRepository(db)
	reservationRepo := repository.NewReservationRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	testEngine = router.Setup(userRepo, restaurantRepo, menuRepo, reservationRepo, orderRepo, jwtSvc)
}

func doRequest(method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	testEngine.ServeHTTP(w, req)
	return w
}

// ─── Integration Tests ────────────────────────────────────────────────────────

func TestIntegration_Health(t *testing.T) {
	setupIntegration(t)
	w := doRequest(http.MethodGet, "/health", nil, "")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIntegration_Register(t *testing.T) {
	setupIntegration(t)
	email := fmt.Sprintf("user_%d@test.com", time.Now().UnixNano())
	body := map[string]interface{}{
		"name": "Test User", "email": email, "password": "password123", "role": "client",
	}
	w := doRequest(http.MethodPost, "/auth/register", body, "")
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp["token"])
}

func TestIntegration_Register_DuplicateEmail(t *testing.T) {
	setupIntegration(t)
	email := fmt.Sprintf("dup_%d@test.com", time.Now().UnixNano())
	body := map[string]interface{}{
		"name": "User", "email": email, "password": "password123", "role": "client",
	}
	doRequest(http.MethodPost, "/auth/register", body, "")
	w := doRequest(http.MethodPost, "/auth/register", body, "")
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestIntegration_Login(t *testing.T) {
	setupIntegration(t)
	email := fmt.Sprintf("login_%d@test.com", time.Now().UnixNano())
	doRequest(http.MethodPost, "/auth/register", map[string]interface{}{
		"name": "User", "email": email, "password": "mypassword", "role": "client",
	}, "")

	w := doRequest(http.MethodPost, "/auth/login", map[string]interface{}{
		"email": email, "password": "mypassword",
	}, "")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIntegration_Login_WrongPassword(t *testing.T) {
	setupIntegration(t)
	w := doRequest(http.MethodPost, "/auth/login", map[string]interface{}{
		"email": "nobody@test.com", "password": "wrong",
	}, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIntegration_GetMe_Unauthorized(t *testing.T) {
	setupIntegration(t)
	w := doRequest(http.MethodGet, "/users/me", nil, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIntegration_GetMe_Authorized(t *testing.T) {
	setupIntegration(t)
	email := fmt.Sprintf("me_%d@test.com", time.Now().UnixNano())
	regW := doRequest(http.MethodPost, "/auth/register", map[string]interface{}{
		"name": "Me User", "email": email, "password": "password123", "role": "client",
	}, "")

	var regResp map[string]interface{}
	json.Unmarshal(regW.Body.Bytes(), &regResp)
	token := regResp["token"].(string)

	w := doRequest(http.MethodGet, "/users/me", nil, token)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIntegration_CreateRestaurant_AdminOnly(t *testing.T) {
	setupIntegration(t)
	// Client token should be forbidden
	email := fmt.Sprintf("client_%d@test.com", time.Now().UnixNano())
	regW := doRequest(http.MethodPost, "/auth/register", map[string]interface{}{
		"name": "Client", "email": email, "password": "password123", "role": "client",
	}, "")
	var regResp map[string]interface{}
	json.Unmarshal(regW.Body.Bytes(), &regResp)
	clientToken := regResp["token"].(string)

	w := doRequest(http.MethodPost, "/restaurants", map[string]interface{}{
		"name": "My Restaurant", "address": "123 St", "phone": "1234", "capacity": 50,
	}, clientToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestIntegration_CreateRestaurant_Admin(t *testing.T) {
	setupIntegration(t)
	email := fmt.Sprintf("admin_%d@test.com", time.Now().UnixNano())
	regW := doRequest(http.MethodPost, "/auth/register", map[string]interface{}{
		"name": "Admin", "email": email, "password": "password123", "role": "admin",
	}, "")
	var regResp map[string]interface{}
	json.Unmarshal(regW.Body.Bytes(), &regResp)
	adminToken := regResp["token"].(string)

	w := doRequest(http.MethodPost, "/restaurants", map[string]interface{}{
		"name": "Gourmet Place", "address": "456 Ave", "phone": "5678", "capacity": 80,
	}, adminToken)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestIntegration_ListRestaurants(t *testing.T) {
	setupIntegration(t)
	email := fmt.Sprintf("list_%d@test.com", time.Now().UnixNano())
	regW := doRequest(http.MethodPost, "/auth/register", map[string]interface{}{
		"name": "User", "email": email, "password": "password123", "role": "client",
	}, "")
	var regResp map[string]interface{}
	json.Unmarshal(regW.Body.Bytes(), &regResp)
	token := regResp["token"].(string)

	w := doRequest(http.MethodGet, "/restaurants", nil, token)
	assert.Equal(t, http.StatusOK, w.Code)
}

func registerAndGetToken(t *testing.T, role string) string {
	t.Helper()
	email := fmt.Sprintf("%s_%d@test.com", role, time.Now().UnixNano())
	regW := doRequest(http.MethodPost, "/auth/register", map[string]interface{}{
		"name": role + "User", "email": email, "password": "password123", "role": role,
	}, "")
	require.Equal(t, http.StatusCreated, regW.Code)
	var resp map[string]interface{}
	json.Unmarshal(regW.Body.Bytes(), &resp)
	return resp["token"].(string)
}

func createRestaurantAndGetID(t *testing.T, adminToken string) string {
	t.Helper()
	w := doRequest(http.MethodPost, "/restaurants", map[string]interface{}{
		"name": "Test Resto", "address": "789 Blvd", "phone": "0000", "capacity": 100,
	}, adminToken)
	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp["id"].(string)
}

func TestIntegration_FullFlow_MenuAndReservation(t *testing.T) {
	setupIntegration(t)
	adminToken := registerAndGetToken(t, "admin")
	clientToken := registerAndGetToken(t, "client")
	restID := createRestaurantAndGetID(t, adminToken)

	// Create menu
	menuW := doRequest(http.MethodPost, "/menus", map[string]interface{}{
		"restaurant_id": restID,
		"name":          "Dinner Menu",
		"items": []map[string]interface{}{
			{"name": "Pasta", "price": 12.99, "available": true},
		},
	}, adminToken)
	assert.Equal(t, http.StatusCreated, menuW.Code)

	var menuResp map[string]interface{}
	json.Unmarshal(menuW.Body.Bytes(), &menuResp)
	menuID := menuResp["id"].(string)

	// Get menu
	getMenuW := doRequest(http.MethodGet, "/menus/"+menuID, nil, clientToken)
	assert.Equal(t, http.StatusOK, getMenuW.Code)

	// Create reservation
	resW := doRequest(http.MethodPost, "/reservations", map[string]interface{}{
		"restaurant_id": restID,
		"date":          time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"party_size":    4,
	}, clientToken)
	assert.Equal(t, http.StatusCreated, resW.Code)

	var resResp map[string]interface{}
	json.Unmarshal(resW.Body.Bytes(), &resResp)
	resID := resResp["id"].(string)

	// Cancel reservation
	cancelW := doRequest(http.MethodDelete, "/reservations/"+resID, nil, clientToken)
	assert.Equal(t, http.StatusOK, cancelW.Code)
}

func TestIntegration_DeleteMenu(t *testing.T) {
	setupIntegration(t)
	adminToken := registerAndGetToken(t, "admin")
	restID := createRestaurantAndGetID(t, adminToken)

	menuW := doRequest(http.MethodPost, "/menus", map[string]interface{}{
		"restaurant_id": restID, "name": "ToDelete",
	}, adminToken)
	require.Equal(t, http.StatusCreated, menuW.Code)
	var menuResp map[string]interface{}
	json.Unmarshal(menuW.Body.Bytes(), &menuResp)
	menuID := menuResp["id"].(string)

	w := doRequest(http.MethodDelete, "/menus/"+menuID, nil, adminToken)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestIntegration_UpdateUser(t *testing.T) {
	setupIntegration(t)
	email := fmt.Sprintf("upd_%d@test.com", time.Now().UnixNano())
	regW := doRequest(http.MethodPost, "/auth/register", map[string]interface{}{
		"name": "Original", "email": email, "password": "password123", "role": "client",
	}, "")
	var regResp map[string]interface{}
	json.Unmarshal(regW.Body.Bytes(), &regResp)
	token := regResp["token"].(string)
	userID := regResp["user"].(map[string]interface{})["id"].(string)

	w := doRequest(http.MethodPut, "/users/"+userID, map[string]interface{}{
		"name": "Updated Name",
	}, token)
	assert.Equal(t, http.StatusOK, w.Code)
}
