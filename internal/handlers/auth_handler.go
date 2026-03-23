package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"restaurant-api/internal/auth"
	"restaurant-api/internal/middleware"
	"restaurant-api/internal/models"
)

// AuthHandler handles HTTP requests related to authentication (register, login, profile).
type AuthHandler struct {
	userRepo UserRepository
	jwtSvc   *auth.JWTService
}

// NewAuthHandler creates an AuthHandler with the given user repository and JWT service.
func NewAuthHandler(userRepo UserRepository, jwtSvc *auth.JWTService) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, jwtSvc: jwtSvc}
}

// Register creates a new user account.
// Steps: validate the request body → check the email is not taken → hash the password
// → save the user in the DB → return a JWT token so the user is logged in right away.
func (h *AuthHandler) Register(c *gin.Context) {
	// Parse and validate the JSON request body.
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Reject registration if the email is already used by another account.
	existing, err := h.userRepo.FindByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
		return
	}

	// Hash the plain-text password before storing it — never save passwords as plain text.
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error hashing password"})
		return
	}

	// Build the user model with the hashed password.
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashed),
		Role:     req.Role,
	}

	// Persist the new user to the database.
	if err := h.userRepo.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error creating user"})
		return
	}

	// Generate a JWT token so the user is authenticated immediately after registering.
	token, err := h.jwtSvc.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error generating token"})
		return
	}

	c.JSON(http.StatusCreated, models.LoginResponse{Token: token, User: *user})
}

// Login authenticates an existing user.
// Steps: validate the request body → look up the user by email → verify the password
// → return a JWT token on success.
func (h *AuthHandler) Login(c *gin.Context) {
	// Parse and validate the JSON request body.
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Look up the user by email. Return a generic error to avoid leaking whether
	// the email exists in the system.
	user, err := h.userRepo.FindByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Compare the provided password against the stored hash.
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Credentials are valid — issue a JWT token for the session.
	token, err := h.jwtSvc.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error generating token"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{Token: token, User: *user})
}

// Me returns the profile of the currently logged-in user.
// It reads the user ID from the JWT claims that the auth middleware already validated.
func (h *AuthHandler) Me(c *gin.Context) {
	// Extract the claims injected into the context by the JWT middleware.
	claims := middleware.ExtractClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Fetch the full user record from the database using the ID in the token.
	user, err := h.userRepo.FindByID(claims.UserID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
