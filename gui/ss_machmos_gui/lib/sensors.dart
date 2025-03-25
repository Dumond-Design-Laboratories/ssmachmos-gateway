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
          padding: const EdgeInsets.all(8),
          child:
              Row(mainAxisAlignment: MainAxisAlignment.spaceEvenly, children: [
            DropdownMenu(
              label: Text("Sensor selection:"),
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
            Text(conn.displayedSensor?.address ?? ""),
            Text(conn.displayedSensor?.model.string ?? "")
          ])),
      // Container(
      //   height: 0.5,
      //   color: Colors.grey,
      // ),
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

enum SensorModel {
  machmo,
  machmomini,
  unknown;

  static SensorModel fromString(String str) {
    if (str == "machmo") return SensorModel.machmo;
    if (str == "machmomini") return SensorModel.machmomini;
    return SensorModel.unknown;
  }

  String get string {
    switch (this) {
      case SensorModel.machmo:
        return "MachMo";
      case SensorModel.machmomini:
        return "MachMo-Mini";
      default:
        return "Unknown";
    }
  }
}

// Data class
class Sensor {
  Uint8List mac;
  String get address => macToString(mac);
  String name;
  SensorModel model;
  List<String> types;
  int batteryLevel;
  int collectionCapacity;
  int wakeUpInterval;
  int wakeUpIntervalMaxOffset;
  DateTime nextWakeUp;
  bool deviceActive; // Should the sensor start sampling or stay idle
  //DateTime? lastSeen;
  SensorStatus? status;
  Map<String, SensorSettings> settings;

  Sensor(
      {required this.mac,
      required this.name,
      required this.model,
      required this.types,
      required this.batteryLevel,
      required this.collectionCapacity,
      required this.settings,
      required this.wakeUpInterval,
      required this.wakeUpIntervalMaxOffset,
      required this.deviceActive,
      this.status,
      // required this.lastSeen,
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
      model: SensorModel.fromString(s["model"] ?? ""),
      types: s["types"] != null
          ? s["types"].cast<String>()
          : [], // Can be null sometimes
      collectionCapacity: s["collection_capacity"],
      wakeUpInterval: s["wake_up_interval"],
      wakeUpIntervalMaxOffset: s["wake_up_interval_max_offset"],
      nextWakeUp: DateTime.parse(s["next_wake_up"]),
      batteryLevel: s["battery_level"],
      deviceActive: s["device_active"] as bool,
      //lastSeen: DateTime.parse(s["last_seen"]), // KEEP THIS AS-IS, stored as local time
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
      if (status != null) {
        DateTime next =
            status!.lastSeen.toLocal().add(Duration(seconds: wakeUpInterval));
        return DateFormat('yyyy-MM-dd HH:mm:ss').format(next);
      } else {
        return "Never seen before!";
      }
    } catch (e) {
      log(e.toString());
      return "Unknown";
    }
  }

  List<int> samplingFreqsForSetting(String setting) {
    if (!settings.containsKey(setting)) {
      log("Setting $setting no sampling frequencies returned");
      return [];
    }
    switch (setting) {
      case "vibration":
        return accelerometerSamplingFrequencies;
      case "flux":
        return fluxSamplingFrequencies;
      case "audio":
        return audioSamplingFrequencies;
    }
    log("Setting $setting has no sampling frequencies");
    return [];
  }

  // FIXME: this should be backend maybe?
  List<int> get accelerometerSamplingFrequencies {
    if (model == SensorModel.machmo) {
      // KX132 ODR values
      return [
        25,
        50,
        100,
        200,
        400,
        800,
        1600,
        3200,
        6400,
        12800,
        25600,
      ];
    }

    if (model == SensorModel.machmomini) {
      // FIXME: some are non-existent
      // IIM42352 ODR Values
      return [
        25,
        50,
        100,
        125,
        200,
        500,
        625,
        1000,
        2000,
        3125,
        4000,
        8000,
        15625,
        16000,
        32000,
      ];
    }
    return [];
  }

  List<int> get fluxSamplingFrequencies {
    if (model == SensorModel.machmo) {
      // ADS1120 output data rate values
      return [
        20,
        45,
        90,
        175,
        330,
        600,
        1000,
      ];
    }

    return [];
  }

  List<int> get audioSamplingFrequencies {
    if (model == SensorModel.machmo) {
      // FIXME: What data rates are we looking for?
      return [500, 2200, 8000];
    }
    return [];
  }
}

String macToString(Uint8List mac) {
  return mac
      .map((b) => b.toRadixString(16).padLeft(2, "0"))
      .join(":")
      .toUpperCase();
}
