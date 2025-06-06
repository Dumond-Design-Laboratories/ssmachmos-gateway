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
  String password;
  String httpEndpoint;
  Gateway({required this.id, required this.password, required this.httpEndpoint});

  factory Gateway.fromJson(Map<String, dynamic> json) {
    return Gateway(id: json["id"], password: json["password"], httpEndpoint: json["http_endpoint"]);
  }
}

class SensorStatus {
  String name;
  String address;
  bool connected;
  DateTime _lastSeen;
  String activity;

  SensorStatus(this.name, this.address, this.connected, this._lastSeen, this.activity);
  factory SensorStatus.fromJson(Map<String, dynamic> ss) {
    return SensorStatus(
      ss['name'] ?? "no name",
      ss['address'] ?? "no address",
      ss['connected'],
      DateTime.parse(ss["last_seen"]),
      ss['activity'],
    );
  }

  DateTime get lastSeen => _lastSeen.toLocal();
}

// UI state. Manages connection to gateway and collecting information through a socket connection
// Unix named sockets here, if we're planning on supporting windows we need named pipes too
class Connection with ChangeNotifier {
  final String socketPath = "/tmp/ss_machmos.sock";
  int? _serverPid;
  int? get serverPid => _serverPid;
  // Have a way to pass toast notifications to UI
  // gateway notifies when this is pushed to, and AppRoot pops and displays each
  Queue<String> toastQueue = Queue<String>();
  void showMessage(String message) => toastQueue.add(message);

  // State of connection to backend
  ConnState _state = ConnState.disabled;
  ConnState get state => _state;

  late Gateway gateway;
  late bool gatewayValid = true;

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

  // List of devices in store, last seen and if connected
  //final List<SensorStatus> connectedSensors = [];

  Socket? _socket;
  late Map<String, ConnectionCallback> _waitingFor; // callback should return if we should remove the callback

  // REVIEW: delete these functions?
  void Function(String message)? onLog;
  void Function()? onError;
  List<String> logs = [];

  Connection() {
    _state = ConnState.disabled;
    _waitingFor = {};
    startBackend();
  }

  // Entrypoint to ensure backend is started
  Future<void> startBackend() async {
    startServer();
    _state = ConnState.inprogress;
    try {
      _socket = await Socket.connect(InternetAddress(socketPath, type: InternetAddressType.unix), 0, timeout: Duration(seconds: 8));
      send("PING");
    } catch (e) {
      _socket = null;
      _state = ConnState.failed;
      log(e.toString());
    }

    if (_state == ConnState.failed) {
      log("Failed to create connection");
      return;
    }

    // Server should be up and running, get PID
    _serverPid = null;
    on("PID", (pid, __) {
      log("PID returned: $pid");
      _serverPid = int.parse(pid);
      return false;
    });
    send("PID");
    log("PID sent");

    // Block until PID is returned
    //while(_serverPid == null){};

    // If we make it this far I'm assuming we're connected
    // Start listen loop
    _state = ConnState.connected;
    listen();

    log("Server connected");

    // Attach various listeners to responses
    _attachListeners();
    _attachPairingListeners();

    // Load initial data
    loadGateway();
    loadSensors();

    // Notify provider to rebuild widget
    notifyListeners();
    return;
  }

  void startServer() {
    // Start process, process dies if duplicate
    // Use Process.runSync to start non-interactively and block
    // We later ask for PID and read stdout from that
    Process.runSync(
      kDebugMode ? "${Directory.current.path}/ssmachmos" : "ssmachmos",
      ["serve", "--no-console"],
    );

    Process.start("${Directory.current.path}/monitor_ssmachmos", []).then((proc) {
      proc.stdout.transform(utf8.decoder).listen((l) {
        logs.add(l);
        notifyListeners();
      });
      // proc.stderr.transform(utf8.decoder).listen((l) {
      //   logs.add(l);
      //   notifyListeners();
      // });
    });
  }

  void stopServer() {
    send("STOP");
  }

  Future<void> openConnection() async {
    _state = ConnState.inprogress;

    // Try 5 times
    for (var i = 0; i < 5; i++) {
      try {
        _socket = await Socket.connect(InternetAddress(socketPath, type: InternetAddressType.unix), 0, timeout: Duration(seconds: 8));
      } catch (e) {
        // If socket error, wait a second
        await Future.delayed(Duration(seconds: 1));
        continue;
      }
      // No error, setup state and exit
      _state = ConnState.connected;
      listen();
      log("Server connection created");
      _attachListeners();
      _attachPairingListeners();
      loadGateway();
      loadSensors();
      notifyListeners();
      return;
    }
    // Tried 5 times but no connection created
    _state = ConnState.failed;
    log("Failed to connect to server");
    notifyListeners();
    if (onError != null) {
      onError!();
    }
  }

  Future<void> close() async {
    await send("REMOVE-LOGGER");
    if (_socket != null) {
      await _socket!.close();
    }
  }

  // FIXME: Move this to a separate file, as an extension?
  void _attachListeners() {
    // Attach ourselves to a feed
    send("ADD-LOGGER");

    testGateway((_, __) {
      return false;
    });

    on("UPLOAD-FAILED", (_, __) {
      // An upload failed
      // Get all uploads pending
      send("LIST-PENDING-UPLOADS");
      // TODO: add badges to mark that...
      toastQueue.add("Failed to upload to gateway");
      notifyListeners();
      return false;
    });

    on("LIST-PENDING-UPLOADS", (info, _) {
      Map<String, dynamic> uploads = jsonDecode(info);
      uploads["count"];
      return false;
    });

    on("SENSOR-CONNECTED", (info, _) {
      send("LIST-CONNECTED");
      return false;
    });
    on("SENSOR-DISCONNECTED", (info, _) {
      send("LIST-CONNECTED");
      return false;
    });
    on("SENSOR-UPDATED", (_, __) {
      send("LIST-CONNECTED");
      return false;
    });

    // List actual sensor data
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

    // FIXME: combine this into LIST
    on("LIST-CONNECTED", (info, _) {
      // Clear stored devices
      //connectedSensors.clear();
      // Parse data
      for (dynamic j in jsonDecode(info)) {
        SensorStatus ss = SensorStatus.fromJson(j);
        // find associated sensor and plug in
        for (Sensor sens in sensors) {
          if (macToString(sens.mac) == ss.address) {
            sens.status = ss;
          }
        }
      }
      //connectedSensors.addAll(jsons.map((j) {}));
      notifyListeners();
      return false;
    });
    send("LIST-CONNECTED");
  }

  void _attachPairingListeners() {
    // Attempted pairing with a sensor already paired
    on("REQUEST-SENSOR-EXISTS", (mac, _) {
      showMessage("Pairing request for already paired sensor $mac. First \"Forget\" sensor $mac before pairing again.");
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
    // TODO rename this to PAIR-NEW
    on("REQUEST-NEW", (mac, _) {
      send("PAIR-LIST");
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
      send("PAIR-LIST");
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

  // Load gateway settings from server
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

  // Ask server to reset gateway endpoint to default.
  void setGatewayDefault(ConnectionCallback callback) {
    on("SET-GATEWAY-HTTP-ENDPOINT", (verb, err) {
      loadGateway();
      return callback(verb, err);
    });
    send("SET-GATEWAY-HTTP-ENDPOINT default");
  }

  // Set gateway login ID
  void setGatewayID(String id, ConnectionCallback callback) {
    on("SET-GATEWAY-ID", (verb, err) {
      loadGateway();
      return callback(verb, err);
    });
    send("SET-GATEWAY-ID $id");
  }

  // Set gateway login password
  void setGatewayPassword(String password, ConnectionCallback callback) {
    on("SET-GATEWAY-PASSWORD", (verb, err) {
      loadGateway();
      return callback(verb, err);
    });
    send("SET-GATEWAY-PASSWORD $password");
  }

  // Set gateway endpoint
  void setGatewayHttpEndpoint(String endpoint, ConnectionCallback callback) {
    on("SET-GATEWAY-HTTP-ENDPOINT", (verb, err) {
      loadGateway();
      return callback(verb, err);
    });
    send("SET-GATEWAY-HTTP-ENDPOINT $endpoint");
  }

  // Helper function to set all three parameters at once
  void setGateway(Gateway newgate, ConnectionCallback idCallback, ConnectionCallback passwordCallback, ConnectionCallback endpointCalllback) {
    setGatewayID(newgate.id, idCallback);
    setGatewayPassword(newgate.password, passwordCallback);
    setGatewayHttpEndpoint(newgate.httpEndpoint, endpointCalllback);
  }

  void testGateway(ConnectionCallback callback) {
    on("TEST-GATEWAY", (a, err) {
      gatewayValid = (err == null);
      return callback(a, err);
    });
    send("TEST-GATEWAY");
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

  void loadConnectedSensors() {
    send("LIST-CONNECTED");
  }

  void loadSensors() async {
    send("LIST");
    send("PAIR-LIST");
    loadConnectedSensors();
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
    on("SET-SENSOR-SETTINGS", (a, b) {
      loadSensors();
      return callback(a, b);
    });
    send("SET-SENSOR-SETTINGS ${sensor.sensorSettingsCommand}");
  }

  void resetSensor(Sensor sensor, ConnectionCallback callback) {
    on("SET-SENSOR-SETTINGS", (a, b) {
      loadSensors();
      return callback(a, b);
    });
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
          print("SERVER: $message");
          if (kDebugMode == true) {}
          List<String> found = [];
          List<String> parts = message.split(":");
          if (parts.length > 1) {
            // Get to first part, skip LOG/ERR/OK
            // FIXME make an actual parser, this is unreadable
            String command = parts[1];
            if (_waitingFor.containsKey(command)) {
              // If callback returns true, mark and remove later
              if (_waitingFor[command]!(
                  parts.length > 2 ? parts.sublist(2).join(":") : "", parts[0] == "ERR" ? "Error: ${parts.sublist(1).join(":")}" : null)) {
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
