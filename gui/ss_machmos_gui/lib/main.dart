import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:ss_machmos_gui/bluetooth.dart';
import 'package:ss_machmos_gui/connection.dart';
import 'package:ss_machmos_gui/gateway.dart';
import 'package:ss_machmos_gui/logs.dart';
import 'package:ss_machmos_gui/sensors.dart';
import 'package:ss_machmos_gui/utils.dart';

void main() {
  runApp(const MainApp());
}

class MainApp extends StatelessWidget {
  const MainApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        fontFamily: "OpenSans",
        textTheme: const TextTheme(
          bodyMedium: TextStyle(fontSize: 14),
        ),
        dropdownMenuTheme: const DropdownMenuThemeData(
          textStyle: TextStyle(fontSize: 14),
        ),
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF326496),
        ),
      ),
      home: const DefaultTabController(
        length: 3,
        child: Scaffold(
          appBar: TabBar(
            isScrollable: true,
            labelPadding: EdgeInsets.symmetric(horizontal: 25),
            tabAlignment: TabAlignment.start,
            tabs: [
              Tab(text: "Sensors"),
              Tab(text: "Gateway"),
              Tab(text: "Logs"),
            ],
          ),
          body: Root(),
        ),
      ),
    );
  }
}

class Root extends StatefulWidget {
  const Root({
    super.key,
  });

  @override
  State<Root> createState() => _RootState();
}

class _RootState extends State<Root> {
  late bool _pairingEnabled;
  late List<String> _sensorsNearby;
  late String? _pairingWith;

  late List<Sensor> _sensorsPaired;

  late Connection _connection;

  late String _logs;
  late ScrollController _logsScrollController;

  @override
  void initState() {
    super.initState();
    setState(() {
      _connection = Connection();
      _logsScrollController = ScrollController();
      _logs = "";
      _connection.onLog = (message) {
        setState(() {
          _logs += message;
          if (_logsScrollController.hasClients) {
            _logsScrollController
                .jumpTo(_logsScrollController.position.maxScrollExtent);
          }
        });
      };
      _sensorsNearby = [];
      _pairingWith = null;
      _pairingEnabled = false;
      _sensorsPaired = [];
    });
    openConnection().then((_) => startLogger());
  }

  Future<void> openConnection() {
    return _connection
        .openConnection()
        .then((_) => _connection.listen())
        .then((_) {
      setState(() {
        _connection = _connection;
      });
    });
  }

  void startLogger() {
    _connection.send("ADD-LOGGER");
  }

  @override
  void dispose() async {
    await _connection.close();
    await _connection.send("REMOVE-LOGGER");
    super.dispose();
  }

  Future<void> onPairingToggle(bool p) async {
    if (p) {
      await _connection.send("PAIR-ENABLE");
      _connection.on("PAIR-ENABLE", (_, __) {
        setState(() {
          _pairingEnabled = p;
          _sensorsNearby = [];
          _pairingWith = null;
        });
        return true;
      });
      _connection.on("REQUEST-TIMEOUT", (mac, _) {
        if (_pairingWith != mac) {
          setState(() {
            _sensorsNearby.remove(mac);
            if (_pairingWith == mac) {
              _pairingWith = null;
            }
          });
        }
        return false;
      });
      _connection.on("REQUEST-NEW", (mac, _) {
        setState(() {
          _sensorsNearby.add(mac);
        });
        return false;
      });
      _connection.on("PAIR-SUCCESS", (mac, _) {
        _sensorsNearby.remove(mac);
        if (_pairingWith == mac) {
          setState(() {
            _sensorsNearby.remove(mac);
            _pairingWith = null;
          });
        }
        showMessage("Sensor $mac has been paired with the gateway", context);
        loadSensors();
        return false;
      });
      _connection.on("PAIRING-DISABLED", (_, __) {
        setState(() => _pairingEnabled = false);
        return false;
      });
      _connection.on("REQUEST-NOT-FOUND", (mac, _) {
        setState(() {
          _sensorsNearby.remove(mac);
          if (_pairingWith == mac) {
            _pairingWith = null;
          }
        });
        return false;
      });
      _connection.on("PAIRING-CANCELED", (mac, _) {
        if (_pairingWith == mac) {
          setState(() {
            _sensorsNearby.remove(mac);
            _pairingWith = null;
          });
        }
        return false;
      });
      _connection.on("PAIRING-WITH", (mac, _) {
        setState(() => _pairingWith = mac);
        return false;
      });
      _connection.on("PAIRING-TIMEOUT", (mac, _) {
        if (_pairingWith == mac) {
          setState(() {
            _sensorsNearby.remove(mac);
            _pairingWith = null;
          });
        }
        showMessage("Pairing timed out with sensor $mac", context);
        return false;
      });
    } else {
      await _connection.send("PAIR-DISABLE");
      _connection.on("PAIR-DISABLE", (_, __) {
        setState(() => _pairingEnabled = p);
        return true;
      });
      _connection.off("REQUEST-TIMEOUT-");
      _connection.off("REQUEST-NEW-");
      _connection.off("PAIR-SUCCESS-");
      _connection.off("PAIRING-DISABLED");
      _connection.off("REQUEST-NOT-FOUND-");
      _connection.off("PAIRING-CANCELED-");
      _connection.off("PAIRING-WITH-");
      _connection.off("PAIRING-TIMEOUT-");
    }
  }

  Future<void> loadSensors() async {
    await _connection.send("LIST");
    _connection.on("LIST", (json, err) {
      if (err != null) {
        showMessage("Failed to load sensors", context);
        return true;
      }
      try {
        List<Sensor> sensors = jsonDecode(json).map<Sensor>((s) {
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
            types: s["types"].cast<String>(),
            collectionCapacity: s["collection_capacity"],
            wakeUpInterval: s["wake_up_interval"],
            wakeUpIntervalMaxOffset: s["wake_up_interval_max_offset"],
            nextWakeUp: DateTime.parse(s["next_wake_up"]),
            batteryLevel: s["battery_level"],
            settings: settings,
          );
        }).toList();
        setState(() {
          _sensorsPaired = sensors;
        });
        return true;
      } catch (e) {
        showMessage("Failed to load sensors: $e", context);
        return true;
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    if (_connection.state == 0) {
      return const Center(child: CircularProgressIndicator());
    } else if (_connection.state == 1) {
      return Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Center(
            child: Text(
                "Error: Could not connect to server. Type \"ssmachmos serve\" in the terminal and try again."),
          ),
          const SizedBox(height: 20),
          TextButton(
            onPressed: () {
              openConnection();
            },
            child: const Text("Try Again"),
          ),
        ],
      );
    } else if (_connection.state == 2) {
      return TabBarView(
        children: [
          Row(
            children: [
              Expanded(
                  flex: 3,
                  child: Sensors(
                    sensors: _sensorsPaired,
                    loadSensors: loadSensors,
                    connection: _connection,
                  )),
              Container(
                width: 0.5,
                color: Colors.grey,
              ),
              Expanded(
                  flex: 2,
                  child: Bluetooth(
                    pairingEnabled: _pairingEnabled,
                    onPairingToggle: onPairingToggle,
                    sensorsNearby: _sensorsNearby,
                    pairingWith: _pairingWith,
                    onPairingSelected: (mac) async => {
                      await _connection.send("PAIR-ACCEPT $mac"),
                      _connection.on("PAIR-ACCEPT", (_, err) {
                        if (err != null) {
                          return true;
                        }
                        setState(() => _pairingWith = mac);
                        return true;
                      }),
                    },
                  )),
            ],
          ),
          Gateway(connection: _connection),
          Logs(
              logsScrollController: _logsScrollController,
              logs: _logs,
              connection: _connection),
        ],
      );
    } else {
      // should never happen as state is always 0, 1 or 2
      return const Center(
        child: Text(
            "Unknown error occurred. Please restart the application and try again."),
      );
    }
  }
}
