package server

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"time"

	"github.com/jukuly/ss_machmos/server/internal/model"
	"github.com/jukuly/ss_machmos/server/internal/out"
	"tinygo.org/x/bluetooth"
)

const ISO8601 = "2006-01-02T15:04:05.999Z"
const DEFAULT_GATEWAY_HTTP_ENDPOINT = "https://openphm.org/gateway_data"

var adapter = bluetooth.DefaultAdapter

var SERVICE_UUID = [4]uint32{0xA07498CA, 0xAD5B474E, 0x940D16F1, 0xFBE7E8CD}                      // same for every gateway fbe7e8cd-940d-16f1-ad5b-474ea07498ca
var PAIR_REQUEST_CHARACTERISTIC_UUID = [4]uint32{0x37ecbcb9, 0xe2514c40, 0xa1613de1, 0x1ea8c363}  // same for every gateway 1ea8c363-a161-3de1-e251-4c4037ecbcb9
var PAIR_RESPONSE_CHARACTERISTIC_UUID = [4]uint32{0x0598acc3, 0x8564405a, 0xaf67823f, 0x029c79b6} // same for every gateway 029c79b6-af67-823f-8564-405a0598acc3

var DATA_TYPES = map[byte]string{
	0x00: "vibration",
	0x01: "audio",
	0x02: "temperature",
	0x03: "battery",
}

const UNSENT_DATA_PATH = "unsent_data/"

var pairResponseCharacteristic bluetooth.Characteristic
var settingsCharacteristic bluetooth.Characteristic
var Gateway *model.Gateway
var Sensors *[]model.Sensor

func Init(ss *[]model.Sensor, g *model.Gateway) error {
	Gateway = g
	Sensors = ss

	err := adapter.Enable()
	if err != nil {
		return err
	}

	state = pairingState{
		active:    false,
		requested: make(map[[6]byte]request),
		pairing:   [6]byte{},
	}
	transmissions = make(map[[3]byte]Transmission)
	dataCharUUID, err := model.GetDataCharUUID(Gateway)
	if err != nil {
		return err
	}
	settingsCharUUID, err := model.GetSettingsCharUUID(Gateway)
	if err != nil {
		return err
	}

	service := bluetooth.Service{
		UUID: SERVICE_UUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				UUID:  dataCharUUID,
				Flags: bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					if len(value) > 0 && value[0] == 0x00 {
						handleData(client, offset, value)
					}
				},
			},
			{
				Handle: &settingsCharacteristic,
				UUID:   settingsCharUUID,
				Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					if len(value) > 0 && value[0] == 0x00 {
						sendSettings(value)
					}
				},
			},
			{
				UUID:  PAIR_REQUEST_CHARACTERISTIC_UUID,
				Flags: bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					if len(value) > 0 && value[0] == 0x00 {
						pairRequest(value)
					}
				},
			},
			{
				Handle: &pairResponseCharacteristic,
				UUID:   PAIR_RESPONSE_CHARACTERISTIC_UUID,
				Value:  []byte{},
				Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					if len(value) > 0 && value[0] == 0x00 {
						pairConfirmation(value)
					}
				},
			},
		},
	}
	err = adapter.AddService(&service)
	if err != nil {
		return err
	}

	adv := adapter.DefaultAdvertisement()
	if adv == nil {
		return errors.New("advertisement is nil")
	}
	err = adv.Configure(bluetooth.AdvertisementOptions{
		LocalName:    "Gateway Server",
		ServiceUUIDs: []bluetooth.UUID{service.UUID},
	})
	if err != nil {
		return err
	}
	return nil
}

func StartAdvertising() error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			if sig == os.Interrupt {
				err := adapter.DefaultAdvertisement().Stop()
				if err != nil {
					out.Logger.Println("Error:", err)
				}
				out.Logger.Println("Stopping server")
				os.Exit(0)
				return
			}
		}
	}()
	adv := adapter.DefaultAdvertisement()
	if adv == nil {
		return errors.New("advertisement is nil")
	}
	return adv.Start()
}

func StopAdvertising() {
	adapter.DefaultAdvertisement().Stop()
	out.Logger.Println("Stopping server")
	os.Exit(0)
}

// see protocol.md to understand what is going on here
func handleData(_ bluetooth.Connection, _ int, value []byte) {

	// if len(value) < 264 {
	if len(value) < 24 {
		out.Logger.Println("Invalid data format received")
		return
	}
	data := value // [:len(value)-256]
	// signature := value[len(value)-256:]

	macAddress := [6]byte(data[1:7])
	var sensor *model.Sensor
	for i, s := range *Sensors {
		if s.Mac == macAddress {
			sensor = &(*Sensors)[i]
			break
		}
	}
	if sensor == nil {
		out.Logger.Println("Device " + model.MacToString(macAddress) + " tried to send data, but it is not paired with this gateway")
		return
	}

	// if !model.VerifySignature(data, signature, &sensor.PublicKey) {
	// out.Logger.Println("Invalid signature received from " + model.MacToString(macAddress))
	// return
	// }

	batteryLevel := int(int8(data[7]))
	dataType := DATA_TYPES[data[8]]
	samplingFrequency := binary.LittleEndian.Uint32(data[9:13])
	lengthOfData := binary.LittleEndian.Uint32(data[13:17])
	messageID := *(*[3]byte)(data[17:20])
	offset := int(binary.LittleEndian.Uint32(data[20:24]))

	// if len(data) > 16 {
	// measurementData := data[8:]
	// var i uint32 = 0
	// for i <= uint32(len(measurementData))-9 {
	// dataType := DATA_TYPES[measurementData[i]]
	// samplingFrequency := binary.LittleEndian.Uint32(measurementData[i+1 : i+5])
	// lengthOfData := binary.LittleEndian.Uint32(measurementData[i+5 : i+9])
	// if i+9+lengthOfData > uint32(len(measurementData)) || lengthOfData == 0 {
	// break
	// }
	// rawData := measurementData[i+9 : i+9+lengthOfData]
	// i += 9 + lengthOfData

	rawData := savePacket(data[24:], macAddress, batteryLevel, dataType, samplingFrequency, lengthOfData, messageID, offset)
	if rawData == nil {
		return
	}

	timestamp := time.Now().UTC().Format(ISO8601)
	measurements := []map[string]interface{}{}
	if batteryLevel != -1 {
		out.Logger.Println("Received battery data from " + model.MacToString(macAddress) + " (" + sensor.Name + ")")
		sensor.BatteryLevel = batteryLevel
		measurements = []map[string]interface{}{
			{
				"sensor_id":          model.MacToString(macAddress),
				"time":               timestamp,
				"measurement_type":   "battery",
				"sampling_frequency": 0,
				"raw_data":           [1]int{batteryLevel},
			},
		}
	}

	out.Logger.Println("Received " + dataType + " data from " + model.MacToString(macAddress) + " (" + sensor.Name + ")")
	if dataType == "vibration" {
		numberOfMeasurements := len(rawData) / 6 // 3 axes, 2 bytes per axis => 6 bytes per measurement
		x, y, z := make([]int16, numberOfMeasurements), make([]int16, numberOfMeasurements), make([]int16, numberOfMeasurements)
		for i := 0; i < numberOfMeasurements; i++ {
			x[i] = int16(rawData[i*6]) | int16(i*6+1)<<8
			y[i] = int16(rawData[i*6+2]) | int16(i*6+3)<<8
			z[i] = int16(rawData[i*6+4]) | int16(i*6+5)<<8
		}

		measurements = append(measurements,
			map[string]interface{}{
				"sensor_id":          model.MacToString(macAddress),
				"time":               timestamp,
				"measurement_type":   dataType,
				"sampling_frequency": samplingFrequency,
				"axis":               "x",
				"raw_data":           x,
			},
			map[string]interface{}{
				"sensor_id":          model.MacToString(macAddress),
				"time":               timestamp,
				"measurement_type":   dataType,
				"sampling_frequency": samplingFrequency,
				"axis":               "y",
				"raw_data":           y,
			},
			map[string]interface{}{
				"sensor_id":          model.MacToString(macAddress),
				"time":               timestamp,
				"measurement_type":   dataType,
				"sampling_frequency": samplingFrequency,
				"axis":               "z",
				"raw_data":           z,
			},
		)
	} else if dataType == "temperature" {
		// if len(rawData) != 2 {
		// 	out.Logger.Println("Invalid temperature data received")
		//	continue
		//}
		if len(rawData) == 2 {
			temperature, err := parseTemperatureData(binary.LittleEndian.Uint16(rawData))
			if err == nil {
				measurements = append(measurements,
					map[string]interface{}{
						"sensor_id":          model.MacToString(macAddress),
						"time":               timestamp,
						"measurement_type":   dataType,
						"sampling_frequency": samplingFrequency,
						"raw_data":           temperature,
					},
				)
			} else {
				out.Logger.Println("Error:", err)
			}
		} else {
			out.Logger.Println("Invalid temperature data received")
		}

		//if err != nil {
		//	out.Logger.Println("Error:", err)
		//	continue
		//}

	} else if dataType == "audio" {
		if len(rawData)%3 == 0 {
			numberOfMeasurements := len(rawData) / 3
			amplitude := make([]int, numberOfMeasurements)
			for i := 0; i < numberOfMeasurements; i++ {
				amplitude[i] = int(rawData[i*3]) | int(rawData[i*3+1])<<8 | int(rawData[i*3+2])<<16
			}

			measurements = append(measurements,
				map[string]interface{}{
					"sensor_id":          model.MacToString(macAddress),
					"time":               timestamp,
					"measurement_type":   dataType,
					"sampling_frequency": samplingFrequency,
					"raw_data":           amplitude,
				},
			)
		} else {
			out.Logger.Println("Invalid audio data received")
		}
	}

	// }
	// }

	jsonData, err := json.Marshal(measurements)
	if err != nil {
		out.Logger.Println("Error:", err)
		return
	}

	resp, err := sendMeasurements(jsonData, Gateway)

	if err != nil {
		out.Logger.Println("Error:", err)
		if err := saveUnsentMeasurements(jsonData, timestamp); err != nil {
			out.Logger.Println("Error:", err)
		}
		return
	}

	if resp.StatusCode != 200 {
		out.Logger.Println("Error sending data to server")
		body := make([]byte, resp.ContentLength)
		defer resp.Body.Close()
		resp.Body.Read(body)
		out.Logger.Println(string(body))
		if err := saveUnsentMeasurements(jsonData, timestamp); err != nil {
			out.Logger.Println(err)
		}
		return
	}

	sendUnsentMeasurements()
}
