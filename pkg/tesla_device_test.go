package pkg

// type MockSenseClient struct {
// 	returnMsg RealtimeMessage
// }

// func (c *MockSenseClient) getRealTime() (*RealtimeMessage, error) {
// 	m := RealtimeMessage{50, 100}
// 	return &m, nil
// }

// func (c *MockEnergyClient) getRealTime() (*RealtimeMessage, error) {

// 	args := c.Called()
// 	return args.Get(0).(*RealtimeMessage), args.Error(1)
// }

// type MockCarClient struct {
// 	mock.Mock
// }

// func (c *MockCarClient) startCharge() error {

// 	args := c.Called()
// 	return args.Error(0)
// }

// func (c *MockCarClient) stopCharge() error {

// 	args := c.Called()
// 	return args.Error(0)
// }

// func (c *MockCarClient) getChargeState() (tesla.ChargeState, error) {
// 	args := c.Called()
// 	return args.Get(0).(tesla.ChargeState), args.Error(1)
// }

// func (c *MockCarClient) setChargingAmps(amps int) error {
// 	args := c.Called()
// 	return args.Error(0)
// }

// func TestModifyCharge(t *testing.T) {

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
