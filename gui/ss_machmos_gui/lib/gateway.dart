import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:ss_machmos_gui/connection.dart';
import 'package:ss_machmos_gui/help.dart';
import 'package:ss_machmos_gui/utils.dart';

class GatewayView extends StatefulWidget {
  //final Connection connection;
  // final TabController tabController;
  // final GlobalKey idKey;
  // final GlobalKey httpEndpointKey;

  const GatewayView({
    super.key,
    //required this.connection,
    // required this.tabController,
    // required this.idKey,
    // required this.httpEndpointKey,
  });

  @override
  State<GatewayView> createState() => _GatewayViewState();
}

class _GatewayViewState extends State<GatewayView> {
  // FormState is a flutter class
  final _formKey = GlobalKey<FormState>();

  final TextEditingController _idController = TextEditingController();
  final TextEditingController _passwordController = TextEditingController();
  final TextEditingController _httpController = TextEditingController();

  @override
  void initState() {
    super.initState();
  }

  @override
  void dispose() {
    _idController.dispose();
    _passwordController.dispose();
    _httpController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Form(
        key: _formKey,
        child: Center(
          child: SizedBox(
            width: 450,
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    // HelpButton(
                    //   tabController: widget.tabController,
                    //   page: widget.idKey,
                    // ),
                    // const SizedBox(width: 10),
                    const Text("Gateway ID:"),
                    const SizedBox(width: 10),
                    Expanded(
                      child: TextFormField(controller: _idController),
                    ),
                  ],
                ),
                const SizedBox(height: 10),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    // HelpButton(
                    //   tabController: widget.tabController,
                    //   page: widget.idKey,
                    // ),
                    // const SizedBox(width: 10),
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
                    // HelpButton(
                    //   tabController: widget.tabController,
                    //   page: widget.httpEndpointKey,
                    // ),
                    // const SizedBox(width: 10),
                    const Text("HTTP Endpoint:"),
                    const SizedBox(width: 10),
                    Expanded(
                      child: TextField(
                        controller: _httpController,
                      ),
                    ),
                    const SizedBox(width: 5),
                    IconButton(
                      iconSize: 20,
                      // Reset gateway button
                      icon: const Icon(Icons.restore),
                      onPressed: () async {
                        context.read<Connection>().setGatewayDefault((_, err) {
                          if (err != null) {
                            showMessage("Failed to save Gateway HTTP Endpoint",
                                context);
                          } else {
                            showMessage("Gateway HTTP Endpoint saved", context);
                          }
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
                    // Save settings button
                    child: const Text("Save"),
                    onPressed: () {
                      if (_idController.text.isNotEmpty) {
                        context
                            .read<Connection>()
                            .setGatewayID(_idController.text, (_, err) {
                          if (err != null) {
                            showMessage("Failed to save Gateway ID", context);
                          } else {
                            showMessage("Gateway ID saved", context);
                          }
                          return true;
                        });
                      }

                      if (_passwordController.text.isNotEmpty) {
                        context.read<Connection>().setGatewayPassword(
                            _passwordController.text, (_, err) {
                          if (err != null) {
                            showMessage(
                                "Failed to save Gateway Password", context);
                          } else {
                            showMessage("Gateway Password saved", context);
                          }
                          return true;
                        });
                      }

                      if (_httpController.text.isNotEmpty) {
                        context.read<Connection>().setGatewayHttpEndpoint(
                            _httpController.text, (_, err) {
                          if (err != null) {
                            showMessage("Failed to save Gateway HTTP Endpoint",
                                context);
                          } else {
                            showMessage("Gateway HTTP Endpoint saved", context);
                          }
                          return true;
                        });
                      }
                    },
                  ),
                ),
              ],
            ),
          ),
        ));
  }
}
