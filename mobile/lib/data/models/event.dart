import 'package:freezed_annotation/freezed_annotation.dart';
import 'package:json_annotation/json_annotation.dart';
import 'package:latlong2/latlong.dart';

part 'event.freezed.dart';
part 'event.g.dart';

@Freezed()
class Event with _$Event {
  @JsonSerializable(fieldRename: FieldRename.snake)
  const factory Event({
    required String id,
    required String title,
    String? description,
    required String createdBy,
    String? creatorName,
    required double latitude,
    required double longitude,
    String? address,
    required DateTime startDatetime,
    DateTime? endDatetime,
    @Default(0) int participants,
    @Default(false) bool userJoined,
    required DateTime createdAt,
  }) = _Event;

  factory Event.fromJson(Map<String, dynamic> json) => _$EventFromJson(json);
}

extension EventX on Event {
  LatLng get latLng => LatLng(latitude, longitude);
}
