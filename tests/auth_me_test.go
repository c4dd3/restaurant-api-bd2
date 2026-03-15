package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"restaurant-api/internal/auth"
	"restaurant-api/internal/handlers"
	"restaurant-api/internal/models"
)

func TestMe_Unauthorized_NoClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockUserRepo)
	jwtSvc := auth.NewJWTService()
	h := handlers.NewAuthHandler(repo, jwtSvc)

	r := gin.New()
	r.GET("/users/me", h.Me)

	req, _ := http.NewRequest(http.MethodGet, "/users/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMe_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockUserRepo)
	jwtSvc := auth.NewJWTService()
	h := handlers.NewAuthHandler(repo, jwtSvc)

	user := &models.User{ID: "user-1", Name: "Ana", Email: "ana@test.com", Role: models.RoleClient}
	repo.On("FindByID", "user-1").Return(user, nil)

	r := gin.New()
	r.GET("/users/me", func(c *gin.Context) {
		c.Set("claims", &models.Claims{UserID: "user-1", Role: models.RoleClient})
		h.Me(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/users/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ana@test.com")
	repo.AssertExpectations(t)
}
