import 'package:flutter/material.dart';
import 'package:ss_machmos_gui/bluetooth.dart';
import 'package:ss_machmos_gui/connection.dart';
import 'package:ss_machmos_gui/gateway.dart';
import 'package:ss_machmos_gui/sensors.dart';
import 'package:ss_machmos_gui/utils.dart';

void main() {
  runApp(const MainApp());
}

class MainApp extends StatelessWidget {
  const MainApp({super.key});

  @override
  Widget build(BuildContext context) {
    return const MaterialApp(
      home: DefaultTabController(
        length: 2,
        child: Scaffold(
          appBar: TabBar(
            isScrollable: true,
            labelPadding: EdgeInsets.symmetric(horizontal: 25),
            tabAlignment: TabAlignment.start,
            tabs: [
              Tab(text: "Sensors"),
              Tab(text: "Gateway"),
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
  bool _pairingEnabled = false;
  List<String> _sensorsNearby = [
    "AA:BB:CC:DD:EE:FF",
    "00:11:22:33:44:55",
    "CC:DD:AA:00:11:22"
  ];
  String? _pairingWith;

  late Connection _connection;

  @override
  void initState() {
    super.initState();
    setState(() {
      _connection = Connection();
    });
    _connection
        .openConnection()
        .then((_) => _connection.listen())
        .then((_) => setState(() {
              _connection = _connection;
            }));
  }

  @override
  void dispose() async {
    await _connection.close();
    super.dispose();
  }

  void onPairingToggle(bool p) async {
    if (p) {
      await _connection.send("PAIR-ENABLE");
      _connection.on("OK:PAIR-ENABLE", (_) {
        setState(() => _pairingEnabled = p);
        return true;
      });
      _connection.on("MSG:REQUEST-TIMEOUT-", (msg) {
        if (_pairingWith != msg.substring("MSG:REQUEST-TIMEOUT-".length)) {
          setState(() {
            _sensorsNearby.remove(msg.substring("MSG:REQUEST-TIMEOUT-".length));
          });
        }
        return false;
      });
      _connection.on("MSG:REQUEST-NEW-", (msg) {
        setState(() {
          _sensorsNearby.add(msg.substring("MSG:REQUEST-NEW-".length));
        });
        return false;
      });
      _connection.on("MSG:PAIR-SUCCESS-", (msg) {
        _sensorsNearby.remove(msg.substring("MSG:PAIR-SUCCESS-".length));
        if (_pairingWith == msg.substring("MSG:PAIR-SUCCESS-".length)) {
          setState(() => _pairingWith = null);
        }
        showMessage(
            "Sensor ${msg.substring("MSG:PAIR-SUCCESS-".length)} has been paired with the gateway",
            context);
        return false;
      });
      _connection.on("MSG:PAIRING-DISABLED", (_) {
        setState(() => _pairingEnabled = false);
        return false;
      });
      _connection.on("MSG:REQUEST-NOT-FOUND-", (msg) {
        setState(() {
          _sensorsNearby.remove(msg.substring("MSG:REQUEST-NOT-FOUND-".length));
        });
        return false;
      });
      _connection.on("MSG:PAIRING-CANCELED-", (msg) {
        if (_pairingWith == msg.substring("MSG:PAIRING-CANCELED-".length)) {
          setState(() => _pairingWith = null);
        }
        return false;
      });
      _connection.on("MSG:PAIRING-WITH-", (msg) {
        setState(
            () => _pairingWith = msg.substring("MSG:PAIRING-WITH-".length));
        return false;
      });
      _connection.on("MSG:PAIRING-TIMEOUT-", (msg) {
        if (_pairingWith == msg.substring("MSG:PAIRING-TIMEOUT-".length)) {
          setState(() => _pairingWith = null);
        }
        showMessage(
            "Pairing timed out with sensor ${msg.substring("MSG:PAIRING-TIMEOUT-".length)}",
            context);
        return false;
      });
    } else {
      await _connection.send("PAIR-DISABLE");
      _connection.on("OK:PAIR-DISABLE", (_) {
        setState(() => _pairingEnabled = p);
        return true;
      });
      _connection.off("MSG:REQUEST-TIMEOUT-");
      _connection.off("MSG:REQUEST-NEW-");
      _connection.off("MSG:PAIR-SUCCESS-");
      _connection.off("MSG:PAIRING-DISABLED");
      _connection.off("MSG:REQUEST-NOT-FOUND-");
      _connection.off("MSG:PAIRING-CANCELED-");
      _connection.off("MSG:PAIRING-WITH-");
      _connection.off("MSG:PAIRING-TIMEOUT-");
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_connection.state == 0) {
      return const Center(child: CircularProgressIndicator());
    } else if (_connection.state == 1) {
      return Column(
        children: [
          const Center(
              child: Text(
                  "Error: Could not connect to server. Type \"ssmachmos serve\" in the terminal and try again.")),
          ElevatedButton(
            onPressed: () => _connection.openConnection(),
            child: const Text("Try Again"),
          ),
        ],
      );
    } else if (_connection.state == 2) {
      return TabBarView(
        children: [
          Row(
            children: [
              const Expanded(child: Sensors()),
              Container(
                width: 0.5,
                color: Colors.grey,
              ),
              Expanded(
                  child: Bluetooth(
                pairingEnabled: _pairingEnabled,
                onPairingToggle: onPairingToggle,
                sensorsNearby: _sensorsNearby,
                pairingWith: _pairingWith,
                onPairingSelected: (mac) async => {
                  await _connection.send("PAIR-ACCEPT $mac"),
                  _connection.on("OK:PAIR-ACCEPT", (_) {
                    setState(() => _pairingWith = mac);
                    return true;
                  }),
                },
              )),
            ],
          ),
          Gateway(connection: _connection),
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
