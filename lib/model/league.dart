class League {
  final int id;
  final String name;
  final String slug;
  final int season;

  const League({
    required this.id,
    required this.name,
    required this.slug,
    required this.season,
  });

  factory League.fromJson(Map<String, dynamic> json) {
    return League(
      id: json['id'] as int,
      name: json['name'] as String,
      slug: json['slug'] as String,
      season: json['season'] as int,
    );
  }
}
