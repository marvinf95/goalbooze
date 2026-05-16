import 'package:goalbooze/model/athlete.dart';

class Team {
  final int id;
  final String name;
  final int leagueId;
  final List<Athlete> squad;

  const Team({
    required this.id,
    required this.name,
    required this.leagueId,
    this.squad = const [],
  });

  factory Team.fromJson(Map<String, dynamic> json) {
    return Team(
      id: json['id'] as int,
      name: json['name'] as String,
      leagueId: json['league_id'] as int? ?? 0,
      squad: (json['squad'] as List<dynamic>? ?? [])
          .map((e) => Athlete.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }
}
