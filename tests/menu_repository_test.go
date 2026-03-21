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

func createRepoTestRestaurant(t *testing.T) *models.Restaurant {
	t.Helper()
	restRepo := repository.NewRestaurantRepository(testDB)
	admin := createRepoTestUser(t, models.RoleAdmin)
	rest := &models.Restaurant{
		Name:        fmt.Sprintf("Repo Menu Resto %d", time.Now().UnixNano()),
		Address:     "123 Menu Street",
		Phone:       "8888-9999",
		Description: "restaurant for menu repo tests",
		AdminID:     admin.ID,
		Capacity:    60,
	}
	err := restRepo.Create(rest)
	require.NoError(t, err)
	require.NotEmpty(t, rest.ID)
	return rest
}

func TestNewMenuRepository(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewMenuRepository(testDB)
	assert.NotNil(t, repo)
}

func TestMenuRepository_Create_And_FindByID(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewMenuRepository(testDB)
	rest := createRepoTestRestaurant(t)
	menu := &models.Menu{
		RestaurantID: rest.ID,
		Name:         fmt.Sprintf("Repo Menu %d", time.Now().UnixNano()),
		Description:  "test menu",
		Items: []models.MenuItem{
			{Name: "Pasta", Description: "white sauce", Price: 12.50, Available: true},
			{Name: "Pizza", Description: "pepperoni", Price: 15.00, Available: true},
		},
	}
	err := repo.Create(menu)
	require.NoError(t, err)
	require.NotEmpty(t, menu.ID)
	found, err := repo.FindByID(menu.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, menu.ID, found.ID)
	assert.Len(t, found.Items, 2)
}

func TestMenuRepository_FindByID_NotFound(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewMenuRepository(testDB)
	menu, err := repo.FindByID("00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)
	assert.Nil(t, menu)
}

func TestMenuRepository_Update(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewMenuRepository(testDB)
	rest := createRepoTestRestaurant(t)
	menu := &models.Menu{
		RestaurantID: rest.ID,
		Name:         fmt.Sprintf("Update Menu %d", time.Now().UnixNano()),
		Description:  "before update",
		Items:        []models.MenuItem{{Name: "Burger", Price: 8.99, Available: true}},
	}
	require.NoError(t, repo.Create(menu))
	updated, err := repo.Update(menu.ID, &models.UpdateMenuRequest{Name: "Updated Menu Name", Description: "after update"})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "Updated Menu Name", updated.Name)
}

func TestMenuRepository_Update_NotFound(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewMenuRepository(testDB)
	menu, err := repo.Update("00000000-0000-0000-0000-000000000000", &models.UpdateMenuRequest{Name: "Nobody"})
	require.NoError(t, err)
	assert.Nil(t, menu)
}

func TestMenuRepository_FindItemByID(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewMenuRepository(testDB)
	rest := createRepoTestRestaurant(t)
	menu := &models.Menu{
		RestaurantID: rest.ID,
		Name:         fmt.Sprintf("Item Lookup Menu %d", time.Now().UnixNano()),
		Items:        []models.MenuItem{{Name: "Lasagna", Price: 13.25, Available: true}},
	}
	require.NoError(t, repo.Create(menu))
	foundMenu, err := repo.FindByID(menu.ID)
	require.NoError(t, err)
	require.Len(t, foundMenu.Items, 1)
	item, err := repo.FindItemByID(foundMenu.Items[0].ID)
	require.NoError(t, err)
	assert.Equal(t, "Lasagna", item.Name)
}

func TestMenuRepository_FindItemByID_NotFound(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewMenuRepository(testDB)
	item, err := repo.FindItemByID("00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)
	assert.Nil(t, item)
}

func TestMenuRepository_Delete(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewMenuRepository(testDB)
	rest := createRepoTestRestaurant(t)
	menu := &models.Menu{
		RestaurantID: rest.ID,
		Name:         fmt.Sprintf("Delete Menu %d", time.Now().UnixNano()),
		Items:        []models.MenuItem{{Name: "Soup", Price: 4.50, Available: true}},
	}
	require.NoError(t, repo.Create(menu))
	require.NoError(t, repo.Delete(menu.ID))
	found, err := repo.FindByID(menu.ID)
	require.NoError(t, err)
	assert.Nil(t, found)
}
