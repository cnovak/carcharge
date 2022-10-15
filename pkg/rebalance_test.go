/*
Copyright © 2022 Chris Novak <canovak@gmail.com>
*/
package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSenseService struct {
	mock.Mock
}

func (m *MockSenseService) getRealTime() (*PowerUsage, error) {
	args := m.Called()
	return args.Get(0).(*PowerUsage), args.Error(1)
}

type MockTeslaService struct {
	mock.Mock
}

func (m *MockTeslaService) chargeCar(targetWatts int) error {
	args := m.Called(targetWatts)
	return args.Error(0)
}

func TestRebalance(t *testing.T) {

	mockSenseService := &MockSenseService{}
	mockTeslaService := &MockTeslaService{}
	rb := NewRebalancer(mockSenseService, mockTeslaService)

	var myTests = []struct {
		powerUsage  PowerUsage
		chargeWatts int
	}{
		{PowerUsage{100, 500}, -400},
		{PowerUsage{1000, 500}, 500},
		{PowerUsage{2020, 0}, 2020},
		{PowerUsage{-1, 500}, -501},
		{PowerUsage{0, 500}, -500},
	}

	for _, testData := range myTests {
		senseMockCall := mockSenseService.On("getRealTime").Return(&testData.powerUsage, nil).Once()
		teslaMockCall := mockTeslaService.On("chargeCar", testData.chargeWatts).Return(nil).Once()
		err := rb.Rebalance()
		assert.Nil(t, err)
		mockSenseService.AssertExpectations(t)
		mockTeslaService.AssertExpectations(t)
		// remove the handler now so we can add another one that takes precedence
		senseMockCall.Unset()
		teslaMockCall.Unset()

	}

}

func TestRebalanceWhenBalanced(t *testing.T) {

	mockSenseService := &MockSenseService{}
	mockTeslaService := &MockTeslaService{}
	rb := NewRebalancer(mockSenseService, mockTeslaService)

	// If power is balanced do not make a call
	myTests := []struct {
		powerUsage PowerUsage
	}{
		{PowerUsage{2200, 2200}},
		{PowerUsage{10, 10}},
		{PowerUsage{0, 0}},
	}

	for _, testData := range myTests {
		senseMockCall := mockSenseService.On("getRealTime").Return(&testData.powerUsage, nil).Once()
		err := rb.Rebalance()
		assert.Nil(t, err)
		mockSenseService.AssertExpectations(t)
		mockTeslaService.AssertExpectations(t)
		// remove the handler now so we can add another one that takes precedence
		senseMockCall.Unset()
	}

}
