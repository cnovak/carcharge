package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const baseURL = "https://api.sense.com/apiservice/api/v1/authenticate"
const wsURL = "wss://clientrt.sense.com/monitors/%v/realtimefeed?access_token=%s"

const contentType = "application/x-www-form-urlencoded; charset=UTF-8;"

type EnergyService interface {
	getRealTime() (*PowerUsage, error)
}

type SenseService struct {
	token     string
	monitorID float64
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

func (c *SenseService) getToken(username string, password string) error {

	jsonBody := []byte(fmt.Sprintf("email=%s&password=%s", username, password))
	bodyReader := bytes.NewReader(jsonBody)

	resp, err := http.Post(baseURL, contentType, bodyReader)
	if err != nil {
		return fmt.Errorf("error autentication to Sense: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("io.ReadAll() error(%v)", err)
	}

	var httpResult map[string]interface{}
	json.Unmarshal([]byte(body), &httpResult)

	// Check for error in the JSON response
	if httpResult["status"] == "error" {
		msg := fmt.Sprintf("error returned from sense getToken(): %v", httpResult["error_reason"])
		log.Println(msg)
		return errors.New(msg)
	}

	c.token = httpResult["access_token"].(string)
	c.monitorID = httpResult["monitors"].([]interface{})[0].(map[string]interface{})["id"].(float64)
	return nil
}

func (c *SenseService) getRealTime() (*PowerUsage, error) {

	addr := fmt.Sprintf(wsURL, c.monitorID, c.token)

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
