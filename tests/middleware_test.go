package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"restaurant-api/internal/auth"
	"restaurant-api/internal/middleware"
	"restaurant-api/internal/models"
)

func TestAuth_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtSvc := auth.NewJWTService()
	r := gin.New()

	r.GET("/protected", middleware.Auth(jwtSvc), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing or invalid authorization header")
}

func TestAuth_InvalidHeaderFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtSvc := auth.NewJWTService()
	r := gin.New()

	r.GET("/protected", middleware.Auth(jwtSvc), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token abc123")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing or invalid authorization header")
}

func TestAuth_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtSvc := auth.NewJWTService()
	r := gin.New()

	r.GET("/protected", middleware.Auth(jwtSvc), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.token")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
}

func TestAuth_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtSvc := auth.NewJWTService()
	token, err := jwtSvc.GenerateToken(&models.User{
		ID:    "user-1",
		Email: "user@test.com",
		Role:  models.RoleClient,
	})
	assert.NoError(t, err)

	r := gin.New()
	r.GET("/protected", middleware.Auth(jwtSvc), func(c *gin.Context) {
		claims := middleware.ExtractClaims(c)
		if claims == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "claims missing"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"user_id": claims.UserID,
			"role":    claims.Role,
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"user_id":"user-1"`)
	assert.Contains(t, w.Body.String(), `"role":"client"`)
}

func TestAdminOnly_NoClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/admin", middleware.AdminOnly(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "admin access required")
}

func TestAdminOnly_ClientForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		c.Set(middleware.ClaimsKey, &models.Claims{
			UserID: "user-1",
			Role:   models.RoleClient,
		})
		middleware.AdminOnly()(c)
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "admin access required")
}

func TestAdminOnly_AdminAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		c.Set(middleware.ClaimsKey, &models.Claims{
			UserID: "admin-1",
			Role:   models.RoleAdmin,
		})
		middleware.AdminOnly()(c)
	}, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"ok":true`)
}

func TestExtractClaims_WrongType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set(middleware.ClaimsKey, "not-claims")

	claims := middleware.ExtractClaims(c)
	assert.Nil(t, claims)
}
