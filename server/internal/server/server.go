package server

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"
	"os/signal"

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

// Chrc to notify a sensor to start sending data immediately.
var CONFIG_START_SAMPLING_CHRC_UUID, _ = bluetooth.ParseUUID("f6344769-e905-4c4d-a6e8-0aa8b63f1153")

// Chrc to notify a sensor of next wake up time.
var CONFIG_WAKE_AT_CHRC_UUID, _ = bluetooth.ParseUUID("9203c6cb-b4d4-49e2-a84d-415d2cb790f1")

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
var configWakeAtChar bluetooth.Characteristic
var configStartSampleChar bluetooth.Characteristic

// Gateway config
var Gateway *model.Gateway

// List of known sensors, displayed by GUI
var Sensors *[]model.Sensor

// List of devices flagged for collection
var flaggedForCollect []string

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
	//transmissions = make(map[[3]byte]Transmission)

	// Data collection service, requires pairing and bonding for authentication
	// https://lpccs-docs.renesas.com/Tutorial-DA145x-BLE-Security/ble_security.html
	// Out of Band could be done with a USB serial connection
	dataService := bluetooth.Service{
		UUID: DATA_SERVICE_UUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				UUID:  ACCEL_DATA_CHRC_UUID,
				Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
				// offset is used for when the MTU is less than 512 bytes, but maybe bluez handles that???
				WriteEvent: func(client bluetooth.Connection, address string, offset int, value []byte) {
					handleData("vibration", client, address, offset, value)
				},
			},
			{
				UUID:  FLUX_DATA_CHRC_UUID,
				Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
				WriteEvent: func(client bluetooth.Connection, address string, offset int, value []byte) {
					handleData("flux", client, address, offset, value)
				},
			},
			{
				UUID:  MIC_DATA_CHRC_UUID,
				Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
				WriteEvent: func(client bluetooth.Connection, address string, offset int, value []byte) {
					handleData("audio", client, address, offset, value)
				},
			},
			{
				UUID:  RTD_DATA_CHRC_UUID,
				Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
				WriteEvent: func(client bluetooth.Connection, address string, offset int, value []byte) {
					handleData("temperature", client, address, offset, value)
				},
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
				Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicNotifyPermission,
				// WriteEvent to collect device information
				WriteEvent: func(connection bluetooth.Connection, address string, offset int, value []byte) {
					mac, err := bluetooth.ParseMAC(address)
					if err != nil {
						out.Logger.Println("SettingsCharacteristic: Failed to parse MAC", address)
						return
					}
					if len(value) > 0 {
						pairReceiveCapabilities(mac, value)
					} else {
						out.Logger.Println("Settings characteristic: Received zero value from device", address, " how.")
					}
				},
				// ReadEvent to return device-appropriate settings
				// Also how...
				ReadEvent: func(client bluetooth.Connection, address string, offset int) []byte {
					out.Logger.Println("Device", address, "requested settings")
					return getSettingsForSensor(address)
				},
			},
			{
				// Signals to devices if they should start sampling immediately
				// works if device is awake and connected, mostly for debugging transmission speed
				Handle: &configStartSampleChar,
				UUID:   CONFIG_START_SAMPLING_CHRC_UUID,
				Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicNotifyPermission,
				ReadEvent: func(client bluetooth.Connection, address string, offset int) []byte {
					for i, dev := range flaggedForCollect {
						if dev == address {
							// Remove flag
							flaggedForCollect[i] = flaggedForCollect[len(flaggedForCollect)-1]
							flaggedForCollect = flaggedForCollect[:len(flaggedForCollect)-1]
							return []byte{0x1}
						}
					}
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
			// NOTE upstream go bluetooth MAC address arrays are reversed
			// This here works if using the patched branch
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

// Notify all connected devices to read configuration again
func TriggerSettingCollection() {
	out.Logger.Println("Notifying all connected devices of new configs");
	// NOTE: this triggers our own ReadEvent callback
	// because this library has no way to direct notify...
	settingsCharacteristic.Write([]byte{0x0});
}

func TriggerCollection(address string) {
	out.Logger.Println("Notifying " + address + " of collection request")
	// notify a device that we want it to start collecting data
	mac, _ := model.StringToMac(address)
	output := make([]byte, 8)
	for i := range mac {
		// Write mac address reversed
		output[i] = mac[len(mac)-i-1]
	}
	output[6] = 0x0 // Command byte
	output[7] = 0x0 // Targets sensor

	// Notify
	configStartSampleChar.Write(output)
}

/*
 * Receive a data upload from the sensor
 * Each sensor data type would have a dedicated characteristic
 */
func handleData(dataType string, _ bluetooth.Connection, address string, mtu int, value []byte) {
	// if len(value) < 264 {
	if len(value) == 0 {
		out.Logger.Println("Zero byte array received from " + address + " handling data for " + dataType)
		return
	}
	//data := value // [:len(value)-256]
	// signature := value[len(value)-256:]

	// Find sensor that is sending data
	macAddress, _ := model.StringToMac(address)
	var sensor *model.Sensor = nil
	for i, s := range *Sensors {
		if s.Mac == macAddress {
			sensor = &(*Sensors)[i]
			break
		}
	}
	// Ensure sensor is permitted to send data
	// TODO only devices that pair with gateway are allowed to access this chrc anyways
	if sensor == nil {
		// BUG the MAC address received here is reversed somehow...
		out.Logger.Println("Device " + address + " tried to send data, but it is not paired with this gateway")
		return
	}

	// Append data to total data transmission
	transmitData, ok := savePacket(value, macAddress, dataType)
	if !ok {
		// incomplete data, keep waiting for more
		return
	}

	// Done collecting data, serialize to json and attempt immediate transfer after
	out.Logger.Println("Received " + dataType + " data from " + model.MacToString(macAddress) + " (" + sensor.Name + ")")

	// Pick apart data and place into json structures
	var measurements []map[string]interface{}
	if dataType == "vibration" {
		measurements = handleVibrationData(transmitData)
	} else if dataType == "temperature" {
		measurements = handleTemperatureData(transmitData)
	} else if dataType == "audio" {
		measurements = handleAudioData(transmitData)
	}

	jsonData, err := json.Marshal(measurements)
	if err != nil {
		out.Logger.Println("Error:", err)
		return
	}

	// Upload to gateway
	resp, err := sendMeasurements(jsonData, Gateway)

	if err != nil {
		out.Logger.Println("Error:", err)
		if err := saveUnsentMeasurements(jsonData, transmitData.timestamp); err != nil {
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
		if err := saveUnsentMeasurements(jsonData, transmitData.timestamp); err != nil {
			out.Logger.Println(err)
		}
		return
	}

	sendUnsentMeasurements()
}

// Accelerometer json output
func handleVibrationData(transmitData Transmission) []map[string]interface{} {
	rawData := transmitData.packets
	numberOfMeasurements := len(rawData) / 6 // 3 axes, 2 bytes per axis => 6 bytes per measurement
	x, y, z := make([]int16, numberOfMeasurements), make([]int16, numberOfMeasurements), make([]int16, numberOfMeasurements)
	for i := 0; i < numberOfMeasurements; i++ {
		x[i] = int16(rawData[i*6]) | int16(rawData[i*6+1])<<8
		y[i] = int16(rawData[i*6+2]) | int16(rawData[i*6+3])<<8
		z[i] = int16(rawData[i*6+4]) | int16(rawData[i*6+5])<<8
	}

	measurements := []map[string]interface{}{}
	measurements = append(measurements,
		map[string]interface{}{
			"sensor_id":          model.MacToString(transmitData.macAddress),
			"time":               transmitData.timestamp,
			"measurement_type":   transmitData.dataType,
			"sampling_frequency": transmitData.samplingFrequency,
			"axis":               "x",
			"raw_data":           x,
		},
		map[string]interface{}{
			"sensor_id":          model.MacToString(transmitData.macAddress),
			"time":               transmitData.timestamp,
			"measurement_type":   transmitData.dataType,
			"sampling_frequency": transmitData.samplingFrequency,
			"axis":               "y",
			"raw_data":           y,
		},
		map[string]interface{}{
			"sensor_id":          model.MacToString(transmitData.macAddress),
			"time":               transmitData.timestamp,
			"measurement_type":   transmitData.dataType,
			"sampling_frequency": transmitData.samplingFrequency,
			"axis":               "z",
			"raw_data":           z,
		},
	)

	return measurements
}

func handleTemperatureData(transmitData Transmission) []map[string]interface{} {
	measurements := []map[string]interface{}{}

	if len(transmitData.packets) == 2 {
		temperature, err := parseTemperatureData(binary.LittleEndian.Uint16(transmitData.packets))
		if err == nil {
			measurements = append(measurements,
				map[string]interface{}{
					"sensor_id":          model.MacToString(transmitData.macAddress),
					"time":               transmitData.timestamp,
					"measurement_type":   transmitData.dataType,
					"sampling_frequency": transmitData.samplingFrequency,
					"raw_data":           temperature,
				},
			)
		} else {
			out.Logger.Println("Error:", err)
		}
	} else {
		out.Logger.Println("Invalid temperature data received")
	}

	return measurements
}

func handleAudioData(transmitData Transmission) []map[string]interface{} {
	measurements := []map[string]interface{}{}
	if len(transmitData.packets)%3 == 0 {
		numberOfMeasurements := len(transmitData.packets) / 3
		amplitude := make([]int, numberOfMeasurements)
		for i := 0; i < numberOfMeasurements; i++ {
			amplitude[i] = int(transmitData.packets[i*3]) | int(transmitData.packets[i*3+1])<<8 | int(transmitData.packets[i*3+2])<<16
		}

		measurements = append(measurements,
			map[string]interface{}{
				"sensor_id":          model.MacToString(transmitData.macAddress),
				"time":               transmitData.timestamp,
				"measurement_type":   transmitData.dataType,
				"sampling_frequency": transmitData.samplingFrequency,
				"raw_data":           amplitude,
			},
		)
	} else {
		out.Logger.Println("Invalid audio data received")
	}
	return measurements
}
