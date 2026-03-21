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

func createRepoTestMenuWithItems(t *testing.T) (*models.Restaurant, *models.Menu) {
	t.Helper()
	menuRepo := repository.NewMenuRepository(testDB)
	rest := createRepoTestRestaurant(t)
	menu := &models.Menu{
		RestaurantID: rest.ID,
		Name:         fmt.Sprintf("Repo Order Menu %d", time.Now().UnixNano()),
		Items: []models.MenuItem{
			{Name: "Pasta", Price: 12.50, Available: true},
			{Name: "Pizza", Price: 15.00, Available: true},
		},
	}
	err := menuRepo.Create(menu)
	require.NoError(t, err)
	require.NotEmpty(t, menu.Items[0].ID)
	return rest, menu
}

func TestNewReservationRepository(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewReservationRepository(testDB)
	assert.NotNil(t, repo)
}

func TestReservationRepository_Create_And_FindByID(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewReservationRepository(testDB)
	rest := createRepoTestRestaurant(t)
	user := createRepoTestUser(t, models.RoleClient)
	res := &models.Reservation{
		RestaurantID: rest.ID, UserID: user.ID,
		Date: time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second),
		PartySize: 4, Status: models.StatusPending, Notes: "window table",
	}
	err := repo.Create(res)
	require.NoError(t, err)
	require.NotEmpty(t, res.ID)
	found, err := repo.FindByID(res.ID)
	require.NoError(t, err)
	assert.Equal(t, res.ID, found.ID)
	assert.Equal(t, "window table", found.Notes)
}

func TestReservationRepository_FindByID_NotFound(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewReservationRepository(testDB)
	res, err := repo.FindByID("00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)
	assert.Nil(t, res)
}

func TestReservationRepository_Cancel(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewReservationRepository(testDB)
	rest := createRepoTestRestaurant(t)
	user := createRepoTestUser(t, models.RoleClient)
	res := &models.Reservation{
		RestaurantID: rest.ID, UserID: user.ID,
		Date: time.Now().Add(48 * time.Hour).UTC().Truncate(time.Second),
		PartySize: 2, Status: models.StatusPending,
	}
	require.NoError(t, repo.Create(res))
	require.NoError(t, repo.Cancel(res.ID))
	found, err := repo.FindByID(res.ID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusCancelled, found.Status)
}

func TestReservationRepository_CheckAvailability(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewReservationRepository(testDB)
	rest := createRepoTestRestaurant(t)
	available, err := repo.CheckAvailability(rest.ID, 7)
	require.NoError(t, err)
	assert.Equal(t, rest.Capacity-7, available)
}

func TestNewOrderRepository(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewOrderRepository(testDB)
	assert.NotNil(t, repo)
}

func TestOrderRepository_Create_And_FindByID(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewOrderRepository(testDB)
	user := createRepoTestUser(t, models.RoleClient)
	rest, menu := createRepoTestMenuWithItems(t)
	order := &models.Order{
		UserID: user.ID, RestaurantID: rest.ID,
		Total: 42.50, Status: models.StatusPending, Pickup: true,
		Items: []models.OrderItem{
			{MenuItemID: menu.Items[0].ID, Quantity: 1, Price: menu.Items[0].Price},
			{MenuItemID: menu.Items[1].ID, Quantity: 2, Price: menu.Items[1].Price},
		},
	}
	err := repo.Create(order)
	require.NoError(t, err)
	require.NotEmpty(t, order.ID)
	found, err := repo.FindByID(order.ID)
	require.NoError(t, err)
	assert.Equal(t, order.ID, found.ID)
	assert.Len(t, found.Items, 2)
}

func TestOrderRepository_FindByID_NotFound(t *testing.T) {
	setupIntegration(t)
	repo := repository.NewOrderRepository(testDB)
	order, err := repo.FindByID("00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)
	assert.Nil(t, order)
}
