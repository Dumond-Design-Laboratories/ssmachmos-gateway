package commands

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"strconv"
)

type sensor struct {
	Mac            [6]byte                      `json:"mac"`
	Name           string                       `json:"name"`
	Types          []string                     `json:"types"`
	WakeUpInterval int                          `json:"wake_up_interval"`
	BatteryLevel   int                          `json:"battery_level"`
	Settings       map[string]map[string]string `json:"settings"`
	PublicKey      rsa.PublicKey                `json:"key"`
}

func sensorJSONToString(jsonStr []byte) (string, error) {
	s := sensor{}
	err := json.Unmarshal(jsonStr, &s)
	if err != nil {
		return "", err
	}

	str := s.Name + " - " + macToString(s.Mac) + "\n"
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
		str += strconv.Itoa(s.BatteryLevel) + " mV\n"
	}
	str += "Settings:\n"
	for setting, value := range s.Settings {
		str += "\t" + setting + ":\n"
		for k, v := range value {
			str += "\t\t" + k + ": " + v + "\n"
		}
	}
	return str, nil
}

func macToString(mac [6]byte) string {
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}
