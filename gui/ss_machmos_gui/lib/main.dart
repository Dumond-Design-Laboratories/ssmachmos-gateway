import 'dart:io';

import 'package:flutter/material.dart';
import 'package:ss_machmos_gui/bluetooth.dart';
import 'package:ss_machmos_gui/connection.dart';
import 'package:ss_machmos_gui/gateway.dart';
import 'package:ss_machmos_gui/sensors.dart';

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
  bool _pairing = false;
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

  @override
  Widget build(BuildContext context) {
    if (_connection.state == 0) {
      return const Center(child: CircularProgressIndicator());
    } else if (_connection.state == 1) {
      return Column(
        children: [
          const Center(child: Text("Failed to connect to server.")),
          ElevatedButton(
            onPressed: () => _connection.openConnection(),
            child: const Text("Retry"),
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
                pairing: _pairing,
                onPairing: (p) async => {
                  if (p)
                    {
                      await _connection.send("PAIR-ENABLE"),
                      _connection.on(
                          "OK:PAIR-ENABLE", () => setState(() => _pairing = p)),
                    }
                  else
                    {
                      await _connection.send("PAIR-DISABLE"),
                      _connection.on("OK:PAIR-DISABLE",
                          () => setState(() => _pairing = p)),
                    }
                },
              )),
            ],
          ),
          Gateway(connection: _connection),
        ],
      );
    } else {
      return const Center(child: Text("Error"));
    }
  }
}
