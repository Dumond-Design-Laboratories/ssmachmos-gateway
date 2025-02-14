import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:ss_machmos_gui/connection.dart';

class Bluetooth extends StatelessWidget {
  const Bluetooth({super.key});

  @override
  Widget build(BuildContext context) {
    Connection _conn = context.watch<Connection>();
    bool pairingEnabled = _conn.pairingEnabled;
    String pairingWith = _conn.pairingWith;
    List<String> sensorsNearby = _conn.sensorsNearby;
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
                value: pairingEnabled,
                onChanged: (value) => _conn.setPairingState(value),
              ),
            ],
          ),
          const SizedBox(height: 12),
          if (pairingEnabled)
            Expanded(
              child: DecoratedBox(
                decoration: BoxDecoration(
                  border: Border.all(color: Colors.grey, width: 0.5),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: ListView.separated(
                  itemCount: sensorsNearby.length + 1,
                  separatorBuilder: (_, __) => const Divider(
                      color: Colors.grey, height: 0.5, thickness: 0.5),
                  itemBuilder: (BuildContext context, int index) {
                    // Last item is the spinner at the bottom
                    if (index == sensorsNearby.length) {
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
                      mac: sensorsNearby[index],
                      pairing: pairingWith == sensorsNearby[index],
                      onTap: () =>
                          _conn.acceptPairing(sensorsNearby[index]),
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

class SensorListItem extends StatelessWidget {
  final String mac;
  final bool pairing;
  final void Function() onTap;

  const SensorListItem(
      {super.key,
      required this.mac,
      required this.pairing,
      required this.onTap});

  @override
  Widget build(BuildContext context) {
    return ListTile(
      title: Row(
        children: [
          Text(mac, style: const TextStyle(fontSize: 14)),
          const Spacer(),
          if (pairing)
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
      onTap: pairing ? null : onTap,
    );
  }
}
