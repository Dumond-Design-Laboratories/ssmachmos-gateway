package commands

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

func sendCommand(command string) (string, error) {
	socketPath := "/tmp/ss_mach_mos.sock"

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return "", fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(command))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	var buf [512]byte
	n, err := conn.Read(buf[:])
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(buf[:n]), nil
}

func Help(args []string) {
	if len(args) == 0 {
		fmt.Print("+---------+------------+---------------------------------+------------------------------------+\n" +
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
			"| pair    |            | None                            | Enter pairing mode                 |\n" +
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
		fmt.Print("+---------+------------+---------------------------------+------------------------------------+\n" +
			"| help    | None       | None                            | View all commands and their usage  |\n" +
			"|         |            | <command>                       | View usage and description         |\n" +
			"|         |            | <command>                       | of a specific command              |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n")

	case "list":
		fmt.Print("+---------+------------+---------------------------------+------------------------------------+\n" +
			"| list    | None       | None                            | List all sensors                   |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n")

	case "view":
		fmt.Print("+---------+------------+---------------------------------+------------------------------------+\n" +
			"| view    | None       | <mac-address>                   | View a specific sensors' settings  |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n")

	case "pair":
		fmt.Print("+---------+------------+---------------------------------+------------------------------------+\n" +
			"| pair    | --enable   | None                            | Enter  pairing mode                |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n")

	case "forget":
		fmt.Print("+---------+------------+---------------------------------+------------------------------------+\n" +
			"| forget  | None       | <mac-address>                   | Forget a sensor                    |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n")

	case "config":
		fmt.Print("+---------+------------+---------------------------------+------------------------------------+\n" +
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
		fmt.Printf("Unknown command: %s\n", args[0])
	}
}

func List() {
	res, err := sendCommand("LIST")
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	fmt.Println(parseResponse(res, func(s string) string {
		sensors := []sensor{}
		err := json.Unmarshal([]byte(s), &sensors)
		if err != nil {
			return "Error: " + err.Error()
		}
		if len(sensors) == 0 {
			return "No sensors currently paired with the Gateway"
		} else {
			str := ""
			for _, sensor := range sensors {
				str += sensor.Name + " - " + macToString(sensor.Mac) + "\n"
			}
			return str
		}
	}))
}

func View(args []string) {
	res, err := sendCommand("VIEW " + args[0])
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	fmt.Println(parseResponse(res, func(s string) string {
		str, err := sensorJSONToString([]byte(s))
		if err != nil {
			return "Error: " + err.Error()
		}
		return str
	}))
}

func Pair(args []string) {

}

func Forget(args []string) {
	res, err := sendCommand("FORGET " + args[0])
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	parseResponse(res, nil)
}

func Config(options []string, args []string) {
	if len(options) == 0 {
		fmt.Print("\nUsage: config --id <gateway-id>\n" +
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
		res, err := sendCommand("SET-GATEWAY-ID " + args[0])
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}
		parseResponse(res, nil)
	case "--password":
		if len(args) == 0 {
			fmt.Println("Usage: config --password <gateway-password>")
			return
		}
		res, err := sendCommand("SET-GATEWAY-PASSWORD " + args[0])
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}
		parseResponse(res, nil)
	case "--sensor":
		if len(args) < 3 {
			fmt.Println("Usage: config --sensor <mac-address> <setting> <value>")
			return
		}
		res, err := sendCommand("SET-SENSOR-SETTING " + args[0] + " " + args[1] + " " + args[2])
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}
		parseResponse(res, nil)
	default:
		fmt.Printf("Option %s does not exist for command config\n", options[0])
	}
}

func Stop() {
	_, err := sendCommand("STOP")
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
}

func parseResponse(res string, read func(string) string) string {
	parts := strings.Split(res, ":")
	if len(parts) < 2 {
		return res
	}
	if parts[0] == "OK" {
		if read == nil {
			return "OK"
		}
		return read(parts[1])
	}
	return "Error: " + parts[1]
}
