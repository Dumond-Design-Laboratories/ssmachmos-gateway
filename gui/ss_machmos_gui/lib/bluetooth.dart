import 'package:flutter/material.dart';

class Bluetooth extends StatefulWidget {
  final bool pairing;
  final void Function(bool) onPairing;

  const Bluetooth({super.key, required this.pairing, required this.onPairing});

  @override
  State<Bluetooth> createState() => _BluetoothState();
}

class _BluetoothState extends State<Bluetooth> {
  List<String> _sensors = [
    "AA:BB:CC:DD:EE:FF",
    "00:11:22:33:44:55",
    "CC:DD:AA:00:11:22"
  ];

  String? _pairingWith;

  @override
  void initState() {
    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Row(
          children: [
            const Icon(Icons.bluetooth),
            const Text("Discover Sensors"),
            const Spacer(),
            Switch(
              value: widget.pairing,
              onChanged: widget.onPairing,
            ),
          ],
        ),
        if (widget.pairing)
          Expanded(
            child: ListView.separated(
              itemCount: _sensors.length + 1,
              separatorBuilder: (_, __) => const Divider(
                  color: Colors.grey, height: 0.5, thickness: 0.5),
              itemBuilder: (BuildContext context, int index) {
                if (index == _sensors.length) {
                  return const Center(
                    child: Padding(
                      padding: EdgeInsets.symmetric(vertical: 50),
                      child: CircularProgressIndicator(),
                    ),
                  );
                }
                return SensorListItem(
                  mac: _sensors[index],
                  pairing: _pairingWith == _sensors[index],
                  onPairing: () {
                    setState(() {
                      _pairingWith = _sensors[index];
                    });
                  },
                );
              },
            ),
          ),
      ],
    );
  }
}

class SensorListItem extends StatefulWidget {
  final String mac;
  final bool pairing;
  final void Function() onPairing;

  const SensorListItem(
      {super.key,
      required this.mac,
      required this.pairing,
      required this.onPairing});

  @override
  State<SensorListItem> createState() => _SensorListItemState();
}

class _SensorListItemState extends State<SensorListItem> {
  @override
  Widget build(BuildContext context) {
    return ListTile(
      title: Row(
        children: [
          Text(widget.mac),
          const Spacer(),
          if (widget.pairing)
            const SizedBox(
              width: 25,
              height: 25,
              child: CircularProgressIndicator(),
            ),
        ],
      ),
      contentPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 5),
      onTap: widget.pairing ? null : widget.onPairing,
    );
  }
}
