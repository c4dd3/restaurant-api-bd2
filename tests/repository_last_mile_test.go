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

func TestReservationRepository_CheckAvailability_WithExistingReservations(t *testing.T) {
	setupIntegration(t)

	repo := repository.NewReservationRepository(testDB)
	rest := createRepoTestRestaurant(t)
	user := createRepoTestUser(t, models.RoleClient)

	res1 := &models.Reservation{
		RestaurantID: rest.ID,
		UserID:       user.ID,
		Date:         time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second),
		PartySize:    4,
		Status:       models.StatusPending,
	}
	res2 := &models.Reservation{
		RestaurantID: rest.ID,
		UserID:       user.ID,
		Date:         time.Now().Add(48 * time.Hour).UTC().Truncate(time.Second),
		PartySize:    6,
		Status:       models.StatusPending,
	}

	require.NoError(t, repo.Create(res1))
	require.NoError(t, repo.Create(res2))

	available, err := repo.CheckAvailability(rest.ID, 5)
	require.NoError(t, err)

	assert.Equal(t, rest.Capacity-5, available)
}

func TestMenuRepository_Create_WithoutItems_And_FindByID(t *testing.T) {
	setupIntegration(t)

	repo := repository.NewMenuRepository(testDB)
	rest := createRepoTestRestaurant(t)

	menu := &models.Menu{
		RestaurantID: rest.ID,
		Name:         fmt.Sprintf("No Items %d", time.Now().UnixNano()),
		Description:  "menu without items",
		Items:        []models.MenuItem{},
	}

	require.NoError(t, repo.Create(menu))
	require.NotEmpty(t, menu.ID)

	found, err := repo.FindByID(menu.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, menu.ID, found.ID)
	assert.Equal(t, menu.Name, found.Name)
	assert.Len(t, found.Items, 0)
}

func TestOrderRepository_Create_And_FindByID_WithReservationID(t *testing.T) {
	setupIntegration(t)

	orderRepo := repository.NewOrderRepository(testDB)
	resRepo := repository.NewReservationRepository(testDB)

	user := createRepoTestUser(t, models.RoleClient)
	rest, menu := createRepoTestMenuWithItems(t)

	res := &models.Reservation{
		RestaurantID: rest.ID,
		UserID:       user.ID,
		Date:         time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second),
		PartySize:    2,
		Status:       models.StatusPending,
	}
	require.NoError(t, resRepo.Create(res))
	require.NotEmpty(t, res.ID)

	reservationID := res.ID

	order := &models.Order{
		UserID:        user.ID,
		RestaurantID:  rest.ID,
		ReservationID: &reservationID,
		Total:         menu.Items[0].Price,
		Status:        models.StatusPending,
		Pickup:        false,
		Items: []models.OrderItem{
			{
				MenuItemID: menu.Items[0].ID,
				Quantity:   1,
				Price:      menu.Items[0].Price,
			},
		},
	}

	require.NoError(t, orderRepo.Create(order))
	require.NotEmpty(t, order.ID)

	found, err := orderRepo.FindByID(order.ID)
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, order.ID, found.ID)
	require.NotNil(t, found.ReservationID)
	assert.Equal(t, res.ID, *found.ReservationID)
	assert.Len(t, found.Items, 1)
}
