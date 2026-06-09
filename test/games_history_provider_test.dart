import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:goalbooze/model/game.dart';
import 'package:goalbooze/provider/games_history_provider.dart';

void main() {
  test('gamesHistoryProvider builds with an empty data state', () {
    final container = ProviderContainer();
    addTearDown(container.dispose);

    final state = container.read(gamesHistoryProvider);

    expect(state, isA<AsyncData<List<Game>>>());
    expect(state.value, isEmpty);
  });

  test('gamesHistoryProvider exposes its notifier (wiring)', () {
    final container = ProviderContainer();
    addTearDown(container.dispose);

    final notifier = container.read(gamesHistoryProvider.notifier);

    expect(notifier, isA<GamesHistoryNotifier>());
  });
}
