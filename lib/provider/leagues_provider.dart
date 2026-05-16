import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:goalbooze/model/league.dart';
import 'package:goalbooze/model/event.dart';
import 'package:goalbooze/service/api_service.dart';

final leaguesProvider = FutureProvider<List<League>>((ref) async {
  final api = ref.read(apiServiceProvider);
  return api.getLeagues();
});

final eventsByLeagueProvider =
    FutureProvider.family<List<SportEvent>, int>((ref, leagueId) async {
  final api = ref.read(apiServiceProvider);
  final now = DateTime.now();
  final season = now.month < 7 ? now.year - 1 : now.year;
  return api.getEvents(leagueId, season);
});
