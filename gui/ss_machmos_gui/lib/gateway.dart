import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:ss_machmos_gui/connection.dart';
import 'package:ss_machmos_gui/utils.dart';

class Gateway extends StatefulWidget {
  final Connection connection;

  const Gateway({super.key, required this.connection});

  @override
  State<Gateway> createState() => _GatewayState();
}

class _GatewayState extends State<Gateway> {
  late TextEditingController _idController;
  late TextEditingController _passwordController;
  late TextEditingController _httpController;

  @override
  void initState() {
    super.initState();
    _idController = TextEditingController();
    _passwordController = TextEditingController();
    _httpController = TextEditingController();
    loadGateway();
  }

  void loadGateway() {
    widget.connection.send("GET-GATEWAY");
    widget.connection.on("GET-GATEWAY", (json, err) {
      if (err != null) {
        showMessage("Failed to load gateway", context);
        return true;
      }
      try {
        Map gateway = jsonDecode(json);
        setState(() {
          if (gateway["id"] != null) {
            _idController.text = gateway["id"];
          }
          if (gateway["http_endpoint"] != null) {
            _httpController.text = gateway["http_endpoint"];
          }
        });
        return true;
      } catch (e) {
        showMessage("Failed to load gateway: $e", context);
        return true;
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Center(
      child: SizedBox(
        width: 450,
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text("Gateway ID:"),
                const SizedBox(width: 10),
                Expanded(
                  child: TextField(
                    controller: _idController,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 10),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text("Gateway Password:"),
                const SizedBox(width: 10),
                Expanded(
                  child: TextField(
                    obscureText: true,
                    enableSuggestions: false,
                    autocorrect: false,
                    controller: _passwordController,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 10),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text("HTTP Endpoint:"),
                const SizedBox(width: 10),
                Expanded(
                  child: TextField(
                    controller: _httpController,
                  ),
                ),
                const SizedBox(width: 10),
                IconButton(
                  iconSize: 20,
                  icon: const Icon(Icons.restore),
                  onPressed: () async {
                    await widget.connection
                        .send("SET-GATEWAY-HTTP-ENDPOINT default");
                    widget.connection.on("SET-GATEWAY-HTTP-ENDPOINT", (_, err) {
                      if (err != null) {
                        showMessage(
                            "Failed to save Gateway HTTP Endpoint", context);
                      } else {
                        showMessage("Gateway HTTP Endpoint saved", context);
                      }
                      loadGateway();
                      return true;
                    });
                  },
                ),
              ],
            ),
            const SizedBox(height: 10),
            Align(
              alignment: Alignment.centerRight,
              child: TextButton(
                onPressed: () async {
                  await widget.connection
                      .send("SET-GATEWAY-ID ${_idController.text}");
                  widget.connection.on("SET-GATEWAY-ID", (_, err) {
                    if (err != null) {
                      showMessage("Failed to save Gateway ID", context);
                    } else {
                      showMessage("Gateway ID saved", context);
                    }
                    loadGateway();
                    return true;
                  });
                  if (_passwordController.text.isNotEmpty) {
                    await widget.connection.send(
                        "SET-GATEWAY-PASSWORD ${_passwordController.text}");
                    widget.connection.on("SET-GATEWAY-PASSWORD", (_, err) {
                      if (err != null) {
                        showMessage("Failed to save Gateway Password", context);
                      } else {
                        showMessage("Gateway Password saved", context);
                      }
                      loadGateway();
                      return true;
                    });
                  }
                  await widget.connection.send(
                      "SET-GATEWAY-HTTP-ENDPOINT ${_httpController.text}");
                  widget.connection.on("SET-GATEWAY-HTTP-ENDPOINT", (_, err) {
                    if (err != null) {
                      showMessage(
                          "Failed to save Gateway HTTP Endpoint", context);
                    } else {
                      showMessage("Gateway HTTP Endpoint saved", context);
                    }
                    loadGateway();
                    return true;
                  });
                },
                child: const Text("Save"),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
