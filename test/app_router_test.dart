import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:goalbooze/router/app_router.dart';

/// Navigation had no coverage at all, so a go_router major could only be
/// validated by hand. These exercise the real route table: the ShellRoute
/// chrome, navigation between shell children, and path parameter parsing.
Widget _app(GoRouter router) => ProviderScope(
      child: MaterialApp.router(routerConfig: router),
    );

void main() {
  group('AppRouter', () {
    testWidgets('initial location renders the shell with its bottom nav',
        (tester) async {
      await tester.pumpWidget(_app(AppRouter().router));
      // HomeScreen kicks off a data load, so settle rather than a bare pump
      // to avoid leaving a pending timer behind.
      await tester.pumpAndSettle();

      expect(find.text('Spiel'), findsOneWidget);
      expect(find.text('Regeln'), findsOneWidget);
      expect(find.text('Verlauf'), findsOneWidget);
    });

    testWidgets('go("/rules") renders RulesScreen and keeps the shell',
        (tester) async {
      final router = AppRouter().router;
      await tester.pumpWidget(_app(router));

      router.go('/rules');
      await tester.pumpAndSettle();

      expect(find.text('Zuweisung'), findsOneWidget);
      expect(find.text('Tor!'), findsOneWidget);
      // Shell chrome must survive the navigation.
      expect(find.byType(BottomNavigationBar), findsOneWidget);
    });

    testWidgets('go("/legal") renders the legal notice screen',
        (tester) async {
      final router = AppRouter().router;
      await tester.pumpWidget(_app(router));

      router.go('/legal');
      await tester.pumpAndSettle();

      expect(find.text('Impressum'), findsWidgets);
    });

    testWidgets('tapping the bottom nav switches shell children',
        (tester) async {
      await tester.pumpWidget(_app(AppRouter().router));
      await tester.pump();

      await tester.tap(find.text('Regeln'));
      await tester.pumpAndSettle();

      expect(find.text('Tor!'), findsOneWidget);
    });

    testWidgets('detail routes parse their :id path parameter',
        (tester) async {
      final seen = <String>[];
      final router = GoRouter(
        initialLocation: '/',
        routes: [
          GoRoute(path: '/', builder: (_, _) => const SizedBox.shrink()),
          // Mirrors the real '/history/:id' and '/game/:id' shape without
          // pulling in the screens' provider and API dependencies.
          GoRoute(
            path: '/history/:id',
            builder: (context, state) {
              seen.add(state.pathParameters['id']!);
              return const SizedBox.shrink();
            },
          ),
        ],
      );
      await tester.pumpWidget(_app(router));

      router.go('/history/42');
      await tester.pumpAndSettle();

      expect(seen, ['42']);
    });
  });
}
