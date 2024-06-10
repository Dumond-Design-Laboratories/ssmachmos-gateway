import 'package:flutter/material.dart';
import 'package:ss_machmos_gui/connection.dart';
import 'package:ss_machmos_gui/sensors.dart';
import 'package:ss_machmos_gui/utils.dart';

class SensorDetails extends StatefulWidget {
  final Sensor sensor;
  final Connection connection;
  final void Function() onForget;

  const SensorDetails({
    super.key,
    required this.sensor,
    required this.connection,
    required this.onForget,
  });

  @override
  State<SensorDetails> createState() => _SensorDetailsState();
}

class _SensorDetailsState extends State<SensorDetails> {
  @override
  Widget build(BuildContext context) {
    var settingsWidget = [
      for (String key in widget.sensor.settings.keys)
        Padding(
          padding: const EdgeInsets.only(left: 20, top: 10),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text("$key:",
                  style: const TextStyle(fontWeight: FontWeight.bold)),
              Padding(
                padding: const EdgeInsets.only(left: 20),
                child: Column(
                  children: [
                    Row(
                      children: [
                        const Text("Active:",
                            style: TextStyle(fontWeight: FontWeight.bold)),
                        const SizedBox(width: 10),
                        Checkbox(
                          value: widget.sensor.settings[key]!.active,
                          onChanged: (value) {
                            setState(() {
                              widget.sensor.settings[key]!.active =
                                  value ?? false;
                            });
                          },
                        ),
                      ],
                    ),
                    SensorDetailField(
                        name: "Wake-Up Interval",
                        value: widget.sensor.settings[key]!.wakeUpInterval
                            .toString(),
                        onChanged: (value) {
                          widget.sensor.settings[key]!.wakeUpInterval =
                              int.parse(value);
                        },
                        units: "seconds"),
                    SensorDetailField(
                      name: "Next Wake-Up",
                      value: widget.sensor.settings[key]!.nextWakeUp
                          .toIso8601String(),
                      readOnly: true,
                    ),
                    if (key != "temperature")
                      SensorDetailField(
                        name: "Sampling Frequency",
                        value: widget.sensor.settings[key]!.samplingFrequency
                            .toString(),
                        onChanged: (value) {
                          widget.sensor.settings[key]!.samplingFrequency =
                              int.parse(value);
                        },
                        units: "Hz",
                      ),
                    if (key != "temperature")
                      SensorDetailField(
                        name: "Sampling Duration",
                        value: widget.sensor.settings[key]!.samplingDuration
                            .toString(),
                        onChanged: (value) {
                          widget.sensor.settings[key]!.samplingDuration =
                              int.parse(value);
                        },
                        units: "seconds",
                      ),
                  ],
                ),
              ),
            ],
          ),
        ),
      Padding(
        padding: const EdgeInsets.symmetric(vertical: 20),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.end,
          children: [
            TextButton(
              onPressed: () async {
                await widget.connection
                    .send("FORGET ${macToString(widget.sensor.mac)}");
                widget.connection.on("FORGET", (_, err) {
                  if (err != null) {
                    showMessage(
                        "Failed to forget sensor ${macToString(widget.sensor.mac)}: $err",
                        context);
                  } else {
                    showMessage(
                        "Forgot sensor ${macToString(widget.sensor.mac)}",
                        context);
                    widget.onForget();
                  }
                  return true;
                });
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
          ],
        ),
      ),
    ];

    return Expanded(
      child: SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 8),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              SensorDetailField(
                name: "Name",
                value: widget.sensor.name,
                onChanged: (value) {
                  widget.sensor.name = value;
                },
              ),
              SensorDetailField(
                name: "MAC",
                value: macToString(widget.sensor.mac),
                readOnly: true,
              ),
              SensorDetailField(
                name: "Types",
                value: widget.sensor.types.join(", "),
                readOnly: true,
              ),
              SensorDetailField(
                name: "Battery Level",
                value: widget.sensor.batteryLevel == -1
                    ? "Unknown"
                    : widget.sensor.batteryLevel.toString(),
                readOnly: true,
                units: widget.sensor.batteryLevel == -1 ? "" : "mV",
              ),
              Container(
                height: 0.5,
                color: Colors.grey,
                margin: const EdgeInsets.only(top: 10, bottom: 20),
              ),
              const Text("Settings:",
                  style: TextStyle(fontWeight: FontWeight.bold)),
              ...settingsWidget,
            ],
          ),
        ),
      ),
    );
  }
}

class SensorDetailField extends StatefulWidget {
  final String name;
  final String value;
  final void Function(String)? onChanged;
  final bool readOnly;
  final String units;

  const SensorDetailField({
    super.key,
    required this.name,
    required this.value,
    this.onChanged,
    this.readOnly = false,
    this.units = "",
  });

  @override
  State<SensorDetailField> createState() => _SensorDetailFieldState();
}

class _SensorDetailFieldState extends State<SensorDetailField> {
  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.symmetric(vertical: widget.readOnly ? 10 : 0),
      child: Row(
        children: [
          Text("${widget.name}:",
              style: const TextStyle(fontWeight: FontWeight.bold)),
          const SizedBox(width: 10),
          if (widget.readOnly)
            Text(widget.value)
          else
            Expanded(
              child: TextField(
                controller: TextEditingController(text: widget.value),
                onChanged: widget.onChanged,
              ),
            ),
          Text(widget.units),
        ],
      ),
    );
  }
}
