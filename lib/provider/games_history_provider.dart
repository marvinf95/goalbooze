import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:goalbooze/model/game.dart';
import 'package:goalbooze/service/api_service.dart';

class GamesHistoryNotifier extends StateNotifier<AsyncValue<List<Game>>> {
  final ApiService _api;

  GamesHistoryNotifier(this._api) : super(const AsyncValue.data([]));

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
    StateNotifierProvider<GamesHistoryNotifier, AsyncValue<List<Game>>>((ref) {
  final api = ref.read(apiServiceProvider);
  return GamesHistoryNotifier(api);
});
