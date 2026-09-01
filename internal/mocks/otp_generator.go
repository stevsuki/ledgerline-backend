package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type OTPGenerator struct {
	mock.Mock
}

var _ domain.OTPGenerator = (*OTPGenerator)(nil)

func (m *OTPGenerator) Generate() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}
