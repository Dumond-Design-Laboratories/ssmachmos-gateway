import 'dart:io';

import 'package:flutter/material.dart';
import 'package:ss_machmos_gui/bluetooth.dart';
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
  bool _connectedToServer = false;
  bool _pairing = false;
  Socket? _socket;

  @override
  void initState() {
    super.initState();
    openConnection().then((socket) async {
      if (socket == null) {
        // launch the server
      } else {
        setState(() {
          _socket = socket;
          _connectedToServer = true;
        });
      }
    });
  }

  Future<Socket?> openConnection() async {
    String socketPath = "/tmp/ss_mach_mos.sock";
    try {
      final socket = await Socket.connect(
          InternetAddress(socketPath, type: InternetAddressType.unix), 0);
      return socket;
    } catch (e) {
      return null;
    }
  }

  @override
  Widget build(BuildContext context) {
    if (!_connectedToServer) {
      return const Center(child: CircularProgressIndicator());
    } else {
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
                onPairing: (p) => setState(() => _pairing = p),
              )),
            ],
          ),
          const Gateway(),
        ],
      );
    }
  }
}
