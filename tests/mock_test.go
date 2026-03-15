package tests

import (
	"github.com/stretchr/testify/mock"

	"restaurant-api/internal/models"
)

type MockUserRepo struct{ mock.Mock }

func (m *MockUserRepo) FindByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) FindByID(id string) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) Create(user *models.User) error {
	args := m.Called(user)
	// Si tu repo real setea ID, aquí puedes simularlo:
	if user != nil && user.ID == "" {
		user.ID = "user-123"
	}
	return args.Error(0)
}

func (m *MockUserRepo) Update(id string, req *models.UpdateUserRequest) (*models.User, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) Delete(id string) error {
	return m.Called(id).Error(0)
}
