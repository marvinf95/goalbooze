import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:goalbooze/provider/game_provider.dart';
import 'package:goalbooze/service/api_service.dart';
import 'package:goalbooze/model/game.dart';
import 'package:goalbooze/model/player.dart';
import 'package:goalbooze/model/event.dart';

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

    notifier.toggleEvent(100);
    expect(container.read(gameProvider).selectedEventIds, [100]);

    notifier.toggleEvent(100);
    expect(container.read(gameProvider).selectedEventIds, isEmpty);

    notifier.toggleEvent(100);
    notifier.toggleEvent(200);
    expect(container.read(gameProvider).selectedEventIds, [100, 200]);
  });

  test('toggleEvent should not exceed 10 events', () {
    final container = ProviderContainer();
    final notifier = container.read(gameProvider.notifier);

    for (var i = 0; i < 11; i++) {
      notifier.toggleEvent(i);
    }

    final state = container.read(gameProvider);
    expect(state.selectedEventIds.length, 10);
  });

  test('reset should clear all state', () {
    final container = ProviderContainer();
    final notifier = container.read(gameProvider.notifier);

    notifier.addPlayer('Alice');
    notifier.toggleEvent(100);
    notifier.reset();

    final state = container.read(gameProvider);
    expect(state.players, isEmpty);
    expect(state.selectedEventIds, isEmpty);
    expect(state.currentGame, isNull);
  });
}
