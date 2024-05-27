package model

import (
	"encoding/json"
	"errors"
	"os"
)

const GATEWAY_FILE = "gateway.json"

type Gateway struct {
	Id           string    `json:"id"`
	Password     string    `json:"password"`
	DataCharUUID [4]uint32 `json:"data_char_uuid"`
}

func LoadSettings(path string, gateway *Gateway) error {
	jsonStr, err := os.ReadFile(path)
	if err != nil {
		gateway = &Gateway{}
		return err
	}
	err = json.Unmarshal(jsonStr, gateway)
	if err != nil {
		gateway = &Gateway{}
		return err
	}
	return nil
}

func SetGatewayId(id string, gateway *Gateway) error {
	gateway.Id = id
	return saveSettings(GATEWAY_FILE, gateway)
}

func SetGatewayPassword(password string, gateway *Gateway) error {
	gateway.Password = password
	return saveSettings(GATEWAY_FILE, gateway)
}

func GetDataCharUUID(gateway *Gateway) ([4]uint32, error) {
	if gateway == nil {
		return [4]uint32{}, errors.New("gateway is nil")
	}
	if gateway.DataCharUUID == [4]uint32{0, 0, 0, 0} {
		var err error
		gateway.DataCharUUID, err = GenerateUUID()
		if err != nil {
			return [4]uint32{}, err
		}
		err = saveSettings(GATEWAY_FILE, gateway)
		if err != nil {
			return [4]uint32{}, err
		}
	}
	return gateway.DataCharUUID, nil
}

func saveSettings(path string, gateway *Gateway) error {
	if gateway == nil {
		return errors.New("gateway is nil")
	}

	jsonStr, err := json.Marshal(gateway)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, jsonStr, 0666)
	if err != nil {
		return err
	}
	return nil
}
