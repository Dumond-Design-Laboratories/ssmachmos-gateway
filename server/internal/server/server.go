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

var DATA_SERVICE_UUID = bluetooth.MustParseUUID("2deacc71-7b29-4ff4-8fc2-59461c7a73f5")
var FLUX_DATA_CHRC_UUID = bluetooth.MustParseUUID("68e92ad3-0fb5-4c93-8b99-0d21771576fd")
var RTD_DATA_CHRC_UUID = bluetooth.MustParseUUID("e64d1230-86ba-46aa-a62d-736d6f58226c")
var ACCEL_DATA_CHRC_UUID = bluetooth.MustParseUUID("e70ada20-ac8e-45f8-9f5d-593226bb7284")
var MIC_DATA_CHRC_UUID = bluetooth.MustParseUUID("fee1ed78-2a76-490e-8a7c-9b698c9202d1")

var CONFIG_SERVICE_UUID, _ = bluetooth.ParseUUID("0ffd06bd-5f9c-4583-b852-e92fdbe8e862")
var CONFIG_IDENTIFY_CHRC_UUID, _ = bluetooth.ParseUUID("4a488208-f3b9-414f-85c7-17eb16c653b0")
var configStartSampleCharUUID, _ = bluetooth.ParseUUID("f6344769-e905-4c4d-a6e8-0aa8b63f1153")
var configWakeAtUUID, _ = bluetooth.ParseUUID("9203c6cb-b4d4-49e2-a84d-415d2cb790f1")

var DATA_TYPES = map[byte]string{
	0x00: "vibration",
	0x01: "audio",
	0x02: "temperature",
	0x03: "battery",
	0x04: "flux",
}

const UNSENT_DATA_PATH = "unsent_data/"

var pairResponseCharacteristic bluetooth.Characteristic
var settingsCharacteristic bluetooth.Characteristic
var Gateway *model.Gateway

// List of known sensors, displayed by GUI
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
	}
	transmissions = make(map[[3]byte]Transmission)

	// Data collection service, requires pairing and bonding for authentication
	// https://lpccs-docs.renesas.com/Tutorial-DA145x-BLE-Security/ble_security.html
	// Out of Band could be done with a USB serial connection
	dataService := bluetooth.Service{
		UUID: DATA_SERVICE_UUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				UUID:  ACCEL_DATA_CHRC_UUID,
				Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
				WriteEvent: func(connection bluetooth.Connection, address string, offset int, value []byte) {
					// offset is used for when the MTU is less than 512 bytes, but maybe bluez handles that???

				},
			},
			{
				UUID:  FLUX_DATA_CHRC_UUID,
				Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
			},
			{
				UUID:  MIC_DATA_CHRC_UUID,
				Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
			},
			{
				UUID:  RTD_DATA_CHRC_UUID,
				Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
			},
		},
	}
	adapter.AddService(&dataService)

	configService := bluetooth.Service{
		UUID: CONFIG_SERVICE_UUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				// Keep this, use for devices to manage settings
				Handle: &settingsCharacteristic,
				UUID:   CONFIG_IDENTIFY_CHRC_UUID,
				Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
				// WriteEvent to collect device information
				WriteEvent: func(connection bluetooth.Connection, address string, offset int, value []byte) {
					mac, err := bluetooth.ParseMAC(address)
					if err != nil {
						out.Logger.Println("Failed to parse MAC", address)
						return
					}
					if len(value) > 0 {
						pairReceiveCapabilities(mac, value)
					} else {
						out.Logger.Println("Received zero value write. how.")
					}
				},
				// ReadEvent to return device-appropriate settings
				// Also how...
				ReadEvent: func(client bluetooth.Connection, address string, offset int) []byte {
					// FIXME sendSettings(address)
					return []byte{0x0}
				},
			},
		},
	}
	err = adapter.AddService(&configService)
	if err != nil {
		return err
	}

	/*
	 * Signal pairing procedure over serial.
	 * Share pairing code over serial
	 * This will need a separate BlueZ BLE agent to handle pairing.
	 * Could the default agent handle it? Think of supplying a code other than 0000
	 * Alternative would be to accept pairing only if device sends MAC address over serial
	 */
	adapter.SetConnectHandler(func(device bluetooth.Device, connected bool) {
		if connected {
			// On connect add to list of devices pending pairing
			pairDeviceConnected(device.Address.MAC)
			out.Logger.Println("Bluetooth connection with device", device.Address.MAC.String())
		} else {
			pairDeviceDisconnected(device.Address.MAC)
			out.Logger.Println("Bluetooth disconnected", device.Address.MAC.String())
		}
	})

	adapter.Enable()

	adv := adapter.DefaultAdvertisement()
	if adv == nil {
		return errors.New("advertisement is nil")
	}
	err = adv.Configure(bluetooth.AdvertisementOptions{
		LocalName: "Gateway Server",
		// BUG can't advertise more than one service UUID
		ServiceUUIDs: []bluetooth.UUID{DATA_SERVICE_UUID},
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
			x[i] = int16(rawData[i*6]) | int16(rawData[i*6+1])<<8
			y[i] = int16(rawData[i*6+2]) | int16(rawData[i*6+3])<<8
			z[i] = int16(rawData[i*6+4]) | int16(rawData[i*6+5])<<8
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
