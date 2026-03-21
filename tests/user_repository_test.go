package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"restaurant-api/internal/models"
	"restaurant-api/internal/repository"
)

func TestUserRepository_CRUD(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewUserRepository(testDB)
	email := fmt.Sprintf("repo_user_%d@test.com", time.Now().UnixNano())
	user := &models.User{Name: "Repo User", Email: email, Password: "hashed-password", Role: models.RoleClient}
	err := repo.Create(user)
	require.NoError(t, err)
	require.NotEmpty(t, user.ID)
	foundByEmail, err := repo.FindByEmail(email)
	require.NoError(t, err)
	require.NotNil(t, foundByEmail)
	assert.Equal(t, user.ID, foundByEmail.ID)
	foundByID, err := repo.FindByID(user.ID)
	require.NoError(t, err)
	require.NotNil(t, foundByID)
	updated, err := repo.Update(user.ID, &models.UpdateUserRequest{Name: "Repo User Updated"})
	require.NoError(t, err)
	assert.Equal(t, "Repo User Updated", updated.Name)
	err = repo.Delete(user.ID)
	require.NoError(t, err)
	foundAfterDelete, err := repo.FindByID(user.ID)
	require.NoError(t, err)
	assert.Nil(t, foundAfterDelete)
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewUserRepository(testDB)
	user, err := repo.FindByEmail(fmt.Sprintf("missing_%d@test.com", time.Now().UnixNano()))
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewUserRepository(testDB)
	user, err := repo.FindByID("00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestUserRepository_Update_NotFound(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewUserRepository(testDB)
	user, err := repo.Update("00000000-0000-0000-0000-000000000000", &models.UpdateUserRequest{Name: "Nobody"})
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestNewUserRepository(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewUserRepository(testDB)
	assert.NotNil(t, repo)
}
