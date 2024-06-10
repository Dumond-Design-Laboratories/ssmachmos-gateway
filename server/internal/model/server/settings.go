package server

import (
	"encoding/binary"
	"time"
)

func sendSettings(value []byte) {
	if len(value) < 6 {
		return
	}
	mac := [6]byte(value[:6])

	for _, sensor := range *Sensors {
		if sensor.Mac != mac {
			continue
		}

		response := mac[:]
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
				response = binary.LittleEndian.AppendUint32(response, uint32(settings.SamplingFrequency))
				response = binary.LittleEndian.AppendUint16(response, uint16(settings.SamplingDuration))
			case "audio":
				response = append(response, 0x01, active)
				response = binary.LittleEndian.AppendUint32(response, uint32(settings.SamplingFrequency))
				response = binary.LittleEndian.AppendUint16(response, uint16(settings.SamplingDuration))
			case "temperature":
				response = append(response, 0x02, active, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00)
			}
			timeUntilNextWakeUp := settings.NextWakeUp.UnixMilli() - time.Now().UnixMilli()
			response = binary.LittleEndian.AppendUint32(response, uint32(timeUntilNextWakeUp))
			settings.NextWakeUp = time.Now().Add(time.Duration(settings.WakeUpInterval) * time.Second)
		}

		settingsCharacteristic.Write(response)

		return
	}
}
