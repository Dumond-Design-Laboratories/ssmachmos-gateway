package cli

/*
 * Command line application to communicate with the server process
 * if argument is serve, creates new server
 * if not, sends commands to a currently running server
 */

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"

	"github.com/jukuly/ss_machmos/server/internal/model"
	"github.com/jukuly/ss_machmos/server/internal/out"
)

var messagesToPrint = map[string]string{
	"REQUEST-SENSOR-EXISTS": "Pairing request for already paired sensor. First \"Forget\" the sensor before pairing again.",
	"REQUEST-TIMEOUT":       "Pairing request timed out for sensor ",
	"REQUEST-NEW":           "New pairing request (\"accept <mac-address>\" to accept) from sensor ",
	"PAIR-SUCCESS":          "Pairing successful with sensor ",
	"PAIRING-DISABLED":      "Error: Pairing mode disabled",
	"REQUEST-NOT-FOUND":     "Error: Pairing request not found for sensor ",
	"PAIRING-CANCELED":      "Pairing canceled with sensor ",
	"PAIRING-WITH":          "Pairing with sensor ",
	"PAIRING-TIMEOUT":       "Pairing timed out with sensor ",
}

var waitingFor = map[string]chan<- bool{}

func waitFor(command ...string) {
	done := make(chan bool)
	for _, p := range command {
		waitingFor[p] = done
	}
	<-done
}

func OpenConnection() (net.Conn, error) {
	socketPath := "/tmp/ss_machmos.sock"

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func Listen(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		str, err := reader.ReadString('\x00')
		if err != nil {
			os.Exit(0)
			return
		}
		ress := strings.Split(str, "\x00")
		for _, res := range ress {
			found := []string{}
			if msg := parseResponse(res); msg != "" {
				fmt.Println(msg)
			}
			for prefix, done := range waitingFor {
				if strings.HasPrefix(res, prefix) {
					done <- true
					found = append(found, prefix)
				}
			}
			for _, f := range found {
				delete(waitingFor, f)
			}
		}
	}
}

func sendCommand(command string, conn net.Conn) error {
	_, err := conn.Write([]byte(command + "\x00"))

	if err != nil {
		return err
	}

	return nil
}

func Help(args []string) {
	if len(args) == 0 {
		fmt.Print("+---------+--------------+---------------------------------+------------------------------------+\n" +
			"| Command | Options      | Arguments                       | Description                        |\n" +
			"+---------+--------------+---------------------------------+------------------------------------+\n" +
			"| help    | None         | None                            | View this table                    |\n" +
			"|         |              | <command>                       | View usage and description         |\n" +
			"|         |              |                                 | of a specific command              |\n" +
			"+---------+--------------+---------------------------------+------------------------------------+\n" +
			"| serve   | None         | None                            | Start the server and see the       |\n" +
			"|         |              |                                 | live stream of logs in the console |\n" +
			"|         |              |                                 |                                    |\n" +
			"|         | --no-console | None                            | Start the server without the       |\n" +
			"|         |              |                                 | live stream of logs                |\n" +
			"+---------+--------------+---------------------------------+------------------------------------+\n" +
			"| logs    | None         | None                            | View the live stream of logs       |\n" +
			"+---------+--------------+---------------------------------+------------------------------------+\n" +
			"| stop    | None         | None                            | Stop the server                    |\n" +
			"+---------+--------------+---------------------------------+------------------------------------+\n" +
			"| list    | None         | None                            | List all sensors                   |\n" +
			"+---------+--------------+---------------------------------+------------------------------------+\n" +
			"| view    | --sensor     | <mac-address>                   | View a specific sensors' settings  |\n" +
			"|         | --gateway    | None                            | View the Gateway settings          |\n" +
			"+---------+--------------+---------------------------------+------------------------------------+\n" +
			"| pair    | None         | None                            | Enter pairing mode                 |\n" +
			"+---------+--------------+---------------------------------+------------------------------------+\n" +
			"| forget  | None         | <mac-address>                   | Forget a sensor                    |\n" +
			"+---------+--------------+---------------------------------+------------------------------------+\n" +
			"| config  | --id         | <gateway-id>                    | Set the Gateway Id                 |\n" +
			"|         | --password   | <gateway-password>              | Set the Gateway Password           |\n" +
			"|         | --http       | <http-endpoint>                 | Set the HTTP Endpoint where the    |\n" +
			"|         |              | default                         |   data will be sent                |\n" +
			"|         |              |                                 |   default is openphm.org           |\n" +
			"|         |              |                                 |                                    |\n" +
			"|         | --sensor     | <mac-address> <setting> <value> | Set a setting of a sensor          |\n" +
			"|         |              |                                 |   Type \"help config\"               |\n" +
			"|         |              |                                 |   for more information             |\n" +
			"+---------+--------------+---------------------------------+------------------------------------+\n")
		return
	}

	switch args[0] {
	case "help":
		fmt.Print("+---------+------------+---------------------------------+------------------------------------+\n" +
			"| help    | None       | None                            | View all commands and their usage  |\n" +
			"|         |            | <command>                       | View usage and description         |\n" +
			"|         |            | <command>                       | of a specific command              |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n")

	case "serve":
		fmt.Print("+---------+--------------+---------------------------------+------------------------------------+\n" +
			"| serve   | None         | None                            | Start the server and see the       |\n" +
			"|         |              |                                 | live stream of logs in the console |\n" +
			"|         |              |                                 |                                    |\n" +
			"|         | --no-console | None                            | Start the server without the       |\n" +
			"|         |              |                                 | live stream of logs                |\n" +
			"+---------+--------------+---------------------------------+------------------------------------+\n")

	case "logs":
		fmt.Print("+---------+------------+---------------------------------+------------------------------------+\n" +
			"| logs    | None       | None                            | View the live stream of logs       |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n")

	case "stop":
		fmt.Print("+---------+------------+---------------------------------+------------------------------------+\n" +
			"| stop    | None       | None                            | Stop the server                    |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n")

	case "list":
		fmt.Print("+---------+------------+---------------------------------+------------------------------------+\n" +
			"| list    | None       | None                            | List all sensors                   |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n")

	case "view":
		fmt.Print("+---------+------------+---------------------------------+------------------------------------+\n" +
			"| view    | --sensor   | <mac-address>                   | View a specific sensors' settings  |\n" +
			"|         | --gateway  | None                            | View the Gateway settings          |\n" +
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
			"|         | --http     | <http-endpoint>                 | Set the HTTP Endpoint where the    |\n" +
			"|         |            |                                 | 	data will be sent                |\n" +
			"|         |            |                                 |                                    |\n" +
			"|         | --sensor   | <mac-address> <setting> <value> | Set a setting of a sensor          |\n" +
			"|         |            | <setting> can be \"name\",        |                                    |\n" +
			"|         |            | \"description\" or composed of    |                                    |\n" +
			"|         |            | the measurement type and the    |                                    |\n" +
			"|         |            | setting separated by an \"_\"     |                                    |\n" +
			"|         |            | eg.: \"audio_wake_up_interval\"|                                    |\n" +
			"+---------+------------+---------------------------------+------------------------------------+\n")

	default:
		fmt.Printf("Unknown command: %s\n", args[0])
	}
}

func Logs(conn net.Conn) {
	err := sendCommand("ADD-LOGGER", conn)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	waitFor("OK:ADD-LOGGER")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			if sig == os.Interrupt {
				err := sendCommand("REMOVE-LOGGER", conn)
				if err != nil {
					out.Logger.Println("Error:", err)
					os.Exit(0)
				}
				return
			}
		}
	}()
	waitFor("OK:REMOVE-LOGGER")
}

func List(conn net.Conn) {
	err := sendCommand("LIST", conn)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	waitFor("OK:LIST", "ERR:LIST")
}

func View(options []string, args []string, conn net.Conn) {
	if len(options) == 0 {
		fmt.Print("\nUsage: view --sensor <mac-address>\n" +
			"              --gateway\n")
		return
	}
	switch options[0] {
	case "--sensor":
		if len(args) == 0 {
			fmt.Println("Usage: view --sensor <mac-address>")
			return
		}
		err := sendCommand("VIEW "+args[0], conn)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		waitFor("OK:VIEW", "ERR:VIEW")
	case "--gateway":
		err := sendCommand("GET-GATEWAY", conn)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		waitFor("OK:GET-GATEWAY", "ERR:GET-GATEWAY")
	default:
		fmt.Printf("Option %s does not exist for command view\n", options[0])
	}
}

func Pair(args []string, conn net.Conn) {
	err := sendCommand("PAIR-ENABLE", conn)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	waitFor("OK:PAIR-ENABLE", "ERR:PAIR-ENABLE")
	fmt.Println("Entering pairing mode. Press Ctrl+C to exit pairing mode.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			if sig == os.Interrupt {
				err := sendCommand("PAIR-DISABLE", conn)
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
			err := sendCommand("PAIR-ACCEPT "+parts[1], conn)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
		}
	}
}

func Forget(args []string, conn net.Conn) {
	if len(args) == 0 {
		fmt.Println("Usage: forget <mac-address>")
		return
	}
	err := sendCommand("FORGET "+args[0], conn)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	waitFor("OK:FORGET", "ERR:FORGET")
}

func Config(options []string, args []string, conn net.Conn) {
	if len(options) == 0 {
		fmt.Print("\nUsage: config --id <gateway-id>\n" +
			"              --password <gateway-password>\n" +
			"              --http <http-endpoint> | default\n" +
			"              --sensor <mac-address> <setting> <value>\n")
		return
	}
	switch options[0] {
	case "--id":
		if len(args) == 0 {
			fmt.Println("Usage: config --id <gateway-id>")
			return
		}
		err := sendCommand("SET-GATEWAY-ID "+args[0], conn)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		waitFor("OK:SET-GATEWAY-ID", "ERR:SET-GATEWAY-ID")
	case "--password":
		if len(args) == 0 {
			fmt.Println("Usage: config --password <gateway-password>")
			return
		}
		err := sendCommand("SET-GATEWAY-PASSWORD "+args[0], conn)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		waitFor("OK:SET-GATEWAY-PASSWORD", "ERR:SET-GATEWAY-PASSWORD")
	case "--http":
		if len(args) == 0 {
			fmt.Println("Usage: config --http <http-endpoint> | default")
			return
		}
		err := sendCommand("SET-GATEWAY-HTTP-ENDPOINT "+args[0], conn)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		waitFor("OK:SET-GATEWAY-HTTP-ENDPOINT", "ERR:SET-GATEWAY-HTTP-ENDPOINT")
	case "--sensor":
		if len(args) < 3 {
			fmt.Println("Usage: config --sensor <mac-address> <setting> <value>")
			return
		}
		err := sendCommand("SET-SENSOR-SETTINGS "+args[0]+" "+args[1]+" "+args[2], conn)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		waitFor("OK:SET-SENSOR-SETTINGS", "ERR:SET-SENSOR-SETTINGS")
	default:
		fmt.Printf("Option %s does not exist for command config\n", options[0])
	}
}

func Stop(conn net.Conn) {
	err := sendCommand("STOP", conn)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

func Read(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		// Strip the newline
		sendCommand(command[:len(command)-1], conn)
	}
}

func parseResponse(res string) string {
	parts := strings.Split(res, ":")
	if len(parts) == 0 || len(parts) == 1 {
		return ""
	}
	if parts[0] == "OK" {
		if len(parts) < 3 {
			return strings.Join(parts[1:], ":")
		}
		parts[2] = strings.Join(parts[2:], ":")
		switch parts[1] {
		case "LIST":
			sensors := []model.Sensor{}
			err := json.Unmarshal([]byte(parts[2]), &sensors)
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
		case "VIEW":
			str, err := sensorJSONToString([]byte(parts[2]))
			if err != nil {
				return "Error: " + err.Error()
			}
			return str
		case "GET-GATEWAY":
			gateway := model.Gateway{}
			err := json.Unmarshal([]byte(parts[2]), &gateway)
			if err != nil {
				return "Error: " + err.Error()
			}
			return "Gateway ID: " + gateway.Id + "\nHTTP Endpoint: " + gateway.HTTPEndpoint
		}
	} else if parts[0] == "ERR" {
		if len(parts) < 3 {
			return "Error: " + strings.Join(parts[1:], ":")
		}
		return "Error: " + strings.Join(parts[2:], ":")
	} else if parts[0] == "MSG" {
		for command, msg := range messagesToPrint {
			if parts[1] == command {
				if len(parts) < 3 {
					return msg
				}
				return msg + strings.Join(parts[2:], ":")
			}
		}
	} else if parts[0] == "LOG" {
		line := strings.Join(parts[1:], ":")
		if last := len(line) - 1; last >= 0 && line[last] == '\n' {
			line = line[:last]
		}
		return line
	}
	return ""
}
