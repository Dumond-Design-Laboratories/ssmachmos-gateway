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
    List<_Device> devices = context.watch<Connection>().connectedSensors.map((ss) => _Device(ss)).toList();

    return Container(
        margin: const EdgeInsets.fromLTRB(8, 12, 8, 8),
        child: Row(spacing: 8, children: [
          Expanded(
              child: Column(children: [
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
    String timestamp = ss.lastSeenTimestamp;
    try {
      //print(ss.lastSeenTimestamp);
      timestamp = DateFormat('yyyy-MM-dd HH:mm:ss').format(DateTime.parse(ss.lastSeenTimestamp));
    } catch(e) {
      print(e);
    }
    return ListTile(
      //decoration: BorderDeco
      //onTap: () {},
      //padding: EdgeInsets.all(8),
      //minTileHeight: 30,
      leading:
          ss.connected ? Icon(Icons.cell_tower, color: Colors.green) : Icon(Icons.portable_wifi_off),
      title: Text("${ss.name} - ${ss.address}"),
      subtitle: Text("Last seen: ${timestamp}"),
      trailing: Column(children: [
          ss.connected ? Text("Device connected.") : Text("Device not connected."),
          if (ss.connected) Text("Idle"),
          ]),
    );
  }
}
