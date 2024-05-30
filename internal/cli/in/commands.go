package in

import (
	"github.com/jukuly/ss_mach_mo/internal/cli/out"
	"github.com/jukuly/ss_mach_mo/internal/model"
	"github.com/jukuly/ss_mach_mo/internal/model/server"
)

func help(args []string) {
	if len(args) == 0 {
		out.Logger.Print("\n+---------+------------+---------------------------------+------------------------------------+\n" +
			"| Command | Options    | Arguments                       | Description                        |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n" +
			"| help    | None       | None                            | View this table                    |\n" +
			"|         |            | <command>                       | View usage and description         |\n" +
			"|         |            |                                 | of a specific command              |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n" +
			"| list    | None       | None                            | List all sensors                   |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n" +
			"| view    | None       | <mac-address>                   | View a specific sensors' settings  |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n" +
			"| pair    | --enable   | None                            | Enable pairing mode                |\n" +
			"|         | --disable  | None                            | Disable pairing mode               |\n" +
			"|         | --accept   | <mac-address>                   | Accept a pairing request           |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n" +
			"| forget  | None       | <mac-address>                   | Forget a sensor                    |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n" +
			"| config  | --id       | <gateway-id>                    | Set the Gateway Id                 |\n" +
			"|         | --password | <gateway-password>              | Set the Gateway Password           |\n" +
			"|         | --sensor   | <mac-address> <setting> <value> | Set a setting of a sensor          |\n" +
			"|         |            |                                 |   Type \"help config\"               |\n" +
			"|         |            |                                 |   for more information             |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n")
		return
	}

	switch args[0] {
	case "help":
		out.Logger.Print("\n+---------+------------+---------------------------------+------------------------------------+\n" +
			"| help    | None       | None                            | View all commands and their usage  |\n" +
			"|         |            | <command>                       | View usage and description         |\n" +
			"|         |            | <command>                       | of a specific command              |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n")

	case "list":
		out.Logger.Print("\n+---------+------------+---------------------------------+------------------------------------+\n" +
			"| list    | None       | None                            | List all sensors                   |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n")

	case "view":
		out.Logger.Print("\n+---------+------------+---------------------------------+------------------------------------+\n" +
			"| view    | None       | <mac-address>                   | View a specific sensors' settings  |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n")

	case "pair":
		out.Logger.Print("\n+---------+------------+---------------------------------+------------------------------------+\n" +
			"| pair    | --enable   | None                            | Enable pairing mode                |\n" +
			"|         | --disable  | None                            | Disable pairing mode               |\n" +
			"|         | --accept   | <mac-address>                   | Accept a pairing request           |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n")

	case "forget":
		out.Logger.Print("\n+---------+------------+---------------------------------+------------------------------------+\n" +
			"| forget  | None       | <mac-address>                   | Forget a sensor                    |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n")

	case "config":
		out.Logger.Print("\n+---------+------------+---------------------------------+------------------------------------+\n" +
			"| config  | --id       | <gateway-id>                    | Set the Gateway Id                 |\n" +
			"|         | --password | <gateway-password>              | Set the Gateway Password           |\n" +
			"|         |            |                                 |                                    |\n" +
			"|         | --sensor   | <mac-address> <setting> <value> | Set a setting of a sensor          |\n" +
			"|         |            | <setting> can be \"name\",        |                                    |\n" +
			"|         |            | \"description\" or composed of    |                                    |\n" +
			"|         |            | the measurement type and the    |                                    |\n" +
			"|         |            | setting separated by an \"_\"     |                                    |\n" +
			"|         |            | eg.: \"acoustic_next_wake_up\"    |                                    |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n")

	default:
		out.Logger.Printf("Unknown command: %s\n", args[0])
	}
}

func list(sensors *[]model.Sensor) {
	out.DisplaySensors(*sensors)
}

func view(args []string, sensors *[]model.Sensor) {
	if len(args) == 0 {
		out.Logger.Println("Usage: view <mac-address>")
		return
	}
	for _, sensor := range *sensors {
		if sensor.IsMacEqual(args[0]) {
			out.DisplaySensor(sensor)
			return
		}
	}
	out.Logger.Printf("Sensor with MAC address %s not found\n", args[0])
}

func pair(options []string, args []string, gateway *model.Gateway) {
	if len(options) == 0 {
		out.Logger.Print("\nUsage: pair --enable\n" +
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
			out.Logger.Println("Usage: pair --accept <mac-address>")
			return
		}
		mac, _ := model.StringToMac(args[0])
		server.Pair(mac, gateway)
	default:
		out.Logger.Printf("Option %s does not exist for command pair\n", options[0])
	}
}

func forget(args []string, sensors *[]model.Sensor) {
	if len(args) == 0 {
		out.Logger.Println("Usage: forget <mac-address>")
		return
	}
	mac, err := model.StringToMac(args[0])
	if err != nil {
		out.Error(err)
		return
	}
	err = model.RemoveSensor(mac, sensors)
	if err != nil {
		out.Error(err)
		return
	}
}

func config(options []string, args []string, sensors *[]model.Sensor, gateway *model.Gateway) {
	if len(options) == 0 {
		out.Logger.Print("\nUsage: config --id <gateway-id>\n" +
			"              --password <gateway-password>\n" +
			"              --sensor <mac-address> <setting> <value>\n")
		return
	}
	switch options[0] {
	case "--id":
		if len(args) == 0 {
			out.Logger.Println("Usage: config --id <gateway-id>")
			return
		}
		err := model.SetGatewayId(gateway, args[0])
		if err != nil {
			out.Error(err)
		}
	case "--password":
		if len(args) == 0 {
			out.Logger.Println("Usage: config --password <gateway-password>")
			return
		}
		err := model.SetGatewayPassword(gateway, args[0])
		if err != nil {
			out.Error(err)
		}
	case "--sensor":
		if len(args) < 3 {
			out.Logger.Println("Usage: config --sensor <mac-address> <setting> <value>")
			return
		}
	default:
		out.Logger.Printf("Option %s does not exist for command config\n", options[0])
	}
}
