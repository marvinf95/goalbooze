import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:goalbooze/model/game.dart';
import 'package:goalbooze/model/league.dart';
import 'package:goalbooze/model/event.dart';
import 'package:goalbooze/model/athlete.dart';
import 'package:goalbooze/model/player.dart';
import 'package:goalbooze/model/team.dart';

class ApiService {
  final Dio _dio;

  ApiService({required String baseUrl})
      : _dio = Dio(BaseOptions(
          baseUrl: baseUrl,
          connectTimeout: const Duration(seconds: 10),
          receiveTimeout: const Duration(seconds: 60),
          headers: {'Content-Type': 'application/json'},
        ));

  Future<List<League>> getLeagues() async {
    final response = await _dio.get('/api/v1/leagues');
    return (response.data as List)
        .map((e) => League.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<SportEvent>> getEvents(int leagueId, int season) async {
    final response =
        await _dio.get('/api/v1/leagues/$leagueId/events?season=$season');
    return (response.data as List)
        .map((e) => SportEvent.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<Team>> getTeams(int leagueId, {int? season}) async {
    final s = season ?? _currentSeason();
    final response =
        await _dio.get('/api/v1/leagues/$leagueId/teams?season=$s');
    return (response.data as List)
        .map((e) => Team.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<Athlete>> getSquad(int teamId) async {
    final response = await _dio.get('/api/v1/teams/$teamId/squad');
    return (response.data as List)
        .map((e) => Athlete.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<Map<String, dynamic>> getEventLineup(
    int eventId, {
    required int homeTeamId,
    required int awayTeamId,
    required String homeTeam,
    required String awayTeam,
    String? date,
  }) async {
    final params = {
      'home_team_id': homeTeamId,
      'away_team_id': awayTeamId,
      'home_team': homeTeam,
      'away_team': awayTeam,
      if (date != null) 'date': date,
    };
    final response =
        await _dio.get('/api/v1/events/$eventId/lineup', queryParameters: params);
    final data = response.data as Map<String, dynamic>;
    return {
      'home': (data['home'] as List)
          .map((e) => Athlete.fromJson(e as Map<String, dynamic>))
          .toList(),
      'away': (data['away'] as List)
          .map((e) => Athlete.fromJson(e as Map<String, dynamic>))
          .toList(),
      'is_squad_pick': data['is_squad_pick'] as bool? ?? true,
    };
  }

  Future<Game> createGame({
    required List<Player> players,
    required List<SportEvent> events,
    Map<int, List<Athlete>>? homeLineups,
    Map<int, List<Athlete>>? awayLineups,
  }) async {
    final eventsPayload = events.map((e) {
      final payload = <String, dynamic>{
        'id': e.id,
        'league_id': e.leagueId,
        'home_team': e.homeTeam,
        'home_team_id': e.homeTeamId,
        'away_team': e.awayTeam,
        'away_team_id': e.awayTeamId,
        'date': e.date.toIso8601String(),
      };
      if (homeLineups != null && homeLineups.containsKey(e.id)) {
        payload['home_lineup'] =
            homeLineups[e.id]!.map((a) => a.toJson()).toList();
      }
      if (awayLineups != null && awayLineups.containsKey(e.id)) {
        payload['away_lineup'] =
            awayLineups[e.id]!.map((a) => a.toJson()).toList();
      }
      return payload;
    }).toList();

    final response = await _dio.post('/api/v1/games', data: {
      'players': players.map((p) => p.toJson()).toList(),
      'events': eventsPayload,
    });
    return Game.fromJson(response.data as Map<String, dynamic>);
  }

  Future<List<Game>> getGames() async {
    final response = await _dio.get('/api/v1/games');
    return (response.data as List)
        .map((e) => Game.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<Game> getGame(int id) async {
    final response = await _dio.get('/api/v1/games/$id');
    return Game.fromJson(response.data as Map<String, dynamic>);
  }

  Future<void> deleteGame(int id) async {
    await _dio.delete('/api/v1/games/$id');
  }

  int _currentSeason() {
    final now = DateTime.now();
    return now.month < 7 ? now.year - 1 : now.year;
  }
}

final apiServiceProvider = Provider<ApiService>((ref) {
  const baseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://localhost:8080',
  );
  return ApiService(baseUrl: baseUrl);
});
