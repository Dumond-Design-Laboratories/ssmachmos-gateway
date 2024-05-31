import 'package:flutter/material.dart';

class Gateway extends StatefulWidget {
  const Gateway({super.key});

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
          onPressed: () {},
          child: const Text("Save"),
        ),
      ],
    );
  }
}
