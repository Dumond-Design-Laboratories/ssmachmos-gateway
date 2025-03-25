import 'dart:developer';

import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import 'package:ss_machmos_gui/connection.dart';

// Aims to be main panel, a more user-friendly logs pane
// Shows connected devices and their status
// Show failed uploads
class Status extends StatelessWidget {
  const Status({super.key});

  @override
  Widget build(BuildContext context) {
    Connection conn = context.read<Connection>();
    List<_Device> devices = [];
    for (SensorStatus? ss in context.watch<Connection>().sensors.map((ss) => ss.status)) {
      if (ss != null) {
        devices.add(_Device(ss));
      } else {
        log("ss is null");
      }
    }

    return Container(
        margin: const EdgeInsets.fromLTRB(8, 12, 8, 8),
        child: Row(spacing: 8, children: [
          Expanded(
              child: Column(children: [
            TextButton(child: Text("Refresh"), onPressed: () => conn.loadConnectedSensors()),
            Text("Connected devices"),
            ListView(
              shrinkWrap: true,
              padding: const EdgeInsets.all(8),
              children: devices,
            )
          ]))
        ]));
  }
}

class _Device extends StatelessWidget {
  final SensorStatus ss;
  const _Device(this.ss);

  @override
  Widget build(BuildContext context) {
    String timestamp = DateFormat('yyyy-MM-dd HH:mm:ss').format(ss.lastSeen);
    //String timestamp = ss.lastSeen;
    // try {
    //   //timestamp = DateFormat('yyyy-MM-dd HH:mm:ss').format(DateTime.parse(ss.lastSeen).toLocal());
    //   timestamp = DateFormat('yyyy-MM-dd HH:mm:ss').format(ss.lastSeen);
    // } catch(e) {
    //   log(e.toString());
    // }
    return ListTile(
      leading: ss.connected ? Icon(Icons.cell_tower, color: Colors.green) : Icon(Icons.portable_wifi_off),
      title: Text("${ss.name} - ${ss.address}"),
      subtitle: Text("Last seen: $timestamp"),
      trailing: Column(children: [
        ss.connected ? Text("Device connected.") : Text("Device not connected."),
        Text(ss.activity),
      ]),
    );
  }
}
