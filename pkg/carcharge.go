package carcharge

import (
	"fmt"
	"log"
	"os"

	"github.com/jsgoecke/tesla"
	"github.com/spf13/viper"
)

// have around 500
const tolerance = 500

type EnergyClient interface {
	getRealTime() (*RealtimeMessage, error)
}

type CarClient interface {
	getChargeState() (tesla.ChargeState, error)
	stopCharge() error
	startCharge() error
	setChargingAmps(amps int) error
}

func GetEnergy() {

	// jsonData := `{
	// 	"email":"abhirockzz@gmail.com",
	// 	"username":"abhirockzz",
	// 	"blogs":[
	// 		{"name":"devto","url":"https://dev.to/abhirockzz/"},
	// 		{"name":"medium","url":"https://medium.com/@abhishek1987/"}
	// 	]}THIS IS INTENTIONALLY MALFORMED NOW`

	// jsonDataReader := strings.NewReader(jsonData)
	// decoder := json.NewDecoder(bodyReader)
	// var profile map[string]interface{}
	// for {
	// 	err := decoder.Decode(&profile)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	if err == io.EOF {
	// 		break
	// 	}
	// }

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
	fmt.Println("USERNAME:", viper.Get("USERNAME"))

	client, _ := NewClient(viper.GetString("USERNAME"), viper.GetString("PASSWORD"))

	modifyCharge(client, nil)
}

func modifyCharge(senseClient EnergyClient, carClient CarClient) error {
	isBalanced := false

	// keep trying to balance
	for !isBalanced {
		realtimeMessage, err := senseClient.getRealTime()

		if err != nil {
			log.Fatalf("ERROR 32: %v", err)
		}
		fmt.Printf("\nUsage %+v, Production:%v\n", realtimeMessage.energyUsage, realtimeMessage.solarProduction)

		powerNeeded := realtimeMessage.solarProduction - realtimeMessage.energyUsage

		delta := tolerance / 2
		negDelta := delta * -1

		if powerNeeded < float64(negDelta) {
			// modify charging to be less
			chargeState, _ := carClient.getChargeState()
			fmt.Printf("reducing charging car to reduce usage delta: %vw\n", realtimeMessage.solarProduction-realtimeMessage.energyUsage)
			carClient.setChargingAmps(int(chargeState.ChargeRate) - 1)
		} else if powerNeeded > tolerance/2 {
			// modify charging to be more
			fmt.Printf("Charging car to cover usage delta %vw\n", realtimeMessage.solarProduction-realtimeMessage.energyUsage)
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
