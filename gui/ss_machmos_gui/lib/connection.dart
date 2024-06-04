import 'dart:async';
import 'dart:io';

class Connection {
  late int _state; // 0: connecting, 1: failed, 2: connected
  Socket? _socket;
  late Map<String, bool Function(String, String?)>
      _waitingFor; // callback should return if we should remove the callback

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
          List<String> parts = message.split(":");
          if (parts.length > 1) {
            String command = parts[1];
            if (_waitingFor.containsKey(command)) {
              if (_waitingFor[command]!(
                  parts.length > 2 ? parts.sublist(2).join(":") : "",
                  parts[0] == "ERR" ? "Error: ${parts[1]}" : null)) {
                found.add(command);
              }
            }
          }
          for (String command in found) {
            _waitingFor.remove(command);
          }
        }
      });
    }
  }

  void on(String command, bool Function(String, String?) callback) {
    _waitingFor[command] = callback;
  }

  void off(String command) {
    _waitingFor.remove(command);
  }

  Future<void> close() async {
    if (_socket != null) {
      await _socket!.close();
    }
  }
}
