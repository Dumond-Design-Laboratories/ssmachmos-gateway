import 'package:flutter/material.dart';

class Sensors extends StatefulWidget {
  const Sensors({super.key});

  @override
  State<Sensors> createState() => _SensorsState();
}

class _SensorsState extends State<Sensors> {
  String? _selectedSensor;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        DropdownMenu(
          hintText: "Select Sensor",
          onSelected: (value) {
            setState(() {
              _selectedSensor = value;
            });
          },
          dropdownMenuEntries: const [
            DropdownMenuEntry(value: "AA:BB:CC:DD:EE:FF", label: "Sensor 1"),
            DropdownMenuEntry(value: "00:11:22:33:44:55", label: "Sensor 2"),
            DropdownMenuEntry(value: "CC:DD:AA:00:11:22", label: "Sensor 3"),
          ],
        ),
        const Text("Details"),
      ],
    );
  }
}
