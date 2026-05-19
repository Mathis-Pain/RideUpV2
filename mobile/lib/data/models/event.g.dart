// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'event.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_$EventImpl _$$EventImplFromJson(Map<String, dynamic> json) => _$EventImpl(
  id: json['id'] as String,
  title: json['title'] as String,
  description: json['description'] as String?,
  createdBy: json['created_by'] as String,
  creatorName: json['creator_name'] as String?,
  latitude: (json['latitude'] as num).toDouble(),
  longitude: (json['longitude'] as num).toDouble(),
  address: json['address'] as String?,
  startDatetime: DateTime.parse(json['start_datetime'] as String),
  endDatetime: json['end_datetime'] == null
      ? null
      : DateTime.parse(json['end_datetime'] as String),
  participants: (json['participants'] as num?)?.toInt() ?? 0,
  userJoined: json['user_joined'] as bool? ?? false,
  createdAt: DateTime.parse(json['created_at'] as String),
);

Map<String, dynamic> _$$EventImplToJson(_$EventImpl instance) =>
    <String, dynamic>{
      'id': instance.id,
      'title': instance.title,
      'description': instance.description,
      'created_by': instance.createdBy,
      'creator_name': instance.creatorName,
      'latitude': instance.latitude,
      'longitude': instance.longitude,
      'address': instance.address,
      'start_datetime': instance.startDatetime.toIso8601String(),
      'end_datetime': instance.endDatetime?.toIso8601String(),
      'participants': instance.participants,
      'user_joined': instance.userJoined,
      'created_at': instance.createdAt.toIso8601String(),
    };
