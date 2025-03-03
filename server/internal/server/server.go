package server

import (
	// "encoding/binary"
	// "encoding/json"
	"errors"
	"os"
	"os/signal"
	"strings"

	"github.com/jukuly/ss_machmos/server/internal/model"
	"github.com/jukuly/ss_machmos/server/internal/out"
	"tinygo.org/x/bluetooth"
)

const DEFAULT_GATEWAY_HTTP_ENDPOINT = "https://openphm.org/gateway_data"

var adapter = bluetooth.DefaultAdapter

var DATA_SERVICE_UUID = bluetooth.MustParseUUID("2deacc71-7b29-4ff4-8fc2-59461c7a73f5")
var DEBUG_DATA_CHRC_UUID = bluetooth.MustParseUUID("ad690aaa-cfd4-4b4a-96a6-1110cb6782f6")
var CONTROL_DATA_CHRC_UUID = bluetooth.MustParseUUID("0b7f9057-38ef-4db5-8e25-64bc66fb1963")

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

var pairResponseCharacteristic bluetooth.Characteristic
var settingsCharacteristic bluetooth.Characteristic
var configWakeAtChar bluetooth.Characteristic
var configStartSampleChar bluetooth.Characteristic

// Gateway config
var Gateway *model.Gateway

// List of known sensors, displayed by GUI
// FIXME: move these to the model package
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

	// Data collection service, requires pairing and bonding for authentication
	// https://lpccs-docs.renesas.com/Tutorial-DA145x-BLE-Security/ble_security.html
	// Out of Band could be done with a USB serial connection
	dataService := bluetooth.Service{
		UUID: DATA_SERVICE_UUID,
		Characteristics: []bluetooth.CharacteristicConfig{
			{
				UUID:  DEBUG_DATA_CHRC_UUID,
				Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
				WriteEvent: func(client bluetooth.Connection, address string, offset int, value []byte) {
					// Dump any data received into a bin file
					handleDebugData(address, value)
				},
			},
			{
				UUID:  CONTROL_DATA_CHRC_UUID,
				Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicWriteWithoutResponsePermission,
				WriteEvent: func(client bluetooth.Connection, address string, offset int, value []byte) {
					// Sensor wants to start/end upload. To start upload sensor
					// sends a header with expected data size and type. To end
					// an upload sensor sends a flag, then server validates
					// data. Maybe for the future we could support upload
					// resuming in case of a crash?

					// TODO: lol
				},
			},
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
							// Remove flag FIXME: this is just invalid code.
							// Replace with a map for the nicer API
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
	out.Logger.Println("Notifying all connected devices of new configs")
	// NOTE: this triggers our own ReadEvent callback
	// because this library has no way to direct notify...
	settingsCharacteristic.Write([]byte{0x0})
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

// Map device address to byte cache
var bufferCache map[string][]byte = make(map[string][]byte)

func handleDebugData(address string, value []byte) {
	buffer, ok := bufferCache[address]
	//println(value)
	// If value sent is a single zero, transmission ended
	if len(value) == 1 && value[0] == 0x00 {
		if ok {
			// Write out cache to disk
			err := os.WriteFile("./"+strings.ReplaceAll(address, ":", "_")+"_debug.bin", buffer, 0644)
			if err != nil {
				println("Failed to write out file:")
				println(err.Error())
			}
			// Clear out buffer
			delete(bufferCache, address)
			out.Logger.Println("Debug data received of size", len(buffer))
		} else {
			out.Logger.Println("Device", address, "attempted to write empty array to disk")
			//out.Logger.Println(bufferCache)
		}
	} else {
		if ok {
			// Append received packet to slice
			bufferCache[address] = append(bufferCache[address], value...)
		} else {
			bufferCache[address] = value
		}
	}
}
