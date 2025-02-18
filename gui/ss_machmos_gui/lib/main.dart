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
      home: Consumer<Connection>(
          builder: (context, connection, child) =>
              AppRoot(connState: connection.state)),
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
