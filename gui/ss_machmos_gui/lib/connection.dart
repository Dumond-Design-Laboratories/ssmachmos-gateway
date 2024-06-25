import 'dart:async';
import 'dart:io';

import 'package:flutter/foundation.dart';

class Connection {
  late int _state; // 0: connecting, 1: failed, 2: connected
  Socket? _socket;
  late Map<String, bool Function(String, String?)>
      _waitingFor; // callback should return if we should remove the callback
  void Function(String message)? onLog;
  void Function()? onError;

  int get state => _state;

  Connection() {
    _state = 1;
    _waitingFor = {};
  }

  Future<void> startServer() async {
    await Process.run(
      kDebugMode
          ? "${Directory.current.parent.parent.path}/server/ssmachmos"
          : "ssmachmos",
      ["serve", "--no-console"],
    );
  }

  Future<void> openConnection() async {
    _state = 0;
    String socketPath = "/tmp/ss_machmos.sock";
    try {
      _socket = null;
      final socket = await Socket.connect(
          InternetAddress(socketPath, type: InternetAddressType.unix), 0);
      _socket = socket;
      _state = 2;
    } catch (e) {
      _state = 1;
      if (onError != null) {
        onError!();
      }
    }
  }

  Future<void> send(String message) async {
    if (_socket == null) {
      _state = 1;
      if (onError != null) {
        onError!();
      }
      return;
    }
    try {
      _socket!.write("$message\x00");
      if (message == "STOP") {
        await _socket!.close();
        _state = 1;
        if (onError != null) {
          onError!();
        }
      }
    } catch (_) {
      _state = 1;
      if (onError != null) {
        onError!();
      }
    }
  }

  Future<void> listen() async {
    if (_socket != null) {
      _socket!.listen((event) {
        List<String> messages = String.fromCharCodes(event).split('\x00');
        for (String message in messages) {
          if (message.isEmpty) {
            continue;
          }
          List<String> found = [];
          List<String> parts = message.split(":");
          if (parts.length > 1) {
            if (parts[0] == "LOG") {
              if (onLog != null) {
                onLog!(parts.sublist(1).join(":"));
              }
              continue;
            }
            String command = parts[1];
            if (_waitingFor.containsKey(command)) {
              if (_waitingFor[command]!(
                  parts.length > 2 ? parts.sublist(2).join(":") : "",
                  parts[0] == "ERR"
                      ? "Error: ${parts.sublist(1).join(":")}"
                      : null)) {
                found.add(command);
              }
            }
          }
          for (String command in found) {
            _waitingFor.remove(command);
          }
        }
      }, cancelOnError: true).onError((error) {
        _state = 1;
        if (onError != null) {
          onError!();
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
