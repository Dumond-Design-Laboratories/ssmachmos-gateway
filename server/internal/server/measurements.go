package server

/*
 * Management of measurements on disk
 */

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/jukuly/ss_machmos/server/internal/model"
	"github.com/jukuly/ss_machmos/server/internal/out"
)

// Used to notify of any data that wasn't uploaded
// serialized and sent over pipe
type UnsentDataError struct {
	//Reason              string    `json:"reason"`
	LastAttemptedUpload time.Time `json:"last_attempted_upload"`
}
var unsentData []UnsentDataError = []UnsentDataError{};

func unsentDataDir() string {
	dir := path.Join(os.TempDir(), "/ss_machmos/", "/unsent_data/")
	os.MkdirAll(dir, 0777)
	return dir
}

func archivedDataDir() string {
	dir := path.Join(os.TempDir(), "/ss_machmos/", "/sent_data/")
	os.MkdirAll(dir, 0777)
	return dir
}

// FIXME: this should be the only entry point to gateway uploading. Server
// should NOT handle saving unsent measurements
func sendMeasurements(jsonData []byte, gateway *model.Gateway) (*http.Response, error) {
	//gateway.AuthError = false
	body := model.RequestBody{
		GatewayId:       gateway.Id,
		GatewayPassword: gateway.Password,
	}
	err := json.Unmarshal(jsonData, &body.Measurements)
	if err != nil {
		return nil, err
	}
	json, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(gateway.HTTPEndpoint, "application/json", bytes.NewBuffer([]byte(json)))
	if err == nil && resp.StatusCode == http.StatusOK {
		// if success, archive
		err := archiveMeasurements(jsonData, time.Now())
		if err != nil {
			out.Logger.Println(err.Error())
		}
		out.Broadcast("UPLOAD-SUCCESS")
	}

	if resp.StatusCode == http.StatusUnauthorized {
		// Unauthorized
		out.Broadcast("GATEWAY-INVALID")
	}

	return resp, err
}

// Sending failed, save to disk for later
func saveUnsentMeasurements(data []byte, timestamp time.Time) error {
	err := os.MkdirAll(unsentDataDir(), 0777)
	if err != nil {
		return err
	}

	unsentData = append(unsentData, UnsentDataError{
		//Reason: "",
		LastAttemptedUpload: timestamp,
	})

	// Notify GUI of a new unsent measurement
	out.Broadcast("UPLOAD-FAILED")
	return os.WriteFile(path.Join(unsentDataDir(), timestamp.String()+".json"), data, 0777)
}

func archiveMeasurements(data []byte, timestamp time.Time) error {
	return os.WriteFile(path.Join(archivedDataDir(), timestamp.String()+".json"), data, 0777)
}

func PendingUploads() []UnsentDataError {
	return unsentData
}

func sendUnsentMeasurements() {
	files, err := os.ReadDir(unsentDataDir())
	if err != nil {
		out.Logger.Println("Error:", err)
		return
	}

	for _, file := range files {
		data, err := os.ReadFile(path.Join(unsentDataDir(), file.Name()))
		if err != nil {
			out.Logger.Println("Error:", err)
			continue
		}

		resp, err := sendMeasurements(data, Gateway)
		if err != nil {
			out.Logger.Println("Error:", err)
			continue
		}

		if resp.StatusCode == 200 {
			// Don't delete, keep around for debugging
			os.Rename(path.Join(unsentDataDir(), file.Name()), path.Join(archivedDataDir(), file.Name()))
		}
	}
}

func parseTemperatureData(data uint16) (float64, error) {
	adc_fs := math.Pow(2, 15) - 1.0
	const r_ref = 1500.0
	const r_0 = 1000.0

	adc_in := float64(data)
	rtd_resistance := adc_in / adc_fs * r_ref

	if rtd_resistance >= 1000 {
		const A = 3.9083e-3
		const B = -5.775e-7

		// Callendar-Van Dusen equation
		sqrt := math.Sqrt(math.Pow(A, 2) - 4*B*(1-rtd_resistance/r_0))
		if sqrt < 0 {
			return 0, errors.New("negative square root")
		}
		return (-A + sqrt) / (2 * B), nil
	} else {
		// Callendar-Van Dusen equation approximation with quadratic equation
		const A = -0.00061414
		const B = 3.907359803
		const C = 999.9979

		sqrt := math.Sqrt(math.Pow(B, 2) - 4*A*(C-rtd_resistance))
		if sqrt < 0 {
			return 0, errors.New("negative square root")
		}
		return (-B + sqrt) / (2 * A), nil
	}
}
