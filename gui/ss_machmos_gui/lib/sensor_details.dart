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
          child: SizedBox(
            width: 350,
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
                  name: "Wake-up Interval",
                  value: widget.sensor.wakeUpInterval.toString(),
                  onChanged: (value) {
                    widget.sensor.wakeUpInterval = int.parse(value);
                  },
                  units: "s",
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
                for (String key in widget.sensor.settings.keys)
                  Padding(
                    padding: const EdgeInsets.only(left: 20),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Padding(
                          padding: const EdgeInsets.only(top: 10),
                          child: Text("$key:",
                              style:
                                  const TextStyle(fontWeight: FontWeight.bold)),
                        ),
                        for (String subKey in (widget.sensor.settings
                                .cast<String, dynamic>()[key]!
                                .cast<String, String>())
                            .keys)
                          Padding(
                            padding: const EdgeInsets.only(left: 20),
                            child: SensorDetailField(
                                name: subKey,
                                value: widget.sensor.settings
                                    .cast<String, dynamic>()[key]!
                                    .cast<String, String>()[subKey],
                                onChanged: (value) {
                                  widget.sensor.settings
                                      .cast<String, dynamic>()[key]!
                                      .cast<String, String>()[subKey] = value;
                                }),
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
                    ],
                  ),
                ),
              ],
            ),
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
