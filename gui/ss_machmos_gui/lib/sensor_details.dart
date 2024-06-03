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
        Text("Settings: ${widget.sensor.settings}"),
      ],
    );
  }
}
