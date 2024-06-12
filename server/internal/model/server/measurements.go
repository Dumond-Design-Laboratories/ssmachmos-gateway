package server

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/jukuly/ss_machmos/server/internal/model"
)

func sendMeasurements(jsonData []byte, gateway *model.Gateway) (*http.Response, error) {
	json := fmt.Sprintf("{\"gateway_id\":\"%s\",\"gateway_password\":\"%s\",\"measurements\":%s}", gateway.Id, gateway.Password, jsonData)
	return http.Post("https://openphm.org/gateway_data", "application/json", bytes.NewBuffer([]byte(json)))
}

func saveUnsentMeasurements(data []byte, timestamp int64) error {
	_, err := os.Stat(path.Join(os.TempDir(), "ss_machmos", UNSENT_DATA_PATH))
	if os.IsNotExist(err) {
		os.MkdirAll(path.Join(os.TempDir(), "ss_machmos", UNSENT_DATA_PATH), os.ModePerm)
	}

	err = os.MkdirAll(path.Join(os.TempDir(), "ss_machmos", UNSENT_DATA_PATH, fmt.Sprintf("%d.json", timestamp)), 0777)
	if err != nil {
		return err
	}

	return os.WriteFile(path.Join(os.TempDir(), "ss_machmos", UNSENT_DATA_PATH, fmt.Sprintf("%d.json", timestamp)), data, 0777)
}

func sendUnsentMeasurements() {
	files, err := os.ReadDir(path.Join(os.TempDir(), "ss_machmos", UNSENT_DATA_PATH))
	if err != nil {
		return
	}

	for _, file := range files {
		data, err := os.ReadFile(path.Join(os.TempDir(), "ss_machmos", UNSENT_DATA_PATH, file.Name()))
		if err != nil {
			continue
		}

		resp, err := sendMeasurements(data, Gateway)
		if err != nil {
			continue
		}

		if resp.StatusCode == 200 {
			os.Remove(UNSENT_DATA_PATH + file.Name())
		}
	}
}
