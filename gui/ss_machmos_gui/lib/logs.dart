import 'package:flutter/material.dart';

class Logs extends StatelessWidget {
  const Logs({
    super.key,
    required ScrollController logsScrollController,
    required String logs,
  })  : _logsScrollController = logsScrollController,
        _logs = logs;

  final ScrollController _logsScrollController;
  final String _logs;

  @override
  Widget build(BuildContext context) {
    return Container(
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
    );
  }
}
