package server

import (
	"encoding/binary"
	"time"

	"github.com/jukuly/ss_machmos/server/internal/model"
)

// see protocol.md to understand what is going on here
func sendSettings(value []byte) {
	if len(value) < 6 {
		return
	}
	mac := [6]byte(value[:6])

	var sensor *model.Sensor
	for _, s := range *Sensors {
		if s.Mac == mac {
			sensor = &s
			break
		}
	}
	if sensor == nil {
		return
	}

	response := mac[:]
	response = binary.LittleEndian.AppendUint32(response, uint32(sensor.WakeUpInterval)*1000)
	sensor.NextWakeUp = sensor.NextWakeUp.Add(time.Duration(sensor.WakeUpInterval) * time.Second)

	for dataType, settings := range sensor.Settings {
		var active byte
		if settings.Active {
			active = 0x01
		} else {
			active = 0x00
		}
		switch dataType {
		case "vibration":
			response = append(response, 0x00, active)
			response = binary.LittleEndian.AppendUint32(response, settings.SamplingFrequency)
			response = binary.LittleEndian.AppendUint16(response, settings.SamplingDuration)
		case "audio":
			response = append(response, 0x01, active)
			response = binary.LittleEndian.AppendUint32(response, settings.SamplingFrequency)
			response = binary.LittleEndian.AppendUint16(response, settings.SamplingDuration)
		case "temperature":
			response = append(response, 0x02, active, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00)
		}
	}

	settingsCharacteristic.Write(response)
}
