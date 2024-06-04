import 'package:flutter/material.dart';
import 'package:ss_machmos_gui/sensors.dart';

class SensorDetails extends StatefulWidget {
  final Sensor sensor;

  const SensorDetails({super.key, required this.sensor});

  @override
  State<SensorDetails> createState() => _SensorDetailsState();
}

class _SensorDetailsState extends State<SensorDetails> {
  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.all(8.0),
          child: Column(
            children: [
              Row(
                children: [
                  const Text("Name: "),
                  Expanded(
                    child: TextField(
                      controller:
                          TextEditingController(text: widget.sensor.name),
                      onChanged: (value) {
                        widget.sensor.name = value;
                      },
                    ),
                  ),
                ],
              ),
              Text("MAC: ${macToString(widget.sensor.mac)}"),
              Text("Types: ${widget.sensor.types.join(", ")}"),
              Row(
                children: [
                  const Text("Wake-up Interval: "),
                  Expanded(
                    child: TextField(
                      controller: TextEditingController(
                          text: widget.sensor.wakeUpInterval.toString()),
                      onChanged: (value) {
                        widget.sensor.wakeUpInterval = int.parse(value);
                      },
                    ),
                  ),
                ],
              ),
              Text(
                  "Battery Level: ${widget.sensor.batteryLevel == -1 ? "Unknown" : widget.sensor.batteryLevel}"),
              const SizedBox(height: 10),
              const Text("Settings: "),
              for (String key in widget.sensor.settings.keys)
                Column(
                  children: [
                    Text("$key:"),
                    for (String subKey in (widget.sensor.settings
                            .cast<String, dynamic>()[key]!
                            .cast<String, String>())
                        .keys)
                      Row(
                        children: [
                          Text("  $subKey: "),
                          Expanded(
                            child: TextField(
                              controller: TextEditingController(
                                  text: widget.sensor.settings
                                      .cast<String, dynamic>()[key]!
                                      .cast<String, String>()[subKey]),
                              onChanged: (value) {
                                widget.sensor.settings
                                    .cast<String, dynamic>()[key]!
                                    .cast<String, String>()[subKey] = value;
                              },
                            ),
                          ),
                        ],
                      ),
                    const SizedBox(height: 10),
                  ],
                ),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  TextButton(
                    onPressed: () {
                      // Delete sensor
                    },
                    child: const Text("Forget"),
                  ),
                  const SizedBox(width: 10),
                  TextButton(
                    onPressed: () {
                      // Save sensor
                    },
                    child: const Text("Save"),
                  ),
                  const SizedBox(width: 10),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
