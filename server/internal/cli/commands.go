package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"

	"github.com/jukuly/ss_mach_mo/server/internal/model"
)

func OpenConnection() (net.Conn, error) {
	socketPath := "/tmp/ss_mach_mos.sock"

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	return conn, nil
}

func Listen(conn net.Conn) {
	for {
		var buf [512]byte
		n, err := conn.Read(buf[:])
		if err != nil {
			return
		}
		if msg := parseResponse(string(buf[:n]), nil); msg != "" {
			fmt.Println(msg)
		}
	}
}

// response should always be false if Listen is running at the same time
func sendCommand(command string, conn net.Conn, response bool) (string, error) {
	_, err := conn.Write([]byte(command))

	if err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	if !response {
		return "", nil
	}

	var buf [512]byte
	n, err := conn.Read(buf[:])

	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(buf[:n]), nil
}

func Help(args []string, conn net.Conn) {
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

func List(conn net.Conn) {
	res, err := sendCommand("LIST", conn, true)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(parseResponse(res, func(s string) string {
		sensors := []model.Sensor{}
		err := json.Unmarshal([]byte(s), &sensors)
		if err != nil {
			return "Error: " + err.Error()
		}
		if len(sensors) == 0 {
			return "No sensors currently paired with the Gateway"
		} else {
			str := ""
			for _, sensor := range sensors {
				str += sensor.Name + " - " + model.MacToString(sensor.Mac) + "\n"
			}
			return str
		}
	}))
}

func View(args []string, conn net.Conn) {
	res, err := sendCommand("VIEW "+args[0], conn, true)
	if err != nil {
		fmt.Println("Error:", err)
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

func Pair(args []string, conn net.Conn) {
	res, err := sendCommand("PAIR-ENABLE", conn, true)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(parseResponse(res, func(s string) string {
		switch {
		case strings.HasPrefix(s, "REQUEST-TIMEOUT-"):
			return "Pair request from " + strings.Replace(s, "REQUEST-TIMEOUT-", "", 1) + " has timed out"
		case strings.HasPrefix(s, "REQUEST-NEW-"):
			return "Pair request from " + strings.Replace(s, "REQUEST-NEW-", "", 1) + " | accept <mac-address> to accept"
		case strings.HasPrefix(s, "PAIR-SUCCESS-"):
			return strings.Replace(s, "PAIR-SUCCESS-", "", 1) + " has been paired with the Gateway"
		case strings.HasPrefix(s, "PAIR-TIMEOUT-"):
			return "Pairing with " + strings.Replace(s, "PAIR-TIMEOUT-", "", 1) + " has timed out"
		default:
			return s
		}
	}))
	fmt.Println("Entering pairing mode. Press Ctrl+C to exit pairing mode.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go Listen(conn)
	go func() {
		for sig := range c {
			if sig == os.Interrupt {
				_, err := sendCommand("PAIR-DISABLE", conn, false)
				if err != nil {
					fmt.Println("Error:", err)
					os.Exit(0)
					return
				}
				fmt.Println("Exiting pairing mode")
				os.Exit(0)
				return
			}
		}
	}()

	for {
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error:", err)
		}
		text = strings.TrimSpace(text)
		if strings.HasPrefix(text, "accept") {
			parts := strings.Split(text, " ")
			if len(parts) < 2 {
				fmt.Println("Usage: accept <mac-address>")
				continue
			}
			_, err := sendCommand("PAIR-ACCEPT "+parts[1], conn, false)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
		}
	}
}

func Forget(args []string, conn net.Conn) {
	res, err := sendCommand("FORGET "+args[0], conn, true)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	parseResponse(res, nil)
}

func Config(options []string, args []string, conn net.Conn) {
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
		res, err := sendCommand("SET-GATEWAY-ID "+args[0], conn, true)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		parseResponse(res, nil)
	case "--password":
		if len(args) == 0 {
			fmt.Println("Usage: config --password <gateway-password>")
			return
		}
		res, err := sendCommand("SET-GATEWAY-PASSWORD "+args[0], conn, true)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		parseResponse(res, nil)
	case "--sensor":
		if len(args) < 3 {
			fmt.Println("Usage: config --sensor <mac-address> <setting> <value>")
			return
		}
		res, err := sendCommand("SET-SENSOR-SETTING "+args[0]+" "+args[1]+" "+args[2], conn, true)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		parseResponse(res, nil)
	default:
		fmt.Printf("Option %s does not exist for command config\n", options[0])
	}
}

func Stop(conn net.Conn) {
	_, err := sendCommand("STOP", conn, false)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

func parseResponse(res string, read func(string) string) string {
	parts := strings.Split(res, ":")
	if len(parts) == 0 || len(parts) == 1 {
		return ""
	}
	if parts[0] == "OK" {
		if read == nil {
			return parts[1]
		}
		return read(parts[1])
	} else if parts[0] == "ERR" {
		return "Error: " + strings.Join(parts[1:], ":")
	} else if parts[0] == "MSG" {
		return strings.Join(parts[1:], ":")
	}
	return ""
}
