# Transmission Protocol

- Uses 2048 bits RSA keys with SHA256 hash
- Uses RSA PKCS #1 v1.5 signatures
- All values must be written in Little Endian
- "Header": 0x00 if request (from sensor to gateway), 0x01 if response (from gateway to sensor) (1 byte) | mac address (FF:FF:FF:FF:FF:FF if broadcasting to everyone) (6 bytes) | content as described below 

## Pairing:

- The sensor generates a key pair and sends his public key, the data types it can collect, the maximum size in bytes of data it can send, and its mac address to the server => data types (1 byte) | collection capacity in bytes (4 bytes) | public key
- Data types: b(0 0 0 0 0 vibration temperature audio)
- The user has 30 seconds to accept the pairing request
- The server writes to the "pairing response" characteristic with the UUID of the data transmission characteristic, the UUID of the settings characteristic and the mac address of the sender (to tell the sensors which one has been accepted) => data characteristic uuid (16 bytes) | settings characteristic uuid (16 bytes)
- The sensor sends an ACK to tell the server he indeed received the UUIDs. From now on, every communication will be signed by the sensor. If the ACK is not received in a delay of 30 seconds by the server, the pairing is cancelled. => data characteristic uuid (16 bytes) | settings characteristic uuid (16 bytes) | signature (256 bytes)

- For now:
- Request: 0x00 | sensor mac address | 0b00000111 | collection capacity
- Response: 0x01 | sensor mac address | data char | settings char
- Confirmation: 0x00 | sensor mac address | data char | settings char

## Data transmission

- The sensor sends the data with a couple of metadata and signs it => battery level in % (1 byte) | data type (1 byte) | sampling frequency in Hz (4 bytes) | length of data (4 bytes) | data | signature (256 bytes)
- Can send multiple data types at once
- Data type: 0x00 => vibration, 0x01 => audio, 0x02 => temperature
- Send -1 as battery level if don't want to send it
- Sampling frequency must be present even if the data type doesn't have a sampling frequency (temperature) (can be anything since it will not be used at all when decoding the data)

- For now:
- 0x00 | sensor mac address | battery level (or -1) | data type | sampling frequency | length of data | message id (3 bytes) | offset in bytes (4 bytes) | data

## Settings changes

- Whenever a sensor wakes up and right after pairing (to get the first wake up time) he sends a request to the server to fetch his settings. => nothing
- Server response: time until next wake up in milliseconds (4 bytes => max 50 days) | for each data type: { 0b00000 | type (2 bits) | active (1 bit) | sampling frequency in Hz (4 bytes) | sampling duration in ms (2 bytes) }
- Data type: 0x00 => vibration, 0x01 => audio, 0x02 => temperature
- It is possible that the settings characteristic is overwritten before the sensor can read from it. To solve this, the sensor should ask for his settings again every 10 seconds + 10 seconds for each time it didn't get the answer in time (to avoid multiple sensors always fighting to get their settings)

- For now:
- Request: 0x00 | sensor mac address
- Response: 0x01 | sensor mac address | time until next wake up | for each data type: { type | active | sampling frequency | sampling duration }




- vibration => 6 bytes/samples => 2 bytes/axis => multiply * float => in G (x, y, z)
- audio => 3 bytes => 24 bit integer => pcm24
