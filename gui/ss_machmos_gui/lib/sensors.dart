import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:ss_machmos_gui/connection.dart';
import 'package:ss_machmos_gui/sensor_details.dart';
import 'package:ss_machmos_gui/utils.dart';

class Sensors extends StatelessWidget {
  //final List<Sensor> sensors;

  const Sensors({
    super.key,
    //required this.sensors,
    //required this.loadSensors,
    //required this.connection,
    //required this.tabController,
    //required this.typesKey,
    //required this.wakeUpIntervalKey,
  });

  @override
  Widget build(BuildContext context) {
    List<Sensor> sensors = context.watch<Connection>().sensors;
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
          initialSelection: context.read<Connection>().displayedSensor,
          onSelected: (selectedSensor) {
            // Read provider state
            context.read<Connection>().displayedSensor =
                selectedSensor as Sensor;
          },
          dropdownMenuEntries: context
              .read<Connection>()
              .sensors
              .map((s) => DropdownMenuEntry(value: s, label: s.name))
              .toList(),
        ),
      ),
      Container(
        height: 0.5,
        color: Colors.grey,
      ),
      if (context.read<Connection>().displayedSensor != null)
        SensorDetails(
          //sensor: context.read<Connection>().displayedSensor!,
          // onForget: () {
          //   setState(() {
          //     widget.sensors.remove(_selectedSensor);
          //     _selectedSensor = null;
          //   });
          // },
          // loadSensors: () async {
          //   await widget.loadSensors();
          //   widget.connection.on("VIEW", (json, err) {
          //     if (err != null) {
          //       showMessage("Failed to load sensors", context);
          //       return true;
          //     }
          //     try {
          //       dynamic s = jsonDecode(json);
          //       Map<String, SensorSettings> settings = {};
          //       for (var k in s["settings"].keys) {
          //         settings[k] = SensorSettings(
          //           active: s["settings"][k]["active"],
          //           samplingFrequency: s["settings"][k]["sampling_frequency"],
          //           samplingDuration: s["settings"][k]["sampling_duration"],
          //         );
          //       }
          //       setState(() {
          //         _selectedSensor = Sensor(
          //           mac: Uint8List.fromList(s["mac"].cast<int>()),
          //           name: s["name"],
          //           types: s["types"].cast<String>(),
          //           collectionCapacity: s["collection_capacity"],
          //           wakeUpInterval: s["wake_up_interval"],
          //           wakeUpIntervalMaxOffset: s["wake_up_interval_max_offset"],
          //           nextWakeUp: DateTime.parse(s["next_wake_up"]),
          //           batteryLevel: s["battery_level"],
          //           settings: settings,
          //         );
          //       });
          //       return true;
          //     } catch (e) {
          //       showMessage("Failed to load sensors: $e", context);
          //       return true;
          //     }
          //   });
          //   await widget.connection
          //       .send("VIEW ${macToString(_selectedSensor!.mac)}");
          // },
        )
    ]);
  }

  // @override
  // State<Sensors> createState() => _SensorsState();
}

/*
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
          const Padding(
            padding: EdgeInsets.only(top: 100.0),
            child: Text("No sensors currently paired with the Gateway"),
          ),
        if (widget.sensors.isNotEmpty)
          Padding(
            padding: const EdgeInsets.all(8.0),
            child: DropdownMenu(
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
          ),
        if (widget.sensors.isNotEmpty)
          Container(
            height: 0.5,
            color: Colors.grey,
          ),
        if (_selectedSensor != null)
          SensorDetails(
            sensor: _selectedSensor!,
            connection: widget.connection,
            onForget: () {
              setState(() {
                widget.sensors.remove(_selectedSensor);
                _selectedSensor = null;
              });
            },
            loadSensors: () async {
              await widget.loadSensors();
              widget.connection.on("VIEW", (json, err) {
                if (err != null) {
                  showMessage("Failed to load sensors", context);
                  return true;
                }
                try {
                  dynamic s = jsonDecode(json);
                  Map<String, SensorSettings> settings = {};
                  for (var k in s["settings"].keys) {
                    settings[k] = SensorSettings(
                      active: s["settings"][k]["active"],
                      samplingFrequency: s["settings"][k]["sampling_frequency"],
                      samplingDuration: s["settings"][k]["sampling_duration"],
                    );
                  }
                  setState(() {
                    _selectedSensor = Sensor(
                      mac: Uint8List.fromList(s["mac"].cast<int>()),
                      name: s["name"],
                      types: s["types"].cast<String>(),
                      collectionCapacity: s["collection_capacity"],
                      wakeUpInterval: s["wake_up_interval"],
                      wakeUpIntervalMaxOffset: s["wake_up_interval_max_offset"],
                      nextWakeUp: DateTime.parse(s["next_wake_up"]),
                      batteryLevel: s["battery_level"],
                      settings: settings,
                    );
                  });
                  return true;
                } catch (e) {
                  showMessage("Failed to load sensors: $e", context);
                  return true;
                }
              });
              await widget.connection
                  .send("VIEW ${macToString(_selectedSensor!.mac)}");
            },
            setState: setState,
            //tabController: widget.tabController,
            //typesKey: widget.typesKey,
            //wakeUpIntervalKey: widget.wakeUpIntervalKey,
          ),
      ],
    );
  }
}
*/

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

class Sensor {
  Uint8List mac;
  String name;
  List<String> types;
  int batteryLevel;
  int collectionCapacity;
  int wakeUpInterval;
  int wakeUpIntervalMaxOffset;
  DateTime nextWakeUp;
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
      types: s["types"] != null ? s["types"].cast<String>() : [], // Can be null sometimes
      collectionCapacity: s["collection_capacity"],
      wakeUpInterval: s["wake_up_interval"],
      wakeUpIntervalMaxOffset: s["wake_up_interval_max_offset"],
      nextWakeUp: DateTime.parse(s["next_wake_up"]),
      batteryLevel: s["battery_level"],
      settings: settings,
    );
  }
}

String macToString(Uint8List mac) {
  return mac.map((b) => b.toRadixString(16).padLeft(2, "0")).join(":");
}
