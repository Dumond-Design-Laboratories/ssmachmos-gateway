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
    return Column(
      children: [
        Text("Name: ${widget.sensor.name}"),
        Text("MAC: ${widget.sensor.mac}"),
        Text("Types: ${widget.sensor.types.join(", ")}"),
        Text("Wake-up Interval: ${widget.sensor.wakeUpInterval}"),
        Text("Battery Level: ${widget.sensor.batteryLevel}"),
        const Text("Settings: "),
        for (String key in widget.sensor.settings.keys)
          Column(
            children: [
              Text(key),
              for (String subKey in (widget.sensor.settings
                      .cast<String, dynamic>()[key]!
                      .cast<String, String>())
                  .keys)
                Text(
                    "  $subKey: ${widget.sensor.settings.cast<String, dynamic>()[key]!.cast<String, String>()[subKey]}"),
            ],
          ),
      ],
    );
  }
}
