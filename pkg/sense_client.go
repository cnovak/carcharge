package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const baseUrl = "https://api.sense.com/apiservice/api/v1/authenticate"
const wsUrl = "wss://clientrt.sense.com/monitors/%v/realtimefeed?access_token=%s&"
const contentType = "application/x-www-form-urlencoded; charset=UTF-8;"

type EnergyService interface {
	getRealTime() (*PowerUsage, error)
}

type SenseService struct {
	token     string
	monitorId float64
}

type PowerUsage struct {
	solarProduction float64
	energyUsage     float64
}

func NewSenseService(username string, password string) (me *SenseService, err error) {

	me = &SenseService{}

	me.getToken(username, password)

	return me, nil
}

func (c *SenseService) getToken(username string, password string) {

	jsonBody := []byte(fmt.Sprintf("email=%s&password=%s", username, password))
	bodyReader := bytes.NewReader(jsonBody)

	resp, err := http.Post(baseUrl, contentType, bodyReader)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("ioutil.ReadAll() error(%v)", err)
	}

	var httpResult map[string]interface{}
	json.Unmarshal([]byte(body), &httpResult)

	c.token = httpResult["access_token"].(string)
	c.monitorId = httpResult["monitors"].([]interface{})[0].(map[string]interface{})["id"].(float64)

}

func (c *SenseService) getRealTime() (*PowerUsage, error) {

	//addr := fmt.Sprintf(wsUrl, c.monitorId, c.token)
	addr := fmt.Sprintf(wsUrl, c.monitorId, c.token)

	//u := url.URL{Scheme: "ws", Host: addr, Path: "/echo"}
	log.Printf("\nconnecting to %s\n", addr)

	wsConnnection, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer wsConnnection.Close()

	var realTimeMsg *PowerUsage = nil

	var rawResponse map[string]interface{}
	for realTimeMsg == nil {
		_, message, err := wsConnnection.ReadMessage()
		if err != nil {
			log.Println("wsConnnection.ReadMessage error:", err)
			return nil, err
		}

		json.Unmarshal(message, &rawResponse)

		if rawResponse["type"] == "error" {
			payload := rawResponse["payload"].(map[string]interface{})
			errMsg := fmt.Sprintf("Server error: %s", payload["error_reason"])
			return nil, errors.New(errMsg)
		}

		if rawResponse["type"] == "realtime_update" {
			payload := rawResponse["payload"].(map[string]interface{})
			usage := payload["w"].(float64)
			solar := payload["solar_w"].(float64)
			realTimeMsg = &PowerUsage{solarProduction: solar, energyUsage: usage}

		}
	}

	return realTimeMsg, nil

}
