package server

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
				response = append(response, 0x00, active, byte(settings.SamplingFrequency>>24), byte(settings.SamplingFrequency>>16), byte(settings.SamplingFrequency>>8), byte(settings.SamplingFrequency))
			case "audio":
				response = append(response, 0x01, active, byte(settings.SamplingFrequency>>24), byte(settings.SamplingFrequency>>16), byte(settings.SamplingFrequency>>8), byte(settings.SamplingFrequency))
			case "temperature":
				response = append(response, 0x02, active, 0x00, 0x00, 0x00, 0x00)
			}
		}

		settingsCharacteristic.Write(response)

		return
	}
}
