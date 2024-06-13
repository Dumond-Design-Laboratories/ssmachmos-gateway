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
            child: SelectableText(
              _logs,
            ),
          ),
        ),
        Positioned(
          top: 48,
          right: 48,
          child: TextButton(
            style: ButtonStyle(
              backgroundColor: const WidgetStatePropertyAll(Colors.red),
              shape: WidgetStatePropertyAll(
                RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(10),
                ),
              ),
            ),
            onPressed: () {
              _connection.send("STOP");
            },
            child: const Padding(
              padding: EdgeInsets.symmetric(horizontal: 4, vertical: 8),
              child: Text("Stop",
                  style: TextStyle(
                    color: Colors.white,
                  )),
            ),
          ),
        ),
      ],
    );
  }
}
