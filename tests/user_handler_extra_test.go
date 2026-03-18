package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"restaurant-api/internal/handlers"
)

func TestUserUpdate_Unauthorized_NoClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockUserRepo)
	h := handlers.NewUserHandler(repo)

	r := gin.New()
	r.PUT("/users/:id", h.Update)

	req, _ := http.NewRequest(http.MethodPut, "/users/user-1", bytes.NewBufferString(`{"name":"X"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUserDelete_Unauthorized_NoClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockUserRepo)
	h := handlers.NewUserHandler(repo)

	r := gin.New()
	r.DELETE("/users/:id", h.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/users/user-1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
