import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:supabase_flutter/supabase_flutter.dart';

import 'package:rideup/presentation/screens/auth/login_screen.dart';
import 'package:rideup/presentation/screens/auth/register_screen.dart';
import 'package:rideup/presentation/screens/chat/chat_screen.dart';
import 'package:rideup/presentation/screens/events/create_event_screen.dart';
import 'package:rideup/presentation/screens/feed/create_post_screen.dart';
import 'package:rideup/presentation/screens/feed/feed_screen.dart';
import 'package:rideup/presentation/screens/map/map_screen.dart';
import 'package:rideup/presentation/screens/profile/profile_screen.dart';

part 'router.g.dart';

@riverpod
GoRouter router(Ref ref) {
  return GoRouter(
    initialLocation: '/map',
    redirect: (context, state) {
      final session = Supabase.instance.client.auth.currentSession;
      final isAuth = session != null;
      final isAuthRoute = state.matchedLocation == '/login' ||
          state.matchedLocation == '/register';

      if (!isAuth && !isAuthRoute) return '/login';
      if (isAuth && isAuthRoute) return '/map';
      return null;
    },
    routes: [
      GoRoute(
        path: '/login',
        name: 'login',
        builder: (ctx, state) => const LoginScreen(),
      ),
      GoRoute(
        path: '/register',
        name: 'register',
        builder: (ctx, state) => const RegisterScreen(),
      ),
      ShellRoute(
        builder: (ctx, state, child) => _AppShell(child: child),
        routes: [
          GoRoute(
            path: '/map',
            name: 'map',
            builder: (ctx, state) => const MapScreen(),
          ),
          GoRoute(
            path: '/feed',
            name: 'feed',
            builder: (ctx, state) => const FeedScreen(),
          ),
        ],
      ),
      GoRoute(
        path: '/events/new',
        name: 'new-event',
        builder: (ctx, state) => const CreateEventScreen(),
      ),
      GoRoute(
        path: '/feed/new',
        name: 'new-post',
        builder: (ctx, state) => const CreatePostScreen(),
      ),
      GoRoute(
        path: '/profile',
        name: 'profile',
        builder: (ctx, state) => const ProfileScreen(),
      ),
      GoRoute(
        path: '/events/:id/chat',
        name: 'event-chat',
        builder: (ctx, state) => ChatScreen(
          eventId: state.pathParameters['id']!,
          eventTitle: state.uri.queryParameters['title'] ?? 'Chat',
        ),
      ),
    ],
  );
}

class _AppShell extends StatelessWidget {
  const _AppShell({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    final location = GoRouterState.of(context).matchedLocation;
    final currentIndex = location.startsWith('/feed') ? 1 : 0;

    return Scaffold(
      body: child,
      bottomNavigationBar: NavigationBar(
        selectedIndex: currentIndex,
        onDestinationSelected: (i) {
          if (i == 0) context.go('/map');
          if (i == 1) context.go('/feed');
        },
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.map_outlined),
            selectedIcon: Icon(Icons.map),
            label: 'Carte',
          ),
          NavigationDestination(
            icon: Icon(Icons.photo_library_outlined),
            selectedIcon: Icon(Icons.photo_library),
            label: 'Feed',
          ),
        ],
      ),
    );
  }
}
