package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
	AuthError        bool      `json:"-"` // memory only flag for error reporting, special tag to omit from json
}

type RequestBody struct {
	GatewayId       string                   `json:"gateway_id"`
	GatewayPassword string                   `json:"gateway_password"`
	Measurements    []map[string]interface{} `json:"measurements"`
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

func SetGatewayHTTPEndpoint(gateway *Gateway, endpoint string) error {
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

func TestGateway(gateway *Gateway) error {
	data, _ := json.Marshal(RequestBody{
		GatewayId:       gateway.Id,
		GatewayPassword: gateway.Password,
	})
	resp, err := http.Post(gateway.HTTPEndpoint, "application/json", bytes.NewBuffer(data))

	if err == nil && resp.StatusCode == http.StatusOK {
		return nil
	}
	if err != nil {
		return err
	}
	bytes, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	return errors.New(fmt.Sprintf("HTTP Status %d - %s", resp.StatusCode, string(bytes)))
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

	jsonStr, err := json.MarshalIndent(gateway, "", "\t")
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
