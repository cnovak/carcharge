package pkg

import (
	"context"
	"fmt"
	"os"
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

func Rebalance() {
	// senseClient, _ := NewClient(config.Sense.Username, config.Sense.Password)
	// realtimeMessage, err := senseClient.getRealTime()
	powerUsage := PowerUsage{
		1200,
		621,
	}

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
		// chargeState, _ := carClient.getChargeState()
		// fmt.Printf("reducing charging car to reduce usage delta: %vw\n", power.solarProduction-power.energyUsage)
		// carClient.setChargingAmps(int(chargeState.ChargeRate) - 1)
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
	// chargeCar(2000)
}

func chargeCar(watts int) {
	token := new(oauth2.Token)

	token.AccessToken = util.Config.Tesla.AccessToken
	token.RefreshToken = util.Config.Tesla.RefreshToken
	token.TokenType = "Bearer"
	token.Expiry = time.Now()

	log.WithFields(log.Fields{
		"AccessToken":  "XXXX" + token.AccessToken[len(token.AccessToken)-7:],
		"RefreshToken": "XXXX" + token.RefreshToken[len(token.RefreshToken)-7:],
	}).Debug("chargeCar() called")

	carClient, err := tesla.NewClient(context.Background(), tesla.WithToken(token))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	vehicles, err := carClient.Vehicles()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ctx := log.WithFields(log.Fields{"vehicleCount": len(vehicles)})
	ctx.Debug("got Tesla Vehicles")

	var vehicle *tesla.Vehicle
	for _, v := range vehicles {
		if v.Vin == util.Config.Tesla.VehicleVin {
			vehicle = v
		}
	}

	ctx = log.WithFields(log.Fields{
		"vin":     vehicle.Vin,
		"carName": vehicle.DisplayName,
	})

	ctx.Debug("Vehicle found")

	volts := util.Config.Tesla.ChargerVolts
	amps := watts / volts
	ctx.WithFields(log.Fields{"watts": watts, "amps": amps, "volts": volts}).Info("calculated amps")

	vehicle, err = teslaWakeup(vehicle)
	if err != nil {
		ctx.WithError(err).Error("waking vehicle")
		return
	}

	if amps > 0 {
		ctx.WithField("amps", amps).Info("setting charging amps")
		err = vehicle.SetChargingAmps(amps)
		if err != nil {
			ctx.WithError(err).Error("setting charging amps")
			return
		}
	} else {
		ctx.WithField("amps", amps).Debug("skipping setting charging amps")
	}

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
