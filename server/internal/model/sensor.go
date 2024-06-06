package model

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const SENSORS_FILE = "sensors.json"

type Sensor struct {
	Mac                [6]byte                      `json:"mac"`
	Name               string                       `json:"name"`
	Types              []string                     `json:"types"`
	WakeUpInterval     int                          `json:"wake_up_interval"`
	BatteryLevel       int                          `json:"battery_level"`
	CollectionCapacity int                          `json:"collection_capacity"`
	Settings           map[string]map[string]string `json:"settings"`
	PublicKey          rsa.PublicKey                `json:"key"`
}

func (s *Sensor) ToString() string {
	str := s.Name + " - " + MacToString(s.Mac) + "\n"
	str += "Sensor Types: "
	for i, t := range s.Types {
		if i < len(s.Types)-1 {
			str += t + ", "
		} else {
			str += t + "\n"
		}
	}
	str += "Wake Up Interval: " + strconv.Itoa(s.WakeUpInterval) + " seconds\n"
	str += "Battery Level: "
	if s.BatteryLevel == -1 {
		str += "Unknown\n"
	} else {
		str += strconv.Itoa(s.BatteryLevel) + " %\n"
	}
	str += "Collection Capacity: " + strconv.Itoa(s.CollectionCapacity) + " bytes\n"
	str += "Settings:\n"
	for setting, value := range s.Settings {
		str += "\t" + setting + ":\n"
		for k, v := range value {
			str += "\t\t" + k + ": " + v + "\n"
		}
	}
	return str
}

func (s *Sensor) IsMacEqual(mac string) bool {
	m, err := StringToMac(mac)
	if err != nil {
		return false
	}
	return s.Mac == m
}

func StringToMac(mac string) ([6]byte, error) {
	var m [6]byte
	_, err := fmt.Sscanf(mac, "%02X:%02X:%02X:%02X:%02X:%02X", &m[5], &m[4], &m[3], &m[2], &m[1], &m[0])
	if err != nil {
		return [6]byte{}, err
	}
	return m, nil
}

func MacToString(mac [6]byte) string {
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", mac[5], mac[4], mac[3], mac[2], mac[1], mac[0])
}

func LoadSensors(path string, sensors *[]Sensor) error {
	jsonStr, err := os.ReadFile(path)
	if err != nil {
		*sensors = make([]Sensor, 0)
		return err
	}
	err = json.Unmarshal(jsonStr, sensors)
	if err != nil {
		*sensors = make([]Sensor, 0)
		return err
	}
	return nil
}

func RemoveSensor(mac [6]byte, sensors *[]Sensor) error {
	if sensors == nil {
		return errors.New("sensors is nil")
	}
	for i, s := range *sensors {
		if s.Mac == mac {
			*sensors = append((*sensors)[:i], (*sensors)[i+1:]...)
			err := saveSensors(SENSORS_FILE, sensors)
			return err
		}
	}
	return nil
}

func AddSensor(mac [6]byte, publicKey *rsa.PublicKey, sensors *[]Sensor) error {
	if sensors == nil {
		return errors.New("sensors is nil")
	}
	// Default settings
	sensor := Sensor{
		Mac:                mac,
		Name:               "Sensor " + MacToString(mac),
		Types:              []string{"vibration", "temperature", "acoustic"},
		WakeUpInterval:     3600,
		BatteryLevel:       -1,
		CollectionCapacity: 100000,
		Settings: map[string]map[string]string{
			"vibration": {
				"active":             "true",
				"sampling_frequency": "1000",
			},
			"temperature": {
				"active": "true",
			},
			"acoustic": {
				"active":             "true",
				"sampling_frequency": "44100",
			},
		},
		PublicKey: *publicKey,
	}

	*sensors = append(*sensors, sensor)
	err := saveSensors(SENSORS_FILE, sensors)
	return err
}

func UpdateSensorSetting(mac [6]byte, setting string, value string, sensors *[]Sensor) error {
	if sensors == nil {
		return errors.New("sensors is nil")
	}
	for _, s := range *sensors {
		if s.Mac != mac {
			continue
		}

		if setting == "name" {
			s.Name = value
			err := saveSensors(SENSORS_FILE, sensors)
			return err
		}

		settingParts := strings.Split(setting, "_")
		if len(settingParts) < 2 {
			return errors.New("invalid setting format")
		}

		dataType := settingParts[0]
		if dataType != "vibration" && dataType != "temperature" && dataType != "acoustic" {
			return errors.New("invalid setting data type")
		}
		setting = strings.Join(settingParts[1:], "_")

		switch setting {
		case "active":
			if value != "true" && value != "false" {
				return errors.New("invalid value for active setting")
			}
			s.Settings[dataType][setting] = value
		case "sampling_frequency":
			_, err := strconv.Atoi(value)
			if err != nil {
				return errors.New("invalid value for sampling_frequency setting")
			}
			s.Settings[dataType][setting] = value
		case "sampling_duration":
			return errors.New("unimplemented setting")
		case "wake_up_interval":
			intValue, err := strconv.Atoi(value)
			if err != nil {
				return errors.New("invalid value for wake_up_interval setting")
			}
			s.WakeUpInterval = intValue
		case "next_wake_up":
			return errors.New("unimplemented setting")
		}

		err := saveSensors(SENSORS_FILE, sensors)
		return err

	}
	return nil
}

func saveSensors(path string, sensors *[]Sensor) error {
	if sensors == nil {
		return errors.New("sensors is nil")
	}

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
