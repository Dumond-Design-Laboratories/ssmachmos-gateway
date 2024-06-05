# Transmission Protocol

Uses 2048 bits RSA keys with SHA256 hash<br />
Uses RSA PKCS #1 v1.5 signatures

## Pairing:

- The sensor generates a key pair and sends his public key as well as his mac address to the server => mac (6 bytes) | public key
- The user has 30 seconds to accept the pairing request
- The server writes to the "pairing response" characteristic with the UUID of the data transmission characteristic and the mac address of the sender (to tell the sensor which one has been accepted) => mac (6 bytes) | uuid (16 bytes)
- The sensor sends an ACK to tell the server he indeed received the UUID. From now on, every communication will be signed by the sensor. If the ACK is not received in a delay of 30 seconds by the server, the pairing is cancelled. => mac (6 bytes) | uuid (16 bytes) | signature (256 bytes)

### Add to pairing

- Pair request should contain the measurement types the sensor is capable of collecting => 1 byte (0 0 0 0 0 vibration temperature acoustic)
- add: sensor will send transmission maximum size => cap on the ui the different settings to fit this number

## Data transmission

- The sensor sends the data with a couple of metadata and signs it => mac (6 bytes) | data type (1 byte) | sampling frequency (3 bytes) | data | signature (256 bytes)
- Data type: 0x00 => vibration, 0x01 => audio, 0x02 => temperature, 0x03 => battery,
- Sampling frequency must be present even if the data type doesn't have a sampling frequency (temperature and battery) (can be anything since it will not be used at all when decoding the data)

### Add to data transmission

- Add battery information when data is transmitted (1 byte) (%)
- Instead of having the battery be a separate measurement, add a battery level reading to each measurement sent to the server

## Settings changes

- Whenever a sensor wakes up he sends a request to the server to fetch his settings
- Server response: mac of sensor (6 bytes) | for each data type: { type (1 byte) | active (1 byte) | sampling frequency (4 bytes) | sampling duration (2 bytes) | time until next wake up in milliseconds (4 bytes => max 50 days) }
- Data type: 0x00 => vibration, 0x01 => audio, 0x02 => temperature
- Still have to make sure with Zetai that this is OK
