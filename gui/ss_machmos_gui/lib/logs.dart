import 'package:flutter/material.dart';
import 'package:ss_machmos_gui/connection.dart';

class Logs extends StatelessWidget {
  const Logs({
    super.key,
    required ScrollController logsScrollController,
    required String logs,
    required Connection connection,
  })  : _logsScrollController = logsScrollController,
        _logs = logs,
        _connection = connection;

  final ScrollController _logsScrollController;
  final String _logs;
  final Connection _connection;

  @override
  Widget build(BuildContext context) {
    return Stack(
      children: [
        Container(
          decoration: BoxDecoration(
            border: Border.all(color: Colors.grey, width: 0.5),
            borderRadius: BorderRadius.circular(10),
          ),
          margin: const EdgeInsets.all(32),
          padding: const EdgeInsets.all(16),
          alignment: Alignment.bottomLeft,
          child: SingleChildScrollView(
            controller: _logsScrollController,
            child: Row(
              children: [
                SelectableText(
                  _logs,
                ),
              ],
            ),
          ),
        ),
        Positioned(
          top: 48.5,
          right: 48.5,
          child: TextButton(
            style: const ButtonStyle(
              backgroundColor: WidgetStatePropertyAll(Colors.red),
            ),
            onPressed: () {
              _connection.send("STOP");
            },
            child: const Text("Stop",
                style: TextStyle(
                  color: Colors.white,
                )),
          ),
        ),
      ],
    );
  }
}
