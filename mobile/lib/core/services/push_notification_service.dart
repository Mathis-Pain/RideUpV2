import 'dart:io';

import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import 'package:supabase_flutter/supabase_flutter.dart';

@pragma('vm:entry-point')
Future<void> _onBackgroundMessage(RemoteMessage message) async {}

class PushNotificationService {
  static final _fcm = FirebaseMessaging.instance;
  static final _localNotif = FlutterLocalNotificationsPlugin();

  static Future<void> init() async {
    FirebaseMessaging.onBackgroundMessage(_onBackgroundMessage);

    await _requestPermission();
    await _initLocalNotifications();
    await _saveToken();

    FirebaseMessaging.onMessage.listen(_showForegroundNotification);

    _fcm.onTokenRefresh.listen(_upsertToken);
  }

  static Future<void> _requestPermission() async {
    await _fcm.requestPermission(
      alert: true,
      badge: true,
      sound: true,
    );
  }

  static Future<void> _initLocalNotifications() async {
    const android = AndroidInitializationSettings('@mipmap/ic_launcher');
    const ios = DarwinInitializationSettings();
    await _localNotif.initialize(
      const InitializationSettings(android: android, iOS: ios),
    );

    if (Platform.isAndroid) {
      await _localNotif
          .resolvePlatformSpecificImplementation<
              AndroidFlutterLocalNotificationsPlugin>()
          ?.createNotificationChannel(const AndroidNotificationChannel(
            'rideup_default',
            'RideUp',
            importance: Importance.high,
          ));
    }
  }

  static Future<void> _saveToken() async {
    final token = await _fcm.getToken();
    if (token != null) await _upsertToken(token);
  }

  static Future<void> _upsertToken(String token) async {
    final supabase = Supabase.instance.client;
    final userId = supabase.auth.currentUser?.id;
    if (userId == null) return;

    await supabase.from('device_tokens').upsert({
      'user_id': userId,
      'token': token,
      'platform': Platform.isIOS ? 'ios' : 'android',
      'updated_at': DateTime.now().toIso8601String(),
    }, onConflict: 'user_id');
  }

  static void _showForegroundNotification(RemoteMessage message) {
    final notification = message.notification;
    if (notification == null) return;

    _localNotif.show(
      notification.hashCode,
      notification.title,
      notification.body,
      NotificationDetails(
        android: AndroidNotificationDetails(
          'rideup_default',
          'RideUp',
          importance: Importance.high,
          priority: Priority.high,
        ),
        iOS: const DarwinNotificationDetails(),
      ),
    );
  }

  static Future<void> deleteToken() async {
    final supabase = Supabase.instance.client;
    final userId = supabase.auth.currentUser?.id;
    if (userId == null) return;

    await supabase.from('device_tokens').delete().eq('user_id', userId);
    await _fcm.deleteToken();
    debugPrint('FCM token supprimé');
  }
}
