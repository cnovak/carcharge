/*
Copyright © 2022 Chris Novak <canovak@gmail.com>
*/
package pkg

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/apex/log"
	"github.com/bogosj/tesla"
	"github.com/cnovak/carcharge/util"
	"golang.org/x/oauth2"
)

type CarService interface {
	ChargeCar(targetWatts int) error
}

type TeslaService struct {
	vehicle *tesla.Vehicle
	logCtx  *log.Entry
}

func NewTeslaService(vehicle *tesla.Vehicle) (*TeslaService, error) {
	ctx := log.WithFields(log.Fields{
		"vin":     vehicle.Vin,
		"carName": vehicle.DisplayName,
	})

	return &TeslaService{
		vehicle: vehicle,
		logCtx:  ctx,
	}, nil
}

func init() {

}

func GetVehicleClient() (*tesla.Vehicle, error) {

	token := new(oauth2.Token)
	token.AccessToken = util.Config.Tesla.AccessToken
	token.RefreshToken = util.Config.Tesla.RefreshToken
	token.TokenType = "Bearer"
	token.Expiry = time.Now()

	log.WithFields(log.Fields{
		"AccessToken":  "XXXX" + token.AccessToken[len(token.AccessToken)-7:],
		"RefreshToken": "XXXX" + token.RefreshToken[len(token.RefreshToken)-7:],
	}).Debug("getVehicleClient() called")

	carClient, err := tesla.NewClient(context.Background(), tesla.WithToken(token))
	if err != nil {
		log.WithError(err).Error("cannot get car client")
		return nil, err
	}

	vehicles, err := carClient.Vehicles()
	if err != nil {
		log.WithError(err).Error("cannot get vehcile list")
		return nil, err
	}
	ctx := log.WithFields(log.Fields{"vehicleCount": len(vehicles)})
	ctx.Debug("got Tesla Vehicles")

	var vehicle *tesla.Vehicle
	for _, v := range vehicles {
		if v.Vin == util.Config.Tesla.VehicleVin {
			vehicle = v
		}
	}

	if reflect.ValueOf(vehicle).IsZero() {
		err := errors.New("vehicle not found with VIN in configuration")
		ctx.WithField("vin", util.Config.Tesla.VehicleVin).WithError(err).Error("Vehicle not found in Vehicle list")
		return nil, err
	}

	ctx = log.WithFields(log.Fields{
		"vin":     vehicle.Vin,
		"carName": vehicle.DisplayName,
	})

	ctx.Debug("Vehicle found")
	return vehicle, err
}

func (t *TeslaService) stopChargingCar() error {

	vehicle, err := t.teslaWakeup(t.vehicle)
	if err != nil {
		t.logCtx.WithError(err).Error("waking vehicle")
		return err
	}

	chargeState, err := vehicle.ChargeState()
	if err != nil {
		t.logCtx.WithError(err).Error("error getting charge state")
		return err
	}

	// if charging is stopped, we are done
	if chargeState.ChargingState == "Stopped" {
		t.logCtx.WithField("ChargeState", chargeState.ChargingState).Debug("charging already stopped")
		return nil
	}

	err = vehicle.StopCharging()
	if err != nil {
		t.logCtx.WithError(err).Error("Error stopping car charging")
		return err
	}

	// // Set charging amps back up to max
	// err = vehicle.SetChargingAmps(chargeState.ChargeCurrentRequestMax)

	t.logCtx.Debug("Stopped charging")
	return nil
}

func (t *TeslaService) ChargeCar(targetWatts int) error {

	vehicle, err := t.teslaWakeup(t.vehicle)
	if err != nil {
		t.logCtx.WithError(err).Error("waking vehicle")
		return err
	}

	chargeState, err := vehicle.ChargeState()
	if err != nil {
		t.logCtx.WithError(err).Error("error getting charge state")
		return err
	}

	// figure out how many amps needed to match target
	volts := util.Config.Tesla.ChargerVolts
	targetAmps := targetWatts / volts
	actualWatts := targetAmps * volts
	t.logCtx.WithFields(log.Fields{"actualWatts": actualWatts, "targetWatts": targetWatts, "targetAmps": targetAmps, "volts": volts}).Info("calculated amps")

	// if amps < 4 since min charge is 5, stop charging
	// else match target amps
	if targetAmps <= 0 {
		// stop charging
		t.logCtx.Info("stop charging")
		t.stopChargingCar()
		return nil
	}

	// set target amps
	t.logCtx.WithField("amps", targetAmps).Info("setting charging amps")

	t.logCtx.Debug("setting charging amps")
	err = vehicle.SetChargingAmps(targetAmps)
	if err != nil {
		t.logCtx.WithError(err).Error("setting charging amps")
		return err
	}

	// if charging is already happening, we are done
	if chargeState.ChargingState == "Charging" {
		t.logCtx.WithField("ChargeState", chargeState.ChargingState).Debug("charging already started")
		return nil
	}

	err = vehicle.StartCharging()
	if err != nil {
		t.logCtx.WithError(err).Error("starting charge failed")
		return err
	}
	t.logCtx.Info("charging started")
	return nil
}

func (t *TeslaService) teslaWakeup(vehicle *tesla.Vehicle) (*tesla.Vehicle, error) {
	timeout := 15
	startTime := time.Now()

	ctx := log.WithFields(log.Fields{
		"VehicleState": vehicle.State,
		"VIN":          vehicle.Vin,
	})
	ctx.Debug("teslaWakeup()")

	var err error
	for int(time.Since(startTime).Seconds()) < timeout {

		vehicle, err := vehicle.Wakeup()
		ctx.Debug("wakeup called")
		if err == nil && vehicle.State == "online" {
			ctx.WithDuration(time.Since(startTime)).Debug("vehicle online")
			break
		}
		ctx.WithDuration(time.Since(startTime)).WithField("sleep", "3").Debug("vehicle not online, sleeping...")
		time.Sleep(3 * time.Second)
	}
	ctx.WithDuration(time.Since(startTime)).Debug("teslaWakeup() done")
	return vehicle, err
}
