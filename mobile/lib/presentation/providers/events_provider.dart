import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:rideup/data/models/event.dart';
import 'package:rideup/data/repositories/event_repository.dart';

part 'events_provider.g.dart';

@riverpod
class EventsNotifier extends _$EventsNotifier {
  @override
  Future<List<Event>> build() =>
      ref.read(eventRepositoryProvider).getEvents();

  Future<void> join(String eventId) async {
    await ref.read(eventRepositoryProvider).joinEvent(eventId);
    ref.invalidateSelf();
  }

  Future<void> leave(String eventId) async {
    await ref.read(eventRepositoryProvider).leaveEvent(eventId);
    ref.invalidateSelf();
  }

  Future<void> delete(String eventId) async {
    await ref.read(eventRepositoryProvider).deleteEvent(eventId);
    ref.invalidateSelf();
  }
}
