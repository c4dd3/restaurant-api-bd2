package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"restaurant-api/internal/auth"
	"restaurant-api/internal/models"
)

// ClaimsKey is the key used to store and retrieve JWT claims in the Gin context.
// Handlers call ExtractClaims(c) instead of using this key directly.
const ClaimsKey = "claims"

// Auth is a Gin middleware that protects routes by requiring a valid JWT token.
// It reads the token from the "Authorization: Bearer <token>" header, validates it,
// and stores the decoded claims in the context so handlers can read them.
// If the header is missing or the token is invalid, the request is rejected with 401.
func Auth(jwtSvc *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read the Authorization header and make sure it follows the "Bearer <token>" format.
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}

		// Strip the "Bearer " prefix to get the raw token string, then validate it.
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwtSvc.ValidateToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Store the validated claims in the context so downstream handlers can access them.
		c.Set(ClaimsKey, claims)
		c.Next()
	}
}

// AdminOnly is a Gin middleware that restricts a route to admin users only.
// It must be placed after the Auth middleware, since it relies on the claims
// that Auth stores in the context. Rejects non-admin requests with 403.
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := ExtractClaims(c)
		if claims == nil || claims.Role != models.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}

// ExtractClaims retrieves the JWT claims stored in the Gin context by the Auth middleware.
// Returns nil if the claims are not present or have an unexpected type,
// which handlers treat as "not authenticated".
func ExtractClaims(c *gin.Context) *models.Claims {
	val, exists := c.Get(ClaimsKey)
	if !exists {
		return nil
	}
	// Type-assert to *models.Claims — returns nil if the stored value has a different type.
	claims, ok := val.(*models.Claims)
	if !ok {
		return nil
	}
	return claims
}
