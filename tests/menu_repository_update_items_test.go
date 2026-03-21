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

func TestMenuRepository_Update_ReplacesItems(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewMenuRepository(testDB)
	rest := createRepoTestRestaurant(t)
	menu := &models.Menu{
		RestaurantID: rest.ID,
		Name:         fmt.Sprintf("Replace Items %d", time.Now().UnixNano()),
		Items:        []models.MenuItem{{Name: "Old Burger", Price: 8.50, Available: true}},
	}
	require.NoError(t, repo.Create(menu))
	updated, err := repo.Update(menu.ID, &models.UpdateMenuRequest{
		Name: "After Replace",
		Items: []models.MenuItemRequest{
			{Name: "New Pasta", Price: 12.25, Available: true},
			{Name: "New Pizza", Price: 15.75, Available: false},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Len(t, updated.Items, 2)
	found, err := repo.FindByID(menu.ID)
	require.NoError(t, err)
	var names []string
	for _, it := range found.Items { names = append(names, it.Name) }
	assert.Contains(t, names, "New Pasta")
	assert.NotContains(t, names, "Old Burger")
}
