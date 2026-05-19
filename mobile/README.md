# RideUp — App mobile Flutter

App mobile pour motards : carte d'events, live location, chat de groupe, alertes SOS.

- **Frontend :** Flutter (iOS + Android)
- **Backend :** Go REST API (racine du repo)
- **DB / Auth :** Supabase (Postgres + Auth JWT + Realtime)

---

## Prérequis

- Flutter 3.44+ (`flutter --version`)
- Dart 3.12+
- Un compte [Supabase](https://supabase.com) (gratuit)
- Xcode (iOS) ou Android Studio (Android)

---

## 1. Créer le projet Supabase

1. Va sur [supabase.com](https://supabase.com) → **New project**
2. Note les valeurs suivantes (disponibles dans **Settings → API**) :
   - `Project URL` → ex. `https://abcdefgh.supabase.co`
   - `anon public key` → clé JWT publique

---

## 2. Initialiser la base de données

Dans le dashboard Supabase → **SQL Editor** → colle et exécute le contenu de :

```
supabase/schema.sql
```

Ce script crée :
- `profiles` — extension de `auth.users` (username, position, préférences)
- `events` — events moto (titre, coords, date, participants)
- `event_participants` — qui a rejoint quoi
- `messages` — chat par event
- `notifications` — notifications utilisateur
- RLS activé sur toutes les tables
- Trigger : profil créé automatiquement à l'inscription
- Trigger : compteur `participants` mis à jour automatiquement

---

## 3. Configurer les variables d'environnement

Les credentials Supabase sont passés à la compilation via `--dart-define`.

### Option A — VS Code (recommandé)

Crée `.vscode/launch.json` à la racine de `mobile/` :

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "RideUp (dev)",
      "request": "launch",
      "type": "dart",
      "args": [
        "--dart-define=SUPABASE_URL=https://TON_ID.supabase.co",
        "--dart-define=SUPABASE_ANON_KEY=ta_clé_anon"
      ]
    }
  ]
}
```

### Option B — Ligne de commande

```bash
flutter run \
  --dart-define=SUPABASE_URL=https://TON_ID.supabase.co \
  --dart-define=SUPABASE_ANON_KEY=ta_clé_anon
```

> Ne jamais commit les credentials. `.vscode/launch.json` est dans `.gitignore`.

---

## 4. Installer les dépendances

```bash
cd mobile
flutter pub get
```

---

## 5. Générer le code (freezed, riverpod, json_serializable)

À faire après `flutter pub get` et après chaque modification de modèles/providers :

```bash
dart run build_runner build --delete-conflicting-outputs
```

---

## 6. Lancer l'app

```bash
# iOS (simulateur)
flutter run --dart-define=SUPABASE_URL=... --dart-define=SUPABASE_ANON_KEY=...

# Android (émulateur ou device)
flutter run --dart-define=SUPABASE_URL=... --dart-define=SUPABASE_ANON_KEY=...

# Lister les devices disponibles
flutter devices
```

---

## 7. Permissions natives (obligatoires avant build)

### iOS — `ios/Runner/Info.plist`

Ajouter :

```xml
<key>NSLocationWhenInUseUsageDescription</key>
<string>RideUp utilise ta position pour afficher ta localisation sur la carte.</string>
<key>NSLocationAlwaysAndWhenInUseUsageDescription</key>
<string>RideUp utilise ta position en arrière-plan pour le partage de localisation live.</string>
```

### Android — `android/app/src/main/AndroidManifest.xml`

Ajouter dans `<manifest>` :

```xml
<uses-permission android:name="android.permission.ACCESS_FINE_LOCATION" />
<uses-permission android:name="android.permission.ACCESS_COARSE_LOCATION" />
<uses-permission android:name="android.permission.INTERNET" />
```

---

## 8. Build release

```bash
# Android APK
flutter build apk --release \
  --dart-define=SUPABASE_URL=... \
  --dart-define=SUPABASE_ANON_KEY=...

# iOS IPA
flutter build ipa --release \
  --dart-define=SUPABASE_URL=... \
  --dart-define=SUPABASE_ANON_KEY=...
```

---

## Structure du projet

```
lib/
├── core/
│   ├── config/        env.dart, theme.dart
│   └── router/        router.dart (go_router + auth guard)
├── data/
│   ├── models/        Event, Profile (freezed + json_serializable)
│   ├── repositories/  EventRepository + providers Riverpod
│   └── sources/
│       └── remote/    SupabaseEventSource (CRUD Supabase)
├── domain/            (use cases à venir)
├── presentation/
│   ├── providers/     EventsNotifier
│   └── screens/
│       ├── auth/      LoginScreen, RegisterScreen
│       ├── map/       MapScreen (flutter_map + OSM)
│       ├── events/    CreateEventScreen
│       └── profile/   ProfileScreen
supabase/
└── schema.sql         Schéma complet + RLS + triggers
```

---

## Commandes utiles

```bash
# Vérifier les erreurs
flutter analyze

# Lancer les tests
flutter test

# Voir les dépendances obsolètes
flutter pub outdated

# Mettre à jour les dépendances
flutter pub upgrade
```
