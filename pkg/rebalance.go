package pkg

import (
	"github.com/apex/log"
)

type Rebalancer interface {
	Rebalance() error
}

type RebalancerImpl struct {
	energyService EnergyService
	carService    CarService
}

func NewRebalancer(energyService EnergyService, carService CarService) *RebalancerImpl {
	return &RebalancerImpl{
		energyService: energyService,
		carService:    carService,
	}
}

func (rb *RebalancerImpl) Rebalance() error {

	powerUsage, err := rb.energyService.getRealTime()
	if err != nil {
		log.WithError(err).Error("error getting power usage from Sense")
		return err
	}

	ctx := log.WithFields(log.Fields{
		"solarProduction": powerUsage.solarProduction,
		"energyUsage":     powerUsage.energyUsage,
	})
	ctx.Debug("current power usage")

	powerNeeded := powerUsage.solarProduction - powerUsage.energyUsage
	ctx = ctx.WithFields(log.Fields{
		"powerNeeded": powerNeeded,
	})

	ctx.Debug("power needed")

	if powerNeeded < 0 {
		// modify charging to be less
		ctx.Info("reduce usage")
		//stopChargingCar()
		err := rb.carService.chargeCar(int(powerNeeded))
		if err != nil {
			log.WithError(err).Error("error charging car")
		}
	} else if powerNeeded > 0 {
		// modify charging to be more
		ctx.Info("increase usage")
		err := rb.carService.chargeCar(int(powerNeeded))
		if err != nil {
			log.WithError(err).Error("error charging car")
		}
	} else {
		ctx.Info("charging balanced")
	}
	return nil
}
