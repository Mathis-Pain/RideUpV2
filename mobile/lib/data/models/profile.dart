import 'package:freezed_annotation/freezed_annotation.dart';

part 'profile.freezed.dart';
part 'profile.g.dart';

@freezed
class Profile with _$Profile {
  const factory Profile({
    required String id,
    required String username,
    String? profilePic,
    @Default(49.43839) double latitude,
    @Default(1.10160) double longitude,
    String? address,
    @Default(50) int preference,
    @Default('user') String role,
  }) = _Profile;

  factory Profile.fromJson(Map<String, dynamic> json) =>
      _$ProfileFromJson(json);
}

extension ProfileX on Profile {
  bool get isAdmin => role == 'admin';
}
