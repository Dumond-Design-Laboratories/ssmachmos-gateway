package server

import "strconv"

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
			if settings["active"] == "true" {
				active = 0x01
			} else {
				active = 0x00
			}
			switch dataType {
			case "vibration":
				sampling_frequency, err := strconv.ParseInt(settings["sampling_frequency"], 10, 32)
				if err != nil {
					return
				}
				response = append(response, 0x00, active, byte(sampling_frequency>>24), byte(sampling_frequency>>16), byte(sampling_frequency>>8), byte(sampling_frequency))
			case "audio":
				sampling_frequency, err := strconv.ParseInt(settings["sampling_frequency"], 10, 32)
				if err != nil {
					return
				}
				response = append(response, 0x01, active, byte(sampling_frequency>>24), byte(sampling_frequency>>16), byte(sampling_frequency>>8), byte(sampling_frequency))
			case "temperature":
				response = append(response, 0x02, active, 0x00, 0x00, 0x00, 0x00)
			}
		}

		settingsCharacteristic.Write(response)

		return
	}
}
