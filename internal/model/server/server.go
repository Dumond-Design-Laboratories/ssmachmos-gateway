package server

import (
	"crypto/rsa"
	"encoding/json"
	"math"
	"time"

	"github.com/jukuly/ss_mach_mo/internal/cli/out"
	"github.com/jukuly/ss_mach_mo/internal/model"
	"tinygo.org/x/bluetooth"
)

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

func Init(sensors *[]model.Sensor, gateway *model.Gateway) error {
	err := adapter.Enable()
	if err != nil {
		return err
	}

	pairingState = PairingState{
		active:    false,
		requested: make(map[[6]byte]*rsa.PublicKey),
		pairing:   [6]byte{},
	}

	dataCharUUID, err := model.GetDataCharUUID(gateway)
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
					handleData(client, offset, value, sensors, gateway)
				},
			},
			{
				UUID:  PAIR_REQUEST_CHARACTERISTIC_UUID,
				Flags: bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					pairRequest(value)
				},
			},
			{
				Handle: &pairResponseCharacteristic,
				UUID:   PAIR_RESPONSE_CHARACTERISTIC_UUID,
				Value:  []byte{},
				Flags:  bluetooth.CharacteristicReadPermission | bluetooth.CharacteristicWritePermission,
				WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
					pairConfirmation(value, sensors, gateway)
				},
			},
		},
	}
	err = adapter.AddService(&service)
	if err != nil {
		return err
	}

	err = adapter.DefaultAdvertisement().Configure(bluetooth.AdvertisementOptions{
		LocalName:    "Gateway Server",
		ServiceUUIDs: []bluetooth.UUID{service.UUID},
	})
	if err != nil {
		return err
	}
	return nil
}

func StartAdvertising() error {
	err := adapter.DefaultAdvertisement().Start()
	if err != nil {
		return err
	}
	return nil
}

func handleData(_ bluetooth.Connection, _ int, value []byte, sensors *[]model.Sensor, gateway *model.Gateway) {

	if len(value) < 266 {
		out.Log("Invalid data format received")
		return
	}
	data := value[:len(value)-256]
	signature := value[len(value)-256:]

	macAddress := [6]byte(data[:6])
	var sensor *model.Sensor
	for _, s := range *sensors {
		if s.Mac == macAddress {
			sensor = &s
			break
		}
	}
	if sensor == nil {
		out.Log("Device " + model.MacToString(macAddress) + " tried to send data, but it is not paired with this gateway")
		return
	}

	if !model.VerifySignature(data, signature, &sensor.PublicKey) {
		out.Log("Invalid signature received from " + model.MacToString(macAddress))
		return
	}

	dataType := DATA_TYPES[data[6]]
	samplingFrequency := data[7:10]
	timestamp := time.Now().Unix()

	out.Log("Received " + dataType + " data from " + model.MacToString(macAddress) + " (" + sensor.Name + ")")

	var measurements []map[string]interface{}
	if dataType == "vibration" {
		numberOfMeasurements := (len(data) - 8) / 12 // 3 axes, 4 bytes per axis => 12 bytes per measurement
		x, y, z := make([]float32, numberOfMeasurements), make([]float32, numberOfMeasurements), make([]float32, numberOfMeasurements)
		for i := 0; i < numberOfMeasurements; i++ {
			x[i] = math.Float32frombits(uint32(data[10+i*12]) | uint32(data[11+i*12])<<8 | uint32(data[12+i*12])<<16 | uint32(data[13+i*12])<<24)
			y[i] = math.Float32frombits(uint32(data[14+i*12]) | uint32(data[15+i*12])<<8 | uint32(data[16+i*12])<<16 | uint32(data[17+i*12])<<24)
			z[i] = math.Float32frombits(uint32(data[18+i*12]) | uint32(data[19+i*12])<<8 | uint32(data[20+i*12])<<16 | uint32(data[21+i*12])<<24)
		}

		measurements = []map[string]interface{}{
			{
				"sensor_id":          sensor.Mac,
				"time":               timestamp,
				"measurement_type":   dataType,
				"sampling_frequency": samplingFrequency,
				"axis":               "x",
				"raw_data":           x,
			},
			{
				"sensor_id":          sensor.Mac,
				"time":               timestamp,
				"measurement_type":   dataType,
				"sampling_frequency": samplingFrequency,
				"axis":               "y",
				"raw_data":           y,
			},
			{
				"sensor_id":          sensor.Mac,
				"time":               timestamp,
				"measurement_type":   dataType,
				"sampling_frequency": samplingFrequency,
				"axis":               "z",
				"raw_data":           z,
			},
		}
	} else {
		measurements = []map[string]interface{}{
			{
				"sensor_id":          sensor.Mac,
				"time":               timestamp,
				"measurement_type":   dataType,
				"sampling_frequency": samplingFrequency,
				"raw_data":           data[10:],
			},
		}
	}

	jsonData, _ := json.Marshal(measurements)
	resp, err := sendMeasurements(jsonData, gateway)

	if err != nil {
		out.Log("Error sending data to server")
		out.Error(err)
		saveUnsentMeasurements(jsonData, timestamp)
		return
	}

	if resp.StatusCode != 200 {
		out.Log("Error sending data to server")

		body := make([]byte, resp.ContentLength)
		defer resp.Body.Close()
		resp.Body.Read(body)
		out.Log(string(body))
		saveUnsentMeasurements(jsonData, timestamp)
		return
	}

	sendUnsentMeasurements(gateway)
}
