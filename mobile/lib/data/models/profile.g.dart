// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'profile.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_$ProfileImpl _$$ProfileImplFromJson(Map<String, dynamic> json) =>
    _$ProfileImpl(
      id: json['id'] as String,
      username: json['username'] as String,
      profilePic: json['profilePic'] as String?,
      latitude: (json['latitude'] as num?)?.toDouble() ?? 49.43839,
      longitude: (json['longitude'] as num?)?.toDouble() ?? 1.10160,
      address: json['address'] as String?,
      preference: (json['preference'] as num?)?.toInt() ?? 50,
      role: json['role'] as String? ?? 'user',
    );

Map<String, dynamic> _$$ProfileImplToJson(_$ProfileImpl instance) =>
    <String, dynamic>{
      'id': instance.id,
      'username': instance.username,
      'profilePic': instance.profilePic,
      'latitude': instance.latitude,
      'longitude': instance.longitude,
      'address': instance.address,
      'preference': instance.preference,
      'role': instance.role,
    };
