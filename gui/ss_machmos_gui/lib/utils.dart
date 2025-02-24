import 'package:flutter/material.dart';

void showMessage(String message, BuildContext context, {Duration duration = const Duration(seconds: 4)}) {
  ScaffoldMessenger.of(context).showSnackBar(
    SnackBar(
      content: Text(message),
      duration: duration,
    ),
  );
}
