module github.com/jukuly/ss_machmos/server

go 1.22

require tinygo.org/x/bluetooth v0.9.0

// Made my own thing
//replace tinygo.org/x/bluetooth => github.com/PencilAmazing/go-bluetooth v0.0.0-20250203213402-0f45e7b58f53
replace tinygo.org/x/bluetooth => ../../bluetooth

require (
	github.com/soypat/cyw43439 v0.0.0-20241116210509-ae1ce0e084c5 // indirect
	github.com/soypat/seqs v0.0.0-20240527012110-1201bab640ef // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	github.com/tinygo-org/pio v0.0.0-20231216154340-cd888eb58899 // indirect
	golang.org/x/exp v0.0.0-20230728194245-b0cb94b80691 // indirect
)

require (
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/google/uuid v1.6.0
	github.com/saltosystems/winrt-go v0.0.0-20240509164145-4f7860a3bd2b // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/tinygo-org/cbgo v0.0.4 // indirect
	golang.org/x/sys v0.13.0 // indirect
)
