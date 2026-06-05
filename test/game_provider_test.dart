import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:goalbooze/provider/game_provider.dart';
import 'package:goalbooze/model/event.dart';

SportEvent _event(int id) => SportEvent(
      id: id,
      leagueId: 1,
      homeTeam: 'Home $id',
      awayTeam: 'Away $id',
      date: DateTime(2026, 5, 16),
      status: 'scheduled',
    );

void main() {
  test('GameNotifier initial state should have empty players and events', () {
    final container = ProviderContainer();
    final state = container.read(gameProvider);

    expect(state.players, isEmpty);
    expect(state.selectedEventIds, isEmpty);
    expect(state.currentGame, isNull);
    expect(state.isLoading, isFalse);
    expect(state.error, isNull);
  });

  test('addPlayer should add a player', () {
    final container = ProviderContainer();

    container.read(gameProvider.notifier).addPlayer('Alice');
    final state = container.read(gameProvider);

    expect(state.players.length, 1);
    expect(state.players.first.name, 'Alice');
  });

  test('removePlayer should remove a player', () {
    final container = ProviderContainer();

    container.read(gameProvider.notifier).addPlayer('Alice');
    container.read(gameProvider.notifier).addPlayer('Bob');
    container.read(gameProvider.notifier).removePlayer('Alice');

    final state = container.read(gameProvider);
    expect(state.players.length, 1);
    expect(state.players.first.name, 'Bob');
  });

  test('addPlayer should not exceed 10 players', () {
    final container = ProviderContainer();
    final notifier = container.read(gameProvider.notifier);

    for (var i = 0; i < 11; i++) {
      notifier.addPlayer('Player $i');
    }

    final state = container.read(gameProvider);
    expect(state.players.length, 10);
  });

  test('toggleEvent should toggle event selection', () {
    final container = ProviderContainer();
    final notifier = container.read(gameProvider.notifier);

    notifier.toggleEvent(_event(100));
    expect(container.read(gameProvider).selectedEventIds, [100]);

    notifier.toggleEvent(_event(100));
    expect(container.read(gameProvider).selectedEventIds, isEmpty);

    notifier.toggleEvent(_event(100));
    notifier.toggleEvent(_event(200));
    expect(container.read(gameProvider).selectedEventIds, [100, 200]);
  });

  test('toggleEvent should not exceed 10 events', () {
    final container = ProviderContainer();
    final notifier = container.read(gameProvider.notifier);

    for (var i = 0; i < 11; i++) {
      notifier.toggleEvent(_event(i));
    }

    final state = container.read(gameProvider);
    expect(state.selectedEventIds.length, 10);
  });

  test('reset should clear all state', () {
    final container = ProviderContainer();
    final notifier = container.read(gameProvider.notifier);

    notifier.addPlayer('Alice');
    notifier.toggleEvent(_event(100));
    notifier.reset();

    final state = container.read(gameProvider);
    expect(state.players, isEmpty);
    expect(state.selectedEventIds, isEmpty);
    expect(state.currentGame, isNull);
  });

  group('setManualMatch', () {
    test('creates a single manual event with the entered teams', () {
      final container = ProviderContainer();
      final notifier = container.read(gameProvider.notifier);

      notifier.setManualMatch(
        homeTeam: 'Rot',
        awayTeam: 'Blau',
        homeAthletes: ['Anna', 'Ben'],
        awayAthletes: ['Cara'],
      );

      final state = container.read(gameProvider);
      expect(state.selectedEvents, hasLength(1));
      final ev = state.selectedEvents.single;
      expect(ev.id, GameNotifier.manualEventId);
      expect(ev.status, 'manual');
      expect(ev.homeTeam, 'Rot');
      expect(ev.awayTeam, 'Blau');
    });

    test('builds athletes from names and assigns the team', () {
      final container = ProviderContainer();
      final notifier = container.read(gameProvider.notifier);

      notifier.setManualMatch(
        homeTeam: 'Rot',
        awayTeam: 'Blau',
        homeAthletes: ['Anna', 'Ben'],
        awayAthletes: ['Cara'],
      );

      final state = container.read(gameProvider);
      final home = state.homeLineups[GameNotifier.manualEventId]!;
      final away = state.awayLineups[GameNotifier.manualEventId]!;
      expect(home.map((a) => a.name), ['Anna', 'Ben']);
      expect(home.every((a) => a.team == 'Rot'), isTrue);
      expect(away.map((a) => a.name), ['Cara']);
      expect(away.single.team, 'Blau');
    });

    test('filters out empty and whitespace-only names', () {
      final container = ProviderContainer();
      final notifier = container.read(gameProvider.notifier);

      notifier.setManualMatch(
        homeTeam: 'Rot',
        awayTeam: 'Blau',
        homeAthletes: ['Anna', '', '  ', 'Ben'],
        awayAthletes: ['Cara'],
      );

      final home =
          container.read(gameProvider).homeLineups[GameNotifier.manualEventId]!;
      expect(home.map((a) => a.name), ['Anna', 'Ben']);
    });
  });
}
