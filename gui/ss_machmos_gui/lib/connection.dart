import 'dart:async';
import 'dart:collection';
import 'dart:convert';
import 'dart:developer';
import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:ss_machmos_gui/sensors.dart';

// 0: connecting, 1: failed, 2: connected
enum ConnState { inprogress, failed, connected, disabled }

typedef ConnectionCallback = bool Function(String, String?);

class Gateway {
  String id;
  //String password;
  String httpEndpoint;
  Gateway({required this.id, required this.httpEndpoint});

  factory Gateway.fromJson(Map<String, dynamic> json) {
    return Gateway(id: json["id"], httpEndpoint: json["http_endpoint"]);
  }
}

// UI state. Manages connection to gateway and collecting information through a socket connection
// Unix named sockets here, if we're planning on supporting windows we need named pipes too
class Connection with ChangeNotifier {
  // Have a way to pass toast notifications to UI
  // gateway notifies when this is pushed to, and AppRoot pops and displays each
  Queue<String> toastQueue = Queue<String>();
  void showMessage(String message) => toastQueue.add(message);

  // State of connection to backend
  ConnState _state = ConnState.disabled;
  ConnState get state => _state;

  late Gateway gateway;

  // Devices pending pairing
  final List<String> sensorsNearby = [];
  // Device that user wants to pair with. Display a spinner until done
  late String pairingWith = "";

  // Is pairing enabled in backend?
  bool _pairingEnabled = false;
  // TODO make this ask backend each time instead?
  bool get pairingEnabled => _pairingEnabled;

  // List of all stored sensor settings
  final List<Sensor> sensors = [];
  // Current sensor displayed
  // Can be null because we haven't selected a sensor yet(?)
  // TODO make this impossible, because we don't create a dropdown unless we do have sensors stored
  Sensor? _displayedSensor;
  Sensor? get displayedSensor => _displayedSensor;
  set displayedSensor(Sensor? value) {
    _displayedSensor = value;
    notifyListeners();
  }

  Socket? _socket;
  late Map<String, ConnectionCallback>
      _waitingFor; // callback should return if we should remove the callback

  void Function(String message)? onLog;
  void Function()? onError;
  String logs = "";

  Connection() {
    _state = ConnState.disabled;
    _waitingFor = {};
  }

  Future<void> startServer() async {
    await Process.start(
      kDebugMode ? "${Directory.current.path}/ssmachmos" : "ssmachmos",
      ["serve", "--no-console"],
    );
  }

  void stopServer() {
    send("STOP");
  }

  Future<void> openConnection() async {
    _state = ConnState.inprogress;
    String socketPath = "/tmp/ss_machmos.sock";
    try {
      // FIXME server doesn't immediately start socket.
      // And this is unreliable
      sleep(Duration(seconds: 8));
      _socket = await Socket.connect(
          InternetAddress(socketPath, type: InternetAddressType.unix), 0);
      _state = ConnState.connected;
      listen();
      log("Server connected");
      _attachPairingListeners();
      loadGateway();
      loadSensors();
      notifyListeners();
    } catch (e) {
      _state = ConnState.failed;
      log("Failed to connect to server $e");
      notifyListeners();
      if (onError != null) {
        onError!();
      }
    }
  }

  Future<void> close() async {
    if (_socket != null) {
      await _socket!.close();
    }
  }

  void _attachPairingListeners() {
    // Attempted pairing with a sensor already paired
    on("REQUEST-SENSOR-EXISTS", (mac, _) {
      showMessage(
          "Pairing request for already paired sensor $mac. First \"Forget\" sensor $mac before pairing again.");
      notifyListeners();
      return false;
    });

    on("REQUEST-TIMEOUT", (mac, _) {
      if (pairingWith != mac) {
        sensorsNearby.remove(mac);
        if (pairingWith == mac) {
          pairingWith = "";
        }
      }
      notifyListeners();
      return false;
    });
    // New pairing request, called on connect for new devices
    on("REQUEST-NEW", (mac, _) {
      log("REQUEST-NEW Called");
      sensorsNearby.add(mac);
      notifyListeners();
      return false;
    });
    on("PAIR-LIST", (devices, _) {
      // Server returns list of devices pending pairing
      sensorsNearby.clear();
      sensorsNearby.addAll(List<String>.from(jsonDecode(devices)));
      notifyListeners();
      // return false to keep in event callback
      return false;
    });
    // BLE agent finished pairing
    on("PAIR-SUCCESS", (mac, _) {
      sensorsNearby.remove(mac);
      if (pairingWith == mac) {
        pairingWith = "";
      }
      showMessage("Sensor $mac has been paired with the gateway");
      loadSensors();
      notifyListeners();
      return false;
    });
    // Pairing disabled, sent by server in response when toggle switch is off
    on("PAIRING-DISABLED", (_, __) {
      _pairingEnabled = false;
      notifyListeners();
      return false;
    });
    // Attempted to start pairing with a device not pending pairing
    on("REQUEST-NOT-FOUND", (mac, _) {
      sensorsNearby.remove(mac);
      if (pairingWith == mac) {
        pairingWith = "";
      }
      notifyListeners();
      return false;
    });
    // Cancelled pairing, either by timeout, BLE agent, or user
    on("PAIRING-CANCELED", (mac, _) {
      if (pairingWith == mac) {
        sensorsNearby.remove(mac);
        pairingWith = "";
      }
      notifyListeners();
      return false;
    });
    // Start pairing with device
    on("PAIRING-WITH", (mac, _) {
      pairingWith = mac;
      notifyListeners();
      return false;
    });
    // Pairing timeout
    on("PAIRING-TIMEOUT", (mac, _) {
      if (pairingWith == mac) {
        sensorsNearby.remove(mac);
        pairingWith = "";
      }
      showMessage("Pairing timed out with sensor $mac");
      notifyListeners();
      return false;
    });

    log("Server event listeners connected");
  }

  void setPairingState(bool enabled) {
    if (enabled) {
      on("PAIR-ENABLE", (_, __) {
        log("Pairing enabled");
        _pairingEnabled = enabled;
        // refresh state
        notifyListeners();
        // Clear this callback
        return false;
      });
    } else {
      on("PAIR-DISABLE", (_, __) {
        log("Pairing disabled");
        _pairingEnabled = enabled;
        // Clear state
        pairingWith = "";
        sensorsNearby.clear();
        notifyListeners();
        return false;
      });
    }
    if (enabled) {
      send("PAIR-ENABLE");
    } else {
      send("PAIR-DISABLE");
    }
    // Then refresh everything lol
    loadSensors();
  }

  void onPairEnable(ConnectionCallback callback) {}

  void loadGateway() {
    on("GET-GATEWAY", (json, err) {
      if (err != null) {
        log("Failed to load gateway");
      }
      try {
        gateway = Gateway.fromJson(jsonDecode(json));
        notifyListeners();
      } catch (e) {
        log("Failed to load gateway: $e");
      }
      return true;
    });
    send("GET-GATEWAY");
  }

  void setGatewayDefault(ConnectionCallback callback) {
    on("SET-GATEWAY-HTTP-ENDPOINT", (verb, err) {
      loadGateway();
      return callback(verb, err);
    });
    send("SET-GATEWAY-HTTP-ENDPOINT default");
  }

  void setGatewayID(String id, ConnectionCallback callback) {
    on("SET-GATEWAY-ID", (verb, err) {
      loadGateway();
      return callback(verb, err);
    });
    send("SET-GATEWAY-ID $id");
  }

  void setGatewayPassword(String password, ConnectionCallback callback) {
    on("SET-GATEWAY-PASSWORD", (verb, err) {
      loadGateway();
      return callback(verb, err);
    });
    send("SET-GATEWAY-PASSWORD $password");
  }

  void setGatewayHttpEndpoint(String endpoint, ConnectionCallback callback) {
    on("SET-GATEWAY-HTTP-ENDPOINT", (verb, err) {
      loadGateway();
      return callback(verb, err);
    });
    send("SET-GATEWAY-HTTP-ENDPOINT $endpoint");
  }

  // Send pair acceptance command
  void acceptPairing(String mac) {
    pairingWith = mac;
    on("PAIR-ACCEPT", (_, err) {
      if (err == null) {
        // Clear current pair mac
        pairingWith = "";
        // TODO  and refresh list
        notifyListeners();
      }

      return true;
    });
    send("PAIR-ACCEPT $mac");
    notifyListeners();
  }

  void loadSensors() async {
    on("LIST", (json, err) {
      if (err != null) {
        showMessage("Failed to load sensors");
        notifyListeners();
      }
      try {
        // Deserialize data into a list of sensors
        // Map all objects in map to a Sensor object
        List<dynamic> decode = jsonDecode(json);
        List<Sensor> decodedSensors = decode.map<Sensor>((dynamic s) => Sensor.fromJson(s)).toList();
        sensors.clear();
        sensors.addAll(decodedSensors);
      } catch (e) {
        showMessage("Failed to load sensors: $e");
        log(e.toString());
        log(json);
        rethrow;
      } finally {
        notifyListeners();
      }
      return false;
    });
    send("LIST");
    send("PAIR-LIST");
  }

  void collectFromSensor(Sensor sensor) {
    send("COLLECT ${macToString(sensor.mac)}");
  }

  // TODO add exceptions
  void forgetSensor(Sensor sensor, ConnectionCallback callback) {
    on("FORGET", (verb, err) {
      // Clear current sensor
      displayedSensor = null;
      // load all sensors
      loadSensors();
      notifyListeners();
      return callback(verb, err);
    });
    send("FORGET ${macToString(sensor.mac)}");
  }

  void saveSensor(Sensor sensor, ConnectionCallback callback) {
    on("SET-SENSOR-SETTINGS", callback);
    // FIXME what is this
    send("SET-SENSOR-SETTINGS ${macToString(sensor.mac)}"
        " name ${sensor.name.replaceAll(" ", "_")}"
        " wake_up_interval ${sensor.wakeUpInterval}"
        " wake_up_interval_max_offset ${sensor.wakeUpIntervalMaxOffset}"
        " ${sensor.settings.keys.map((k) {
      var s = sensor.settings[k]!;
      return "${k}_active ${s.active}"
          " ${k}_sampling_frequency ${s.samplingFrequency}"
          " ${k}_sampling_duration ${s.samplingDuration}";
    }).join(" ")}");
  }

  void resetSensor(Sensor sensor, ConnectionCallback callback) {
    on("SET-SENSOR-SETTINGS", callback);
    send("SET-SENSOR-SETTINGS ${macToString(sensor.mac)} auto auto");
  }

  Future<void> send(String message) async {
    if (_socket == null) {
      _state = ConnState.failed;
      if (onError != null) {
        onError!();
      }
      return;
    }
    try {
      _socket!.write("$message\x00");
      if (message == "STOP") {
        await _socket!.close();
        _state = ConnState.disabled;
        if (onError != null) {
          onError!();
        }
      }
    } catch (_) {
      _state = ConnState.failed;
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
          log("SERVER: $message");
          if (kDebugMode == true) {}
          List<String> found = [];
          List<String> parts = message.split(":");
          if (parts.length > 1) {
            if (parts[0] == "LOG") {
              if (onLog != null) {
                onLog!(parts.sublist(1).join(":"));
              }
              continue;
            }
            // Get to first part, skip LOG/ERR/OK
            // FIXME make an actual parser, this is unreadable
            String command = parts[1];
            if (_waitingFor.containsKey(command)) {
              // If callback returns true, mark and remove later
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
        _state = ConnState.failed;
        if (onError != null) {
          onError!();
        }
      });
    }
  }

  // Add a callback on receiving a prefixed message from server
  // Callback is automatically removed if function returns true
  void on(String command, bool Function(String, String?) callback) {
    // TODO what if multiple widgets listen to the same command?
    _waitingFor[command] = callback;
  }

  void off(String command) {
    _waitingFor.remove(command);
  }
}
