package server

import (
	"encoding/binary"
	"time"

	"github.com/jukuly/ss_machmos/server/internal/model"
	"github.com/jukuly/ss_machmos/server/internal/out"
)

/* 0x01 | mac address | Sleep until | repeat {dataTypeByte | active | Sampling Frequency | SamplingDuration}  */
func getSettingsForSensor(address string) []byte {
	mac, _ := model.StringToMac(address)
	var sensor *model.Sensor
	for i, s := range *model.Sensors {
		if s.Mac == mac {
			sensor = &(*model.Sensors)[i]
			break
		}
	}
	if sensor == nil {
		out.Logger.Println("Device", address, "not found in settings, reject")
		return []byte{0x00}
	}
	// Update last seen log
	sensor.UpdateLastSeen(model.SensorActivityIdle, model.Sensors)

	response := []byte{}
	if sensor.DeviceActive {
		response = append(response, 0x01)
	} else {
		response = append(response, 0x00)
	}
	response = append(response, mac[:]...)
	response = binary.LittleEndian.AppendUint32(response, setNextWakeUp(sensor))

	for dataType, settings := range sensor.Settings {
		var active byte
		if settings.Active {
			active = 0x01
		} else {
			active = 0x00
		}
		switch dataType {
		case "vibration":
			// 1 + 1 + 4 + 2 = 8
			response = append(response, 0x00, active)
			response = binary.LittleEndian.AppendUint32(response, settings.SamplingFrequency)
			response = binary.LittleEndian.AppendUint16(response, settings.SamplingDuration)
		case "audio":
			// 1 + 1 + 4 + 2 = 8
			response = append(response, 0x01, active)
			response = binary.LittleEndian.AppendUint32(response, settings.SamplingFrequency)
			response = binary.LittleEndian.AppendUint16(response, settings.SamplingDuration)
		case "temperature":
			// 							1	  2       3		4	  5	    6	  7     8
			response = append(response, 0x02, active, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00)
		default:
			out.Logger.Println("Unknown data type", dataType)
		}
	}

	return response
}

func setNextWakeUp(sensor *model.Sensor) uint32 {
	maxOffset := time.Duration(sensor.WakeUpIntervalMaxOffset) * time.Second

	wakeUpDurationThis := getWakeUpDuration(sensor)
	nextWakeUpCenter := time.Now().Add(time.Duration(sensor.WakeUpInterval) * time.Second)

	// going backward
	nextWakeUpLow := nextWakeUpCenter
	var offsetLow time.Duration
	i := 0
	for i < len(*model.Sensors) {
		s := (*model.Sensors)[i]
		i++
		wakeUpDurationOther := getWakeUpDuration(&s)

		j := 0
		for {
			difference := s.NextWakeUp.Add(time.Duration(j) * time.Duration(s.WakeUpInterval) * time.Second).Sub(nextWakeUpLow)
			j++

			if difference > wakeUpDurationThis {
				break
			}

			// if conflict
			if difference >= 0 && difference < wakeUpDurationThis || difference < 0 && -difference < wakeUpDurationOther {
				nextWakeUpLow = nextWakeUpLow.Add(difference - wakeUpDurationThis)
				i = 0
				break
			}
		}

		if nextWakeUpCenter.Sub(nextWakeUpLow) > maxOffset {
			offsetLow = time.Duration(-1)
			break
		}
	}
	offsetLow = nextWakeUpCenter.Sub(nextWakeUpLow)

	// going forward
	nextWakeUpHigh := nextWakeUpCenter
	var offsetHigh time.Duration
	i = 0
	for i < len(*model.Sensors) {
		s := (*model.Sensors)[i]
		i++
		wakeUpDurationOther := getWakeUpDuration(&s)

		j := 0
		for {
			difference := s.NextWakeUp.Add(time.Duration(j) * time.Duration(s.WakeUpInterval) * time.Second).Sub(nextWakeUpHigh)
			j++

			if difference > wakeUpDurationThis || nextWakeUpHigh.Sub(nextWakeUpCenter) > maxOffset {
				break
			}

			// if conflict
			if difference >= 0 && difference < wakeUpDurationThis || difference < 0 && -difference < wakeUpDurationOther {
				nextWakeUpHigh = nextWakeUpHigh.Add(difference + wakeUpDurationOther)
				i = 0
			}
		}

		if nextWakeUpHigh.Sub(nextWakeUpCenter) > maxOffset {
			offsetHigh = time.Duration(-1)
			break
		}
	}
	offsetHigh = nextWakeUpHigh.Sub(nextWakeUpCenter)

	if (offsetHigh >= offsetLow || offsetHigh == time.Duration(-1)) && offsetLow != time.Duration(-1) {
		sensor.NextWakeUp = nextWakeUpLow
		return uint32(sensor.WakeUpInterval)*1000 - uint32(offsetLow.Milliseconds())
	} else if offsetHigh != time.Duration(-1) {
		sensor.NextWakeUp = nextWakeUpHigh
		return uint32(sensor.WakeUpInterval)*1000 + uint32(offsetHigh.Milliseconds())
	} else {
		sensor.NextWakeUp = nextWakeUpCenter
		return uint32(sensor.WakeUpInterval) * 1000
	}
}

func getWakeUpDuration(sensor *model.Sensor) time.Duration {
	WAKE_UP_DURATION_BASELINE := time.Second * 30 // baseline to account for transmission time

	var maxSamplingDuration uint16 = 0
	for _, setting := range sensor.Settings {
		if setting.SamplingDuration > maxSamplingDuration {
			maxSamplingDuration = setting.SamplingDuration
		}
	}
	return time.Duration(maxSamplingDuration)*time.Second + WAKE_UP_DURATION_BASELINE
}
