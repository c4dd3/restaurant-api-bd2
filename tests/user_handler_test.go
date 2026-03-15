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

func TestUserUpdate_ForbiddenForOtherUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockUserRepo)
	h := handlers.NewUserHandler(repo)

	r := gin.New()
	r.PUT("/users/:id", func(c *gin.Context) {
		c.Set("claims", &models.Claims{UserID: "user-1", Role: models.RoleClient})
		h.Update(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/users/user-2", bytes.NewBufferString(`{"name":"X"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "cannot update another user")
	repo.AssertExpectations(t)
}

func TestUserUpdate_SuccessSelf(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockUserRepo)
	h := handlers.NewUserHandler(repo)

	updated := &models.User{ID: "user-1", Name: "Updated", Email: "a@b.com", Role: models.RoleClient}
	repo.On("Update", "user-1", mock.AnythingOfType("*models.UpdateUserRequest")).Return(updated, nil)

	r := gin.New()
	r.PUT("/users/:id", func(c *gin.Context) {
		c.Set("claims", &models.Claims{UserID: "user-1", Role: models.RoleClient})
		h.Update(c)
	})

	req, _ := http.NewRequest(http.MethodPut, "/users/user-1", bytes.NewBufferString(`{"name":"Updated"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Updated")
	repo.AssertExpectations(t)
}

func TestUserDelete_ForbiddenForOtherUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockUserRepo)
	h := handlers.NewUserHandler(repo)

	r := gin.New()
	r.DELETE("/users/:id", func(c *gin.Context) {
		c.Set("claims", &models.Claims{UserID: "user-1", Role: models.RoleClient})
		h.Delete(c)
	})

	req, _ := http.NewRequest(http.MethodDelete, "/users/user-2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "cannot delete another user")
	repo.AssertExpectations(t)
}

func TestUserDelete_SuccessSelf(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(MockUserRepo)
	h := handlers.NewUserHandler(repo)

	repo.On("Delete", "user-1").Return(nil)

	r := gin.New()
	r.DELETE("/users/:id", func(c *gin.Context) {
		c.Set("claims", &models.Claims{UserID: "user-1", Role: models.RoleClient})
		h.Delete(c)
	})

	req, _ := http.NewRequest(http.MethodDelete, "/users/user-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	repo.AssertExpectations(t)
}
