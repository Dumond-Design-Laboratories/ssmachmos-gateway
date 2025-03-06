import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:ss_machmos_gui/connection.dart';
import 'package:ss_machmos_gui/utils.dart';

class GatewayView extends StatefulWidget {
  const GatewayView({super.key});

  @override
  State<GatewayView> createState() => _GatewayViewState();
}

class _GatewayViewState extends State<GatewayView> {
  // FormState is a flutter class
  final _formKey = GlobalKey<FormState>();

  late TextEditingController _idController;
  late TextEditingController _passwordController;
  late TextEditingController _httpController;

  @override
  void dispose() {
    _idController.dispose();
    _passwordController.dispose();
    _httpController.dispose();
    super.dispose();
  }

  @override
  void initState() {
    // Default values on init state
    Connection conn = context.read<Connection>();
    _idController = TextEditingController(text: conn.gateway.id);
    _passwordController = TextEditingController(text: conn.gateway.password);
    _httpController = TextEditingController(text: conn.gateway.httpEndpoint);
    super.initState();
  }

  void submitGateway() async {
    if (_formKey.currentState!.validate()) {
      Connection conn = context.read<Connection>();
      // Form data is valid, send to backend
      conn.setGatewayID(_idController.text, (_, err) {
        if (err != null) {
          showMessage("Failed to save Gateway ID", context);
        } // else {
        //   showMessage("Gateway ID saved", context,
        //       duration: Duration(seconds: 1));
        // }
        return true;
      });

      conn.setGatewayPassword(_passwordController.text,
          (_, err) {
        if (err != null) {
          showMessage("Failed to save Gateway Password", context);
        } // else {
        //   showMessage("Gateway Password saved", context,
        //       duration: Duration(seconds: 1));
        // }
        return true;
      });

      conn.setGatewayHttpEndpoint(_httpController.text,
          (_, err) {
        if (err != null) {
          showMessage("Failed to save Gateway HTTP Endpoint", context);
        } // else {
        //   showMessage("Gateway HTTP Endpoint saved", context,
        //       duration: Duration(seconds: 1));
        // }
        return true;
      });

      conn.testGateway((_, err) {
          if (err != null) {
            print(err);
            showMessage("Gateway invalid: $err", context);
          }
          return false;
      });
    }
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
              TextFormField(
                  controller: _idController,
                  decoration: const InputDecoration(labelText: "Gateway ID"),
                  validator: (value) {
                    if (value == null || value.isEmpty) {
                      return "Missing gateway login ID";
                    }
                    return null;
                  }),
              const SizedBox(height: 10),
              TextFormField(
                  controller: _passwordController,
                  obscureText: true,
                  enableSuggestions: false,
                  autocorrect: false,
                  decoration:
                      const InputDecoration(labelText: "Gateway password"),
                  validator: (value) {
                    if (value == null || value.isEmpty) {
                      return "Missing gateway login password";
                    }
                    return null;
                  }),
              const SizedBox(height: 10),
              TextFormField(
                  controller: _httpController,
                  decoration: InputDecoration(
                      labelText: "HTTP Endpoint",
                      suffixIcon: IconButton(
                        icon: Icon(Icons.restore),
                        tooltip: "Restore http endpoint to default value",
                        onPressed: () {
                          // Replace endpoint with default
                          _httpController.text =
                              "https://openphm.org/gateway_data";
                        },
                      ))),
              const SizedBox(height: 10),
              Align(
                alignment: Alignment.centerRight,
                // Save settings button
                child: TextButton(
                    onPressed: submitGateway, child: const Text("Save")),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
