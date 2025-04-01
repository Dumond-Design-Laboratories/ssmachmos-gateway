import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import 'package:ss_machmos_gui/connection.dart';
import 'package:ss_machmos_gui/sensors.dart';

// Aims to be main panel, a more user-friendly logs pane
// Shows connected devices and their status
// Show failed uploads
class Status extends StatelessWidget {
  const Status({super.key});

  @override
  Widget build(BuildContext context) {
    Connection conn = context.read<Connection>();
    List<_Device> devices = context.watch<Connection>().sensors.where((Sensor s) => s.status != null).map((Sensor e) => _Device(e)).toList();

    return Container(
        margin: const EdgeInsets.fromLTRB(8, 12, 8, 8),
        child: Column(children: [
          Row(mainAxisAlignment: MainAxisAlignment.center, children: [
            IconButton(icon: const Icon(Icons.refresh), tooltip: "Refresh device list", onPressed: () => conn.loadConnectedSensors()),
            Text("Process PID ${conn.serverPid ?? 'none'}"),
          ]),
          ListView(
            shrinkWrap: true,
            padding: const EdgeInsets.all(8),
            children: devices,
          )
        ]));
  }
}

class _Device extends StatelessWidget {
  final Sensor sensor;
  const _Device(this.sensor);

  @override
  Widget build(BuildContext context) {
    SensorStatus ss = sensor.status!;
    String timestamp = DateFormat('yyyy-MM-dd HH:mm:ss').format(ss.lastSeen);
    return ListTile(
      leading: ss.connected ? Icon(Icons.cell_tower, color: Colors.green) : Icon(Icons.portable_wifi_off),
      title: Text("${ss.name} - ${sensor.model.name}"),
      subtitle: Flex(direction: Axis.horizontal, children: [
        Text(ss.address),
        VerticalDivider(),
        Text(sensor.model.name),
        VerticalDivider(),
        Text("Last seen at $timestamp"),
        Spacer(),
        //Spacer(flex: 4),
      ]),
      trailing: Column(children: [
        ss.connected ? Text("Device connected.") : Text("Device not connected."),
        Text(ss.activity),
      ]),
    );
  }
}
