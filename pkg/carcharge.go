package carcharge

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

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

	realtimeMessage, err := client.getRealTime()

	if err != nil {
		log.Fatalf("ERROR 32: %v", err)
	}
	fmt.Printf("\nUsage %+v, Production:%v\n", realtimeMessage.energyUsage, realtimeMessage.solarProduction)

	if realtimeMessage.energyUsage < realtimeMessage.solarProduction {
		fmt.Printf("Charging car to cover %vw\n", realtimeMessage.solarProduction-realtimeMessage.energyUsage)
	}
}
