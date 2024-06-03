import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:ss_machmos_gui/sensor_details.dart';

class Sensors extends StatefulWidget {
  final List<Sensor> sensors;
  final Future<void> Function() loadSensors;

  const Sensors({super.key, required this.sensors, required this.loadSensors});

  @override
  State<Sensors> createState() => _SensorsState();
}

class _SensorsState extends State<Sensors> {
  Sensor? _selectedSensor;

  @override
  void initState() {
    super.initState();
    widget.loadSensors();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        if (widget.sensors.isEmpty)
          const Text("No sensors currently paired with the Gateway"),
        if (widget.sensors.isNotEmpty)
          DropdownMenu(
            hintText: "Select Sensor",
            onSelected: (value) {
              setState(() {
                _selectedSensor = value;
              });
            },
            dropdownMenuEntries: widget.sensors
                .map((s) => DropdownMenuEntry(value: s, label: s.name))
                .toList(),
          ),
        if (_selectedSensor != null) SensorDetails(sensor: _selectedSensor!),
      ],
    );
  }
}

class Sensor {
  final Uint8List mac;
  final String name;
  final List<String> types;
  final int wakeUpInterval;
  final int batteryLevel;
  final Map<String, Map<String, String>> settings;

  Sensor({
    required this.mac,
    required this.name,
    required this.types,
    required this.wakeUpInterval,
    required this.batteryLevel,
    required this.settings,
  });

  factory Sensor.fromJson(Map<String, dynamic> json) {
    return Sensor(
      mac: Uint8List.fromList(json["mac"]),
      name: json["name"],
      types: List<String>.from(json["types"]),
      wakeUpInterval: json["wakeUpInterval"],
      batteryLevel: json["battery_level"],
      settings: Map<String, Map<String, String>>.from(json["settings"]),
    );
  }
}
