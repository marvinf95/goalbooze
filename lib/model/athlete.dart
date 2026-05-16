class Athlete {
  final int id;
  final String name;
  final int number;
  final String position;
  final String team;

  const Athlete({
    required this.id,
    required this.name,
    required this.number,
    required this.position,
    required this.team,
  });

  factory Athlete.fromJson(Map<String, dynamic> json) {
    return Athlete(
      id: json['id'] as int? ?? 0,
      name: json['name'] as String,
      number: json['number'] as int? ?? 0,
      position: json['position'] as String? ?? '',
      team: json['team'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'name': name,
        'number': number,
        'position': position,
        'team': team,
      };
}
