import 'player.dart';
import 'event.dart';

class Assignment {
  final String playerName;
  final String athleteName;
  final String teamName;
  final int eventId;
  final String position;

  const Assignment({
    required this.playerName,
    required this.athleteName,
    required this.teamName,
    required this.eventId,
    required this.position,
  });

  factory Assignment.fromJson(Map<String, dynamic> json) {
    return Assignment(
      playerName: json['player_name'] as String,
      athleteName: json['athlete_name'] as String,
      teamName: json['team_name'] as String,
      eventId: json['event_id'] as int,
      position: json['position'] as String? ?? '',
    );
  }
}

class Game {
  final int id;
  final DateTime createdAt;
  final List<Player> players;
  final List<SportEvent> events;
  final List<Assignment> assignments;

  const Game({
    required this.id,
    required this.createdAt,
    required this.players,
    required this.events,
    required this.assignments,
  });

  factory Game.fromJson(Map<String, dynamic> json) {
    return Game(
      id: json['id'] as int,
      createdAt: DateTime.parse(json['created_at'] as String),
      players: (json['players'] as List)
          .map((e) => Player.fromJson(e as Map<String, dynamic>))
          .toList(),
      events: (json['events'] as List)
          .map((e) => SportEvent.fromJson(e as Map<String, dynamic>))
          .toList(),
      assignments: json['assignments'] != null
          ? (json['assignments'] as List)
              .map((e) => Assignment.fromJson(e as Map<String, dynamic>))
              .toList()
          : [],
    );
  }
}
