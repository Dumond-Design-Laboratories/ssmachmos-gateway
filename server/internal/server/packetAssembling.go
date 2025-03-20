package server

/*
 * Collecting bytes from BLE and placing into a json
 */

import (
	"encoding/binary"
	"encoding/json"
	"time"

	"github.com/jukuly/ss_machmos/server/internal/model"
	"github.com/jukuly/ss_machmos/server/internal/out"
	"tinygo.org/x/bluetooth"
)

type Packet struct {
	offset int
	data   []byte
}

type Transmission struct {
	macAddress [6]byte // FIXME make this a string
	// batteryLevel      int 		// UNUSED
	timestamp         time.Time // Transmission start time
	endTimestamp      time.Time // transmission end time
	dataType          string    // Enum-like
	samplingFrequency uint32    // Frequency of samples
	currentLength     int       // Number of packets collected so far
	totalLength       uint32    // Total amount announced by sensor
	packets           []byte
}

// https://go.dev/doc/faq#atomic_maps
// :^)
var transmissions map[[6]byte]Transmission = make(map[[6]byte]Transmission)

func savePacket(data []byte, macAddress [6]byte, dataType string) (t Transmission, ok bool) {
	if _, exists := transmissions[macAddress]; !exists {
		// New transmission
		out.Logger.Println("COLLECT-START:" + model.MacToString(macAddress))
		// First packet is a header
		totalLength := binary.LittleEndian.Uint32(data[0:4])
		samplingFrequency := binary.LittleEndian.Uint32(data[4:8])
		// batteryLevel := -1
		transmissions[macAddress] = Transmission{
			macAddress: macAddress,
			timestamp:  time.Now(),
			//batteryLevel:      batteryLevel,
			dataType:          dataType,
			samplingFrequency: samplingFrequency,
			currentLength:     0,
			totalLength:       totalLength,
			packets:           make([]byte, 0),
		}
		out.Logger.Println("Received collection header")
		out.Logger.Println("DataType:", transmissions[macAddress].dataType)
		out.Logger.Println("Total expected length:", transmissions[macAddress].totalLength)
	} else {
		// Other packets are raw data
		transmission := transmissions[macAddress]
		// Append data to end of stream
		transmission.packets = append(transmission.packets, data...)
		// increase current byte count
		transmission.currentLength += len(data)
		transmissions[macAddress] = transmission
		out.Logger.Println("Received packet from", model.MacToString(macAddress), "total", transmission.currentLength, "/", transmission.totalLength)
	}

	// Header includes expected length, expect more from that
	if transmissions[macAddress].currentLength >= int(transmissions[macAddress].totalLength) {
		fullTransmit := transmissions[macAddress]
		// Plug in end timestamp
		fullTransmit.endTimestamp = time.Now()
		//copy(fullTransmit.packets, transmissions[macAddress].packets)
		delete(transmissions, macAddress)
		out.Logger.Println("Assembled transmission packets for", model.MacToString(macAddress),
			"time taken", fullTransmit.endTimestamp.Sub(fullTransmit.timestamp))

		out.Logger.Println("COLLECT-END:" + model.MacToString(macAddress))
		return fullTransmit, true
	}

	// Nothing to return
	return Transmission{}, false
}

/*
 * Receive a data upload from the sensor
 * Each sensor data type would have a dedicated characteristic
 */
func handleData(dataType string, _ bluetooth.Connection, address string, _mtu int, value []byte) {
	if len(value) == 0 {
		out.Logger.Println("Zero byte array received from " + address + " handling data for " + dataType)
		return
	}

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
	// Keep status updated
	sensor.UpdateLastSeen(model.SensorActivityTransmitting, Sensors)
	// Append data to total data transmission
	transmitData, ok := savePacket(value, macAddress, dataType)
	if !ok {
		// incomplete data, keep waiting for more
		return
	}

	// Done collecting data, serialize to json and attempt immediate transfer after
	out.Logger.Println("Received " + dataType + " data transmission from " + model.MacToString(macAddress) + " (" + sensor.Name + ")")
	sensor.UpdateLastSeen(model.SensorActivityIdle, Sensors)

	// Pick apart data and place into json structures
	var measurements []map[string]interface{}
	if dataType == "vibration" {
		measurements = handleVibrationData(transmitData)
	} else if dataType == "temperature" {
		measurements = handleTemperatureData(transmitData)
	} else if dataType == "audio" {
		measurements = handleAudioData(transmitData)
	} else {
		out.Logger.Println("unknown data type", dataType)
		return
	}

	jsonData, err := json.Marshal(measurements)
	if err != nil {
		out.Logger.Println("Error:", err)
		return
	}

	// Upload to gateway
	resp, err := sendMeasurements(jsonData, Gateway)

	// catch marshal errors
	if err != nil {
		out.Logger.Println("Error:", err)
		// Save to disk instead
		if err := saveUnsentMeasurements(jsonData, transmitData.timestamp); err != nil {
			out.Logger.Println("Error:", err)
		}
		return
	}

	// TODO: we could use this point to verify gateway settings

	// Catch gateway response
	if resp.StatusCode != 200 {
		out.Logger.Println("Error sending data to server")
		// Print out response error
		body := make([]byte, resp.ContentLength)
		defer resp.Body.Close()
		resp.Body.Read(body)
		out.Logger.Println(string(body))
		// Save data
		if err := saveUnsentMeasurements(jsonData, transmitData.timestamp); err != nil {
			out.Logger.Println(err)
		}
		return
	}

	// Send everything else in store
	sendUnsentMeasurements()
}

// Accelerometer json output
func handleVibrationData(transmitData Transmission) []map[string]interface{} {
	convRange8G := .000244
	rawData := transmitData.packets
	numberOfMeasurements := len(rawData) / 6 // 3 axes, 2 bytes per axis => 6 bytes per measurement
	out.Logger.Println("Vibration data consists of", len(rawData), "bytes =", numberOfMeasurements, "measurements.")
	x, y, z := make([]float64, numberOfMeasurements), make([]float64, numberOfMeasurements), make([]float64, numberOfMeasurements)
	for i := 0; i < len(rawData); i += 6 {
		//out.Logger.Println(rawData[i:i+6])
		// We're receiving signed integers, of course.
		x[i/6] = float64(int16(transmitData.packets[i+1])<<8|int16(transmitData.packets[i+0])) * convRange8G
		y[i/6] = float64(int16(transmitData.packets[i+3])<<8|int16(transmitData.packets[i+2])) * convRange8G
		z[i/6] = float64(int16(transmitData.packets[i+5])<<8|int16(transmitData.packets[i+4])) * convRange8G
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

// Temperature data is a single number in an array
func handleTemperatureData(transmitData Transmission) []map[string]any {
	measurements := []map[string]any{}

	if len(transmitData.packets) == 2 {
		temperature, err := parseTemperatureData(int16(transmitData.packets[1])<<8 | int16(transmitData.packets[0]))
		if err == nil {
			measurements = append(measurements,
				map[string]any{
					"sensor_id":          model.MacToString(transmitData.macAddress),
					"time":               transmitData.timestamp,
					"measurement_type":   transmitData.dataType,
					"sampling_frequency": transmitData.samplingFrequency,
					// Has to be an array
					"raw_data": []float64{temperature},
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
		//out.Logger.Println(transmitData.packets[5000])
		numberOfMeasurements := len(transmitData.packets) / 3
		amplitude := make([]int, numberOfMeasurements)
		// Collect bytes into 24 bit integers
		for i := 0; i < numberOfMeasurements; i++ {
			// Data is sent as left aligned, little endian uint32 bytes
			// NOTE: this is how OpenPHM expects the bytes to be assembled. Flipped.
			amplitude[i] = int(transmitData.packets[i*3])<<16 | int(transmitData.packets[i*3+1])<<8 | int(transmitData.packets[i*3+2])
		}

		// NOTE: The mic sensor sends some zeros at the beginning, we want to eliminate those
		for i, amp := range amplitude {
			if i > 512 {
				break
			}
			if amp != 0 {
				// First non-zero byte found
				amplitude = amplitude[i:]
				break
			}
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
		out.Logger.Println("Invalid audio data received. Packets of length", len(transmitData.packets), "not multiple of 3.")
	}
	return measurements
}
