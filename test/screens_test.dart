import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:goalbooze/screen/rules_screen.dart';
import 'package:goalbooze/screen/legal_notice_screen.dart';

void main() {
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
      expect(find.text('Torwart-Gegentor'), findsOneWidget);
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
      expect(
        find.text('GoalBooze – Das Party-Trinkspiel für Sport-Freunde'),
        findsOneWidget,
      );
    });
  });
}
