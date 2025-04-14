package api

import (
	"encoding/json"
	"errors"

	"github.com/jukuly/ss_machmos/server/internal/model"
	"github.com/jukuly/ss_machmos/server/internal/server"
)

func pairEnable() {
	server.EnablePairing()
}

func pairDisable() {
	server.DisablePairing()
}

func pairAccept(mac string) error {
	m, err := model.StringToMac(mac)
	if err != nil {
		return err
	}
	server.Pair(m)
	return nil
}

func pairListPending() (string, error) {
	jsonStr, err := json.Marshal(server.ListDevicesPendingPairing())
	return string(jsonStr), err
}

func list() (string, error) {
	jsonStr, err := json.Marshal(model.Sensors)
	return string(jsonStr), err
}

func listConnected() (string, error) {
	res, err := json.Marshal(server.ConnectedDevices())
	return string(res), err
}

func deviceCollect(address string) error {
	server.TriggerCollection(address)
	return nil
}

// List of all measurements pending upload
func pendingUploads() (string, error) {
	pending := server.PendingUploads()
	// anonymous struct yay
	res, err := json.Marshal(struct {
		Count   int                      `json:"count"`
		Pending []server.UnsentDataError `json:"pending"`
	}{
		Count:   len(pending),
		Pending: pending,
	})
	return string(res), err
}

func view(mac string) (string, error) {
	for _, sensor := range model.Sensors {
		if sensor.IsMacEqual(mac) {
			jsonStr, err := json.Marshal(sensor)
			return string(jsonStr), err
		}
	}
	return "", errors.New("Sensor with MAC address " + mac + " not found")
}

func forget(mac string) error {
	m, err := model.StringToMac(mac)
	if err != nil {
		return err
	}
	err = model.RemoveSensor(m)
	if err != nil {
		return err
	}
	return nil
}

func getGateway() (string, error) {
	jsonStr, err := json.Marshal(*server.Gateway)
	return string(jsonStr), err
}

func stop() {
	server.StopAdvertising()
}
