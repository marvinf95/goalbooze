class SportEvent {
  final int id;
  final int leagueId;
  final String homeTeam;
  final int homeTeamId;
  final String awayTeam;
  final int awayTeamId;
  final DateTime date;
  final String status;

  const SportEvent({
    required this.id,
    required this.leagueId,
    required this.homeTeam,
    this.homeTeamId = 0,
    required this.awayTeam,
    this.awayTeamId = 0,
    required this.date,
    required this.status,
  });

  factory SportEvent.fromJson(Map<String, dynamic> json) {
    return SportEvent(
      id: json['id'] as int,
      leagueId: json['league_id'] as int? ?? 0,
      homeTeam: json['home_team'] as String,
      homeTeamId: json['home_team_id'] as int? ?? 0,
      awayTeam: json['away_team'] as String,
      awayTeamId: json['away_team_id'] as int? ?? 0,
      date: DateTime.parse(json['date'] as String),
      status: json['status'] as String? ?? '',
    );
  }

  String get displayName => '$homeTeam vs $awayTeam';

  /// Whether the kickoff is within the next hour (or already started). Official
  /// lineups are usually published at most ~1 hour before kickoff, so AI lineup
  /// lookup only makes sense once this window is open.
  bool get isLineupWindowOpen =>
      date.difference(DateTime.now()) < const Duration(hours: 1);

  /// Whether the kickoff falls on the same local calendar day as [reference].
  bool isOnSameLocalDay(DateTime reference) {
    final local = date.toLocal();
    return local.year == reference.year &&
        local.month == reference.month &&
        local.day == reference.day;
  }
}
