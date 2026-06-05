import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:goalbooze/model/athlete.dart';
import 'package:goalbooze/model/event.dart';
import 'package:goalbooze/model/player.dart';
import 'package:goalbooze/service/api_service.dart';

/// Builds an ApiService whose Dio short-circuits every request: the outgoing
/// RequestOptions is captured and a canned game response is returned, so no
/// network call happens.
(ApiService, List<RequestOptions> captured) _stubbedService() {
  final captured = <RequestOptions>[];
  final dio = Dio();
  dio.interceptors.add(InterceptorsWrapper(
    onRequest: (options, handler) {
      captured.add(options);
      handler.resolve(Response(
        requestOptions: options,
        statusCode: 201,
        data: {
          'id': 1,
          'created_at': '2026-06-05T12:00:00.000Z',
          'players': [],
          'events': [],
          'assignments': [],
        },
      ));
    },
  ));
  return (ApiService(baseUrl: 'http://test', dio: dio), captured);
}

SportEvent _event() => SportEvent(
      id: -1,
      leagueId: 0,
      homeTeam: 'Rot',
      awayTeam: 'Blau',
      date: DateTime.utc(2026, 6, 5, 18),
      status: 'manual',
    );

Athlete _athlete(String name, String team) =>
    Athlete(id: 1, name: name, number: 0, position: '', team: team);

void main() {
  group('seasonForLeague', () {
    test('World Cup league uses the fixed tournament year', () {
      expect(seasonForLeague(worldCupLeagueId), worldCupSeason);
    });

    test('club leagues use the month-based heuristic', () {
      final now = DateTime.now();
      final expected = now.month < 7 ? now.year - 1 : now.year;
      expect(seasonForLeague(1), expected);
    });
  });

  group('createGame payload', () {
    test('manual game sends manual:true and a UTC (Z) date', () async {
      final (api, captured) = _stubbedService();

      await api.createGame(
        players: [const Player(name: 'Tom')],
        events: [_event()],
        homeLineups: {
          -1: [_athlete('A', 'Rot')]
        },
        awayLineups: {
          -1: [_athlete('B', 'Blau')]
        },
        manual: true,
      );

      expect(captured, hasLength(1));
      expect(captured.first.path, '/api/v1/games');
      final ev = (captured.first.data['events'] as List).first
          as Map<String, dynamic>;
      expect(ev['manual'], true);
      expect(ev['date'], endsWith('Z'),
          reason: 'dates must be UTC so the Go backend accepts RFC3339');
      expect((ev['home_lineup'] as List), hasLength(1));
      expect((ev['away_lineup'] as List), hasLength(1));
    });

    test('schedule game omits the manual flag and absent lineups', () async {
      final (api, captured) = _stubbedService();

      await api.createGame(
        players: [const Player(name: 'Tom')],
        events: [_event()],
      );

      final ev = (captured.first.data['events'] as List).first
          as Map<String, dynamic>;
      expect(ev.containsKey('manual'), isFalse);
      expect(ev.containsKey('home_lineup'), isFalse);
      expect(ev.containsKey('away_lineup'), isFalse);
    });
  });
}
