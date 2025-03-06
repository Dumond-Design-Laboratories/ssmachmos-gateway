import 'dart:developer';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import 'package:ss_machmos_gui/connection.dart';
import 'package:ss_machmos_gui/sensor_details.dart';

class Sensors extends StatelessWidget {
  const Sensors({super.key});

  @override
  Widget build(BuildContext context) {
    Connection conn = context.watch<Connection>();
    List<Sensor> sensors = conn.sensors;
    if (sensors.isEmpty) {
      return Column(children: [
        const Padding(
          padding: EdgeInsets.only(top: 100.0),
          child: Text("No sensors currently paired with the Gateway"),
        )
      ]);
    }

    return Column(children: [
      Padding(
        padding: const EdgeInsets.all(8.0),
        child: DropdownMenu(
          hintText: "Select Sensor",
          initialSelection: conn.displayedSensor,
          onSelected: (selectedSensor) {
            // Read provider state
            conn.displayedSensor = selectedSensor as Sensor;
          },
          dropdownMenuEntries: conn.sensors
              .map((s) => DropdownMenuEntry(value: s, label: s.name))
              .toList(),
        ),
      ),
      Container(
        height: 0.5,
        color: Colors.grey,
      ),
      if (conn.displayedSensor != null) SensorDetails()
    ]);
  }
}

class SensorSettings {
  bool active;
  int? samplingFrequency;
  int? samplingDuration;

  SensorSettings({
    required this.active,
    required this.samplingFrequency,
    required this.samplingDuration,
  });
}

// Data class
class Sensor {
  Uint8List mac;
  String name;
  List<String> types;
  int batteryLevel;
  int collectionCapacity;
  int wakeUpInterval;
  int wakeUpIntervalMaxOffset;
  DateTime nextWakeUp;
  bool deviceActive; // Should the sensor start sampling or stay idle
  DateTime lastSeen;
  Map<String, SensorSettings> settings;

  Sensor(
      {required this.mac,
      required this.name,
      required this.types,
      required this.batteryLevel,
      required this.collectionCapacity,
      required this.settings,
      required this.wakeUpInterval,
      required this.wakeUpIntervalMaxOffset,
      required this.deviceActive,
      required this.lastSeen,
      required this.nextWakeUp});

  factory Sensor.fromJson(Map<String, dynamic> s) {
    Map<String, SensorSettings> settings = {};
    for (var k in s["settings"].keys) {
      settings[k] = SensorSettings(
        active: s["settings"][k]["active"],
        samplingFrequency: s["settings"][k]["sampling_frequency"],
        samplingDuration: s["settings"][k]["sampling_duration"],
      );
    }
    return Sensor(
      mac: Uint8List.fromList(s["mac"].cast<int>()),
      name: s["name"],
      types: s["types"] != null
          ? s["types"].cast<String>()
          : [], // Can be null sometimes
      collectionCapacity: s["collection_capacity"],
      wakeUpInterval: s["wake_up_interval"],
      wakeUpIntervalMaxOffset: s["wake_up_interval_max_offset"],
      nextWakeUp: DateTime.parse(s["next_wake_up"]),
      batteryLevel: s["battery_level"],
      deviceActive: s["device_active"] as bool,
      lastSeen: DateTime.parse(s["last_seen"]), // KEEP THIS AS-IS, stored as
                                                // local time
      settings: settings,
    );
  }

  String get sensorSettingsCommand {
    var subSettings = settings.entries
        .map<String>((e) => "${e.key}_active ${e.value.active}"
            " ${e.key}_sampling_frequency ${e.value.samplingFrequency}"
            " ${e.key}_sampling_duration ${e.value.samplingDuration}")
        .join(" ");

    // Convert name spaces to underscores
    return "${macToString(mac)}"
        " name ${name.replaceAll(' ', '_')}"
        " device_active $deviceActive"
        " wake_up_interval $wakeUpInterval"
        " wake_up_interval_max_offset $wakeUpIntervalMaxOffset"
        " $subSettings";
  }

  String get predictedWakeupTime {
    try {
      DateTime next = lastSeen.toLocal().add(Duration(seconds: wakeUpInterval));
      return DateFormat('yyyy-MM-dd HH:mm:ss').format(next);
    } catch (e) {
      log(e.toString());
      return "Unknown";
    }
  }
}

String macToString(Uint8List mac) {
  return mac.map((b) => b.toRadixString(16).padLeft(2, "0")).join(":");
}
