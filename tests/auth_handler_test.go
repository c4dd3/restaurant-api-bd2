package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"restaurant-api/internal/auth"
	"restaurant-api/internal/handlers"
	"restaurant-api/internal/models"
)

func TestRegister_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockUserRepo)
	jwtSvc := auth.NewJWTService()
	h := handlers.NewAuthHandler(repo, jwtSvc)

	repo.On("FindByEmail", "ana@test.com").Return(nil, nil)
	repo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)

	r := gin.New()
	r.POST("/auth/register", h.Register)

	body := models.RegisterRequest{
		Name:     "Ana",
		Email:    "ana@test.com",
		Password: "secret123",
		Role:     "client",
	}
	b, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"token"`)
	assert.Contains(t, w.Body.String(), `"ana@test.com"`)
	repo.AssertExpectations(t)
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockUserRepo)
	jwtSvc := auth.NewJWTService()
	h := handlers.NewAuthHandler(repo, jwtSvc)

	existing := &models.User{ID: "1", Email: "ana@test.com", Role: "client"}
	repo.On("FindByEmail", "ana@test.com").Return(existing, nil)

	r := gin.New()
	r.POST("/auth/register", h.Register)

	req, _ := http.NewRequest(http.MethodPost, "/auth/register",
		bytes.NewBufferString(`{"name":"Ana","email":"ana@test.com","password":"secret123","role":"client"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "email already in use")
	repo.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockUserRepo)
	jwtSvc := auth.NewJWTService()
	h := handlers.NewAuthHandler(repo, jwtSvc)

	hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	user := &models.User{ID: "u1", Name: "Ana", Email: "ana@test.com", Password: string(hash), Role: "client"}

	repo.On("FindByEmail", "ana@test.com").Return(user, nil)

	r := gin.New()
	r.POST("/auth/login", h.Login)

	req, _ := http.NewRequest(http.MethodPost, "/auth/login",
		bytes.NewBufferString(`{"email":"ana@test.com","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"token"`)
	assert.Contains(t, w.Body.String(), `"ana@test.com"`)
	repo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockUserRepo)
	jwtSvc := auth.NewJWTService()
	h := handlers.NewAuthHandler(repo, jwtSvc)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	user := &models.User{ID: "u1", Email: "ana@test.com", Password: string(hash), Role: "client"}
	repo.On("FindByEmail", "ana@test.com").Return(user, nil)

	r := gin.New()
	r.POST("/auth/login", h.Login)

	req, _ := http.NewRequest(http.MethodPost, "/auth/login",
		bytes.NewBufferString(`{"email":"ana@test.com","password":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid credentials")
	repo.AssertExpectations(t)
}
