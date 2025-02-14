import 'dart:async';
import 'dart:convert';
import 'dart:developer';

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:ss_machmos_gui/bluetooth.dart';
import 'package:ss_machmos_gui/connection.dart';
import 'package:ss_machmos_gui/gateway.dart';
import 'package:ss_machmos_gui/help.dart';
import 'package:ss_machmos_gui/logs.dart';
import 'package:ss_machmos_gui/sensors.dart';
import 'package:ss_machmos_gui/utils.dart';

void main() {
  // Global state because why not
  runApp(ChangeNotifierProvider(
      create: (context) => Connection(), child: const MainApp()));
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
      home: Consumer<Connection>(builder: (context, connection, child) => AppRoot(connState: connection.state)),
    );
  }
}

class AppRoot extends StatelessWidget {
  final ConnState connState;
  const AppRoot({super.key, required this.connState});

  static final GlobalKey _sensorTypesKey = GlobalKey();
  static final GlobalKey _wakeUpIntervalKey = GlobalKey();
  static final GlobalKey _gatewayIdKey = GlobalKey();
  static final GlobalKey _httpEndpointKey = GlobalKey();

  @override
  Widget build(BuildContext context) {
    const List<Tab> tabs = [
      Tab(text: "Sensors"),
      Tab(text: "Gateway"),
      Tab(text: "Logs"),
      Tab(icon: Icon(Icons.help_outline)),
    ];

    var conn = context.read<Connection>();
    Widget body;
    if (connState != ConnState.connected) {
      // Placeholder until connection restarts...
      body = Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Center(
            child: Text("Error: Could not connect to server."),
          ),
          const SizedBox(height: 20),
          // Button to start the gateway server backend
          TextButton(
            onPressed: () async {
              // Start server backend. Once it is up and running, a provider state triggers to redraw tree
              await conn.startServer();
              conn.openConnection();
            },
            child: const Text("Start Server"),
          ),
        ],
      );
    } else {
      body = TabBarView(
        children: [
          // Left column displaying sensors available
          // Right column displaying sensors awaiting pairing
          Row(
            children: [
              Expanded(flex: 3, child: Sensors()),
              Container(width: 0.5, color: Colors.grey),
              Expanded(flex: 2, child: Bluetooth()),
            ],
          ),
          GatewayView(),
          Logs(
              //logsScrollController: _logsScrollController,
              //logs: _logs,
              //connection: _connection
              ),
          Help(
            sensorTypesKey: _sensorTypesKey,
            wakeUpIntervalKey: _wakeUpIntervalKey,
            gatewayIdKey: _gatewayIdKey,
            httpEndpointKey: _httpEndpointKey,
          ),
        ],
      );
    }

    WidgetsBinding.instance.addPostFrameCallback((_) {
      // Can't call this in build
      while (conn.toastQueue.isNotEmpty) {
        showMessage(conn.toastQueue.removeFirst(), context);
      }
    });

    return DefaultTabController(
      animationDuration: Duration.zero,
      length: tabs.length,
      child: Scaffold(
        appBar: AppBar(
          bottom: const TabBar(tabs: tabs),
        ),
        body: body,
      ),
    );
  }
}

// TODO delete this
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

  late Timer _pairListTimer;

  @override
  void initState() {
    super.initState();
    setState(() {
      _tabController = TabController(
          length: tabs.length, vsync: this, animationDuration: Duration.zero);
      // Connection to backend
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

    // Open socket to server
    // This is a unix socket
    openConnection();
    _pairListTimer = Timer.periodic(const Duration(seconds: 2), (Timer t) {
      // Every two seconds poll server for new paired devices
      if (_pairingEnabled) {
        _connection.send("PAIR LIST");
      }
    });

    _connection.on("PAIR-LIST", (devices, _) {
      // Server returns list of devices pending pairing
      setState(() {
        _sensorsNearby = List<String>.from(jsonDecode(devices));
      });
      // return false to keep in event callback
      return false;
    });
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
    _pairListTimer.cancel();
    await _connection.close();
    await _connection.send("REMOVE-LOGGER");
    _logsConnected = false;
    if (_pairingEnabled) {
      await _connection.send("PAIR-DISABLE");
    }
    super.dispose();
  }

  // On pairing enable, attach callbacks from server
  // TODO move this to [Connection]
  Future<void> onPairingToggle(bool p) async {
    if (p) {
      // Clear state on enable
      _connection.on("PAIR-ENABLE", (_, __) {
        setState(() {
          _pairingEnabled = p;
          _sensorsNearby = [];
          _pairingWith = null;
        });
        return true;
      });
      // Attempted pairing with a sensor already paired
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
      // New pairing request, called on connect for new devices
      _connection.on("REQUEST-NEW", (mac, _) {
        setState(() {
          log("REQUEST-NEW Called");
          _sensorsNearby.add(mac);
        });
        return false;
      });
      // BLE agent finished pairing
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
      // Pairing disabled, sent by server in response when toggle switch is off
      _connection.on("PAIRING-DISABLED", (_, __) {
        setState(() => _pairingEnabled = false);
        return false;
      });
      // Attempted to start pairing with a device not pending pairing
      _connection.on("REQUEST-NOT-FOUND", (mac, _) {
        setState(() {
          _sensorsNearby.remove(mac);
          if (_pairingWith == mac) {
            _pairingWith = null;
          }
        });
        return false;
      });
      // Cancelled pairing, either by timeout, BLE agent, or user
      _connection.on("PAIRING-CANCELED", (mac, _) {
        if (_pairingWith == mac) {
          setState(() {
            _sensorsNearby.remove(mac);
            _pairingWith = null;
          });
        }
        return false;
      });
      // Start pairing with device
      _connection.on("PAIRING-WITH", (mac, _) {
        setState(() => _pairingWith = mac);
        return false;
      });
      // Pairing timeout
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
      // Ask server to enable pairing
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
        List<Sensor> sensors =
            jsonDecode(json).map<Sensor>((s) => Sensor.fromJson(s)).toList();
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
      // Button to start server
      body = Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Center(
            child: Text("Error: Could not connect to server."),
          ),
          const SizedBox(height: 20),
          // Button to start the gateway server backend
          TextButton(
            onPressed: () async {
              // await _connection.startServer();
              // for (int i = 0; i < 30; i++) {
              //   await openConnection();
              //   await Future.delayed(const Duration(seconds: 1));
              //   if (_logsConnected) {
              //     break;
              //   }
              // }
            },
            child: const Text("Start Server"),
          ),
        ],
      );
    } else if (_connection.state == 2) {
      body = TabBarView(
        controller: _tabController,
        children: [
          // First tab is the sensor display
          Row(
            children: [
              // Left column displaying sensors available
              Expanded(
                  flex: 3,
                  child: Sensors(
                      // sensors: _sensorsPaired,
                      // loadSensors: loadSensors,
                      // connection: _connection,
                      // tabController: _tabController,
                      // typesKey: _sensorTypesKey,
                      // wakeUpIntervalKey: _wakeUpIntervalKey,
                      )),
              Container(
                width: 0.5,
                color: Colors.grey,
              ),
              // Right column displaying sensors awaiting pairing
              Expanded(
                  flex: 2,
                  child: Bluetooth(
                      // pairingEnabled: _pairingEnabled,
                      //   onPairingToggle: onPairingToggle,
                      //   sensorsNearby: _sensorsNearby,
                      //   pairingWith: _pairingWith,
                      //   // On selecting a device, send pair command
                      //   onPairingSelected: (mac) async => {
                      //     // BUG this doesn't get cleared out on select.
                      //     // Happens when forgetting device then repairing
                      //     _connection.on("PAIR-ACCEPT", (_, err) {
                      //       if (err != null) {
                      //         return true;
                      //       }
                      //       setState(() => _pairingWith = mac);
                      //       return true;
                      //     }),
                      //     await _connection.send("PAIR-ACCEPT $mac"),
                      //   },
                      )),
            ],
          ),
          // Second tab is gateway information
          GatewayView(
              // connection: _connection,
              // tabController: _tabController,
              // idKey: _gatewayIdKey,
              // httpEndpointKey: _httpEndpointKey,
              ),
          // Logs(
          //     logsScrollController: _logsScrollController,
          //     logs: _logs,
          //     connection: _connection),
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
