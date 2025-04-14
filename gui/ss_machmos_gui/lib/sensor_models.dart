// Embedding data as enums
enum SensorModel {
  // To be more precise, the cache partition is 0x2D0000 bytes
  machmo(name: "MachMo", autoMaxCapacity: true, maxCapacity: 0x2D0000, sensorsAvailable: {
    "vibration": (singlePoint: false, samplingFrequencies: [25, 50, 100, 200, 400, 800, 1600, 3200, 6400, 12800, 25600]),
    "audio": (singlePoint: false, samplingFrequencies: [221100]),
    "temperature": (singlePoint: true, samplingFrequencies: []),
    "flux": (singlePoint: true, samplingFrequencies: []),
  }),
  machmomini(name: "MachMoMini", autoMaxCapacity: false, maxCapacity: 2949120, sensorsAvailable: {
    "vibration": (singlePoint: false, samplingFrequencies: [25, 50, 100, 200, 1000, 2000, 4000, 8000, 16000, 32000]),
    "temperature": (singlePoint: true, samplingFrequencies: []),
  }),
  unknown(name: "Unknown", autoMaxCapacity: false, maxCapacity: 0, sensorsAvailable: {});

  const SensorModel({required this.name, required this.autoMaxCapacity, required this.maxCapacity, required this.sensorsAvailable});

  final String name;
  final bool autoMaxCapacity;
  final int maxCapacity;
  final Map<String, ({bool singlePoint, List<int> samplingFrequencies})> sensorsAvailable;

  factory SensorModel.fromString(String str) {
    if (str == "machmo") return SensorModel.machmo;
    if (str == "machmomini") return SensorModel.machmomini;
    return SensorModel.unknown;
  }
}
