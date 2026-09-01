package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type PasswordHasher struct {
	mock.Mock
}

var _ domain.PasswordHasher = (*PasswordHasher)(nil)

func (m *PasswordHasher) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *PasswordHasher) Compare(hash, password string) error {
	return m.Called(hash, password).Error(0)
}
