import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:goalbooze/model/athlete.dart';
import 'package:goalbooze/model/event.dart';
import 'package:goalbooze/model/game.dart';
import 'package:goalbooze/model/player.dart';
import 'package:goalbooze/service/api_service.dart';

class GameState {
  final List<Player> players;
  final List<SportEvent> selectedEvents;
  final Map<int, List<Athlete>> homeLineups;
  final Map<int, List<Athlete>> awayLineups;
  final Game? currentGame;
  final bool isLoading;
  final String? error;

  const GameState({
    this.players = const [],
    this.selectedEvents = const [],
    this.homeLineups = const {},
    this.awayLineups = const {},
    this.currentGame,
    this.isLoading = false,
    this.error,
  });

  GameState copyWith({
    List<Player>? players,
    List<SportEvent>? selectedEvents,
    Map<int, List<Athlete>>? homeLineups,
    Map<int, List<Athlete>>? awayLineups,
    Game? currentGame,
    bool? isLoading,
    String? error,
    bool clearError = false,
  }) {
    return GameState(
      players: players ?? this.players,
      selectedEvents: selectedEvents ?? this.selectedEvents,
      homeLineups: homeLineups ?? this.homeLineups,
      awayLineups: awayLineups ?? this.awayLineups,
      currentGame: currentGame ?? this.currentGame,
      isLoading: isLoading ?? this.isLoading,
      error: clearError ? null : (error ?? this.error),
    );
  }

  List<int> get selectedEventIds => selectedEvents.map((e) => e.id).toList();
}

class GameNotifier extends StateNotifier<GameState> {
  final ApiService _api;

  GameNotifier(this._api) : super(const GameState());

  void addPlayer(String name) {
    if (state.players.length >= 10) return;
    final trimmed = name.trim();
    if (trimmed.isEmpty) return;
    if (state.players.any((p) => p.name == trimmed)) return;
    state = state.copyWith(players: [...state.players, Player(name: trimmed)]);
  }

  void removePlayer(String name) {
    state = state.copyWith(
      players: state.players.where((p) => p.name != name).toList(),
    );
  }

  void toggleEvent(SportEvent event) {
    final current = [...state.selectedEvents];
    final exists = current.any((e) => e.id == event.id);
    if (exists) {
      current.removeWhere((e) => e.id == event.id);
    } else {
      if (current.length >= 10) return;
      current.add(event);
    }
    state = state.copyWith(selectedEvents: current);
  }

  void setLineup(int eventId, List<Athlete> home, List<Athlete> away) {
    final newHome = Map<int, List<Athlete>>.from(state.homeLineups);
    final newAway = Map<int, List<Athlete>>.from(state.awayLineups);
    newHome[eventId] = home;
    newAway[eventId] = away;
    state = state.copyWith(homeLineups: newHome, awayLineups: newAway);
  }

  /// Builds a fully manual match (self-entered teams + athlete names) as the
  /// single selected event. Athlete names are turned into [Athlete]s; positions
  /// and numbers are left empty. Used by the manual setup mode.
  static const manualEventId = -1;

  void setManualMatch({
    required String homeTeam,
    required String awayTeam,
    required List<String> homeAthletes,
    required List<String> awayAthletes,
  }) {
    List<Athlete> toAthletes(List<String> names, String team) {
      final cleaned = names.map((n) => n.trim()).where((n) => n.isNotEmpty);
      var i = 0;
      return cleaned
          .map((n) => Athlete(
                id: manualEventId * 1000 - (i++),
                name: n,
                number: 0,
                position: '',
                team: team,
              ))
          .toList();
    }

    final event = SportEvent(
      id: manualEventId,
      leagueId: 0,
      homeTeam: homeTeam.trim(),
      awayTeam: awayTeam.trim(),
      date: DateTime.now(),
      status: 'manual',
    );

    state = state.copyWith(
      selectedEvents: [event],
      homeLineups: {manualEventId: toAthletes(homeAthletes, homeTeam.trim())},
      awayLineups: {manualEventId: toAthletes(awayAthletes, awayTeam.trim())},
    );
  }

  Future<void> createGame({bool manual = false}) async {
    state = state.copyWith(isLoading: true, clearError: true);
    try {
      final game = await _api.createGame(
        players: state.players,
        events: state.selectedEvents,
        homeLineups: state.homeLineups.isEmpty ? null : state.homeLineups,
        awayLineups: state.awayLineups.isEmpty ? null : state.awayLineups,
        manual: manual,
      );
      state = state.copyWith(currentGame: game, isLoading: false);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  void reset() {
    state = const GameState();
  }
}

final gameProvider = StateNotifierProvider<GameNotifier, GameState>((ref) {
  final api = ref.read(apiServiceProvider);
  return GameNotifier(api);
});
