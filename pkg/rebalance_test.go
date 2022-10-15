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
	mockSenseService.On("getRealTime").Return(&PowerUsage{100, 500}, nil).Once()

	mockTeslaService := &MockTeslaService{}
	mockTeslaService.On("chargeCar", -400).Return(nil).Once()

	rb := NewRebalancer(mockSenseService, mockTeslaService)
	err := rb.Rebalance()
	assert.Nil(t, err)

	// // assert that the expectations were met
	mockSenseService.AssertExpectations(t)
	mockTeslaService.AssertExpectations(t)
}

// mockEnergyClient := &MockEnergyClient{}
// mockEnergyClient.On("getRealTime").Return(&RealtimeMessage{2000, 500}, nil).Once()
// mockEnergyClient.On("getRealTime").Return(&RealtimeMessage{2000, 800}, nil).Once()
// mockEnergyClient.On("getRealTime").Return(&RealtimeMessage{2000, 1600}, nil).Once()
// mockEnergyClient.On("getRealTime").Return(&RealtimeMessage{2000, 1800}, nil).Once()

// mockCarClient := &MockCarClient{}
// mockCarClient.On("startCharge").Return(nil)
// //mockCarClient.On("getChargeState").Return(tesla.ChargeState{ChargeRate: 5}, errors.New("Fail Duckworth"))
// mockCarClient.On("getChargeState").Return(tesla.ChargeState{ChargeRate: 5}, nil)
// mockCarClient.On("setChargingAmps").Return(nil)

// error := modifyCharge(mockEnergyClient, mockCarClient)

// assert.Nil(t, error)

// // assert that the expectations were met
// mockEnergyClient.AssertExpectations(t)
// mockCarClient.AssertExpectations(t)
// }

// func Test_teslaWakeup(t *testing.T) {

// 	v1 := tesla.Vehicle{
// 		Color:       nil,
// 		DisplayName: "v1",
// 		ID:          123,
// 		OptionCodes: "options",

// 		VehicleID:              3422,
// 		Vin:                    "DV2344",
// 		State:                  "sleeps",
// 		IDS:                    "ids",
// 		RemoteStartEnabled:     true,
// 		CalendarEnabled:        false,
// 		NotificationsEnabled:   true,
// 		BackseatToken:          "backseat token",
// 		BackseatTokenUpdatedAt: "updated at",
// 		AccessType:             "Access",
// 		InService:              false,
// 		APIVersion:             1.0,
// 		CommandSigning:         "Cmd",
// 		VehicleConfig:          nil,
// 	}

// 	type args struct {
// 		vehicle *tesla.Vehicle
// 	}

// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    *tesla.Vehicle
// 		wantErr bool
// 	}{
// 		{
// 			name: "test1",
// 			args: args{
// 				vehicle: &v1,
// 			},
// 			want:    &v1,
// 			wantErr: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := teslaWakeup(tt.args.vehicle)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("teslaWakeup() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("teslaWakeup() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
