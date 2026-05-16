import 'package:flutter_test/flutter_test.dart';
import 'package:goalbooze/model/game.dart';
import 'package:goalbooze/model/player.dart';
import 'package:goalbooze/model/event.dart';
import 'package:goalbooze/model/athlete.dart';
import 'package:goalbooze/model/league.dart';

void main() {
  group('Player model', () {
    test('fromJson should create Player', () {
      final player = Player.fromJson({'name': 'Alice'});
      expect(player.name, 'Alice');
    });

    test('toJson should serialize correctly', () {
      final player = Player(name: 'Bob');
      expect(player.toJson(), {'name': 'Bob'});
    });
  });

  group('Athlete model', () {
    test('fromJson should create Athlete with all fields', () {
      final athlete = Athlete.fromJson({
        'id': 42,
        'name': 'Harry Kane',
        'number': 9,
        'position': 'Forward',
        'team': 'Bayern München',
      });
      expect(athlete.id, 42);
      expect(athlete.name, 'Harry Kane');
      expect(athlete.number, 9);
      expect(athlete.position, 'Forward');
      expect(athlete.team, 'Bayern München');
    });

    test('fromJson should handle missing optional fields', () {
      final athlete = Athlete.fromJson({
        'id': 1,
        'name': 'Player X',
      });
      expect(athlete.number, 0);
      expect(athlete.position, '');
      expect(athlete.team, '');
    });
  });

  group('SportEvent model', () {
    test('fromJson should parse date correctly', () {
      final event = SportEvent.fromJson({
        'id': 103,
        'league_id': 78,
        'home_team': 'FC Bayern',
        'away_team': 'BVB',
        'date': '2026-04-25T15:30:00Z',
        'status': 'scheduled',
      });
      expect(event.id, 103);
      expect(event.leagueId, 78);
      expect(event.homeTeam, 'FC Bayern');
      expect(event.awayTeam, 'BVB');
      expect(event.status, 'scheduled');
      expect(event.date.year, 2026);
      expect(event.date.month, 4);
      expect(event.date.day, 25);
    });

    test('displayName should format correctly', () {
      final event = SportEvent.fromJson({
        'id': 1,
        'home_team': 'FC Bayern',
        'away_team': 'BVB',
        'date': '2026-01-01T00:00:00Z',
        'status': 'scheduled',
      });
      expect(event.displayName, 'FC Bayern vs BVB');
    });
  });

  group('League model', () {
    test('fromJson should create League', () {
      final league = League.fromJson({
        'id': 78,
        'name': '1. Bundesliga',
        'slug': 'bl1',
        'season': 2025,
      });
      expect(league.id, 78);
      expect(league.name, '1. Bundesliga');
      expect(league.slug, 'bl1');
      expect(league.season, 2025);
    });
  });

  group('Assignment model', () {
    test('fromJson should create Assignment', () {
      final assignment = Assignment.fromJson({
        'player_name': 'Alice',
        'athlete_name': 'Kane',
        'team_name': 'FCB',
        'event_id': 100,
        'position': 'Forward',
      });
      expect(assignment.playerName, 'Alice');
      expect(assignment.athleteName, 'Kane');
      expect(assignment.teamName, 'FCB');
      expect(assignment.eventId, 100);
      expect(assignment.position, 'Forward');
    });

    test('fromJson should handle missing position', () {
      final assignment = Assignment.fromJson({
        'player_name': 'Bob',
        'athlete_name': 'Müller',
        'team_name': 'FCB',
        'event_id': 100,
      });
      expect(assignment.position, '');
    });
  });

  group('Game model', () {
    test('fromJson should parse complete game', () {
      final game = Game.fromJson({
        'id': 1,
        'created_at': '2026-04-25T15:00:00Z',
        'players': [
          {'name': 'Alice'},
          {'name': 'Bob'},
        ],
        'events': [
          {
            'id': 103,
            'league_id': 78,
            'home_team': 'FCB',
            'away_team': 'BVB',
            'date': '2026-04-25T15:30:00Z',
            'status': 'scheduled',
          },
        ],
        'assignments': [
          {
            'player_name': 'Alice',
            'athlete_name': 'Kane',
            'team_name': 'FCB',
            'event_id': 103,
            'position': 'Forward',
          },
          {
            'player_name': 'Bob',
            'athlete_name': 'Müller',
            'team_name': 'FCB',
            'event_id': 103,
            'position': 'Midfielder',
          },
        ],
      });
      expect(game.id, 1);
      expect(game.players.length, 2);
      expect(game.events.length, 1);
      expect(game.assignments.length, 2);
    });

    test('fromJson should handle missing assignments', () {
      final game = Game.fromJson({
        'id': 1,
        'created_at': '2026-04-25T15:00:00Z',
        'players': [{'name': 'Alice'}],
        'events': [
          {
            'id': 103,
            'league_id': 78,
            'home_team': 'FCB',
            'away_team': 'BVB',
            'date': '2026-04-25T15:30:00Z',
            'status': 'scheduled',
          },
        ],
      });
      expect(game.assignments, isEmpty);
    });
  });
}
