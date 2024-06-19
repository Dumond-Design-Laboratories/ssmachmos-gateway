package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/jukuly/ss_machmos/server/internal/model"
	"github.com/jukuly/ss_machmos/server/internal/out"
)

type requestBody struct {
	GatewayId       string `json:"gateway_id"`
	GatewayPassword string `json:"gateway_password"`
	Measurements    string `json:"measurements"`
}

func sendMeasurements(jsonData []byte, gateway *model.Gateway) (*http.Response, error) {
	body := requestBody{
		GatewayId:       gateway.Id,
		GatewayPassword: gateway.Password,
		Measurements:    string(jsonData),
	}
	json, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return http.Post("https://openphm.org/gateway_data", "application/json", bytes.NewBuffer([]byte(json)))
}

func saveUnsentMeasurements(data []byte, timestamp int64) error {
	err := os.MkdirAll(path.Join(os.TempDir(), "ss_machmos", UNSENT_DATA_PATH), 0777)
	if err != nil {
		return err
	}

	return os.WriteFile(path.Join(os.TempDir(), "ss_machmos", UNSENT_DATA_PATH, fmt.Sprintf("%d.json", timestamp)), data, 0777)
}

func sendUnsentMeasurements() {
	files, err := os.ReadDir(path.Join(os.TempDir(), "ss_machmos", UNSENT_DATA_PATH))
	if err != nil {
		out.Logger.Println(err)
		return
	}

	for _, file := range files {
		data, err := os.ReadFile(path.Join(os.TempDir(), "ss_machmos", UNSENT_DATA_PATH, file.Name()))
		if err != nil {
			out.Logger.Println(err)
			continue
		}

		resp, err := sendMeasurements(data, Gateway)
		if err != nil {
			out.Logger.Println(err)
			continue
		}

		if resp.StatusCode == 200 {
			os.Remove(path.Join(os.TempDir(), "ss_machmos", UNSENT_DATA_PATH, file.Name()))
		}
	}
}
