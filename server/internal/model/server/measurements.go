package server

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jukuly/ss_mach_mo/server/internal/model"
)

func sendMeasurements(jsonData []byte, gateway *model.Gateway) (*http.Response, error) {
	json := fmt.Sprintf("{\"gateway_id\":\"%s\",\"gateway_password\":\"%s\",\"measurements\":%s}", gateway.Id, gateway.Password, jsonData)
	return http.Post("https://openphm.org/gateway_data", "application/json", bytes.NewBuffer([]byte(json)))
}

func saveUnsentMeasurements(data []byte, timestamp int64) {
	_, err := os.Stat(UNSENT_DATA_PATH)
	if os.IsNotExist(err) {
		os.MkdirAll(UNSENT_DATA_PATH, os.ModePerm)
	}

	path, _ := filepath.Abs(fmt.Sprintf("%s%d.json", UNSENT_DATA_PATH, timestamp))

	os.WriteFile(path, data, 0644)
}

func sendUnsentMeasurements() {
	files, err := os.ReadDir(UNSENT_DATA_PATH)
	if err != nil {
		return
	}

	for _, file := range files {
		data, err := os.ReadFile(UNSENT_DATA_PATH + file.Name())
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
