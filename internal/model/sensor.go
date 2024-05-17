package model

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

const SENSORS_FILE = "sensors.json"

var sensors []Sensor

type Sensor struct {
	Mac            [6]byte                      `json:"mac"`
	Name           string                       `json:"name"`
	Types          []string                     `json:"types"`
	WakeUpInterval int                          `json:"wake_up_interval"`
	BatteryLevel   int                          `json:"battery_level"`
	Settings       map[string]map[string]string `json:"settings"`
}

func (s *Sensor) ToString() string {
	str := s.Name + " - " + s.MacToString() + "\n"
	str += "Sensor Types: "
	for i, t := range s.Types {
		if i < len(s.Types)-1 {
			str += t + ", "
		} else {
			str += t + "\n"
		}
	}
	str += "Wake Up Interval: " + strconv.Itoa(s.WakeUpInterval) + " seconds\n"
	str += "Battery Level: " + strconv.Itoa(s.BatteryLevel) + " mV\n"
	str += "Settings:\n"
	for setting, value := range s.Settings {
		str += "\t" + setting + ":\n"
		for k, v := range value {
			str += "\t\t" + k + ": " + v + "\n"
		}
	}
	return str
}

func (s *Sensor) MacToString() string {
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", s.Mac[0], s.Mac[1], s.Mac[2], s.Mac[3], s.Mac[4], s.Mac[5])
}

func (s *Sensor) IsMacEqual(mac string) bool {
	var m [6]byte
	_, err := fmt.Sscanf(mac, "%02X:%02X:%02X:%02X:%02X:%02X", &m[0], &m[1], &m[2], &m[3], &m[4], &m[5])
	if err != nil {
		return false
	}
	return s.Mac == m
}

func GetSensors() ([]Sensor, error) {
	var err error
	if sensors == nil {
		err = loadSensors(SENSORS_FILE)
	}
	return sensors, err
}

func loadSensors(path string) error {
	jsonStr, err := os.ReadFile(path)
	if err != nil {
		sensors = make([]Sensor, 0)
		return err
	}
	err = json.Unmarshal(jsonStr, &sensors)
	if err != nil {
		sensors = make([]Sensor, 0)
		return err
	}
	return nil
}

func RemoveSensor(mac string) error {
	if sensors == nil {
		err := loadSensors(SENSORS_FILE)
		if err != nil {
			return err
		}
	}
	for i, s := range sensors {
		if s.IsMacEqual(mac) {
			sensors = append(sensors[:i], sensors[i+1:]...)
			err := saveSensors(SENSORS_FILE)
			return err
		}
	}
	return nil
}

func saveSensors(path string) error {
	if sensors == nil {
		err := loadSensors(SENSORS_FILE)
		if err != nil {
			return err
		}
	}

	jsonStr, err := json.Marshal(sensors)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, jsonStr, 0666)
	if err != nil {
		return err
	}
	err = loadSensors(SENSORS_FILE)
	return err
}
