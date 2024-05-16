package model

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
)

const SENSORS_FILE = "sensors.json"

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

func LoadSensors(path string) ([]Sensor, error) {
	jsonStr, err := os.ReadFile(path)
	if err != nil {
		return make([]Sensor, 0), err
	}
	var sensors []Sensor
	err = json.Unmarshal(jsonStr, &sensors)
	if err != nil {
		return make([]Sensor, 0), err
	}
	return sensors, nil
}

func RemoveSensor(sensors []Sensor, mac string) error {
	for i, s := range sensors {
		if s.Mac == mac {
			sensors = append(sensors[:i], sensors[i+1:]...)
			saveSensors(SENSORS_FILE, sensors)
			return nil
		}
	}
	return errors.New("Sensor with MAC address: " + mac + " not found")
}

func saveSensors(path string, sensors []Sensor) error {
	jsonStr, err := json.Marshal(sensors)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, jsonStr, 0666)
	if err != nil {
		return err
	}
	return nil
}
