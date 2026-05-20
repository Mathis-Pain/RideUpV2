import 'package:firebase_core/firebase_core.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:supabase_flutter/supabase_flutter.dart';

import 'package:rideup/core/config/env.dart';
import 'package:rideup/core/config/theme.dart';
import 'package:rideup/core/router/router.dart';
import 'package:rideup/core/services/push_notification_service.dart';
import 'package:rideup/firebase_options.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  await initializeDateFormatting('fr_FR');

  await Firebase.initializeApp(options: DefaultFirebaseOptions.currentPlatform);

  await Supabase.initialize(
    url: Env.supabaseUrl,
    anonKey: Env.supabaseAnonKey,
  );

  // Init push notifications si l'utilisateur est connecté
  if (Supabase.instance.client.auth.currentUser != null) {
    await PushNotificationService.init();
  }

  // Écouter les connexions futures pour init les notifs
  Supabase.instance.client.auth.onAuthStateChange.listen((data) {
    if (data.event == AuthChangeEvent.signedIn) {
      PushNotificationService.init();
    } else if (data.event == AuthChangeEvent.signedOut) {
      PushNotificationService.deleteToken();
    }
  });

  runApp(const ProviderScope(child: RideUpApp()));
}

class RideUpApp extends ConsumerWidget {
  const RideUpApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);
    return MaterialApp.router(
      title: 'RideUp',
      theme: appTheme,
      routerConfig: router,
      debugShowCheckedModeBanner: false,
    );
  }
}
