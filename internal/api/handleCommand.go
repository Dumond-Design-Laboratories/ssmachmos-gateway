package api

import (
	"encoding/json"
	"errors"

	"github.com/jukuly/ss_mach_mo/internal/model"
	"github.com/jukuly/ss_mach_mo/internal/model/server"
)

func pairEnable() error {
	server.EnablePairing()
	return nil
}

func pairDisable() error {
	server.DisablePairing()
	return nil
}

func pairAccept(mac string) error {
	m, err := model.StringToMac(mac)
	if err != nil {
		return err
	}
	server.Pair(m)
	return nil
}

func list() (string, error) {
	jsonStr, err := json.Marshal(*server.Sensors)
	return string(jsonStr), err
}

func view(mac string) (string, error) {
	for _, sensor := range *server.Sensors {
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
	err = model.RemoveSensor(m, server.Sensors)
	if err != nil {
		return err
	}
	return nil
}
