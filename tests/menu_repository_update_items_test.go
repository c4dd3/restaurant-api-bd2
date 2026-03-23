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
		Description:  "before",
		Items: []models.MenuItem{
			{
				Name:        "Old Burger",
				Description: "old",
				Price:       8.50,
				Available:   true,
			},
		},
	}

	require.NoError(t, repo.Create(menu))
	require.NotEmpty(t, menu.ID)

	updateReq := &models.UpdateMenuRequest{
		Name:        "After Replace",
		Description: "after",
		Items: []models.MenuItemRequest{
			{
				Name:        "New Pasta",
				Description: "new 1",
				Price:       12.25,
				Available:   true,
			},
			{
				Name:        "New Pizza",
				Description: "new 2",
				Price:       15.75,
				Available:   false,
			},
		},
	}

	updated, err := repo.Update(menu.ID, updateReq)
	require.NoError(t, err)
	require.NotNil(t, updated)

	assert.Equal(t, "After Replace", updated.Name)
	assert.Equal(t, "after", updated.Description)
	assert.Len(t, updated.Items, 2)

	found, err := repo.FindByID(menu.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, "After Replace", found.Name)
	assert.Equal(t, "after", found.Description)
	assert.Len(t, found.Items, 2)

	var names []string
	for _, it := range found.Items {
		names = append(names, it.Name)
	}

	assert.Contains(t, names, "New Pasta")
	assert.Contains(t, names, "New Pizza")
	assert.NotContains(t, names, "Old Burger")
}
