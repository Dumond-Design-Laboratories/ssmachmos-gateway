import 'dart:async';
import 'dart:io';

class Connection {
  late int _state; // 0: connecting, 1: failed, 2: connected
  Socket? _socket;
  late Map<String, void Function()> _waitingFor;

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
        _socket!.write(message);
      } catch (e) {
        _state = 1;
      }
    }
  }

  Future<void> listen() async {
    if (_socket != null) {
      _socket!.listen((event) {
        String message = String.fromCharCodes(event);
        List<String> found = [];
        for (String prefix in _waitingFor.keys) {
          if (message.startsWith(prefix)) {
            _waitingFor[prefix]!();
            found.add(prefix);
          }
        }
        for (String prefix in found) {
          _waitingFor.remove(prefix);
        }
      });
    }
  }

  void on(String prefix, void Function() callback) {
    _waitingFor[prefix] = callback;
  }

  Future<void> close() async {
    if (_socket != null) {
      await _socket!.close();
    }
  }
}
