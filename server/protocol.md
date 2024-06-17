# Transmission Protocol

- Uses 2048 bits RSA keys with SHA256 hash
- Uses RSA PKCS #1 v1.5 signatures
- All values must be written in Little Endian

## Pairing:

- The sensor generates a key pair and sends his public key, the data types it can collect, the maximum size in bytes of data it can send, and its mac address to the server => mac (6 bytes) | data types (1 byte) | collection capacity in bytes (4 bytes) | public key
- Data types: b(0 0 0 0 0 vibration temperature acoustic)
- The user has 30 seconds to accept the pairing request
- The server writes to the "pairing response" characteristic with the UUID of the data transmission characteristic, the UUID of the settings characteristic and the mac address of the sender (to tell the sensors which one has been accepted) => mac (6 bytes) | data characteristic uuid (16 bytes) | settings characteristic uuid (16 bytes)
- The sensor sends an ACK to tell the server he indeed received the UUIDs. From now on, every communication will be signed by the sensor. If the ACK is not received in a delay of 30 seconds by the server, the pairing is cancelled. => mac (6 bytes) | data characteristic uuid (16 bytes) | settings characteristic uuid (16 bytes) | signature (256 bytes)

## Data transmission

- The sensor sends the data with a couple of metadata and signs it => mac (6 bytes) | battery level in % (1 byte) | for each data type: { data type (1 byte) | sampling frequency in Hz (4 bytes) | length of data (4 bytes) | data } | signature (256 bytes)
- Can send multiple data types at once
- Data type: 0x00 => vibration, 0x01 => audio, 0x02 => temperature
- Send -1 as battery level if don't want to send it
- Sampling frequency must be present even if the data type doesn't have a sampling frequency (temperature) (can be anything since it will not be used at all when decoding the data)

## Settings changes

- Whenever a sensor wakes up he sends a request to the server and right after pairing (to get the first wake up time) to fetch his settings. => mac (6 bytes)
- Server response: mac of sensor (6 bytes) | time until next wake up in milliseconds (4 bytes => max 50 days) | for each data type: { type (1 byte) | active (1 byte) | sampling frequency in Hz (4 bytes) | sampling duration in s (2 bytes) }
- Data type: 0x00 => vibration, 0x01 => audio, 0x02 => temperature
- It is possible that the settings characteristic is overwritten before the sensor can read from it. To solve this, the sensor should ask for his settings again every 10 seconds + 10 seconds for each time it didn't get the answer in time (to avoid multiple sensors always fighting to get their settings)
