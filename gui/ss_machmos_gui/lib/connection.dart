import 'dart:async';
import 'dart:io';

class Connection {
  late int _state; // 0: connecting, 1: failed, 2: connected
  Socket? _socket;
  late Map<String, bool Function(String)> _waitingFor;

  Connection() {
    _state = 1;
    _waitingFor = {};
  }

  int get state => _state;

  Future<void> openConnection() async {
    _state = 0;
    String socketPath = "/tmp/ss_mach_mos.sock";
    try {
      final socket = await Socket.connect(
          InternetAddress(socketPath, type: InternetAddressType.unix), 0);
      _socket = socket;
      _state = 2;
    } catch (e) {
      _state = 1;
    }
  }

  Future<void> send(String message) async {
    if (_socket != null) {
      try {
        _socket!.write("\n$message\n");
      } catch (e) {
        _state = 1;
      }
    }
  }

  Future<void> listen() async {
    if (_socket != null) {
      _socket!.listen((event) {
        List<String> messages = String.fromCharCodes(event).split("\n");
        for (String message in messages) {
          if (message.isEmpty) {
            continue;
          }
          List<String> found = [];
          for (String prefix in _waitingFor.keys) {
            if (message.startsWith(prefix)) {
              if (_waitingFor[prefix]!(message)) {
                found.add(prefix);
              }
            }
          }
          for (String prefix in found) {
            _waitingFor.remove(prefix);
          }
        }
      });
    }
  }

  void on(String prefix, bool Function(String) callback) {
    _waitingFor[prefix] = callback;
  }

  void off(String prefix) {
    _waitingFor.remove(prefix);
  }

  Future<void> close() async {
    if (_socket != null) {
      await _socket!.close();
    }
  }

  void setUpPairingResponses(void Function(String) print) {
    on("MSG:REQUEST-TIMEOUT-", (msg) {
      print(
          "Pairing request timed out for sensor ${msg.substring("MSG:REQUEST-TIMEOUT-".length)}");
      return false;
    });
    on("MSG:REQUEST-NEW-", (msg) {
      print(
          "New pairing request from sensor ${msg.substring("MSG:REQUEST-NEW-".length)}");
      return false;
    });
    on("MSG:PAIR-SUCCESS-", (msg) {
      print(
          "Pairing successful with sensor ${msg.substring("MSG:PAIR-SUCCESS-".length)}");
      return false;
    });
    on("MSG:PAIRING-DISABLED", (_) {
      print("Error: Pairing mode disabled");
      return false;
    });
    on("MSG:REQUEST-NOT-FOUND-", (msg) {
      print(
          "Error: Pairing request not found for sensor ${msg.substring("MSG:REQUEST-NOT-FOUND-".length)}");
      return false;
    });
    on("MSG:PAIRING-CANCELED-", (msg) {
      print(
          "Pairing canceled with sensor ${msg.substring("MSG:PAIRING-CANCELED-".length)}");
      return false;
    });
    on("MSG:PAIRING-WITH-", (msg) {
      print("Pairing with sensor ${msg.substring("MSG:PAIRING-WITH-".length)}");
      return false;
    });
    on("MSG:PAIRING-TIMEOUT-", (msg) {
      print(
          "Pairing timed out with sensor ${msg.substring("MSG:PAIRING-TIMEOUT-".length)}");
      return false;
    });
  }

  void takeDownPairingResponses() {
    off("MSG:REQUEST-TIMEOUT-");
    off("MSG:REQUEST-NEW-");
    off("MSG:PAIR-SUCCESS-");
    off("MSG:PAIRING-DISABLED");
    off("MSG:REQUEST-NOT-FOUND-");
    off("MSG:PAIRING-CANCELED-");
    off("MSG:PAIRING-WITH-");
    off("MSG:PAIRING-TIMEOUT-");
  }
}
