import 'package:flutter/material.dart';
import 'package:ss_machmos_gui/connection.dart';

class Gateway extends StatefulWidget {
  final Connection connection;

  const Gateway({super.key, required this.connection});

  @override
  State<Gateway> createState() => _GatewayState();
}

class _GatewayState extends State<Gateway> {
  late TextEditingController _idController;
  late TextEditingController _passwordController;

  @override
  void initState() {
    super.initState();
    _idController = TextEditingController();
    _passwordController = TextEditingController();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Row(
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
        Row(
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
        TextButton(
          onPressed: () async {
            await widget.connection
                .send("SET-GATEWAY-ID ${_idController.text}");
            await widget.connection
                .send("SET-GATEWAY-PASSWORD ${_passwordController.text}");
            widget.connection.on("OK:SET-GATEWAY-ID", () {
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(
                  content: Text("Gateway ID saved."),
                ),
              );
            });
            widget.connection.on("OK:SET-GATEWAY-PASSWORD", () {
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(
                  content: Text("Gateway password saved."),
                ),
              );
            });
          },
          child: const Text("Save"),
        ),
      ],
    );
  }
}
