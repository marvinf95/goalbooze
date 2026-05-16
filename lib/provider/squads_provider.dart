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
