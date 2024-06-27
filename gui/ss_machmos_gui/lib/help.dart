import 'package:flutter/material.dart';

class Help extends StatelessWidget {
  const Help({super.key});

  @override
  Widget build(BuildContext context) {
    return const SingleChildScrollView(
      child: Padding(
        padding: EdgeInsets.all(32),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            H1(
              "Overview",
              first: true,
            ),
            Text(
              "This application serves as the gateway for SS MachMoS. It allows the sensors to transmit their information to a cloud platform like openMachMoS. "
              "The app is made up of three parts: a server that runs in the background, a command line interface, and a graphical user interface.\n\n"
              "The server is a separate process that runs in the background independently from both interfaces. It always listens for sensors sending data, and will relay it "
              "to the cloud platform depending on the gateway settings. To Start the server, the user can either use this command in the terminal: \"ssmachmos serve\" "
              "or start the graphical user interface and press \"Start Server\". As it it independent from the interfaces, closing an interface won't close the server. To close the server, "
              "the user can either use this command in the terminal: \"ssmachmos stop\" or press the stop button in the \"Logs\" tab of the graphical user interface.\n\n"
              "The command line interface allows the user to monitor the sensors and to change the settings of the server. For more information on the commands available, the user can "
              "use this command in the terminal: \"ssmachmos help\"\n\n"
              "The graphical user interface allows the user to do everything he can do with the command line interface, only it is easier to understand and to use. "
              "There are four tabs at the top of the window: Sensors, Gateway, Logs, and Help. The \"Sensors\" tab allows the user to add/modify/monitor/forget sensors, "
              "the \"Gateway\" tab allows the user to change the gateway settings, and the \"Logs\" tab allows the user to view the logs of the server and to monitor its activity.",
            ),
            H1(
              "Sensors Tab",
            ),
            Text(
              "The sensors are paired like described in the \"Pairing\" section. A sensor's settings can be ajusted by selecting it in the dropdown menu to the left of the page. "
              "This will display the different sensor properties described in the \"Sensor Properties\" section",
            ),
            H2(
              "Pairing",
            ),
            Text(
              "The sensors connect to the gateway via Bluetooth. For a sensor to be able to comunicate with the gateway, it needs to be paired. "
              "To pair a sensor, the user should first turn pairing on by clicking on the toggle switch next to \"Discover Sensors\" "
              "He will then need to click on the button on the physical sensor for it to start looking for nearby gateways. When the MAC address of the sensor "
              "appears in the list under \"Discover Sensors\", the user can click on it to initiate pairing with the sensor. Once paired, the sensor will disappear from "
              "this list and appear in the dropdown to the left of the page.",
            ),
            H2(
              "Sensor Properties",
            ),
            Padding(
              padding: EdgeInsets.only(left: 16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  H3(
                    "Name",
                    first: true,
                  ),
                  Text(
                    "A name to be given to a sensor. Allows the user to more easily identify a sensor.",
                  ),
                  H3(
                    "MAC",
                  ),
                  Text(
                    "The MAC address of the sensor. Used as the id on the cloud platform and on the gateway.",
                  ),
                  H3(
                    "Types",
                  ),
                  Text(
                    "The different types of measurement the sensor is capable of. They are transmitted by the sensor when initially pairing.",
                  ),
                  H3(
                    "Battery Level",
                  ),
                  Text(
                    "The battery level in mV of the sensor. It is transmitted by the sensor every time a measurement is taken.",
                  ),
                  H3(
                    "Wake-Up Interval",
                  ),
                  Text(
                    "The duration between measurements (seconds). The server will synchronize the sensors to try to make them wake-up not at the same time. "
                    "To control this, the user can set the \"Wake-Up Interval Max Offset\" field to specify what is the maximum deviation from the next expected wake-up that "
                    "is acceptable for this sensor. If there is no available time interval between "
                    "(Wake-Up Interval - Wake-Up Interval Max Offset) and (Wake-Up Interval + Wake-Up Interval Max Offset), the sensor will wake-up exactly the wake-up interval duration "
                    "after its last wake-up. The Wake-Up Interval field must be greater than the Wake-Up Interval Max Offest field. Finally, the next wake-up time is displayed in the "
                    "Next Wake-Up field in local time.",
                  ),
                  H3(
                    "Measurement Specific Settings",
                  ),
                  Padding(
                    padding: EdgeInsets.only(left: 16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        H4(
                          "Active",
                          first: true,
                        ),
                        Text(
                          "When on, the sensor will record this type of measurement.",
                        ),
                        H4(
                          "Sampling Frequency",
                        ),
                        Text(
                          "The sampling frequency of the measurements taken (samples per seconds). Not applicable for temperature measurements.",
                        ),
                        H4(
                          "Sampling Duration",
                        ),
                        Text(
                          "The duration of the measurements taken (seconds). Not applicable for temperature measurements.",
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            H1(
              "Gateway Tab",
            ),
            H2(
              "Id and Password",
              first: true,
            ),
            Text(
              "This is the id and password that will be transmitted to the server when sending data. For the openMachMoS cloud platform, they correspond to the id and password "
              "set when creating the gateway on https://openphm.org",
            ),
            H2(
              "HTTP Endpoint",
            ),
            Text(
              "This is the endpoint to which the data will be sent. By default, it is set to send to the openMachMoS cloud platform. "
              "If, when a sensor sends its data to the server, the server is not connected to the internet or any error occurs in the data transmission to the cloud, "
              "the unsent data will be temporarily saved and the server will try to send it to the cloud the next time data is collected.\n\n"
              "This is the format of the HTTP POST request sent to this endpoint in JSON:\n\n"
              "{\n"
              "    \"gateway_id\": <gateway-id>\n"
              "    \"gateway_passord\": <gateway-password>\n"
              "    \"measurements\": [ (can contain multiple measurements)\n"
              "        {\n"
              "            \"sensor_id\": <sensor-mac-address>\n"
              "            \"time\": <timestamp-formated-like-2006-01-02T15:04:05.999Z> (ISO8601)\n"
              "            \"measurement_type\": \"vibration\" or \"temperature\" or \"audio\" or \"battery\"\n"
              "            \"sampling_frequency\": <sampling-frequency-in-Hz> (only for audio and vibration)\n"
              "            \"axis\": \"x\" or \"y\" or \"z\" (only for vibration)\n"
              "            \"raw_data\": [] (array of numbers)\n"
              "        }\n"
              "    ]\n"
              "}",
            ),
            H1(
              "Logs Tab",
            ),
            Text(
              "This tab allows the user to monitor server activity. In particular, the logs will display any error (important or not) that will occur during the process. "
              "They will also display the interactions between the sensors and the server. For example, when a sensor sends data to the server and when pairing is taking place.",
            ),
          ],
        ),
      ),
    );
  }
}

class H1 extends StatelessWidget {
  final String text;
  final bool first;

  const H1(this.text, {super.key, this.first = false});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(top: first ? 0 : 32, bottom: 16),
      child: Text(
        text,
        style: const TextStyle(
          fontSize: 24,
          fontWeight: FontWeight.bold,
        ),
      ),
    );
  }
}

class H2 extends StatelessWidget {
  final String text;
  final bool first;

  const H2(this.text, {super.key, this.first = false});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(top: first ? 0 : 16, bottom: 8),
      child: Text(
        text,
        style: const TextStyle(
          fontSize: 20,
          fontWeight: FontWeight.bold,
        ),
      ),
    );
  }
}

class H3 extends StatelessWidget {
  final String text;
  final bool first;

  const H3(this.text, {super.key, this.first = false});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(top: first ? 0 : 8, bottom: 4),
      child: Text(
        text,
        style: const TextStyle(
          fontSize: 16,
          fontWeight: FontWeight.bold,
        ),
      ),
    );
  }
}

class H4 extends StatelessWidget {
  final String text;
  final bool first;

  const H4(this.text, {super.key, this.first = false});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(top: first ? 0 : 4, bottom: 2),
      child: Text(
        text,
        style: const TextStyle(
          fontSize: 14,
          fontWeight: FontWeight.bold,
        ),
      ),
    );
  }
}
