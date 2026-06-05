import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:goalbooze/model/game.dart';
import 'package:goalbooze/model/league.dart';
import 'package:goalbooze/model/event.dart';
import 'package:goalbooze/model/athlete.dart';
import 'package:goalbooze/model/player.dart';
import 'package:goalbooze/model/team.dart';

/// Internal league ID for the FIFA World Cup (matches backend config).
const int worldCupLeagueId = 4;

/// The World Cup runs in summer, so its season is the tournament year rather
/// than the club-season heuristic used for domestic leagues.
const int worldCupSeason = 2026;

/// Returns the football-data season to query for a league. Summer tournaments
/// use a fixed tournament year; club leagues use the month-based heuristic.
int seasonForLeague(int leagueId) {
  if (leagueId == worldCupLeagueId) return worldCupSeason;
  final now = DateTime.now();
  return now.month < 7 ? now.year - 1 : now.year;
}

class ApiService {
  final Dio _dio;

  /// [dio] can be injected in tests; production uses a configured client.
  ApiService({required String baseUrl, Dio? dio})
      : _dio = dio ??
            Dio(BaseOptions(
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
    final s = season ?? seasonForLeague(leagueId);
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
    bool manual = false,
  }) async {
    final eventsPayload = events.map((e) {
      final payload = <String, dynamic>{
        'id': e.id,
        'league_id': e.leagueId,
        'home_team': e.homeTeam,
        'home_team_id': e.homeTeamId,
        'away_team': e.awayTeam,
        'away_team_id': e.awayTeamId,
        // Always send UTC so the string carries a timezone (…Z); the Go backend
        // parses strictly as RFC3339, which requires an offset. A local
        // DateTime (e.g. DateTime.now()) would serialize without one and 400.
        'date': e.date.toUtc().toIso8601String(),
        if (manual) 'manual': true,
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
}

final apiServiceProvider = Provider<ApiService>((ref) {
  const baseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://localhost:8080',
  );
  return ApiService(baseUrl: baseUrl);
});
