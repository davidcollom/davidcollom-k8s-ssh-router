package mocks

import (
	"github.com/stretchr/testify/mock"
	"k8s.io/client-go/tools/remotecommand"
)

type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) Stream(options remotecommand.StreamOptions) error {
	args := m.Called(options)
	return args.Error(0)
}
