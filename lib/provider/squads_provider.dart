import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:goalbooze/model/team.dart';
import 'package:goalbooze/service/api_service.dart';

final teamsByLeagueProvider =
    FutureProvider.family<List<Team>, int>((ref, leagueId) async {
  final api = ref.read(apiServiceProvider);
  return api.getTeams(leagueId);
});

/// Fetches a single team (with full squad) by loading all teams for the given
/// league. This ensures the backend squad cache is warmed before game creation.
final squadByTeamProvider =
    FutureProvider.family<Team?, ({int leagueId, int teamId})>((ref, args) async {
  if (args.teamId == 0 || args.leagueId == 0) return null;
  final teams = await ref.watch(teamsByLeagueProvider(args.leagueId).future);
  final matches = teams.where((t) => t.id == args.teamId);
  return matches.isEmpty ? null : matches.first;
});

typedef LineupArgs = ({
  int eventId,
  int homeTeamId,
  int awayTeamId,
  String homeTeam,
  String awayTeam,
  String date,
});

/// Fetches the live starting lineup for an event via Claude AI + WebSearch.
/// Returns null on error so the caller can fall back to squad selection.
/// Result: `{'home': List<Athlete>, 'away': List<Athlete>, 'is_squad_pick': bool}`
final eventLineupProvider =
    FutureProvider.family<Map<String, dynamic>?, LineupArgs>((ref, args) async {
  try {
    return await ref.read(apiServiceProvider).getEventLineup(
          args.eventId,
          homeTeamId: args.homeTeamId,
          awayTeamId: args.awayTeamId,
          homeTeam: args.homeTeam,
          awayTeam: args.awayTeam,
          date: args.date,
        );
  } catch (_) {
    return null;
  }
});
