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

func TestMenuRepository_Update_OnlyName(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewMenuRepository(testDB)
	rest := createRepoTestRestaurant(t)
	menu := &models.Menu{RestaurantID: rest.ID, Name: fmt.Sprintf("Only Name Before %d", time.Now().UnixNano()), Description: "desc before", Items: []models.MenuItem{{Name: "Burger", Price: 8.99, Available: true}}}
	require.NoError(t, repo.Create(menu))
	updated, err := repo.Update(menu.ID, &models.UpdateMenuRequest{Name: "Only Name After"})
	require.NoError(t, err)
	assert.Equal(t, "Only Name After", updated.Name)
	assert.Equal(t, "desc before", updated.Description)
}

func TestMenuRepository_Update_OnlyDescription(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewMenuRepository(testDB)
	rest := createRepoTestRestaurant(t)
	menu := &models.Menu{RestaurantID: rest.ID, Name: fmt.Sprintf("Only Desc Before %d", time.Now().UnixNano()), Description: "desc before", Items: []models.MenuItem{{Name: "Pizza", Price: 12.99, Available: true}}}
	require.NoError(t, repo.Create(menu))
	updated, err := repo.Update(menu.ID, &models.UpdateMenuRequest{Description: "desc after"})
	require.NoError(t, err)
	assert.Equal(t, menu.Name, updated.Name)
	assert.Equal(t, "desc after", updated.Description)
}

func TestMenuRepository_Update_EmptyValuesKeepOriginal(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewMenuRepository(testDB)
	rest := createRepoTestRestaurant(t)
	menu := &models.Menu{RestaurantID: rest.ID, Name: fmt.Sprintf("Keep Original %d", time.Now().UnixNano()), Description: "original description", Items: []models.MenuItem{{Name: "Soup", Price: 4.99, Available: true}}}
	require.NoError(t, repo.Create(menu))
	updated, err := repo.Update(menu.ID, &models.UpdateMenuRequest{})
	require.NoError(t, err)
	assert.Equal(t, menu.Name, updated.Name)
	assert.Equal(t, menu.Description, updated.Description)
}
