package server

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"math"
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

	if len(value) < 264 {
		out.Logger.Println("Invalid data format received")
		return
	}
	data := value[:len(value)-256]
	signature := value[len(value)-256:]

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

	if !model.VerifySignature(data, signature, &sensor.PublicKey) {
		out.Logger.Println("Invalid signature received from " + model.MacToString(macAddress))
		return
	}

	batteryLevel := int(int8(data[7]))
	timestamp := time.Now().UTC().Format(ISO8601)

	measurements := []map[string]interface{}{}

	if batteryLevel != -1 {
		out.Logger.Println("Received battery data from " + model.MacToString(macAddress) + " (" + sensor.Name + ")")
		sensor.BatteryLevel = batteryLevel
		measurements = []map[string]interface{}{
			{
				"sensor_id":          sensor.Mac,
				"time":               timestamp,
				"measurement_type":   "battery",
				"sampling_frequency": 0,
				"raw_data":           [1]int{batteryLevel},
			},
		}
	}
	if len(data) > 16 {
		measurementData := data[8:]
		var i uint32 = 0
		for i <= uint32(len(measurementData))-9 {
			dataType := DATA_TYPES[measurementData[i]]
			samplingFrequency := binary.LittleEndian.Uint32(measurementData[i+1 : i+5])
			lengthOfData := binary.LittleEndian.Uint32(measurementData[i+5 : i+9])
			if i+9+lengthOfData > uint32(len(measurementData)) || lengthOfData == 0 {
				break
			}
			rawData := measurementData[i+9 : i+9+lengthOfData]
			i += 9 + lengthOfData

			if dataType == "vibration" {
				out.Logger.Println("Received vibration data from " + model.MacToString(macAddress) + " (" + sensor.Name + ")")
				numberOfMeasurements := len(rawData) / 12 // 3 axes, 4 bytes per axis => 12 bytes per measurement
				x, y, z := make([]float32, numberOfMeasurements), make([]float32, numberOfMeasurements), make([]float32, numberOfMeasurements)
				for i := 0; i < numberOfMeasurements; i++ {
					x[i] = math.Float32frombits(binary.LittleEndian.Uint32(rawData[i*12 : 4+i*12]))
					y[i] = math.Float32frombits(binary.LittleEndian.Uint32(rawData[4+i*12 : 8+i*12]))
					z[i] = math.Float32frombits(binary.LittleEndian.Uint32(rawData[8+i*12 : 12+i*12]))
				}

				measurements = append(measurements,
					map[string]interface{}{
						"sensor_id":          model.MacToString(sensor.Mac),
						"time":               timestamp,
						"measurement_type":   dataType,
						"sampling_frequency": samplingFrequency,
						"axis":               "x",
						"raw_data":           x,
					},
					map[string]interface{}{
						"sensor_id":          model.MacToString(sensor.Mac),
						"time":               timestamp,
						"measurement_type":   dataType,
						"sampling_frequency": samplingFrequency,
						"axis":               "y",
						"raw_data":           y,
					},
					map[string]interface{}{
						"sensor_id":          model.MacToString(sensor.Mac),
						"time":               timestamp,
						"measurement_type":   dataType,
						"sampling_frequency": samplingFrequency,
						"axis":               "z",
						"raw_data":           z,
					},
				)
			} else {
				out.Logger.Println("Received " + dataType + " data from " + model.MacToString(macAddress) + " (" + sensor.Name + ")")
				measurements = append(measurements,
					map[string]interface{}{
						"sensor_id":          model.MacToString(sensor.Mac),
						"time":               timestamp,
						"measurement_type":   dataType,
						"sampling_frequency": samplingFrequency,
						"raw_data":           rawData, // TODO check the format of the raw data for acoustic and temperature
					},
				)
			}
		}
	}

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
