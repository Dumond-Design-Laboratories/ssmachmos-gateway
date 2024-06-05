package model

import (
	"encoding/json"
	"errors"
	"os"
)

const GATEWAY_FILE = "gateway.json"

type Gateway struct {
	Id               string    `json:"id"`
	Password         string    `json:"password"`
	DataCharUUID     [4]uint32 `json:"data_char_uuid"`
	SettingsCharUUID [4]uint32 `json:"settings_char_uuid"`
}

func LoadSettings(gateway *Gateway, path string) error {
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

func SetGatewayId(gateway *Gateway, id string) error {
	gateway.Id = id
	return saveSettings(gateway, GATEWAY_FILE)
}

func SetGatewayPassword(gateway *Gateway, password string) error {
	gateway.Password = password
	return saveSettings(gateway, GATEWAY_FILE)
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
		err = saveSettings(gateway, GATEWAY_FILE)
		if err != nil {
			return [4]uint32{}, err
		}
	}
	return gateway.DataCharUUID, nil
}

func GetSettingsCharUUID(gateway *Gateway) ([4]uint32, error) {
	if gateway == nil {
		return [4]uint32{}, errors.New("gateway is nil")
	}
	if gateway.DataCharUUID == [4]uint32{0, 0, 0, 0} {
		var err error
		gateway.SettingsCharUUID, err = GenerateUUID()
		if err != nil {
			return [4]uint32{}, err
		}
		err = saveSettings(gateway, GATEWAY_FILE)
		if err != nil {
			return [4]uint32{}, err
		}
	}
	return gateway.SettingsCharUUID, nil
}

func saveSettings(gateway *Gateway, path string) error {
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
