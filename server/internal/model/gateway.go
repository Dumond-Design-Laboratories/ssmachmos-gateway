package model

import (
	"encoding/json"
	"errors"
	"os"
	"path"
)

const GATEWAY_FILE = "gateway.json"

type Gateway struct {
	Id               string    `json:"id"`
	Password         string    `json:"password"`
	DataCharUUID     [4]uint32 `json:"data_char_uuid"`
	SettingsCharUUID [4]uint32 `json:"settings_char_uuid"`
	HTTPEndpoint     string    `json:"http_endpoint"`
}

func LoadSettings(gateway *Gateway, fileName string) error {
	configPath, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Join(configPath, "ss_machmos"), 0777)
	if err != nil {
		return err
	}

	jsonStr, err := os.ReadFile(path.Join(configPath, "ss_machmos", fileName))
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

func SetHTTPEndpoint(gateway *Gateway, endpoint string) error {
	gateway.HTTPEndpoint = endpoint
	return saveSettings(gateway, GATEWAY_FILE)
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
	if gateway.SettingsCharUUID == [4]uint32{0, 0, 0, 0} {
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

func saveSettings(gateway *Gateway, fileName string) error {
	if gateway == nil {
		return errors.New("gateway is nil")
	}

	jsonStr, err := json.Marshal(gateway)
	if err != nil {
		return err
	}
	configPath, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Join(configPath, "ss_machmos"), 0777)
	if err != nil {
		return err
	}

	return os.WriteFile(path.Join(configPath, "ss_machmos", fileName), jsonStr, 0777)
}
