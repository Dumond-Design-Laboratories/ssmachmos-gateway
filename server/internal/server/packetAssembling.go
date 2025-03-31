package server

/*
 * Collecting bytes from BLE and placing into a json
 */

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jukuly/ss_machmos/server/internal/model"
	"github.com/jukuly/ss_machmos/server/internal/out"
	"tinygo.org/x/bluetooth"
)

// Maximum time in seconds a transmission can stay idle before being cancelled
// NOTE: Should this be less than the wake-up time? Or simply extrememly short?
const TRANSMISSION_TIMEOUT = 30

type Packet struct {
	offset int
	data   []byte
}

type Transmission struct {
	macAddress        [6]byte   // FIXME make this a string
	sensorModel       string    // Board model to choose conversion algorithm
	timestamp         time.Time // Transmission start time
	endTimestamp      time.Time // transmission end time
	dataType          string    // Enum-like
	samplingFrequency uint32    // Frequency of samples
	currentLength     int       // To compare with totalLength promised by sensor
	totalLength       uint32    // Total amount announced by sensor
	packets           []byte    // Byte stream of sent numbers
	lastActivity      int64     // Last time activity was seen here
	stale             bool      // Transmission timed out, discard on next touch
}

// https://go.dev/doc/faq#atomic_maps
// TODO: See if another package like concurrent-map might help here
var transmissionMutex sync.RWMutex

type TransmissionMap map[[6]byte]Transmission

var transmissions TransmissionMap = make(TransmissionMap)

// func (tm *TransmissionMap) Put(key [6]byte, value Transmission) {
// 	transmissionMutex.Lock()
// 	(*tm)[key] = value
// 	transmissionMutex.Unlock()
// }

// func (tm *TransmissionMap) Read(key [6]byte) Transmission {
// 	transmissionMutex.RLock()
// 	value := (*tm)[key]
// 	transmissionMutex.RUnlock()
// 	return value
// }

// Activity watchdog goroutine timer to delete stale pending transmissions. If a
// sensor fails during upload or crashes, the gateway would still keep the data
// in memory. If the sensor starts a new transmission, the new data would be
// appended into the old data and uploaded, then the rest would get queued in
// and ruin subsequent uploads.
func startWatchdog() {
	for {
		// Sleep for a second
		time.Sleep(1 * time.Second)
		now := time.Now().Unix()

		transmissionMutex.RLock() // Lock read
		for mac, transmission := range transmissions {
			transmissionMutex.RUnlock() // unlock after read
			// Time is up, mark for deletion
			if now-transmission.lastActivity >= TRANSMISSION_TIMEOUT {
				transmissionMutex.Lock() // write lock
				t := transmissions[mac]
				t.stale = true
				transmissions[mac] = t
				transmissionMutex.Unlock() // write unlock
				out.Logger.Printf("Idle timeout transmission for %s datatype %s", model.MacToString(mac), transmission.dataType)
			}
			transmissionMutex.RLock() // lock before read
		}
		transmissionMutex.RUnlock()
	}
}

func savePacket(data []byte, macAddress [6]byte, dataType string) (t Transmission, ok bool) {
	transmissionMutex.RLock()
	transmission, exists := transmissions[macAddress]
	transmissionMutex.RUnlock()
	// If new transmission, or replacing stale transmission
	if exists == false || transmission.stale == true {
		// New transmission
		out.Logger.Println("COLLECT-START:" + model.MacToString(macAddress))
		// First packet is a header, unpack
		totalLength := binary.LittleEndian.Uint32(data[0:4])
		samplingFrequency := binary.LittleEndian.Uint32(data[4:8])
		transmissionMutex.Lock()
		transmissions[macAddress] = Transmission{
			macAddress:        macAddress,
			sensorModel:       sensorExists(macAddress).Model,
			timestamp:         time.Now(),
			dataType:          dataType,
			samplingFrequency: samplingFrequency,
			currentLength:     0,
			totalLength:       totalLength,
			packets:           make([]byte, 0),
			lastActivity:      time.Now().Unix(),
			stale:             false,
		}
		transmissionMutex.Unlock()
		out.Logger.Println("Received collection header")
		out.Logger.Println("DataType:", transmissions[macAddress].dataType)
		out.Logger.Println("Total expected length:", transmissions[macAddress].totalLength)
	} else {
		transmissionMutex.Lock()
		// Other packets are raw data
		transmission := transmissions[macAddress]
		transmission.packets = append(transmission.packets, data...) // Append data to end of stream
		transmission.currentLength += len(data)                      // increase current byte count
		transmission.lastActivity = time.Now().Unix()                // Update idle timer
		transmissions[macAddress] = transmission
		transmissionMutex.Unlock()
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
	for i, s := range *model.Sensors {
		if s.Mac == macAddress {
			sensor = &(*model.Sensors)[i]
			break
		}
	}
	// Ensure sensor is permitted to send data
	// TODO only devices that pair with gateway are allowed to access this chrc anyways
	if sensor == nil {
		out.Logger.Println("Device " + address + " tried to send data, but it is not paired with this gateway")
		return
	}
	// Keep status updated
	sensor.UpdateLastSeen(model.SensorActivityTransmitting, model.Sensors)
	// Append data to total data transmission
	transmitData, ok := savePacket(value, macAddress, dataType)
	if !ok {
		// incomplete data, keep waiting for more
		return
	}

	// Done collecting data, serialize to json and attempt immediate transfer after
	out.Logger.Println("Received " + dataType + " data transmission from " + model.MacToString(macAddress) + " (" + sensor.Name + ")")
	sensor.UpdateLastSeen(model.SensorActivityIdle, model.Sensors)

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
		// TODO: Get sensor model from transmission
		var temperature float64
		var err error
		if transmitData.sensorModel == "machmo" {
			var digitalTemp int16 = int16(transmitData.packets[1])<<8 | int16(transmitData.packets[0])
			temperature, err = parseTemperatureData(digitalTemp)
			out.Logger.Println("MachMo temperature reading")
		} else if transmitData.sensorModel == "machmomini" {
			var digitalTemp int16 = int16(transmitData.packets[1])<<8 | int16(transmitData.packets[0])
			err = nil
			temperature = float64(digitalTemp) * 0.0625
			out.Logger.Println(transmitData.packets)
			out.Logger.Printf("MachMo mini temperature digital %d celsius %f", digitalTemp, temperature)
		} else {
			err = errors.New("Unknown board model " + transmitData.sensorModel)
		}

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
		out.Logger.Println("Invalid temperature data received, expected 2 bytes but received", len(transmitData.packets))
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
