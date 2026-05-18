# CLAUDE.md

## Project Overview

Mobile app for motorcyclists — maps, live location sharing, group itineraries, photo posts, social feed, real-time chat, and SOS alerts.

- **Backend:** Go (REST API + WebSocket)
- **Frontend:** Flutter (iOS + Android)
- **Target:** Mobile only

---

## Architecture

```
lib/
├── core/
│   ├── config/          # env, constants, theme
│   ├── router/          # go_router routes + guards
│   ├── network/         # dio client, interceptors, ws client
│   └── utils/           # extensions, helpers
├── data/
│   ├── models/          # freezed + json_serializable DTOs
│   ├── repositories/    # abstract interfaces
│   └── sources/
│       ├── remote/      # API calls (Dio)
│       └── local/       # cache (Hive / SharedPrefs)
├── domain/
│   ├── entities/        # pure business objects
│   └── usecases/        # single-responsibility use cases
├── presentation/
│   ├── providers/       # Riverpod providers
│   └── screens/
│       ├── map/
│       ├── feed/
│       ├── groups/
│       ├── chat/
│       ├── itinerary/
│       └── alert/
└── main.dart
```

---

## Tech Stack

| Domain                    | Package                                              | Version       |
| ------------------------- | ---------------------------------------------------- | ------------- |
| State management          | `riverpod` + `flutter_riverpod`                      | ^2.x          |
| Routing                   | `go_router`                                          | ^14.x         |
| Models                    | `freezed` + `json_serializable`                      | latest        |
| HTTP                      | `dio`                                                | ^5.x          |
| Maps                      | `flutter_map` + `latlong2`                           | ^7.x / ^0.9.x |
| Tiles                     | OSM (default) — Mapbox optionnel                     | —             |
| Geolocation               | `geolocator`                                         | ^13.x         |
| Background tracking       | `flutter_background_geolocation`                     | ^4.x          |
| Auth / Presence / Storage | `supabase_flutter`                                   | ^2.x          |
| Chat WebSocket            | `web_socket_channel`                                 | ^3.x          |
| Push notifications        | `firebase_messaging` + `flutter_local_notifications` | latest        |
| Photo / Camera            | `image_picker` + `cached_network_image`              | latest        |
| Storage (media)           | Supabase Storage                                     | —             |
| Secure storage            | `flutter_secure_storage`                             | ^9.x          |
| Env                       | `envied`                                             | latest        |

---

## Code Conventions

### Models

Always use `freezed`. Generate with `dart run build_runner build`.

```dart
@freezed
class Rider with _$Rider {
  const factory Rider({
    required String id,
    required String username,
    required LatLng position,
    DateTime? lastSeen,
  }) = _Rider;

  factory Rider.fromJson(Map<String, dynamic> json) => _$RiderFromJson(json);
}
```

### Providers

Use `AsyncNotifierProvider` for async state. Never put business logic in widgets.

```dart
@riverpod
class GroupNotifier extends _$GroupNotifier {
  @override
  Future<List<Group>> build() => ref.read(groupRepositoryProvider).getAll();

  Future<void> create(CreateGroupDto dto) async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() =>
      ref.read(groupRepositoryProvider).create(dto),
    );
  }
}
```

### Routing

All routes are named. Use `redirect` for auth guard.

```dart
GoRoute(
  path: '/map',
  name: 'map',
  builder: (ctx, state) => const MapScreen(),
),
```

### Network

- Base URL from `envied` env var, never hardcoded
- Dio interceptor handles JWT refresh automatically
- WebSocket client reconnects with exponential backoff

---

## Key Features & Implementation Notes

### Auth & Password Reset

- Auth entièrement gérée par **Supabase Auth** (JWT, sessions, refresh token)
- Reset mot de passe : flux natif Supabase, zéro code Go nécessaire
  1. `supabase.auth.resetPasswordForEmail(email)` → mail envoyé avec lien signé
  2. Deep link vers l'app configuré dans le dashboard Supabase (`app.yourapp.com/reset`)
  3. Flutter intercepte via `go_router` + `supabase.auth.onAuthStateChange` (event `passwordRecovery`)
  4. `supabase.auth.updateUser(password: newPassword)` → mot de passe mis à jour
- Template du mail et URL de redirect configurés dans Supabase Dashboard → Auth → Email Templates

- `flutter_map` with `TileLayer` (OSM tiles)
- Live rider markers via Supabase Realtime (presence channels)
- Itinerary rendered as `PolylineLayer`
- Cluster markers when zoomed out (`flutter_map_marker_cluster`)

### Live Location

- `geolocator` for foreground
- `flutter_background_geolocation` for background (requires entitlements on iOS)
- Position pushed to Supabase Realtime presence every 5s when active
- User can toggle visibility (stealth mode)

### Groups & Itineraries

- A group has members, a departure point, waypoints, and a destination
- Itinerary stored as ordered `List<LatLng>` serialized to GeoJSON on the backend
- Group admin can lock the route (read-only for members)

### Chat

- Per-group chat room via **Go WebSocket** (`wss://api.yourapp.com/ws/chat/:groupId`)
- Connexion authentifiée : JWT passé en query param à l'upgrade (`?token=...`)
- Messages persistés en DB côté Go, chargement initial paginé via REST (cursor-based)
- Optimistic UI : message affiché immédiatement, confirmé sur ACK du serveur
- Reconnexion automatique avec backoff exponentiel (`web_socket_channel` + retry logic dans le repository)
- Format message (JSON) :
  ```json
  {
    "id": "uuid",
    "group_id": "uuid",
    "sender_id": "uuid",
    "body": "...",
    "sent_at": "ISO8601"
  }
  ```
- Supabase Realtime **non utilisé** pour le chat — réservé à la présence carte et aux alertes SOS

### Photos / Feed

- Upload via `image_picker` → compress → Supabase Storage
- Feed is a paginated list of posts (infinite scroll)
- Posts can be tagged to a group or public

### SOS Alert

- Single-tap button accessible from map screen (FAB, always visible)
- Sends: current GPS position + rider ID + timestamp to backend
- Backend broadcasts to all group members + nearby riders (radius configurable)
- Received alerts displayed as a pulsing marker + push notification
- Alert persists until manually dismissed by sender

---

## Backend Contract (Go API)

- **Base URL:** `$API_BASE_URL` (env)
- **Auth:** Bearer JWT dans le header `Authorization` (REST) ou query param `?token=` (WS upgrade)
- **Chat WebSocket:** `wss://$API_BASE_URL/ws/chat/:groupId` — géré par Go
- **Présence carte / SOS broadcast:** Supabase Realtime (client Flutter connecte directement)
- **Errors:** `{ "error": { "code": "RIDER_NOT_FOUND", "message": "..." } }`
- **Pagination:** `?cursor=<id>&limit=20`

All Go struct field names map 1:1 to `json_serializable` keys (snake_case).

---

## Environment

```
# .env (generated via envied)
API_BASE_URL=https://api.yourapp.com
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_ANON_KEY=...
MAPBOX_TOKEN=...          # optional, if switching from OSM
FCM_SENDER_ID=...
```

Never commit `.env`. Use `envied` with `@Envied(obfuscate: true)`.

---

## Build & Run

```bash
# Install deps
flutter pub get

# Generate code (freezed, json_serializable, envied, riverpod)
dart run build_runner build --delete-conflicting-outputs

# Run dev
flutter run --flavor dev --target lib/main_dev.dart

# Build release
flutter build apk --release --flavor prod
flutter build ipa --release --flavor prod
```

---

## Testing

- **Unit:** `flutter_test` — repositories, use cases, utils
- **Widget:** `flutter_test` + `mocktail` for providers
- **Integration:** `patrol` for E2E on real device

```bash
flutter test
flutter test integration_test/
```

---

## Critical Rules

- Never call Supabase or Dio directly from a widget — always through a repository
- Never hardcode strings visible to the user — use `l10n` (ARB files)
- All `async` calls in providers wrapped in `AsyncValue.guard`
- Background location requires explicit user consent screen before activation
- SOS feature must work offline (queued and sent on reconnect)
- Compress images before upload (`flutter_image_compress`, target < 500 KB)
