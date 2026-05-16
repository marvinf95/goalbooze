import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:goalbooze/app.dart';
import 'package:goalbooze/screen/home_screen.dart';
import 'package:goalbooze/screen/add_players_screen.dart';
import 'package:goalbooze/screen/rules_screen.dart';
import 'package:goalbooze/screen/legal_notice_screen.dart';

void main() {
  group('HomeScreen', () {
    testWidgets('should display app title and new game button', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(home: Scaffold(body: HomeScreen())),
        ),
      );

      expect(find.text('GoalBooze'), findsOneWidget);
      expect(find.text('Das Party-Trinkspiel für Sport-Freunde'), findsOneWidget);
      expect(find.text('Neues Spiel'), findsOneWidget);
    });
  });

  group('AddPlayersScreen', () {
    testWidgets('should show text field and add button', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(home: Scaffold(body: AddPlayersScreen())),
        ),
      );

      expect(find.text('Spieler hinzufügen'), findsOneWidget);
      expect(find.byType(TextField), findsOneWidget);
      expect(find.text('Noch keine Spieler hinzugefügt'), findsOneWidget);
      expect(find.text('Spieltermine wählen'), findsOneWidget);
    });

    testWidgets('should add player and show in list', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(home: Scaffold(body: AddPlayersScreen())),
        ),
      );

      await tester.enterText(find.byType(TextField), 'Alice');
      await tester.tap(find.text('Hinzufügen'));
      await tester.pump();

      expect(find.text('Alice'), findsOneWidget);
      expect(find.text('Noch keine Spieler hinzugefügt'), findsNothing);
    });

    testWidgets('should remove player from list', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(home: Scaffold(body: AddPlayersScreen())),
        ),
      );

      await tester.enterText(find.byType(TextField), 'Alice');
      await tester.tap(find.text('Hinzufügen'));
      await tester.pump();

      await tester.tap(find.byIcon(Icons.remove_circle_outline));
      await tester.pump();

      expect(find.text('Alice'), findsNothing);
      expect(find.text('Noch keine Spieler hinzugefügt'), findsOneWidget);
    });
  });

  group('RulesScreen', () {
    testWidgets('should show all 5 rules', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(home: Scaffold(body: RulesScreen())),
        ),
      );

      expect(find.text('Zuweisung'), findsOneWidget);
      expect(find.text('Tor!'), findsOneWidget);
      expect(find.text('Rote Karte!'), findsOneWidget);
      expect(find.text('Auswechslung'), findsOneWidget);
      expect(find.text('Torwart-Tor'), findsOneWidget);
    });
  });

  group('LegalNoticeScreen', () {
    testWidgets('should display legal information', (tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: MaterialApp(home: Scaffold(body: LegalNoticeScreen())),
        ),
      );

      expect(find.text('Impressum'), findsNWidgets(2));
      expect(find.text('GoalBooze – Das Party-Trinkspiel für Sport-Freunde'), findsOneWidget);
    });
  });
}
