import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';
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
  final List<Widget> configs = [];

  @override
  void initState() {
    super.initState();
    Sensor initSensor = context.read<Connection>().displayedSensor!;
    name = TextEditingController(text: initSensor.name);
    deviceActive = initSensor.deviceActive;
    wakeupInterval =
        TextEditingController(text: initSensor.wakeUpInterval.toString());
    wakeupOffset = TextEditingController(
        text: initSensor.wakeUpIntervalMaxOffset.toString());
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
    configs.addAll(sensor.model.sensorsAvailable.entries.map((entry) => entry.value.singlePoint
        ? SensorConfigSingleSample(
            specific: entry.key,
            value: sensor.settings[entry.key]!.active,
            onActivate: (enable) =>
                setState(() => sensor.settings[entry.key]!.active = enable!),
          )
        : SensorConfigMultipleSamples(sensor: sensor, specific: entry.key)));

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
              DividerText("Sensor configuration"),
              Row(spacing: 8, children: [
                Expanded(
                    flex: 2,
                    child: TextFormField(
                        controller: name,
                        decoration: InputDecoration(labelText: "Name"))),
                Expanded(
                    child: CheckboxListTile(
                        contentPadding: EdgeInsets.zero,
                        controlAffinity: ListTileControlAffinity.leading,
                        title: Text("Device active?"),
                        value: deviceActive,
                        onChanged: (value) =>
                            setState(() => deviceActive = value!))),
                TextButton(
                    child: const Text("Collect now"),
                    onPressed: () =>
                        context.read<Connection>().collectFromSensor(sensor)),
              ]),
              Text("Types: ${sensor.types.join(' | ')}"),
              Row(spacing: 8, children: [
                Expanded(
                    child: TextFormField(
                        controller: wakeupInterval,
                        decoration: InputDecoration(
                            labelText: "Wake-Up Interval",
                            suffixText: "seconds"))),
                Expanded(
                    child: TextFormField(
                        controller: wakeupOffset,
                        decoration: InputDecoration(
                            labelText: "Wake-up interval offset",
                            suffixText: "seconds"))),
              ]),
              Text("Predicted next wake-up: ${sensor.predictedWakeupTime}"),
              DividerText("Individual sensor configurations"),
              Text("Memory usage: ${sensor.memoryUsage}/1200"),
              ...configs,
              Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  spacing: 8,
                  children: [
                    TextButton(
                        onPressed: () => onForget(context),
                        child: Text("Forget Sensor")),
                    TextButton(
                        onPressed: () => onSave(context), child: Text("Save")),
                  ])
            ],
          ),
        ),
      ),
    ));
  }
}

class SensorConfigSingleSample extends StatelessWidget {
  final String specific;
  final Function(bool?) onActivate;
  final bool value;

  const SensorConfigSingleSample(
      {super.key,
      required this.specific,
      required this.value,
      required this.onActivate});

  @override
  Widget build(BuildContext context) {
    return Container(
        padding: EdgeInsets.symmetric(horizontal: 8, vertical: 0),
        child: CheckboxListTile(
          controlAffinity: ListTileControlAffinity.leading,
          value: value,
          title: Text(toBeginningOfSentenceCase(specific),
              style: TextStyle(fontWeight: FontWeight.bold)),
          subtitle: Text("Activate on next collection?"),
          onChanged: onActivate,
        ));
  }
}

class SensorConfigMultipleSamples extends StatefulWidget {
  final String specific;
  final Sensor sensor;
  const SensorConfigMultipleSamples(
      {super.key, required this.sensor, required this.specific});

  @override
  State<SensorConfigMultipleSamples> createState() =>
      _SensorConfigMultipleSamplesState();
}

class _SensorConfigMultipleSamplesState
    extends State<SensorConfigMultipleSamples> {
  _SensorConfigMultipleSamplesState();

  @override
  Widget build(BuildContext context) {
    SensorSettings display = widget.sensor.settings[widget.specific]!;
    return Container(
        padding: EdgeInsets.symmetric(horizontal: 8, vertical: 0),
        child: Row(
          //spacing: 16,
          children: [
            Expanded(
              flex: 2,
                child: CheckboxListTile(
                    controlAffinity: ListTileControlAffinity.leading,
                    value: display.active,
                    title: Text(toBeginningOfSentenceCase(widget.specific),
                        style: TextStyle(fontWeight: FontWeight.bold)),
                    subtitle: Text("Activate on next collection?"),
                    // Make sure to use setState to rebuild widgets
                    onChanged: (value) =>
                        setState(() => display.active = value!))),
            Expanded(
                child: FrequencySelectionDropdown(
                    freqs:
                        widget.sensor.samplingFreqsForSetting(widget.specific),
                    onSelected: (val) =>
                        setState(() => display.samplingFrequency = val ?? 0),
                    initialSelection: display.samplingFrequency)),
            Expanded(
                child: TextFormField(
              initialValue: display.samplingDuration.toString(),
              onChanged: (value) => setState(
                  () => display.samplingDuration = int.tryParse(value) ?? 0),
              decoration: InputDecoration(
                  labelText: "Sampling duration", suffixText: "seconds"),
              inputFormatters: [FilteringTextInputFormatter.digitsOnly],
            )),
          ],
        ));
  }
}

class FrequencySelectionDropdown extends StatelessWidget {
  final List<int> freqs;
  final Function(int?) onSelected;
  final int initialSelection;

  const FrequencySelectionDropdown(
      {super.key,
      required this.freqs,
      required this.onSelected,
      required this.initialSelection});

  @override
  Widget build(BuildContext context) {
    return DropdownMenu<int>(
        label: Text("Max detectable frequency"),
        onSelected: onSelected,
        initialSelection: freqs.last,
        dropdownMenuEntries: freqs
            .map((x) => DropdownMenuEntry<int>(value: x, label: "${x/2} Hz"))
            .toList());
  }
}

class DividerText extends StatelessWidget {
  final String label;
  const DividerText(this.label, {super.key});

  @override
  Widget build(BuildContext context) =>
      Row(spacing: 4, crossAxisAlignment: CrossAxisAlignment.center, children: [
        Expanded(child: Divider()),
        Text(label,
            style: TextStyle(
                fontWeight: FontWeight.bold,
                fontSize: MediaQuery.textScalerOf(context).scale(15))),
        Divider(),
        Expanded(child: Divider())
      ]);
}
