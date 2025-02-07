import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:ss_machmos_gui/bluetooth.dart';
import 'package:ss_machmos_gui/connection.dart';
import 'package:ss_machmos_gui/gateway.dart';
import 'package:ss_machmos_gui/help.dart';
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
          bodyLarge: TextStyle(fontSize: 14),
        ),
        dropdownMenuTheme: const DropdownMenuThemeData(
          textStyle: TextStyle(fontSize: 14),
        ),
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF326496),
        ),
        inputDecorationTheme: const InputDecorationTheme(
          border: OutlineInputBorder(),
          contentPadding: EdgeInsets.symmetric(horizontal: 10, vertical: 10),
          isDense: true,
        ),
        textButtonTheme: TextButtonThemeData(
          style: ButtonStyle(
            shape: WidgetStateProperty.all(
              RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(5),
              ),
            ),
            foregroundColor: WidgetStateProperty.all(Colors.white),
            backgroundColor: WidgetStateProperty.all(const Color(0xFF326496)),
            padding: WidgetStateProperty.all(
              const EdgeInsets.symmetric(horizontal: 10, vertical: 10),
            ),
          ),
        ),
      ),
      home: const Root(),
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

class _RootState extends State<Root> with SingleTickerProviderStateMixin {
  final List<Tab> tabs = [
    const Tab(text: "Sensors"),
    const Tab(text: "Gateway"),
    const Tab(text: "Logs"),
    const Tab(icon: Icon(Icons.help_outline)),
  ];
  late TabController _tabController;

  late bool _pairingEnabled;
  late List<String> _sensorsNearby;
  late String? _pairingWith;

  late List<Sensor> _sensorsPaired;

  late Connection _connection;

  late String _logs;
  late bool _logsConnected;
  late ScrollController _logsScrollController;

  late GlobalKey _sensorTypesKey;
  late GlobalKey _wakeUpIntervalKey;
  late GlobalKey _gatewayIdKey;
  late GlobalKey _httpEndpointKey;

  @override
  void initState() {
    super.initState();
    setState(() {
      _tabController = TabController(
          length: tabs.length, vsync: this, animationDuration: Duration.zero);
      _connection = Connection();
      _logsScrollController = ScrollController();
      _logs = "";
      _logsConnected = false;
      _connection.onLog = (message) {
        setState(() {
          _logs += message;
          if (_logsScrollController.hasClients) {
            _logsScrollController
                .jumpTo(_logsScrollController.position.maxScrollExtent);
          }
        });
      };
      _connection.onError = () {
        _logsConnected = false;
      };
      _sensorsNearby = [];
      _pairingWith = null;
      _pairingEnabled = false;
      _sensorsPaired = [];
      _sensorTypesKey = GlobalKey();
      _wakeUpIntervalKey = GlobalKey();
      _gatewayIdKey = GlobalKey();
      _httpEndpointKey = GlobalKey();
    });
    openConnection();
  }

  Future<void> openConnection() {
    return _connection
        .openConnection()
        .then((_) => _connection.listen())
        .then((_) {
      setState(() {
        _connection = _connection;
      });
    }).then((_) {
      if (!_logsConnected) {
        _logsConnected = true;
        startLogger();
      }
    });
  }

  void startLogger() {
    _connection.send("ADD-LOGGER");
  }

  @override
  void dispose() async {
    await _connection.close();
    await _connection.send("REMOVE-LOGGER");
    _logsConnected = false;
    if (_pairingEnabled) {
      await _connection.send("PAIR-DISABLE");
    }
    super.dispose();
  }

  Future<void> onPairingToggle(bool p) async {
    if (p) {
      _connection.on("PAIR-ENABLE", (_, __) {
        setState(() {
          _pairingEnabled = p;
          _sensorsNearby = [];
          _pairingWith = null;
        });
        return true;
      });
      _connection.on("REQUEST-SENSOR-EXISTS", (mac, _) {
        showMessage(
            "Pairing request for already paired sensor $mac. First \"Forget\" sensor $mac before pairing again.",
            context);
        return false;
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
      await _connection.send("PAIR-ENABLE");
    } else {
      _connection.on("PAIR-DISABLE", (_, __) {
        setState(() {
          _pairingEnabled = p;
          _sensorsNearby = [];
          _pairingWith = null;
        });
        _connection.off("REQUEST-SENSOR-EXISTS");
        _connection.off("REQUEST-TIMEOUT");
        _connection.off("REQUEST-NEW");
        _connection.off("PAIR-SUCCESS");
        _connection.off("PAIRING-DISABLED");
        _connection.off("REQUEST-NOT-FOUND");
        _connection.off("PAIRING-CANCELED");
        _connection.off("PAIRING-WITH");
        _connection.off("PAIRING-TIMEOUT");
        return true;
      });
      await _connection.send("PAIR-DISABLE");
    }
  }

  Future<void> loadSensors() async {
    _connection.on("LIST", (json, err) {
      if (err != null) {
        showMessage("Failed to load sensors", context);
        return true;
      }
      try {
        // Deserialize data into a list of sensors
        // Map all objects in map to a Sensor object
        List<Sensor> sensors = jsonDecode(json).map<Sensor>((s) {
          return Sensor.fromJson(s);
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
    await _connection.send("LIST");
  }

  @override
  Widget build(BuildContext context) {
    Widget body;
    if (_connection.state == 0) {
      body = const Center(child: CircularProgressIndicator());
    } else if (_connection.state == 1) {
      body = Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Center(
            child: Text("Error: Could not connect to server."),
          ),
          const SizedBox(height: 20),
          TextButton(
            onPressed: () async {
              await _connection.startServer();
              for (int i = 0; i < 30; i++) {
                await openConnection();
                await Future.delayed(const Duration(seconds: 1));
                if (_logsConnected) {
                  break;
                }
              }
            },
            child: const Text("Start Server"),
          ),
        ],
      );
    } else if (_connection.state == 2) {
      body = TabBarView(
        controller: _tabController,
        children: [
          Row(
            children: [
              Expanded(
                  flex: 3,
                  child: Sensors(
                    sensors: _sensorsPaired,
                    loadSensors: loadSensors,
                    connection: _connection,
                    tabController: _tabController,
                    typesKey: _sensorTypesKey,
                    wakeUpIntervalKey: _wakeUpIntervalKey,
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
                      _connection.on("PAIR-ACCEPT", (_, err) {
                        if (err != null) {
                          return true;
                        }
                        setState(() => _pairingWith = mac);
                        return true;
                      }),
                      await _connection.send("PAIR-ACCEPT $mac"),
                    },
                  )),
            ],
          ),
          Gateway(
            connection: _connection,
            tabController: _tabController,
            idKey: _gatewayIdKey,
            httpEndpointKey: _httpEndpointKey,
          ),
          Logs(
              logsScrollController: _logsScrollController,
              logs: _logs,
              connection: _connection),
          Help(
            sensorTypesKey: _sensorTypesKey,
            wakeUpIntervalKey: _wakeUpIntervalKey,
            gatewayIdKey: _gatewayIdKey,
            httpEndpointKey: _httpEndpointKey,
          ),
        ],
      );
    } else {
      // should never happen as state is always 0, 1 or 2
      body = const Center(
        child: Text(
            "Unknown error occurred. Please restart the application and try again."),
      );
    }

    return DefaultTabController(
      animationDuration: Duration.zero,
      length: tabs.length,
      child: Scaffold(
          appBar: TabBar(
            isScrollable: true,
            labelPadding: const EdgeInsets.symmetric(horizontal: 25),
            tabAlignment: TabAlignment.start,
            controller: _tabController,
            tabs: tabs,
          ),
          body: body),
    );
  }
}
