import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:ss_machmos_gui/connection.dart';

class Logs extends StatelessWidget {
  const Logs({
    super.key,
    // required ScrollController logsScrollController,
    required this.logs,
  });
  final List<String> logs;
  static final ScrollController _logsScrollController = ScrollController();

  @override
  Widget build(BuildContext context) {
    //String logs = context.watch<Connection>().logs;
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
          child: SelectionArea(
            child: ListView.builder(
              itemCount: logs.length,
              reverse: true,
              //prototypeItem: Text(logs.last),
              controller: _logsScrollController,
              itemBuilder: (context, index) => Text(logs[logs.length - 1 - index].trim().trim()),
            ),
          ),
        ),
        /*
        Positioned(
          top: 48.5,
          right: 48.5,
          child: TextButton(
            style: const ButtonStyle(
              backgroundColor: WidgetStatePropertyAll(Colors.red),
            ),
            onPressed: () {
              showDialog(
                context: context,
                builder: (BuildContext context) {
                  return AlertDialog(
                    title: const Text(
                      "Stop the server",
                    ),
                    actions: [
                      TextButton(
                        onPressed: () {
                          Navigator.of(context).pop();
                        },
                        child: const Text("Cancel"),
                      ),
                      TextButton(
                        onPressed: () {
                          context.read<Connection>().stopServer();
                          //_connection.send("STOP");
                          Navigator.of(context).pop();
                        },
                        child: const Text("Confirm"),
                      ),
                    ],
                  );
                },
              );
            },
            child: const Text("Stop",
                style: TextStyle(
                  color: Colors.white,
                )),
          ),
        )
        */
      ],
    );
  }
}
