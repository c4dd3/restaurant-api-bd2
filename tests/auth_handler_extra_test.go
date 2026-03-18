package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"restaurant-api/internal/auth"
	"restaurant-api/internal/handlers"
)

func TestRegister_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockUserRepo)
	jwtSvc := auth.NewJWTService()
	h := handlers.NewAuthHandler(repo, jwtSvc)

	r := gin.New()
	r.POST("/auth/register", h.Register)

	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"name":`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockUserRepo)
	jwtSvc := auth.NewJWTService()
	h := handlers.NewAuthHandler(repo, jwtSvc)

	repo.On("FindByEmail", "ghost@test.com").Return(nil, nil)

	r := gin.New()
	r.POST("/auth/login", h.Login)

	req, _ := http.NewRequest(http.MethodPost, "/auth/login",
		bytes.NewBufferString(`{"email":"ghost@test.com","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid credentials")
	repo.AssertExpectations(t)
}
