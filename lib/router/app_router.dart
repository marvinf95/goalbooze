import 'package:go_router/go_router.dart';
import 'package:goalbooze/screen/home_screen.dart';
import 'package:goalbooze/screen/setup_screen.dart';
import 'package:goalbooze/screen/squad_selection_screen.dart';
import 'package:goalbooze/screen/game_overview_screen.dart';
import 'package:goalbooze/screen/rules_screen.dart';
import 'package:goalbooze/screen/history_screen.dart';
import 'package:goalbooze/screen/legal_notice_screen.dart';
import 'package:goalbooze/widget/bottom_nav.dart';

class AppRouter {
  final GoRouter router = GoRouter(
    initialLocation: '/',
    routes: [
      ShellRoute(
        builder: (context, state, child) =>
            ScaffoldWithBottomNav(child: child),
        routes: [
          GoRoute(
            path: '/',
            builder: (context, state) => const HomeScreen(),
          ),
          GoRoute(
            path: '/rules',
            builder: (context, state) => const RulesScreen(),
          ),
          GoRoute(
            path: '/history',
            builder: (context, state) => const HistoryScreen(),
          ),
          GoRoute(
            path: '/history/:id',
            builder: (context, state) => GameOverviewScreen(
              gameId: int.parse(state.pathParameters['id']!),
              readOnly: true,
            ),
          ),
          GoRoute(
            path: '/legal',
            builder: (context, state) => const LegalNoticeScreen(),
          ),
        ],
      ),
      GoRoute(
        path: '/new-game',
        builder: (context, state) => const SetupScreen(),
      ),
      GoRoute(
        path: '/new-game/lineup',
        builder: (context, state) => const SquadSelectionScreen(),
      ),
      GoRoute(
        path: '/game/:id',
        builder: (context, state) => GameOverviewScreen(
          gameId: int.parse(state.pathParameters['id']!),
        ),
      ),
    ],
  );
}
