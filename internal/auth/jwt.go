package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"restaurant-api/internal/models"
)

// ErrInvalidToken is returned whenever a token cannot be parsed or its signature is invalid.
var ErrInvalidToken = errors.New("invalid token")

// JWTService holds the signing secret and token lifetime used across all JWT operations.
type JWTService struct {
	secret     []byte
	expiration time.Duration
}

// NewJWTService creates a JWTService using the JWT_SECRET env var (falls back to a default
// insecure key if the variable is unset — must be overridden in production).
func NewJWTService() *JWTService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "super-secret-key-change-in-production"
	}
	return &JWTService{
		secret:     []byte(secret),
		expiration: 24 * time.Hour,
	}
}

// jwtClaims extends the standard registered claims with application-specific user fields.
type jwtClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed HS256 JWT for the given user, valid for the configured duration.
func (s *JWTService) GenerateToken(user *models.User) (string, error) {
	claims := jwtClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidateToken parses and verifies a JWT string, returning the embedded claims on success.
func (s *JWTService) ValidateToken(tokenStr string) (*models.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		// Reject tokens signed with any algorithm other than HMAC (e.g. "alg:none" attacks).
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Type-assert to our custom claims struct to access the application fields.
	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return &models.Claims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
	}, nil
}
