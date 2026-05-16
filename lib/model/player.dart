class Player {
  final String name;

  const Player({required this.name});

  factory Player.fromJson(Map<String, dynamic> json) {
    return Player(name: json['name'] as String);
  }

  Map<String, dynamic> toJson() => {'name': name};
}
