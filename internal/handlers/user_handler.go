package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"restaurant-api/internal/middleware"
	"restaurant-api/internal/models"
	"restaurant-api/internal/repository"
)

type UserHandler struct {
	userRepo *repository.UserRepository
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

// UpdateUser godoc
// @Summary      Update a user
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path  string  true  "User ID"
// @Param        body body  models.UpdateUserRequest true "Update data"
// @Success      200  {object}  models.User
// @Router       /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	claims := middleware.ExtractClaims(c)

	// Users can only update themselves; admins can update anyone
	if claims.Role != models.RoleAdmin && claims.UserID != id {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot update another user"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

// DeleteUser godoc
// @Summary      Delete a user
// @Tags         users
// @Security     BearerAuth
// @Param        id   path  string  true  "User ID"
// @Success      204
// @Router       /users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	claims := middleware.ExtractClaims(c)

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
