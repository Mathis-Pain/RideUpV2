import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:geolocator/geolocator.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import 'package:latlong2/latlong.dart';
import 'package:supabase_flutter/supabase_flutter.dart';

import 'package:rideup/data/models/event.dart';
import 'package:rideup/presentation/providers/events_provider.dart';

class MapScreen extends ConsumerStatefulWidget {
  const MapScreen({super.key});

  @override
  ConsumerState<MapScreen> createState() => _MapScreenState();
}

class _MapScreenState extends ConsumerState<MapScreen> {
  final _mapController = MapController();
  final _supabase = Supabase.instance.client;

  Event? _selectedEvent;
  LatLng? _userPosition;

  // SOS
  List<Map<String, dynamic>> _sosAlerts = [];
  String? _mySosId;
  RealtimeChannel? _sosChannel;
  Timer? _blinkTimer;
  bool _sosVisible = true;

  @override
  void initState() {
    super.initState();
    _blinkTimer = Timer.periodic(const Duration(milliseconds: 700), (_) {
      if (mounted) setState(() => _sosVisible = !_sosVisible);
    });
    _locateUser();
    _loadAndSubscribeSos();
  }

  @override
  void dispose() {
    _blinkTimer?.cancel();
    _mapController.dispose();
    _sosChannel?.unsubscribe();
    super.dispose();
  }

  Future<void> _locateUser() async {
    try {
      var permission = await Geolocator.checkPermission();
      if (permission == LocationPermission.denied) {
        permission = await Geolocator.requestPermission();
      }
      if (permission == LocationPermission.denied ||
          permission == LocationPermission.deniedForever) return;

      final pos = await Geolocator.getCurrentPosition(
        locationSettings: const LocationSettings(accuracy: LocationAccuracy.high),
      );
      final latLng = LatLng(pos.latitude, pos.longitude);
      if (!mounted) return;
      setState(() => _userPosition = latLng);
      _mapController.move(latLng, 13);
    } catch (_) {}
  }

  Future<void> _loadAndSubscribeSos() async {
    final data = await _supabase
        .from('sos_alerts')
        .select('id, user_id, latitude, longitude, profiles(username)')
        .eq('is_active', true);

    if (mounted) setState(() => _sosAlerts = List<Map<String, dynamic>>.from(data));

    _sosChannel = _supabase
        .channel('sos')
        .onPostgresChanges(
          event: PostgresChangeEvent.insert,
          schema: 'public',
          table: 'sos_alerts',
          callback: (payload) {
            if (mounted) setState(() => _sosAlerts.add(payload.newRecord));
          },
        )
        .onPostgresChanges(
          event: PostgresChangeEvent.update,
          schema: 'public',
          table: 'sos_alerts',
          callback: (payload) {
            final updated = payload.newRecord;
            if (updated['is_active'] == false && mounted) {
              setState(() {
                _sosAlerts.removeWhere((a) => a['id'] == updated['id']);
                if (_mySosId == updated['id']) _mySosId = null;
              });
            }
          },
        )
        .subscribe();
  }

  Future<void> _triggerSos() async {
    if (_mySosId != null) {
      final sosId = _mySosId!;
      await _supabase
          .from('sos_alerts')
          .update({'is_active': false}).eq('id', sosId);
      if (mounted) {
        setState(() {
          _sosAlerts.removeWhere((a) => a['id'] == sosId);
          _mySosId = null;
        });
      }
      return;
    }

    final pos = _userPosition ?? await _getCurrentPos();
    if (pos == null) return;

    final userId = _supabase.auth.currentUser?.id;
    if (userId == null) return;

    final result = await _supabase.from('sos_alerts').insert({
      'user_id': userId,
      'latitude': pos.latitude,
      'longitude': pos.longitude,
    }).select().single();

    if (mounted) setState(() => _mySosId = result['id'] as String);
  }

  Future<LatLng?> _getCurrentPos() async {
    try {
      final pos = await Geolocator.getCurrentPosition(
        locationSettings: const LocationSettings(accuracy: LocationAccuracy.high),
      );
      return LatLng(pos.latitude, pos.longitude);
    } catch (_) {
      return null;
    }
  }

  void _onMarkerTap(Event event) {
    setState(() => _selectedEvent = event);
    _mapController.move(event.latLng, 14);
  }

  Color _markerColor(Event event) {
    final userId = _supabase.auth.currentUser?.id;
    if (event.createdBy == userId) return Colors.green;
    if (event.userJoined) return Colors.purple;
    return Colors.red;
  }

  @override
  Widget build(BuildContext context) {
    final eventsAsync = ref.watch(eventsNotifierProvider);
    final isSosActive = _mySosId != null;

    return Scaffold(
      appBar: AppBar(
        title: const Text('RideUp'),
        actions: [
          IconButton(
            icon: const Icon(Icons.person),
            onPressed: () => context.push('/profile'),
          ),
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: () async {
              await _supabase.auth.signOut();
              if (context.mounted) context.go('/login');
            },
          ),
        ],
      ),
      body: eventsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Erreur : $e')),
        data: (events) => Stack(
          children: [
            FlutterMap(
              mapController: _mapController,
              options: const MapOptions(
                initialCenter: LatLng(46.6, 2.3),
                initialZoom: 6,
              ),
              children: [
                TileLayer(
                  urlTemplate: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
                  userAgentPackageName: 'com.rideup.rideup',
                ),
                MarkerLayer(
                  markers: [
                    if (_userPosition != null)
                      Marker(
                        point: _userPosition!,
                        width: 20,
                        height: 20,
                        child: Container(
                          decoration: BoxDecoration(
                            color: Colors.blue,
                            shape: BoxShape.circle,
                            border: Border.all(color: Colors.white, width: 2),
                          ),
                        ),
                      ),
                    if (_sosVisible)
                      ..._sosAlerts.map((sos) => Marker(
                            point: LatLng(
                              (sos['latitude'] as num).toDouble(),
                              (sos['longitude'] as num).toDouble(),
                            ),
                            width: 48,
                            height: 48,
                            child: const Icon(
                              Icons.warning_rounded,
                              color: Colors.orange,
                              size: 48,
                            ),
                          )),
                    ...events.map((event) => Marker(
                          point: event.latLng,
                          width: 36,
                          height: 36,
                          child: GestureDetector(
                            onTap: () => _onMarkerTap(event),
                            child: Icon(
                              Icons.location_pin,
                              color: _markerColor(event),
                              size: 36,
                            ),
                          ),
                        )),
                  ],
                ),
              ],
            ),
            if (_selectedEvent != null)
              _EventPopup(
                event: _selectedEvent!,
                onClose: () => setState(() => _selectedEvent = null),
                onAction: () => setState(() => _selectedEvent = null),
              ),
          ],
        ),
      ),
      floatingActionButton: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          FloatingActionButton(
            heroTag: 'sos',
            backgroundColor: isSosActive ? Colors.orange : Colors.red,
            onPressed: () async {
              final confirm = await showDialog<bool>(
                context: context,
                builder: (ctx) => AlertDialog(
                  title: Text(isSosActive ? 'Annuler l\'alerte SOS ?' : 'Envoyer une alerte SOS ?'),
                  content: Text(isSosActive
                      ? 'Ton alerte sera désactivée.'
                      : 'Ta position sera partagée avec tous les riders à proximité.'),
                  actions: [
                    TextButton(
                      onPressed: () => Navigator.pop(ctx, false),
                      child: const Text('Annuler'),
                    ),
                    FilledButton(
                      style: FilledButton.styleFrom(
                        backgroundColor: isSosActive ? Colors.orange : Colors.red,
                      ),
                      onPressed: () => Navigator.pop(ctx, true),
                      child: Text(isSosActive ? 'Désactiver' : 'SOS'),
                    ),
                  ],
                ),
              );
              if (confirm == true) _triggerSos();
            },
            child: Icon(isSosActive ? Icons.warning_rounded : Icons.sos),
          ),
          const SizedBox(height: 8),
          if (_userPosition != null)
            FloatingActionButton.small(
              heroTag: 'locate',
              onPressed: () => _mapController.move(_userPosition!, 14),
              child: const Icon(Icons.my_location),
            ),
          const SizedBox(height: 8),
          FloatingActionButton.extended(
            heroTag: 'new_event',
            onPressed: () => context.push('/events/new'),
            icon: const Icon(Icons.add),
            label: const Text('Nouvel event'),
          ),
        ],
      ),
    );
  }
}

class _EventPopup extends ConsumerWidget {
  const _EventPopup({
    required this.event,
    required this.onClose,
    required this.onAction,
  });

  final Event event;
  final VoidCallback onClose;
  final VoidCallback onAction;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final userId = Supabase.instance.client.auth.currentUser?.id;
    final isOwner = event.createdBy == userId;
    final fmt = DateFormat('dd/MM/yyyy HH:mm', 'fr_FR');

    return Positioned(
      bottom: 0,
      left: 0,
      right: 0,
      child: Card(
        margin: EdgeInsets.zero,
        shape: const RoundedRectangleBorder(
          borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
        ),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              Row(
                children: [
                  Expanded(
                    child: Text(
                      event.title,
                      style: Theme.of(context).textTheme.titleLarge,
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.close),
                    onPressed: onClose,
                  ),
                ],
              ),
              if (event.description != null) ...[
                const SizedBox(height: 4),
                Text(event.description!),
              ],
              const SizedBox(height: 8),
              Text('📅 ${fmt.format(event.startDatetime)}'),
              if (event.address != null) Text('📍 ${event.address}'),
              if (event.creatorName != null) Text('👤 ${event.creatorName}'),
              Text('👥 ${event.participants} participant(s)'),
              const SizedBox(height: 12),
              if (isOwner || event.userJoined)
                OutlinedButton.icon(
                  onPressed: () => context.push(
                    Uri(
                      path: '/events/${event.id}/chat',
                      queryParameters: {'title': event.title},
                    ).toString(),
                  ),
                  icon: const Icon(Icons.chat_bubble_outline),
                  label: const Text('Chat'),
                ),
              const SizedBox(height: 8),
              if (isOwner)
                FilledButton.tonal(
                  style: FilledButton.styleFrom(
                    backgroundColor: Colors.red.shade800,
                  ),
                  onPressed: () async {
                    final confirm = await showDialog<bool>(
                      context: context,
                      builder: (ctx) => AlertDialog(
                        title: const Text('Supprimer l\'event ?'),
                        actions: [
                          TextButton(
                            onPressed: () => Navigator.pop(ctx, false),
                            child: const Text('Annuler'),
                          ),
                          FilledButton(
                            onPressed: () => Navigator.pop(ctx, true),
                            child: const Text('Supprimer'),
                          ),
                        ],
                      ),
                    );
                    if (confirm == true) {
                      await ref
                          .read(eventsNotifierProvider.notifier)
                          .delete(event.id);
                      onAction();
                    }
                  },
                  child: const Text('Supprimer'),
                )
              else
                FilledButton(
                  onPressed: () async {
                    if (event.userJoined) {
                      await ref
                          .read(eventsNotifierProvider.notifier)
                          .leave(event.id);
                    } else {
                      await ref
                          .read(eventsNotifierProvider.notifier)
                          .join(event.id);
                    }
                    onAction();
                  },
                  child: Text(event.userJoined ? 'Quitter' : 'Rejoindre'),
                ),
            ],
          ),
        ),
      ),
    );
  }
}
