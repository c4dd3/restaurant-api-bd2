package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"restaurant-api/internal/middleware"
	"restaurant-api/internal/models"
)

// UserHandler handles HTTP requests for updating and deleting user accounts.
type UserHandler struct {
	userRepo UserRepository
}

// NewUserHandler creates a UserHandler with the given user repository.
func NewUserHandler(userRepo UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

// Update modifies the fields of an existing user identified by the URL parameter ID.
// Only the user themselves or an admin can update the account.
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")

	// Get the caller's identity from the JWT claims injected by the auth middleware.
	claims := middleware.ExtractClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Prevent a regular user from updating someone else's account.
	// Admins can update any account.
	if claims.Role != models.RoleAdmin && claims.UserID != id {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot update another user"})
		return
	}

	// Parse and validate the JSON request body.
	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Apply the update; the repository returns nil if no user with that ID exists.
	user, err := h.userRepo.Update(id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error updating user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Delete removes a user account identified by the URL parameter ID.
// Only the user themselves or an admin can delete the account.
// Returns 204 No Content on success (no body needed).
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	// Get the caller's identity from the JWT claims injected by the auth middleware.
	claims := middleware.ExtractClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Prevent a regular user from deleting someone else's account.
	// Admins can delete any account.
	if claims.Role != models.RoleAdmin && claims.UserID != id {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete another user"})
		return
	}

	if err := h.userRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting user"})
		return
	}

	c.Status(http.StatusNoContent)
}
