/*
Copyright © 2022 Chris Novak <canovak@gmail.com>
*/
package services

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/apex/log"
	"github.com/bogosj/tesla"
	"github.com/cnovak/carcharge/internal/util"
	"golang.org/x/oauth2"
)

type CarService interface {
	ChargeCar(targetWatts int) error
}

type TeslaService struct {
	client TeslaClient
	logCtx *log.Entry
}

type TeslaClient interface {
	Vehicles() ([]*tesla.Vehicle, error)
}

func GetVehicleClient() (*tesla.Client, error) {

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
	return carClient, nil
}

func NewTeslaService(client TeslaClient) (*TeslaService, error) {

	return &TeslaService{
		client: client,
	}, nil
}

func (t *TeslaService) getVehicle() (*tesla.Vehicle, error) {

	vehicles, err := t.client.Vehicles()
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

func (t *TeslaService) stopChargingCar(vehicle *tesla.Vehicle) error {

	ctx := log.WithFields(log.Fields{
		"vin":     vehicle.Vin,
		"carName": vehicle.DisplayName,
	})

	vehicle, err := t.teslaWakeup(vehicle)
	if err != nil {
		ctx.WithError(err).Error("waking vehicle")
		return err
	}

	chargeState, err := vehicle.ChargeState()
	if err != nil {
		ctx.WithError(err).Error("error getting charge state")
		return err
	}

	// if charging is stopped, we are done
	if chargeState.ChargingState == "Stopped" {
		ctx.WithField("ChargeState", chargeState.ChargingState).Debug("charging already stopped")
		return nil
	}

	err = vehicle.StopCharging()
	if err != nil {
		ctx.WithError(err).Error("Error stopping car charging")
		return err
	}

	// // Set charging amps back up to max
	// err = vehicle.SetChargingAmps(chargeState.ChargeCurrentRequestMax)

	ctx.Debug("Stopped charging")
	return nil
}

func (t *TeslaService) ChargeCar(deltaWatts int) error {
	vehicle, err := t.getVehicle()
	if err != nil {
		log.WithError(err).Error("ChargeCar: cannot get vehicle information")
		return err
	}

	ctx := log.WithFields(log.Fields{
		"vin":     vehicle.Vin,
		"carName": vehicle.DisplayName,
	})

	vehicle, err = t.teslaWakeup(vehicle)
	if err != nil {
		ctx.WithError(err).Error("waking vehicle")
		return err
	}

	chargeState, err := vehicle.ChargeState()
	if err != nil {
		ctx.WithError(err).Error("error getting charge state")
		return err
	}

	switch chargeState.ChargingState {
	case "Disconnected":
		ctx.WithField("ChargeState", chargeState.ChargingState).Info("Car disconnected")
		return nil
	case "Complete":
		ctx.WithField("ChargeState", chargeState.ChargingState).Debug("Charging complete")
		return nil
	}

	volts := util.Config.Tesla.ChargerVolts
	currentAmps := chargeState.ChargeAmps
	// How much power are we currently consuming?
	var currentWatts int
	if chargeState.ChargingState == "Stopped" {
		currentWatts = 0
	} else {
		currentWatts = currentAmps * volts
	}
	ctx.WithField("ChargeState", chargeState.ChargingState).Debug("Charge State")

	targetWatts := currentWatts + deltaWatts

	targetAmps := targetWatts / volts
	// actualWatts := targetAmps * volts
	ctx.WithFields(log.Fields{"currentWatts": currentWatts, "deltaWatts": deltaWatts, "targetWatts": targetWatts, "targetAmps": targetAmps, "volts": volts}).Info("calculated amps")

	// if amps < 4 since min charge is 5, stop charging
	// else match target amps
	if targetAmps < 1 {
		// stop charging
		ctx.Info("stop charging")
		t.stopChargingCar(vehicle)
		return nil
	}

	// set target amps
	if targetAmps == currentAmps {
		ctx.WithFields(log.Fields{"targetAmps": targetAmps, "currentAmps": currentAmps}).Debug("target amps equals current amps")
		return nil
	}

	// try setting amps 5 times, sometimes Tesla API seemst to not
	// set amps correctly when amps < 5
	maxAmpRetries := 5
	for i := 0; i < maxAmpRetries; i++ {
		// Set amps
		ctx.WithFields(log.Fields{"targetAmps": targetAmps, "currentAmps": currentAmps}).Debug("setting charging amps")
		err = vehicle.SetChargingAmps(targetAmps)
		if err != nil {
			ctx.WithError(err).Error("setting charging amps")
			return err
		}
		//time.Sleep(time.Duration(2) * time.Second)
		// validate amps are set correctly
		chargeState, err := vehicle.ChargeState()
		ctx.WithFields(log.Fields{"chargeState.ChargeAmps": chargeState.ChargeAmps, "targetAmps": targetAmps}).Error("Check currentAmps match targetAmps")
		if err != nil {
			ctx.WithError(err).Error("error getting charge state")
		} else {
			if chargeState.ChargeAmps == targetAmps {
				break
			} else {
				ctx.WithFields(log.Fields{"chargeState.ChargeAmps": chargeState.ChargeAmps, "targetAmps": targetAmps}).Error("Setting amps did not work")
			}
		}
	}

	// if charging is already happening, we are done
	if chargeState.ChargingState == "Charging" {
		ctx.WithField("ChargeState", chargeState.ChargingState).Debug("charging already started")
		return nil
	}

	err = vehicle.StartCharging()
	if err != nil {
		ctx.WithError(err).Error("starting charge failed")
		return err
	}
	ctx.Info("charging started")
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
