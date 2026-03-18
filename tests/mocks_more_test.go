package tests

import (
	"restaurant-api/internal/models"

	"github.com/stretchr/testify/mock"
)

type MockRestaurantRepo struct{ mock.Mock }

func (m *MockRestaurantRepo) Create(rest *models.Restaurant) error {
	args := m.Called(rest)
	if rest != nil && rest.ID == "" {
		rest.ID = "rest-1"
	}
	return args.Error(0)
}
func (m *MockRestaurantRepo) FindAll() ([]models.Restaurant, error) {
	args := m.Called()
	return args.Get(0).([]models.Restaurant), args.Error(1)
}
func (m *MockRestaurantRepo) FindByID(id string) (*models.Restaurant, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Restaurant), args.Error(1)
}

type MockMenuRepo struct{ mock.Mock }

func (m *MockMenuRepo) Create(menu *models.Menu) error {
	args := m.Called(menu)
	if menu != nil && menu.ID == "" {
		menu.ID = "menu-1"
	}
	return args.Error(0)
}
func (m *MockMenuRepo) FindByID(id string) (*models.Menu, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Menu), args.Error(1)
}
func (m *MockMenuRepo) Update(id string, req *models.UpdateMenuRequest) (*models.Menu, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Menu), args.Error(1)
}
func (m *MockMenuRepo) Delete(id string) error {
	return m.Called(id).Error(0)
}
func (m *MockMenuRepo) FindItemByID(id string) (*models.MenuItem, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MenuItem), args.Error(1)
}

type MockReservationRepo struct{ mock.Mock }

func (m *MockReservationRepo) Create(res *models.Reservation) error {
	args := m.Called(res)
	if res != nil && res.ID == "" {
		res.ID = "res-1"
	}
	return args.Error(0)
}
func (m *MockReservationRepo) FindByID(id string) (*models.Reservation, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Reservation), args.Error(1)
}
func (m *MockReservationRepo) Cancel(id string) error {
	return m.Called(id).Error(0)
}
func (m *MockReservationRepo) CheckAvailability(restaurantID string, partySize int) (int, error) {
	args := m.Called(restaurantID, partySize)
	return args.Int(0), args.Error(1)
}

type MockOrderRepo struct{ mock.Mock }

func (m *MockOrderRepo) Create(order *models.Order) error {
	args := m.Called(order)
	if order != nil && order.ID == "" {
		order.ID = "order-1"
	}
	return args.Error(0)
}
func (m *MockOrderRepo) FindByID(id string) (*models.Order, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}
