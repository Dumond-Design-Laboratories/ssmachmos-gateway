package model

import (
	"encoding/json"
	"os"
	"strconv"
)

const SENSORS_FILE = "sensors.json"

var sensors []Sensor

type Sensor struct {
	Mac            string                       `json:"mac"`
	Name           string                       `json:"name"`
	Types          []string                     `json:"types"`
	WakeUpInterval int                          `json:"wake_up_interval"`
	BatteryLevel   int                          `json:"battery_level"`
	Settings       map[string]map[string]string `json:"settings"`
}

func (s *Sensor) ToString() string {
	str := s.Name + " - " + s.Mac + "\n"
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
		if s.Mac == mac {
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
