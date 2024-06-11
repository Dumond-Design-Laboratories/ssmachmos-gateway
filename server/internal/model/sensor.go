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
	NextWakeUp        time.Time `json:"next_wake_up"` // next wake up is the next time the server will transmit to the sensor to wake up (the next NEXT wake up)
}

type Sensor struct {
	Mac                [6]byte             `json:"mac"`
	Name               string              `json:"name"`
	Types              []string            `json:"types"`
	BatteryLevel       int                 `json:"battery_level"`
	CollectionCapacity uint32              `json:"collection_capacity"`
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
	str += "Collection Capacity: " + strconv.Itoa(int(s.CollectionCapacity)) + " bytes\n"
	str += "Settings:\n"
	for setting, value := range s.Settings {
		str += "\t" + setting + ":\n"
		str += "\t\tActive: " + strconv.FormatBool(value.Active) + "\n"
		str += "\t\tWake Up Interval: " + strconv.Itoa(value.WakeUpInterval) + " seconds\n"
		str += "\t\tNext Wake Up: " + value.NextWakeUp.Add(-1*time.Second*time.Duration(value.WakeUpInterval)).Local().Format(time.RFC3339) + "\n"
		if setting == "temperature" {
			continue
		}
		str += "\t\tSampling Frequency: " + strconv.Itoa(int(value.SamplingFrequency)) + " Hz\n"
		str += "\t\tSampling Duration: " + strconv.Itoa(int(value.SamplingDuration)) + " seconds\n"
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

func AddSensor(mac [6]byte, types []string, collectionCapacity uint32, publicKey *rsa.PublicKey, sensors *[]Sensor) error {
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

	var sensor *Sensor
	for i, s := range *sensors {
		if s.Mac == mac {
			sensor = &(*sensors)[i]
			break
		}
	}
	if sensor == nil {
		return errors.New("sensor not found")
	}

	if setting == "name" {
		sensor.Name = value
		err := saveSensors(SENSORS_FILE, sensors)
		return err
	}

	settingParts := strings.Split(setting, "_")
	if len(settingParts) < 2 {
		fmt.Println(settingParts)
		return errors.New("invalid setting format")
	}

	dataType := settingParts[0]
	if dataType != "vibration" && dataType != "temperature" && dataType != "acoustic" {
		return errors.New("invalid setting data type")
	}
	setting = strings.Join(settingParts[1:], "_")

	switch setting {
	case "active":
		setting := sensor.Settings[dataType]
		setting.Active = value == "true"
		sensor.Settings[dataType] = setting
	case "sampling_frequency":
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return errors.New("invalid value for sampling_frequency setting (must an integer (Hz))")
		}
		// not a uint32
		if intValue < 0 || intValue > 4294967295 {
			return errors.New("invalid value for sampling_frequency setting (must an integer between 0 and 4 294 967 295)")
		}
		setting := sensor.Settings[dataType]
		err = isExceedingCollectionCapacity(sensor, "sampling_frequency", intValue, dataType, setting)
		if err != nil {
			return err
		}
		setting.SamplingFrequency = uint32(intValue)
		sensor.Settings[dataType] = setting
	case "sampling_duration":
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return errors.New("invalid value for sampling_duration setting (must be an integer (seconds))")
		}
		// not a uint16
		if intValue < 0 || intValue > 65535 {
			return errors.New("invalid value for sampling_duration setting (must an integer between 0 and 65 535)")
		}
		setting := sensor.Settings[dataType]
		err = isExceedingCollectionCapacity(sensor, "sampling_duration", intValue, dataType, setting)
		if err != nil {
			return err
		}
		setting.SamplingDuration = uint16(intValue)
		sensor.Settings[dataType] = setting
	case "wake_up_interval":
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return errors.New("invalid value for wake_up_interval setting (must be an integer (seconds))")
		}
		// when converted to milliseconds it will not be a uint32
		if intValue < 0 || intValue > 4294967 {
			return errors.New("invalid value for wake_up_interval setting (must an integer between 0 and 4 294 967)")
		}
		setting := sensor.Settings[dataType]
		setting.WakeUpInterval = intValue
		sensor.Settings[dataType] = setting
	case "next_wake_up":
		timeValue, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return errors.New("invalid value for next_wake_up setting (must be a date in RFC3339 format)")
		}
		setting := sensor.Settings[dataType]
		setting.NextWakeUp = timeValue.Add(time.Duration(setting.WakeUpInterval) * time.Second)
		sensor.Settings[dataType] = setting
	default:
		return errors.New("setting " + setting + " doesn't exist")
	}

	err := saveSensors(SENSORS_FILE, sensors)
	return err
}

func getCollectionSize(sensor *Sensor) int {
	result := 0
	for dataType, settings := range sensor.Settings {
		if dataType == "temperature" {
			result += 4 // TODO check the format of the raw data for acoustic and temperature (assuming a 4 bytes number)
			continue
		}
		term := int(settings.SamplingFrequency) * int(settings.SamplingDuration)
		if dataType == "vibration" {
			term *= 12 // 3 axes of 4 bytes each
		} else {
			term *= 4 // TODO check the format of the raw data for acoustic and temperature (assuming a 4 bytes number)
		}
		result += term
	}
	return result
}

func isExceedingCollectionCapacity(sensor *Sensor, setting string, value int, dataType string, settings settings) error {
	currentCollectionSize := getCollectionSize(sensor)
	if dataType == "temperature" {
		return nil
	}
	if settings.SamplingDuration == 0 || settings.SamplingFrequency == 0 {
		return errors.New("sampling_duration and sampling_frequency must be greater than 0")
	}

	sizeOfData := 1
	if dataType == "vibration" {
		sizeOfData = 12 // 3 axes of 4 bytes each
	} else {
		sizeOfData = 4 // TODO check the format of the raw data for acoustic and temperature (assuming a 4 bytes number for now)
	}
	otherFactor := 1
	if setting == "sampling_frequency" {
		otherFactor = int(settings.SamplingDuration)
	} else {
		otherFactor = int(settings.SamplingFrequency)
	}

	currentCollectionSize -= sizeOfData * otherFactor
	max := (int(sensor.CollectionCapacity) - currentCollectionSize) / (sizeOfData * otherFactor)
	if value > max {
		return errors.New("invalid value for " + setting + " setting (exceeds collection capacity of sensor (current maximum: " + strconv.Itoa(max) + "))")
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
