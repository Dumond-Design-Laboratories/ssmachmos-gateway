package in

import (
	"fmt"

	"github.com/jukuly/ss_mach_mo/internal/model"
	"github.com/jukuly/ss_mach_mo/internal/model/server"
	"github.com/jukuly/ss_mach_mo/internal/view/out"
)

func help(args []string) {
	if len(args) == 0 {
		fmt.Print("\n+---------+------------+---------------------------------+-----------------------------------+\n" +
			"| Command | Options    | Arguments                       | Description                       |\n" +
			"+---------+------------+---------------------------------+-----------------------------------+\n" +
			"| help    | None       | None                            | View this table                   |\n" +
			"|         |            | <command>                       | View usage and description        |\n" +
			"|         |            |                                 | of a specific command             |\n" +
			"+---------+------------+---------------------------------+-----------------------------------+\n" +
			"| list    | None       | None                            | List all sensors                  |\n" +
			"+---------+------------+---------------------------------+-----------------------------------+\n" +
			"| view    | None       | <mac-address>                   | View a specific sensors' settings |\n" +
			"+---------+------------+---------------------------------+-----------------------------------+\n" +
			"| pair    | --enable   | None                            | Enable pairing mode               |\n" +
			"|         | --disable  | None                            | Disable pairing mode              |\n" +
			"|         | --accept   | <mac-address>                   | Accept a pairing request          |\n" +
			"+---------+------------+---------------------------------+-----------------------------------+\n" +
			"| forget  | None       | <mac-address>                   | Forget a sensor                   |\n" +
			"+---------+------------+---------------------------------+-----------------------------------+\n" +
			"| config  | --id       | <gateway-id>                    | Set the Gateway Id                |\n" +
			"|         | --password | <gateway-password>              | Set the Gateway Password          |\n" +
			"|         | --sensor   | <mac-address> <setting> <value> | Set a setting of a sensor         |\n" +
			"+---------+------------+---------------------------------+-----------------------------------+\n")
		return
	}

	switch args[0] {
	case "help":
		fmt.Print("\n+---------+------------+---------------------------------+-----------------------------------+\n" +
			"| help    | None       | None                            | View all commands and their usage |\n" +
			"|         |            | <command>                       | View usage and description        |\n" +
			"|         |            | <command>                       | of a specific command             |\n" +
			"+---------+------------+---------------------------------+-----------------------------------+\n")

	case "list":
		fmt.Print("\n+---------+------------+---------------------------------+-----------------------------------+\n" +
			"| list    | None       | None                            | List all sensors                  |\n" +
			"+---------+------------+---------------------------------+-----------------------------------+\n")

	case "view":
		fmt.Print("\n+---------+------------+---------------------------------+-----------------------------------+\n" +
			"| view    | None       | <mac-address>                   | View a specific sensors' settings |\n" +
			"+---------+------------+---------------------------------+-----------------------------------+\n")

	case "pair":
		fmt.Print("\n+---------+------------+---------------------------------+-----------------------------------+\n" +
			"| pair    | --enable   | None                            | Enable pairing mode               |\n" +
			"|         | --disable  | None                            | Disable pairing mode              |\n" +
			"|         | --accept   | <mac-address>                   | Accept a pairing request          |\n" +
			"+---------+------------+---------------------------------+-----------------------------------+\n")

	case "forget":
		fmt.Print("\n+---------+------------+---------------------------------+-----------------------------------+\n" +
			"| forget  | None       | <mac-address>                   | Forget a sensor                   |\n" +
			"+---------+------------+---------------------------------+-----------------------------------+\n")

	case "config":
		fmt.Print("\n+---------+------------+---------------------------------+-----------------------------------+\n" +
			"| config  | --id       | <gateway-id>                    | Set the Gateway Id                |\n" +
			"|         | --password | <gateway-password>              | Set the Gateway Password          |\n" +
			"|         | --sensor   | <mac-address> <setting> <value> | Set a setting of a sensor         |\n" +
			"+---------+------------+---------------------------------+-----------------------------------+\n")

	default:
		fmt.Printf("Unknown command: %s\n", args[0])
	}
}

func list(sensors *[]model.Sensor) {
	out.DisplaySensors(*sensors)
}

func view(args []string, sensors *[]model.Sensor) {
	if len(args) == 0 {
		fmt.Println("Usage: view <mac-address>")
		return
	}
	for _, sensor := range *sensors {
		if sensor.IsMacEqual(args[0]) {
			out.DisplaySensor(sensor)
			return
		}
	}
	fmt.Printf("Sensor with MAC address %s not found\n", args[0])
}

func pair(options []string, args []string) {
	if len(options) == 0 {
		fmt.Print("Usage: pair --enable\n" +
			"            --disable\n" +
			"            --accept <mac-address>\n")
		return
	}
	switch options[0] {
	case "--enable":
		server.EnablePairing()
	case "--disable":
		server.DisablePairing()
	case "--accept":
		if len(args) == 0 {
			fmt.Println("Usage: pair --accept <mac-address>")
			return
		}
		mac, _ := model.StringToMac(args[0])
		server.Pair(mac)
	default:
		fmt.Printf("Option %s does not exist for command pair\n", options[0])
	}
}

func forget(args []string, sensors *[]model.Sensor) {
	if len(args) == 0 {
		fmt.Println("Usage: forget <mac-address>")
		return
	}
	err := model.RemoveSensor(args[0], sensors)
	if err != nil {
		out.Error(err)
		return
	}
}

func config(options []string, args []string, sensors *[]model.Sensor, gateway *model.Gateway) {
	if len(options) == 0 {
		fmt.Print("Usage: config --id <gateway-id>\n" +
			"              --password <gateway-password>\n" +
			"              --sensor <mac-address> <setting> <value>\n")
		return
	}
	switch options[0] {
	case "--id":
		if len(args) == 0 {
			fmt.Println("Usage: config --id <gateway-id>")
			return
		}
		err := model.SetGatewayId(args[0], gateway)
		if err != nil {
			out.Error(err)
		}
	case "--password":
		if len(args) == 0 {
			fmt.Println("Usage: config --password <gateway-password>")
			return
		}
		err := model.SetGatewayPassword(args[0], gateway)
		if err != nil {
			out.Error(err)
		}
	case "--sensor":
		if len(args) < 3 {
			fmt.Println("Usage: config --sensor <mac-address> <setting> <value>")
			return
		}
	default:
		fmt.Printf("Option %s does not exist for command config\n", options[0])
	}
}
