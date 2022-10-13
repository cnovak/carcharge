package pkg

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/apex/log"
	"github.com/bogosj/tesla"
	"golang.org/x/oauth2"

	"github.com/cnovak/carcharge/util"
)

// have around 500
const tolerance = 500

type EnergyClient interface {
	getCurrentUsage() (*PowerUsage, error)
}

type CarClient interface {
	getChargeState() (tesla.ChargeState, error)
	stopCharge() error
	startCharge() error
	setChargingAmps(amps int) error
}

func Rebalance() error {
	senseClient, _ := NewClient(util.Config.Sense.Username, util.Config.Sense.Password)
	powerUsage, err := senseClient.getRealTime()
	if err != nil {
		log.WithError(err).Error("error getting power usage from Sense")
		return err
	}
	// powerUsage := PowerUsage{
	// 	00,
	// 	1200,
	// }

	ctx := log.WithFields(log.Fields{
		"solarProduction": powerUsage.solarProduction,
		"energyUsage":     powerUsage.energyUsage,
	})
	ctx.Debug("current power usage")

	// get the current solar and energy used
	// fmt.Printf("amps := watts / volts \n")
	// fmt.Printf("%v := %v  / %v  \n", amps, watts, volts)

	// fmt.Printf("Starting to charge %s with %v amps \n", vehicle.DisplayName, amps)
	powerNeeded := powerUsage.solarProduction - powerUsage.energyUsage
	ctx = ctx.WithFields(log.Fields{
		"powerNeeded": powerNeeded,
	})

	ctx.Debug("power needed")

	delta := tolerance / 2
	negDelta := delta * -1

	if powerNeeded < float64(negDelta) {
		// modify charging to be less
		ctx.Info("reduce usage")
		stopChargingCar()
	} else if powerNeeded > tolerance/2 {
		// modify charging to be more
		ctx.Info("increase usage")
		// fmt.Printf("Charging car to cover usage delta %vw\n", power.solarProduction-power.energyUsage)
		chargeCar(int(powerNeeded))
		// chargeState, error := carClient.getChargeState()
		// if error != nil {
		// 	log.Fatalf("ERROR 33: %v", err)
		// }
		// carClient.setChargingAmps(int(chargeState.ChargeRate) + 1)
	} else {
		ctx.Info("charging balanced")
		// fmt.Printf("Charging balanced\n")
	}
	ctx.Info("sleeping for 15 mintues...")
	time.Sleep(15 * time.Minute)
	return nil
}

func stopChargingCar() error {
	err, ctx, vehicle := getVehicleClient()
	if err != nil {
		return err
	}

	ctx = log.WithFields(log.Fields{
		"vin":     vehicle.Vin,
		"carName": vehicle.DisplayName,
	})

	err = vehicle.StopCharging()
	if err != nil {
		ctx.WithError(err).Error("Error stopping car charging")
		return err
	}

	ctx.Debug("Stopped charging")
	return nil
}

func chargeCar(watts int) error {

	err, ctx, vehicle := getVehicleClient()

	volts := util.Config.Tesla.ChargerVolts
	amps := watts / volts
	ctx.WithFields(log.Fields{"watts": watts, "amps": amps, "volts": volts}).Info("calculated amps")

	vehicle, err = teslaWakeup(vehicle)
	if err != nil {
		ctx.WithError(err).Error("waking vehicle")
		return err
	}

	if amps > 0 {
		ctx.WithField("amps", amps).Info("setting charging amps")
		err = vehicle.SetChargingAmps(amps)
		if err != nil {
			ctx.WithError(err).Error("setting charging amps")
			return err
		}
	} else {
		ctx.WithField("amps", amps).Debug("skipping setting charging amps")
	}

	err = vehicle.StartCharging()
	if err != nil {
		ctx.WithError(err).Error("starting charge failed")
		return err
	}
	ctx.Info("charging started")
	return nil
}

func getVehicleClient() (error, *log.Entry, *tesla.Vehicle) {

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
		return err, nil, nil
	}

	vehicles, err := carClient.Vehicles()
	if err != nil {
		log.WithError(err).Error("cannot get vehcile list")
		return err, nil, nil
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
		return err, nil, nil
	}

	ctx = log.WithFields(log.Fields{
		"vin":     vehicle.Vin,
		"carName": vehicle.DisplayName,
	})

	ctx.Debug("Vehicle found")
	return err, ctx, vehicle
}

func modifyCharge(senseClient EnergyClient, carClient CarClient) error {
	isBalanced := false

	// keep trying to balance
	for !isBalanced {
		power, err := senseClient.getCurrentUsage()

		if err != nil {
			log.Fatalf("ERROR 32: %v", err)
		}
		fmt.Printf("\nUsage %+v, Production:%v\n", power.energyUsage, power.solarProduction)

		powerNeeded := power.solarProduction - power.energyUsage

		delta := tolerance / 2
		negDelta := delta * -1

		if powerNeeded < float64(negDelta) {
			// modify charging to be less
			chargeState, _ := carClient.getChargeState()
			fmt.Printf("reducing charging car to reduce usage delta: %vw\n", power.solarProduction-power.energyUsage)
			carClient.setChargingAmps(int(chargeState.ChargeRate) - 1)
		} else if powerNeeded > tolerance/2 {
			// modify charging to be more
			fmt.Printf("Charging car to cover usage delta %vw\n", power.solarProduction-power.energyUsage)
			carClient.startCharge()
			chargeState, error := carClient.getChargeState()
			if error != nil {
				log.Fatalf("ERROR 33: %v", err)
			}
			carClient.setChargingAmps(int(chargeState.ChargeRate) + 1)
		} else {
			isBalanced = true
			fmt.Printf("Charging balanced\n")
		}
	}
	return nil
}

func teslaWakeup(vehicle *tesla.Vehicle) (*tesla.Vehicle, error) {
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
