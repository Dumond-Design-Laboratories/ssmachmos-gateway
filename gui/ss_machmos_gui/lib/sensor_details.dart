import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:ss_machmos_gui/connection.dart';
import 'package:ss_machmos_gui/sensors.dart';
import 'package:ss_machmos_gui/utils.dart';

class SensorDetails extends StatefulWidget {
  const SensorDetails({super.key});

  @override
  State<SensorDetails> createState() => _SensorDetailsState();
}

class _SensorDetailsState extends State<SensorDetails> {
  // NOTE: There's a mix of controllers and provider state management. This is
  // becuase I can't access the state in the subwidgets without specifying
  // everything by hand
  final _formKey = GlobalKey<FormState>();
  late TextEditingController name;
  late bool deviceActive;
  late TextEditingController wakeupInterval;
  late TextEditingController wakeupOffset;
  final List<SensorConfigMultipleSamples> configs = [];

  @override
  void initState() {
    super.initState();
    Sensor initSensor = context.read<Connection>().displayedSensor!;
    name = TextEditingController(text: initSensor.name);
    deviceActive = initSensor.deviceActive;
    wakeupInterval = TextEditingController(text: initSensor.wakeUpInterval.toString());
    wakeupOffset = TextEditingController(text: initSensor.wakeUpIntervalMaxOffset.toString());
  }

  // Clear all to initial
  void onForget(BuildContext context) {
    Connection conn = context.read<Connection>();
    conn.forgetSensor(conn.displayedSensor!, (_, err) {
      if (err == null) {
        showMessage("Sensor removed", context);
      }
      return false;
    });
    return;
  }

  // Reset to defaults
  //void onReset() {}

  // Send settings to backend
  void onSave(BuildContext context) {
    // Verify offsets are integers
    // collect sub sensor settings and verify those are also integers
    if (!_formKey.currentState!.validate()) {
      return;
    }
    Sensor sensor = context.read<Connection>().displayedSensor!;

    sensor.name = name.text;
    sensor.deviceActive = deviceActive;
    sensor.wakeUpInterval = int.tryParse(wakeupInterval.text)!;
    sensor.wakeUpIntervalMaxOffset = int.tryParse(wakeupOffset.text)!;

    Connection conn = context.read<Connection>();
    conn.saveSensor(conn.displayedSensor!, (_, err) {
      if (err == null) {
        showMessage("Successfully saved sensor", context);
      }
      return true;
    });
  }

  @override
  Widget build(BuildContext context) {
    Sensor sensor = context.watch<Connection>().displayedSensor!;
    configs.clear();
    configs.addAll(sensor.settings.keys.map((k) => SensorConfigMultipleSamples(sensor: sensor, specific: k)));

    return Expanded(
        child: SingleChildScrollView(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 8),
        child: Form(
          key: _formKey,
          child: Column(
            spacing: 8,
            //crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(spacing: 8, children: [
                Expanded(flex: 2, child: TextFormField(controller: name, decoration: InputDecoration(labelText: "Name"))),
                // Text(macToString(sensor.mac).toUpperCase()),
                // Expanded(flex: 1, child: Text(sensor.model.string)),
                Expanded(
                    flex: 2,
                    child: CheckboxListTile(
                        contentPadding: EdgeInsets.zero,
                        controlAffinity: ListTileControlAffinity.leading,
                        title: Text("Device active?"),
                        value: deviceActive,
                        onChanged: (value) => setState(() => deviceActive = value!))),
                Spacer(flex: 3),
                TextButton(child: const Text("Collect now"), onPressed: () => context.read<Connection>().collectFromSensor(sensor)),
              ]),
              Text("Types: ${sensor.types.join(', ')}"),
              Row(spacing: 8, children: [
                Expanded(
                    child:
                        TextFormField(controller: wakeupInterval, decoration: InputDecoration(labelText: "Wake-Up Interval", suffixText: "seconds"))),
                Expanded(
                    child: TextFormField(
                        controller: wakeupOffset, decoration: InputDecoration(labelText: "Wake-up interval offset", suffixText: "seconds"))),
              ]),
              Text("Predicted next wake-up: ${sensor.predictedWakeupTime}"),
              //SensorDetailField(name: "Next Wake-Up", value: sensor.nextWakeUp.toLocal().toString(), readOnly: true),
              Divider(),
              const Text("Settings:", style: TextStyle(fontWeight: FontWeight.bold)),
              ...configs,
              Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, spacing: 8, children: [
                TextButton(onPressed: () => onForget(context), child: Text("Forget Sensor")),
                TextButton(onPressed: () => onSave(context), child: Text("Save")),
              ])
            ],
          ),
        ),
      ),
    ));
  }
}

class SensorConfigMultipleSamples extends StatefulWidget {
  final String specific;
  final Sensor sensor;
  const SensorConfigMultipleSamples({super.key, required this.sensor, required this.specific});

  @override
  State<SensorConfigMultipleSamples> createState() => _SensorConfigMultipleSamplesState();
}

class _SensorConfigMultipleSamplesState extends State<SensorConfigMultipleSamples> {
  _SensorConfigMultipleSamplesState();

  @override
  Widget build(BuildContext context) {
    //SensorSettings display = context.read<Connection>().displayedSensor!.settings[widget.specific]!;
    SensorSettings display = widget.sensor.settings[widget.specific]!;
    return Container(
        padding: EdgeInsets.all(8),
        child: Column(
          spacing: 16,
          children: [
            Row(children: [
              Text(widget.specific, style: TextStyle(fontWeight: FontWeight.bold)),
              Expanded(
                  child: CheckboxListTile(
                      controlAffinity: ListTileControlAffinity.leading,
                      value: display.active,
                      title: Text("Activate on next collection?"),
                      // Make sure to use setState to rebuild widgets
                      onChanged: (value) => setState(() => display.active = value!)))
            ]),
            Row(spacing: 8, children: [
              Expanded(
                  child: FrequencySelectionDropdown(
                      freqs: widget.sensor.samplingFreqsForSetting(widget.specific),
                      onSelected: (val) {
                        setState(() => display.samplingFrequency = val);
                      },
                      initialSelection: display.samplingFrequency!)
                  // TextFormField(
                  //   initialValue: display.samplingFrequency.toString(),
                  //   onChanged: (value) => setState(() => display.samplingFrequency = int.tryParse(value) ?? 0),
                  //   decoration: InputDecoration(labelText: "Sampling frequency", suffixText: "Hz"),
                  //   inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                  // ),
                  ),
              Expanded(
                child: TextFormField(
                  initialValue: display.samplingDuration.toString(),
                  onChanged: (value) => setState(() => display.samplingDuration = int.tryParse(value) ?? 0),
                  decoration: InputDecoration(labelText: "Sampling duration", suffixText: "seconds"),
                  inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                ),
              ),
            ]),
            Divider(),
          ],
        ));
  }
}

class FrequencySelectionDropdown extends StatelessWidget {
  final List<int> freqs;
  final Function(int?) onSelected;
  final int initialSelection;

  const FrequencySelectionDropdown({super.key, required this.freqs, required this.onSelected, required this.initialSelection});

  @override
  Widget build(BuildContext context) {
    return DropdownMenu<int>(
        onSelected: onSelected,
        initialSelection: initialSelection,
        dropdownMenuEntries: freqs.map((x) => DropdownMenuEntry<int>(value: x, label: x.toString())).toList());
  }
}
