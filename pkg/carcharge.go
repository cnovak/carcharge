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

var config2 *util.Configuration

func Rebalance(config *util.Configuration) {
	config2 = config
	// senseClient, _ := NewClient(config.Sense.Username, config.Sense.Password)
	// realtimeMessage, err := senseClient.getRealTime()
	powerUsage := PowerUsage{
		5312,
		621,
	}

	ctx := log.WithFields(log.Fields{
		"power": powerUsage,
		"cmd":   "rebalance",
	})
	ctx.Debug("got power usage")

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
		log.Info("reduce usage")
		// chargeState, _ := carClient.getChargeState()
		// fmt.Printf("reducing charging car to reduce usage delta: %vw\n", power.solarProduction-power.energyUsage)
		// carClient.setChargingAmps(int(chargeState.ChargeRate) - 1)
	} else if powerNeeded > tolerance/2 {
		// modify charging to be more
		log.Info("increase usage")
		// fmt.Printf("Charging car to cover usage delta %vw\n", power.solarProduction-power.energyUsage)
		chargeCar(int(powerNeeded))
		// chargeState, error := carClient.getChargeState()
		// if error != nil {
		// 	log.Fatalf("ERROR 33: %v", err)
		// }
		// carClient.setChargingAmps(int(chargeState.ChargeRate) + 1)
	} else {
		log.Info("charging balanced")
		// fmt.Printf("Charging balanced\n")
	}
	// chargeCar(2000)
}

func chargeCar(watts int) {
	token := new(oauth2.Token)

	token.AccessToken = config2.Tesla.AccessToken
	token.RefreshToken = config2.Tesla.RefreshToken
	token.TokenType = "Bearer"
	token.Expiry = time.Now()

	// token.Expiry = {{ From DataBase }}
	// token.TokenType = {{ From DataBase }}

	carClient, err := tesla.NewClient(context.Background(), tesla.WithToken(token))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ctx := log.WithFields(log.Fields{
		"token.AccessToken":  token.AccessToken[0:5],
		"token.RefreshToken": token.RefreshToken[0:5],
	})
	ctx.Debug("got Car Client instance")

	vehicles, err := carClient.Vehicles()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	vin := util.Config.Tesla.VehicleVin
	var vehicle *tesla.Vehicle

	for _, v := range vehicles {
		if v.Vin == vin {
			vehicle = v
		}
	}

	ctx = log.WithFields(log.Fields{
		"vin":     vehicle.Vin,
		"carName": vehicle.DisplayName,
	})

	ctx.Debug("new client2")
	volts := util.Config.Tesla.ChargerVolts
	amps := watts / volts

	ctx.Debug("amps := watts / volts \n")
	ctx.Debugf("%v := %v  / %v  \n", amps, watts, volts)

	ctx.Debugf("Starting to charge %s with %v amps \n", vehicle.DisplayName, amps)

	vehicle, err = teslaWakeup(vehicle)
	if err != nil {
		ctx.WithError(err).Error("waking vehicle")
		return
	}

	ctx.Debug("setting charging amps")
	err = vehicle.SetChargingAmps(amps)
	if err != nil {
		ctx.WithError(err).Error("setting charging amps")
		return
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

	fmt.Printf("time since: %v\n", time.Since(startTime))
	fmt.Printf("time.Since(startTime) < time.Second: %v\n", time.Since(startTime) < time.Second)

	var err error
	for int(time.Since(startTime).Seconds()) < timeout {
		log.WithField("VIN", vehicle.Vin).Debug("waking vehice")

		vehicle, err := vehicle.Wakeup()
		// err = errors.New("40X")
		if err == nil {
			log.WithField("timeElapsed", time.Since(startTime).Seconds()).Debug("vehicle.Wakeup")
			break
		}
		log.WithField("timeElapsed", time.Since(startTime).Seconds()).WithError(err).Error("vehicle.Wakeup")
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("vehicle state: %v\n", vehicle.State)
		time.Sleep(3 * time.Second)
	}
	return vehicle, err
}
