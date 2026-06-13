import 'package:goalbooze/model/game.dart';

/// Builds a compact, WhatsApp-friendly plain-text representation of a game's
/// drawn assignments, so it can be copied to the clipboard and shared with
/// players who don't have the app at hand.
///
/// Layout (player-grouped, mirroring the overview screen):
///
///     🍻 GoalBooze – Auslosung
///     Home vs Away · 13.6.2026, 20:30 Uhr
///
///     Max
///     • Harry Kane (Bayern)
///
///     Lukas
///     • Marco Reus (Dortmund)
String buildAssignmentShareText(Game game) {
  final buffer = StringBuffer('🍻 GoalBooze – Auslosung');

  for (final event in game.events) {
    buffer.write('\n${event.displayName} · ${_formatDate(event.date)}');
  }

  // Group assignments by player, preserving their original order.
  final grouped = <String, List<Assignment>>{};
  for (final a in game.assignments) {
    grouped.putIfAbsent(a.playerName, () => []).add(a);
  }

  for (final entry in grouped.entries) {
    buffer.write('\n\n${entry.key}');
    for (final a in entry.value) {
      buffer.write('\n• ${a.athleteName} (${a.teamName})');
    }
  }

  return buffer.toString().trim();
}

String _formatDate(DateTime dt) {
  final d = dt.toLocal();
  final mm = d.minute.toString().padLeft(2, '0');
  return '${d.day}.${d.month}.${d.year}, ${d.hour}:$mm Uhr';
}
