import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:goalbooze/model/game.dart';
import 'package:goalbooze/service/api_service.dart';

class GamesHistoryNotifier extends Notifier<AsyncValue<List<Game>>> {
  ApiService get _api => ref.read(apiServiceProvider);

  @override
  AsyncValue<List<Game>> build() => const AsyncValue.data([]);

  Future<void> loadGames() async {
    state = const AsyncValue.loading();
    try {
      final games = await _api.getGames();
      state = AsyncValue.data(games);
    } catch (e) {
      state = AsyncValue.error(e, StackTrace.current);
    }
  }

  Future<void> deleteGame(int id) async {
    try {
      await _api.deleteGame(id);
      await loadGames();
    } catch (e) {
      state = AsyncValue.error(e, StackTrace.current);
    }
  }
}

final gamesHistoryProvider =
    NotifierProvider<GamesHistoryNotifier, AsyncValue<List<Game>>>(
        GamesHistoryNotifier.new);
