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
			{
				Name:        "Pasta",
				Description: "white sauce",
				Price:       12.50,
				Available:   true,
			},
			{
				Name:        "Pizza",
				Description: "pepperoni",
				Price:       15.00,
				Available:   true,
			},
		},
	}

	err := repo.Create(menu)
	require.NoError(t, err)
	require.NotEmpty(t, menu.ID)

	found, err := repo.FindByID(menu.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, menu.ID, found.ID)
	assert.Equal(t, rest.ID, found.RestaurantID)
	assert.Equal(t, menu.Name, found.Name)
	assert.Equal(t, menu.Description, found.Description)
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
		Items: []models.MenuItem{
			{
				Name:        "Burger",
				Description: "classic",
				Price:       8.99,
				Available:   true,
			},
		},
	}

	require.NoError(t, repo.Create(menu))
	require.NotEmpty(t, menu.ID)

	updateReq := &models.UpdateMenuRequest{
		Name:        "Updated Menu Name",
		Description: "after update",
	}

	updated, err := repo.Update(menu.ID, updateReq)
	require.NoError(t, err)
	require.NotNil(t, updated)

	assert.Equal(t, menu.ID, updated.ID)
	assert.Equal(t, "Updated Menu Name", updated.Name)
	assert.Equal(t, "after update", updated.Description)

	found, err := repo.FindByID(menu.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "Updated Menu Name", found.Name)
	assert.Equal(t, "after update", found.Description)
}

func TestMenuRepository_Update_NotFound(t *testing.T) {
	setupIntegration(t)

	repo := repository.NewMenuRepository(testDB)

	updateReq := &models.UpdateMenuRequest{
		Name:        "Nobody",
		Description: "nothing",
	}

	menu, err := repo.Update("00000000-0000-0000-0000-000000000000", updateReq)
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
		Description:  "find item test",
		Items: []models.MenuItem{
			{
				Name:        "Lasagna",
				Description: "cheesy",
				Price:       13.25,
				Available:   true,
			},
		},
	}

	require.NoError(t, repo.Create(menu))

	foundMenu, err := repo.FindByID(menu.ID)
	require.NoError(t, err)
	require.NotNil(t, foundMenu)
	require.Len(t, foundMenu.Items, 1)
	require.NotEmpty(t, foundMenu.Items[0].ID)

	item, err := repo.FindItemByID(foundMenu.Items[0].ID)
	require.NoError(t, err)
	require.NotNil(t, item)

	assert.Equal(t, foundMenu.Items[0].ID, item.ID)
	assert.Equal(t, "Lasagna", item.Name)
	assert.Equal(t, 13.25, item.Price)
	assert.Equal(t, true, item.Available)
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
		Description:  "to delete",
		Items: []models.MenuItem{
			{
				Name:        "Soup",
				Description: "hot",
				Price:       4.50,
				Available:   true,
			},
		},
	}

	require.NoError(t, repo.Create(menu))
	require.NotEmpty(t, menu.ID)

	err := repo.Delete(menu.ID)
	require.NoError(t, err)

	found, err := repo.FindByID(menu.ID)
	require.NoError(t, err)
	assert.Nil(t, found)
}
