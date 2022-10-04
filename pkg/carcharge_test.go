package carcharge

import (
	"testing"

	"github.com/jsgoecke/tesla"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// type MockSenseClient struct {
// 	returnMsg RealtimeMessage
// }

// func (c *MockSenseClient) getRealTime() (*RealtimeMessage, error) {
// 	m := RealtimeMessage{50, 100}
// 	return &m, nil
// }

// MyMockedObject is a mocked object that implements an interface
// that describes an object that the code I am testing relies on.
type MockEnergyClient struct {
	mock.Mock
}

func (c *MockEnergyClient) getRealTime() (*RealtimeMessage, error) {

	args := c.Called()
	return args.Get(0).(*RealtimeMessage), args.Error(1)
}

type MockCarClient struct {
	mock.Mock
}

func (c *MockCarClient) startCharge() error {

	args := c.Called()
	return args.Error(0)
}

func (c *MockCarClient) stopCharge() error {

	args := c.Called()
	return args.Error(0)
}

func (c *MockCarClient) getChargeState() (tesla.ChargeState, error) {
	args := c.Called()
	return args.Get(0).(tesla.ChargeState), args.Error(1)
}

func (c *MockCarClient) setChargingAmps(amps int) error {
	args := c.Called()
	return args.Error(0)
}

func TestModifyCharge(t *testing.T) {

	mockEnergyClient := &MockEnergyClient{}
	mockEnergyClient.On("getRealTime").Return(&RealtimeMessage{2000, 500}, nil).Once()
	mockEnergyClient.On("getRealTime").Return(&RealtimeMessage{2000, 800}, nil).Once()
	mockEnergyClient.On("getRealTime").Return(&RealtimeMessage{2000, 1600}, nil).Once()
	mockEnergyClient.On("getRealTime").Return(&RealtimeMessage{2000, 1800}, nil).Once()

	mockCarClient := &MockCarClient{}
	mockCarClient.On("startCharge").Return(nil)
	//mockCarClient.On("getChargeState").Return(tesla.ChargeState{ChargeRate: 5}, errors.New("Fail Duckworth"))
	mockCarClient.On("getChargeState").Return(tesla.ChargeState{ChargeRate: 5}, nil)
	mockCarClient.On("setChargingAmps").Return(nil)

	error := modifyCharge(mockEnergyClient, mockCarClient)

	assert.Nil(t, error)

	// assert that the expectations were met
	mockEnergyClient.AssertExpectations(t)
	mockCarClient.AssertExpectations(t)
}
