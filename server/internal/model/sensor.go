package model

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const SENSORS_FILE = "sensors.json"

type settings struct {
	Active            bool      `json:"active"`
	SamplingFrequency uint32    `json:"sampling_frequency"`
	SamplingDuration  uint16    `json:"sampling_duration"`
	WakeUpInterval    int       `json:"wake_up_interval"`
	NextWakeUp        time.Time `json:"next_wake_up"`
}

type Sensor struct {
	Mac                [6]byte             `json:"mac"`
	Name               string              `json:"name"`
	Types              []string            `json:"types"`
	BatteryLevel       int                 `json:"battery_level"`
	CollectionCapacity int                 `json:"collection_capacity"`
	Settings           map[string]settings `json:"settings"`
	PublicKey          rsa.PublicKey       `json:"key"`
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
		str += "\t\tActive: " + strconv.FormatBool(value.Active) + "\n"
		str += "Wake Up Interval: " + strconv.Itoa(value.WakeUpInterval) + " seconds\n"
		str += "Next Wake Up: " + value.NextWakeUp.Local().Format(time.RFC822) + "\n"
		if setting == "temperature" {
			continue
		}
		str += "\t\tSampling Frequency: " + strconv.Itoa(int(value.SamplingFrequency)) + "\n"
		str += "\t\tSampling Duration: " + strconv.Itoa(int(value.SamplingDuration)) + "\n"
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
	_, err := fmt.Sscanf(mac, "%02X:%02X:%02X:%02X:%02X:%02X", &m[0], &m[1], &m[2], &m[3], &m[4], &m[5])
	if err != nil {
		return [6]byte{}, err
	}
	return m, nil
}

func MacToString(mac [6]byte) string {
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
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

func AddSensor(mac [6]byte, types []string, collectionCapacity int, publicKey *rsa.PublicKey, sensors *[]Sensor) error {
	if sensors == nil {
		return errors.New("sensors is nil")
	}
	// Default settings
	sensor := Sensor{
		Mac:                mac,
		Name:               "Sensor " + MacToString(mac),
		Types:              types,
		BatteryLevel:       -1,
		CollectionCapacity: collectionCapacity,
		Settings:           map[string]settings{},
		PublicKey:          *publicKey,
	}

	for _, t := range types {
		switch t {
		case "vibration":
			sensor.Settings["vibration"] = settings{
				Active:            true,
				SamplingFrequency: 100,
				SamplingDuration:  1,
				WakeUpInterval:    3600,
				NextWakeUp:        time.Now().Add(3600 * time.Second),
			}
		case "temperature":
			sensor.Settings["temperature"] = settings{
				Active:         true,
				WakeUpInterval: 3600,
				NextWakeUp:     time.Now().Add(3600 * time.Second),
			}
		case "acoustic":
			sensor.Settings["acoustic"] = settings{
				Active:            true,
				SamplingFrequency: 8000,
				SamplingDuration:  1,
				WakeUpInterval:    3600,
				NextWakeUp:        time.Now().Add(3600 * time.Second),
			}
		}
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
			setting := s.Settings[dataType]
			setting.Active = value == "true"
			s.Settings[dataType] = setting
		case "sampling_frequency":
			intValue, err := strconv.Atoi(value)
			if err != nil {
				return errors.New("invalid value for sampling_frequency setting (must be an integer (Hz))")
			}
			setting := s.Settings[dataType]
			setting.SamplingFrequency = uint32(intValue)
			s.Settings[dataType] = setting
		case "sampling_duration":
			intValue, err := strconv.Atoi(value)
			if err != nil {
				return errors.New("invalid value for sampling_duration setting (must be an integer (seconds))")
			}
			setting := s.Settings[dataType]
			setting.SamplingDuration = uint16(intValue)
			s.Settings[dataType] = setting
		case "wake_up_interval":
			intValue, err := strconv.Atoi(value)
			if err != nil {
				return errors.New("invalid value for wake_up_interval setting (must be an integer (seconds))")
			}
			setting := s.Settings[dataType]
			setting.WakeUpInterval = intValue
			s.Settings[dataType] = setting
		case "next_wake_up":
			timeValue, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return errors.New("invalid value for next_wake_up setting (must be a date in RFC3339 format)")
			}
			setting := s.Settings[dataType]
			setting.NextWakeUp = timeValue
			s.Settings[dataType] = setting
		default:
			return errors.New("setting " + setting + " doesn't exist")
		}

		err := saveSensors(SENSORS_FILE, sensors)
		return err

	}
	return errors.New("sensor not found")
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
