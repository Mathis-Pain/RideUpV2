import 'package:supabase_flutter/supabase_flutter.dart';
import 'package:rideup/data/models/event.dart';

class SupabaseEventSource {
  final SupabaseClient _client;

  SupabaseEventSource(this._client);

  Future<List<Event>> getEvents() async {
    final userId = _client.auth.currentUser?.id;

    final rows = await _client
        .from('events')
        .select('''
          *,
          profiles!created_by(username),
          event_participants!left(user_id)
        ''')
        .order('start_datetime');

    return rows.map<Event>((row) {
      final participants = (row['event_participants'] as List?) ?? [];
      final userJoined = userId != null &&
          participants.any((p) => p['user_id'] == userId);

      return Event.fromJson({
        ...row,
        'creator_name': row['profiles']?['username'],
        'user_joined': userJoined,
      });
    }).toList();
  }

  Future<void> createEvent({
    required String title,
    String? description,
    required double latitude,
    required double longitude,
    String? address,
    required DateTime startDatetime,
    DateTime? endDatetime,
  }) async {
    final userId = _client.auth.currentUser!.id;
    await _client.from('events').insert({
      'title': title,
      'description': description,
      'created_by': userId,
      'latitude': latitude,
      'longitude': longitude,
      'address': address,
      'start_datetime': startDatetime.toIso8601String(),
      'end_datetime': endDatetime?.toIso8601String(),
    });
  }

  Future<void> joinEvent(String eventId) async {
    final userId = _client.auth.currentUser!.id;
    await _client.from('event_participants').insert({
      'event_id': eventId,
      'user_id': userId,
    });
  }

  Future<void> leaveEvent(String eventId) async {
    final userId = _client.auth.currentUser!.id;
    await _client
        .from('event_participants')
        .delete()
        .eq('event_id', eventId)
        .eq('user_id', userId);
  }

  Future<void> deleteEvent(String eventId) async {
    await _client.from('events').delete().eq('id', eventId);
  }
}
