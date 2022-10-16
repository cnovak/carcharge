/*
// Copyright © 2022 Chris Novak <canovak@gmail.com>
//
*/
package services

// import (
// 	"testing"

// 	"github.com/bogosj/tesla"
// 	"github.com/cnovak/carcharge/internal/services"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// )

// type

// type MockTeslaClient struct {
// }

// func (m *MockTeslaClient) Vehicles() ([]*tesla.Vehicle, error) {

// }

// func newTeslaClientMock() TeslaClient {

// }

// func TestTeslaService(t *testing.T) {
// 	client := mocks.NewTeslaClient(t)

// 	teslaService, err := services.NewTeslaService(client)
// 	assert.Nil(t, err)

// 	var myTests = []struct {
// 		powerUsage  PowerUsage
// 		chargeWatts int
// 	}{
// 		{PowerUsage{100, 500}, -400},
// 		{PowerUsage{1000, 500}, 500},
// 		{PowerUsage{2020, 0}, 2020},
// 		{PowerUsage{-1, 500}, -501},
// 		{PowerUsage{0, 500}, -500},
// 	}
//}
