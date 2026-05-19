import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:supabase_flutter/supabase_flutter.dart';
import 'package:rideup/data/models/event.dart';
import 'package:rideup/data/sources/remote/supabase_event_source.dart';

part 'event_repository.g.dart';

@riverpod
SupabaseEventSource eventSource(Ref ref) =>
    SupabaseEventSource(Supabase.instance.client);

@riverpod
EventRepository eventRepository(Ref ref) =>
    EventRepository(ref.watch(eventSourceProvider));

class EventRepository {
  final SupabaseEventSource _source;

  EventRepository(this._source);

  Future<List<Event>> getEvents() => _source.getEvents();

  Future<void> createEvent({
    required String title,
    String? description,
    required double latitude,
    required double longitude,
    String? address,
    required DateTime startDatetime,
    DateTime? endDatetime,
  }) =>
      _source.createEvent(
        title: title,
        description: description,
        latitude: latitude,
        longitude: longitude,
        address: address,
        startDatetime: startDatetime,
        endDatetime: endDatetime,
      );

  Future<void> joinEvent(String eventId) => _source.joinEvent(eventId);
  Future<void> leaveEvent(String eventId) => _source.leaveEvent(eventId);
  Future<void> deleteEvent(String eventId) => _source.deleteEvent(eventId);
}
