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

func createRepoTestUser(t *testing.T, role string) *models.User {
	t.Helper()

	userRepo := repository.NewUserRepository(testDB)

	user := &models.User{
		Name:     "Repo Admin",
		Email:    fmt.Sprintf("repo_admin_%d@test.com", time.Now().UnixNano()),
		Password: "hashed-password",
		Role:     role,
	}

	err := userRepo.Create(user)
	require.NoError(t, err)
	require.NotEmpty(t, user.ID)

	return user
}

func TestNewRestaurantRepository(t *testing.T) {
	setupIntegration(t)

	repo := repository.NewRestaurantRepository(testDB)
	assert.NotNil(t, repo)
}

func TestRestaurantRepository_Create_And_FindByID(t *testing.T) {
	setupIntegration(t)

	repo := repository.NewRestaurantRepository(testDB)
	admin := createRepoTestUser(t, models.RoleAdmin)

	rest := &models.Restaurant{
		Name:        fmt.Sprintf("Repo Resto %d", time.Now().UnixNano()),
		Address:     "123 Test Street",
		Phone:       "2222-3333",
		Description: "test restaurant",
		AdminID:     admin.ID,
		Capacity:    80,
	}

	err := repo.Create(rest)
	require.NoError(t, err)
	require.NotEmpty(t, rest.ID)
	assert.False(t, rest.CreatedAt.IsZero())
	assert.False(t, rest.UpdatedAt.IsZero())

	found, err := repo.FindByID(rest.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, rest.ID, found.ID)
	assert.Equal(t, rest.Name, found.Name)
	assert.Equal(t, rest.Address, found.Address)
	assert.Equal(t, rest.Phone, found.Phone)
	assert.Equal(t, rest.Description, found.Description)
	assert.Equal(t, rest.AdminID, found.AdminID)
	assert.Equal(t, rest.Capacity, found.Capacity)
}

func TestRestaurantRepository_FindAll(t *testing.T) {
	setupIntegration(t)

	repo := repository.NewRestaurantRepository(testDB)
	admin := createRepoTestUser(t, models.RoleAdmin)

	rest1 := &models.Restaurant{
		Name:        fmt.Sprintf("Repo FindAll A %d", time.Now().UnixNano()),
		Address:     "Address A",
		Phone:       "1111-1111",
		Description: "desc A",
		AdminID:     admin.ID,
		Capacity:    50,
	}
	rest2 := &models.Restaurant{
		Name:        fmt.Sprintf("Repo FindAll B %d", time.Now().UnixNano()),
		Address:     "Address B",
		Phone:       "2222-2222",
		Description: "desc B",
		AdminID:     admin.ID,
		Capacity:    90,
	}

	require.NoError(t, repo.Create(rest1))
	require.NoError(t, repo.Create(rest2))

	restaurants, err := repo.FindAll()
	require.NoError(t, err)
	require.NotNil(t, restaurants)

	var found1, found2 bool
	for _, r := range restaurants {
		if r.ID == rest1.ID {
			found1 = true
		}
		if r.ID == rest2.ID {
			found2 = true
		}
	}

	assert.True(t, found1)
	assert.True(t, found2)
}

func TestRestaurantRepository_FindByID_NotFound(t *testing.T) {
	setupIntegration(t)

	repo := repository.NewRestaurantRepository(testDB)

	rest, err := repo.FindByID("00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)
	assert.Nil(t, rest)
}
