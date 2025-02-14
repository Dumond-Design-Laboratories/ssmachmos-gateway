import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:ss_machmos_gui/connection.dart';
import 'package:ss_machmos_gui/help.dart';
import 'package:ss_machmos_gui/sensors.dart';
import 'package:ss_machmos_gui/utils.dart';

class SensorDetails extends StatelessWidget {
  const SensorDetails({super.key});

  @override
  Widget build(BuildContext context) {
    Sensor sensor = context.read<Connection>().displayedSensor!;
    List<Widget> settingsWidget = [
      // Autogenerate controls for each setting
      for (String key in sensor.settings.keys)
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
                        const SizedBox(width: 30),
                        const Text("Active:",
                            style: TextStyle(fontWeight: FontWeight.bold)),
                        const SizedBox(width: 10),
                        Checkbox(
                          value: sensor.settings[key]!.active,
                          onChanged: (value) {
                            // Store changes in memory, don't commit yet
                            sensor.settings[key]!.active = value ?? false;
                          },
                        ),
                      ],
                    ),
                    if (key != "temperature")
                      SensorDetailField(
                        name: "Sampling Frequency",
                        value:
                            sensor.settings[key]!.samplingFrequency.toString(),
                        onChanged: (value) {
                          try {
                            sensor.settings[key]!.samplingFrequency =
                                int.parse(value);
                          } catch (_) {}
                        },
                        units: "Hz",
                      ),
                    if (key != "temperature")
                      SensorDetailField(
                        name: "Sampling Duration",
                        value:
                            sensor.settings[key]!.samplingDuration.toString(),
                        onChanged: (value) {
                          try {
                            sensor.settings[key]!.samplingDuration =
                                int.parse(value);
                          } catch (_) {}
                        },
                        units: "seconds",
                      ),
                  ],
                ),
              ),
            ],
          ),
        ),
      // Bottom buttons to save, forget, or reset edits
      Padding(
        padding: const EdgeInsets.symmetric(vertical: 20),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.end,
          children: [
            TextButton(
              onPressed: () {
                // Send forget command
                context.read<Connection>().forgetSensor(sensor, (_, err) {
                  if (err != null) {
                    showMessage(
                        "Failed to forget sensor ${macToString(sensor.mac)}: $err",
                        context);
                  } else {
                    showMessage(
                        "Forgot sensor ${macToString(sensor.mac)}", context);
                  }
                  return true;
                });
              },
              child: const Text("Forget"),
            ),
            const SizedBox(width: 10),
            TextButton(
              onPressed: () {
                showDialog(
                  context: context,
                  builder: (BuildContext context) {
                    return AlertDialog(
                      title: Text(
                        "Reset sensor: ${sensor.name}",
                      ),
                      actions: [
                        // Cancel reset popup
                        TextButton(
                          onPressed: () {
                            Navigator.of(context).pop();
                          },
                          child: const Text("Cancel"),
                        ),
                        TextButton(
                          // Commit reset
                          onPressed: () {
                            context.read<Connection>().resetSensor(sensor,
                                (_, err) {
                              if (err != null) {
                                showMessage(
                                    "Failed to save ${sensor.name} settings",
                                    context);
                              } else {
                                showMessage(
                                    "${sensor.name} settings saved", context);
                                // FIXME when does this run now?
                                //loadSensors();
                              }
                              return true;
                            });
                            Navigator.of(context).pop();
                          },
                          child: const Text("Confirm"),
                        ),
                      ],
                    );
                  },
                );
              },
              child: const Text("Reset"),
            ),
            const SizedBox(width: 10),
            // Save current settings of this sensor
            TextButton(
              onPressed: () {
                context.read<Connection>().saveSensor(sensor, (_, err) {
                  if (err != null) {
                    showMessage(
                        "Failed to save ${sensor.name} settings", context);
                  } else {
                    showMessage("${sensor.name} settings saved", context);
                  }
                  return true;
                });
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
              TextButton(child: Text("Collect now"), onPressed: () => context.read<Connection>().collectFromSensor(sensor)),
              SensorDetailField(
                name: "Name",
                value: sensor.name,
                onChanged: (value) {
                  sensor.name = value;
                },
              ),
              SensorDetailField(
                name: "MAC",
                value: macToString(sensor.mac),
                readOnly: true,
              ),
              SensorDetailField(
                name: "Types",
                value: sensor.types.join(", "),
                readOnly: true,
                // tabController: tabController,
                // page: typesKey,
              ),
              SensorDetailField(
                name: "Battery Level",
                value: sensor.batteryLevel == -1
                    ? "Unknown"
                    : sensor.batteryLevel.toString(),
                readOnly: true,
                units: sensor.batteryLevel == -1 ? "" : "%",
              ),
              SensorDetailField(
                name: "Wake-Up Interval",
                value: sensor.wakeUpInterval.toString(),
                onChanged: (value) {
                  try {
                    sensor.wakeUpInterval = int.parse(value);
                  } catch (_) {}
                },
                units: "seconds",
              ),
              SensorDetailField(
                name: "Wake-Up Interval Max Offset",
                value: sensor.wakeUpIntervalMaxOffset.toString(),
                onChanged: (value) {
                  try {
                    sensor.wakeUpIntervalMaxOffset = int.parse(value);
                  } catch (_) {}
                },
                units: "seconds",
                // tabController: tabController,
                // page: wakeUpIntervalKey,
              ),
              SensorDetailField(
                name: "Next Wake-Up",
                value: sensor.nextWakeUp.toLocal().toString(),
                readOnly: true,
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
  final TabController? tabController;
  final GlobalKey? page;

  const SensorDetailField({
    super.key,
    required this.name,
    required this.value,
    this.onChanged,
    this.readOnly = false,
    this.units = "",
    this.tabController,
    this.page,
  });

  @override
  State<SensorDetailField> createState() => _SensorDetailFieldState();
}

class _SensorDetailFieldState extends State<SensorDetailField> {
  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.symmetric(vertical: widget.readOnly ? 10 : 5),
      child: Row(
        children: [
          if (widget.tabController != null && widget.page != null)
            HelpButton(
                tabController: widget.tabController!, page: widget.page!),
          SizedBox(
              width: (widget.tabController != null && widget.page != null)
                  ? 10
                  : 30),
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
          const SizedBox(width: 10),
          Text(widget.units),
        ],
      ),
    );
  }
}
