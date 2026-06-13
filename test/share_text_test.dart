import 'package:flutter_test/flutter_test.dart';
import 'package:goalbooze/model/event.dart';
import 'package:goalbooze/model/game.dart';
import 'package:goalbooze/model/player.dart';
import 'package:goalbooze/util/share_text.dart';

SportEvent _event(int id, String home, String away, DateTime date) => SportEvent(
      id: id,
      leagueId: 1,
      homeTeam: home,
      awayTeam: away,
      date: date,
      status: 'scheduled',
    );

Assignment _assign(String player, String athlete, String team, int eventId) =>
    Assignment(
      playerName: player,
      athleteName: athlete,
      teamName: team,
      eventId: eventId,
      position: '',
    );

void main() {
  test('formats a single-event game grouped by player', () {
    final game = Game(
      id: 1,
      createdAt: DateTime(2026, 6, 13, 18, 0),
      players: const [Player(name: 'Max'), Player(name: 'Lukas')],
      events: [_event(10, 'Bayern', 'Dortmund', DateTime(2026, 6, 13, 20, 30))],
      assignments: [
        _assign('Max', 'Harry Kane', 'Bayern', 10),
        _assign('Max', 'Jamal Musiala', 'Bayern', 10),
        _assign('Lukas', 'Marco Reus', 'Dortmund', 10),
      ],
    );

    expect(
      buildAssignmentShareText(game),
      '🍻 GoalBooze – Auslosung\n'
      'Bayern vs Dortmund · 13.6.2026, 20:30 Uhr\n'
      '\n'
      'Max\n'
      '• Harry Kane (Bayern)\n'
      '• Jamal Musiala (Bayern)\n'
      '\n'
      'Lukas\n'
      '• Marco Reus (Dortmund)',
    );
  });

  test('lists one header line per event for a multi-event game', () {
    final game = Game(
      id: 2,
      createdAt: DateTime(2026, 6, 13, 18, 0),
      players: const [Player(name: 'Max')],
      events: [
        _event(10, 'Bayern', 'Dortmund', DateTime(2026, 6, 13, 20, 30)),
        _event(11, 'Liverpool', 'City', DateTime(2026, 6, 14, 18, 5)),
      ],
      assignments: [
        _assign('Max', 'Harry Kane', 'Bayern', 10),
        _assign('Max', 'Mohamed Salah', 'Liverpool', 11),
      ],
    );

    expect(
      buildAssignmentShareText(game),
      '🍻 GoalBooze – Auslosung\n'
      'Bayern vs Dortmund · 13.6.2026, 20:30 Uhr\n'
      'Liverpool vs City · 14.6.2026, 18:05 Uhr\n'
      '\n'
      'Max\n'
      '• Harry Kane (Bayern)\n'
      '• Mohamed Salah (Liverpool)',
    );
  });

  test('returns just the header when there are no assignments', () {
    final game = Game(
      id: 3,
      createdAt: DateTime(2026, 6, 13, 18, 0),
      players: const [],
      events: [_event(10, 'Bayern', 'Dortmund', DateTime(2026, 6, 13, 20, 30))],
      assignments: const [],
    );

    expect(
      buildAssignmentShareText(game),
      '🍻 GoalBooze – Auslosung\n'
      'Bayern vs Dortmund · 13.6.2026, 20:30 Uhr',
    );
  });
}
