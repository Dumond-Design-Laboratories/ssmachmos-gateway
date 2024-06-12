import 'package:flutter/material.dart';

class Bluetooth extends StatefulWidget {
  final bool pairingEnabled;
  final void Function(bool) onPairingToggle;

  final List<String> sensorsNearby;
  final String? pairingWith;
  final void Function(String) onPairingSelected;

  const Bluetooth(
      {super.key,
      required this.pairingEnabled,
      required this.onPairingToggle,
      required this.sensorsNearby,
      required this.pairingWith,
      required this.onPairingSelected});

  @override
  State<Bluetooth> createState() => _BluetoothState();
}

class _BluetoothState extends State<Bluetooth> {
  @override
  void initState() {
    super.initState();
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(8, 12, 8, 8),
      child: Column(
        children: [
          Row(
            children: [
              const Icon(
                Icons.bluetooth,
                size: 30,
              ),
              const Text("Discover Sensors"),
              const Spacer(),
              Switch(
                value: widget.pairingEnabled,
                onChanged: widget.onPairingToggle,
              ),
            ],
          ),
          const SizedBox(height: 12),
          if (widget.pairingEnabled)
            Expanded(
              child: DecoratedBox(
                decoration: BoxDecoration(
                  border: Border.all(color: Colors.grey, width: 0.5),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: ListView.separated(
                  itemCount: widget.sensorsNearby.length + 1,
                  separatorBuilder: (_, __) => const Divider(
                      color: Colors.grey, height: 0.5, thickness: 0.5),
                  itemBuilder: (BuildContext context, int index) {
                    if (index == widget.sensorsNearby.length) {
                      return const Center(
                        child: Padding(
                          padding: EdgeInsets.symmetric(vertical: 30),
                          child: SizedBox(
                            width: 25,
                            height: 25,
                            child: CircularProgressIndicator(
                              strokeWidth: 3,
                            ),
                          ),
                        ),
                      );
                    }
                    return SensorListItem(
                      mac: widget.sensorsNearby[index],
                      pairing:
                          widget.pairingWith == widget.sensorsNearby[index],
                      onPairing: () =>
                          widget.onPairingSelected(widget.sensorsNearby[index]),
                    );
                  },
                ),
              ),
            ),
        ],
      ),
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
          Text(widget.mac, style: const TextStyle(fontSize: 14)),
          const Spacer(),
          if (widget.pairing)
            const SizedBox(
              width: 15,
              height: 15,
              child: CircularProgressIndicator(
                strokeWidth: 2.5,
              ),
            ),
        ],
      ),
      contentPadding: const EdgeInsets.symmetric(horizontal: 20),
      onTap: widget.pairing ? null : widget.onPairing,
    );
  }
}
