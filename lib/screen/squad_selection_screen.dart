import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:goalbooze/model/athlete.dart';
import 'package:goalbooze/model/event.dart';
import 'package:goalbooze/model/team.dart';
import 'package:goalbooze/provider/game_provider.dart';
import 'package:goalbooze/provider/squads_provider.dart';
import 'package:goalbooze/widget/athlete_card.dart';

class SquadSelectionScreen extends ConsumerStatefulWidget {
  const SquadSelectionScreen({super.key});

  @override
  ConsumerState<SquadSelectionScreen> createState() =>
      _SquadSelectionScreenState();
}

class _SquadSelectionScreenState extends ConsumerState<SquadSelectionScreen> {
  final Map<int, List<Athlete>> _homeSelections = {};
  final Map<int, List<Athlete>> _awaySelections = {};
  // eventId → true means user chose to override the live lineup manually
  final Map<int, bool> _overrideToManual = {};
  bool _isCreating = false;

  LineupArgs _lineupArgs(SportEvent event) => (
        eventId: event.id,
        homeTeamId: event.homeTeamId,
        awayTeamId: event.awayTeamId,
        homeTeam: event.homeTeam,
        awayTeam: event.awayTeam,
        date: event.date.toIso8601String(),
      );

  @override
  Widget build(BuildContext context) {
    final gameState = ref.watch(gameProvider);
    final events = gameState.selectedEvents;

    // Auto-populate selections when a live lineup arrives
    for (final event in events) {
      ref.listen(eventLineupProvider(_lineupArgs(event)), (_, next) {
        next.whenData((data) {
          if (data == null || data['is_squad_pick'] == true) return;
          if (!_homeSelections.containsKey(event.id)) {
            setState(() {
              _homeSelections[event.id] =
                  List<Athlete>.from(data['home'] as List);
              _awaySelections[event.id] =
                  List<Athlete>.from(data['away'] as List);
            });
          }
        });
      });
    }

    return Scaffold(
      appBar: AppBar(
        title: const Text('Aufstellung wählen'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => context.pop(),
        ),
      ),
      body: events.isEmpty
          ? const Center(child: Text('Keine Spiele ausgewählt'))
          : ListView.builder(
              padding: const EdgeInsets.all(16),
              itemCount: events.length + 1,
              itemBuilder: (context, index) {
                if (index == events.length) {
                  return _StartButton(
                    isLoading: _isCreating,
                    onPressed: _startGame,
                  );
                }
                final event = events[index];
                return _EventLineupCard(
                  event: event,
                  homeSelections: _homeSelections[event.id] ?? [],
                  awaySelections: _awaySelections[event.id] ?? [],
                  overrideToManual: _overrideToManual[event.id] ?? false,
                  onHomeChanged: (athletes) => setState(() {
                    _homeSelections[event.id] = athletes;
                  }),
                  onAwayChanged: (athletes) => setState(() {
                    _awaySelections[event.id] = athletes;
                  }),
                  onManualOverride: () => setState(() {
                    _overrideToManual[event.id] = true;
                  }),
                );
              },
            ),
    );
  }

  Future<void> _startGame() async {
    if (_isCreating) return;
    setState(() => _isCreating = true);

    final gameState = ref.read(gameProvider);
    final notifier = ref.read(gameProvider.notifier);

    for (final event in gameState.selectedEvents) {
      final home = _homeSelections[event.id];
      final away = _awaySelections[event.id];
      if (home != null && away != null) {
        notifier.setLineup(event.id, home, away);
      }
    }

    await notifier.createGame();

    final state = ref.read(gameProvider);
    setState(() => _isCreating = false);

    if (!mounted) return;

    if (state.error != null) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(state.error!),
          backgroundColor: Theme.of(context).colorScheme.error,
        ),
      );
      return;
    }

    if (state.currentGame != null) {
      context.go('/game/${state.currentGame!.id}');
    }
  }
}

class _EventLineupCard extends ConsumerWidget {
  final SportEvent event;
  final List<Athlete> homeSelections;
  final List<Athlete> awaySelections;
  final bool overrideToManual;
  final void Function(List<Athlete>) onHomeChanged;
  final void Function(List<Athlete>) onAwayChanged;
  final VoidCallback onManualOverride;

  const _EventLineupCard({
    required this.event,
    required this.homeSelections,
    required this.awaySelections,
    required this.overrideToManual,
    required this.onHomeChanged,
    required this.onAwayChanged,
    required this.onManualOverride,
  });

  LineupArgs get _lineupArgs => (
        eventId: event.id,
        homeTeamId: event.homeTeamId,
        awayTeamId: event.awayTeamId,
        homeTeam: event.homeTeam,
        awayTeam: event.awayTeam,
        date: event.date.toIso8601String(),
      );

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);

    final lineupAsync = ref.watch(eventLineupProvider(_lineupArgs));
    final liveData = lineupAsync.valueOrNull;
    final isLive =
        !overrideToManual && liveData != null && liveData['is_squad_pick'] == false;

    final homeTeamsAsync = event.homeTeamId > 0 && event.leagueId > 0
        ? ref.watch(squadByTeamProvider(
            (leagueId: event.leagueId, teamId: event.homeTeamId)))
        : const AsyncValue<Team?>.data(null);
    final awayTeamsAsync = event.awayTeamId > 0 && event.leagueId > 0
        ? ref.watch(squadByTeamProvider(
            (leagueId: event.leagueId, teamId: event.awayTeamId)))
        : const AsyncValue<Team?>.data(null);

    return Card(
      margin: const EdgeInsets.only(bottom: 16),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Match header
            Row(
              children: [
                Expanded(
                  child: Text(
                    event.homeTeam,
                    style: theme.textTheme.bodyMedium
                        ?.copyWith(fontWeight: FontWeight.w600),
                    textAlign: TextAlign.center,
                  ),
                ),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                  decoration: BoxDecoration(
                    color: theme.colorScheme.primary.withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(6),
                  ),
                  child: Text('VS',
                      style: TextStyle(
                          color: theme.colorScheme.primary,
                          fontWeight: FontWeight.bold,
                          fontSize: 11)),
                ),
                Expanded(
                  child: Text(
                    event.awayTeam,
                    style: theme.textTheme.bodyMedium
                        ?.copyWith(fontWeight: FontWeight.w600),
                    textAlign: TextAlign.center,
                  ),
                ),
              ],
            ),
            // Loading indicator while fetching live lineup
            if (lineupAsync.isLoading) ...[
              const SizedBox(height: 10),
              LinearProgressIndicator(
                minHeight: 2,
                color: theme.colorScheme.primary,
                backgroundColor: theme.colorScheme.primary.withValues(alpha: 0.15),
              ),
              const SizedBox(height: 6),
              Text(
                'Offizielle Aufstellung wird geladen...',
                style: theme.textTheme.bodySmall?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                  fontStyle: FontStyle.italic,
                ),
                textAlign: TextAlign.center,
              ),
            ],
            const SizedBox(height: 16),
            // Show live lineup or squad sections
            if (isLive)
              _LiveLineupSection(
                homeAthletes: List<Athlete>.from(liveData['home'] as List),
                awayAthletes: List<Athlete>.from(liveData['away'] as List),
                homeTeam: event.homeTeam,
                awayTeam: event.awayTeam,
                onManualOverride: onManualOverride,
              )
            else ...[
              _SquadSection(
                teamName: event.homeTeam,
                teamId: event.homeTeamId,
                squadAsync: homeTeamsAsync,
                selected: homeSelections,
                onChanged: onHomeChanged,
                accentColor: const Color(0xFF4FC3F7),
              ),
              const SizedBox(height: 12),
              _SquadSection(
                teamName: event.awayTeam,
                teamId: event.awayTeamId,
                squadAsync: awayTeamsAsync,
                selected: awaySelections,
                onChanged: onAwayChanged,
                accentColor: const Color(0xFFFF6B35),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _LiveLineupSection extends StatelessWidget {
  final List<Athlete> homeAthletes;
  final List<Athlete> awayAthletes;
  final String homeTeam;
  final String awayTeam;
  final VoidCallback onManualOverride;

  const _LiveLineupSection({
    required this.homeAthletes,
    required this.awayAthletes,
    required this.homeTeam,
    required this.awayTeam,
    required this.onManualOverride,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Badge
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
          decoration: BoxDecoration(
            color: const Color(0xFF00D4AA).withValues(alpha: 0.15),
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: const Color(0xFF00D4AA).withValues(alpha: 0.4)),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.verified, size: 14, color: Color(0xFF00D4AA)),
              const SizedBox(width: 6),
              Text(
                'Offizielle Aufstellung',
                style: theme.textTheme.bodySmall?.copyWith(
                  color: const Color(0xFF00D4AA),
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 12),
        // Home team
        _LineupTeamList(
          teamName: homeTeam,
          athletes: homeAthletes,
          accentColor: const Color(0xFF4FC3F7),
        ),
        const SizedBox(height: 12),
        // Away team
        _LineupTeamList(
          teamName: awayTeam,
          athletes: awayAthletes,
          accentColor: const Color(0xFFFF6B35),
        ),
        const SizedBox(height: 12),
        // Manual override button
        SizedBox(
          width: double.infinity,
          child: OutlinedButton.icon(
            onPressed: onManualOverride,
            icon: const Icon(Icons.edit, size: 16),
            label: const Text('Manuell anpassen'),
            style: OutlinedButton.styleFrom(
              padding: const EdgeInsets.symmetric(vertical: 10),
              foregroundColor: theme.colorScheme.onSurfaceVariant,
              side: BorderSide(color: theme.colorScheme.outline),
            ),
          ),
        ),
      ],
    );
  }
}

class _LineupTeamList extends StatelessWidget {
  final String teamName;
  final List<Athlete> athletes;
  final Color accentColor;

  const _LineupTeamList({
    required this.teamName,
    required this.athletes,
    required this.accentColor,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Container(
              width: 10,
              height: 10,
              decoration: BoxDecoration(color: accentColor, shape: BoxShape.circle),
            ),
            const SizedBox(width: 6),
            Text(
              teamName,
              style: theme.textTheme.bodySmall?.copyWith(
                color: accentColor,
                fontWeight: FontWeight.w600,
              ),
            ),
          ],
        ),
        const SizedBox(height: 6),
        ...athletes.map((a) => Padding(
              padding: const EdgeInsets.only(bottom: 4),
              child: Row(
                children: [
                  SizedBox(
                    width: 32,
                    child: Text(
                      a.position,
                      style: theme.textTheme.bodySmall?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                        fontSize: 10,
                      ),
                    ),
                  ),
                  Text(
                    a.name,
                    style: theme.textTheme.bodySmall
                        ?.copyWith(fontWeight: FontWeight.w500),
                  ),
                  if (a.number > 0) ...[
                    const Spacer(),
                    Text(
                      '#${a.number}',
                      style: theme.textTheme.bodySmall?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                      ),
                    ),
                  ],
                ],
              ),
            )),
      ],
    );
  }
}

class _SquadSection extends StatelessWidget {
  final String teamName;
  final int teamId;
  final AsyncValue<Team?> squadAsync;
  final List<Athlete> selected;
  final void Function(List<Athlete>) onChanged;
  final Color accentColor;

  const _SquadSection({
    required this.teamName,
    required this.teamId,
    required this.squadAsync,
    required this.selected,
    required this.onChanged,
    required this.accentColor,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Container(
              width: 10,
              height: 10,
              decoration:
                  BoxDecoration(color: accentColor, shape: BoxShape.circle),
            ),
            const SizedBox(width: 6),
            Text(
              teamName,
              style: theme.textTheme.bodySmall?.copyWith(
                color: accentColor,
                fontWeight: FontWeight.w600,
              ),
            ),
            const Spacer(),
            Text(
              '${selected.length}/11',
              style: theme.textTheme.bodySmall?.copyWith(
                color: selected.length == 11
                    ? theme.colorScheme.primary
                    : theme.colorScheme.onSurfaceVariant,
              ),
            ),
          ],
        ),
        const SizedBox(height: 8),
        squadAsync.when(
          loading: () => const Center(
            child: Padding(
              padding: EdgeInsets.all(8),
              child: CircularProgressIndicator(strokeWidth: 2),
            ),
          ),
          error: (_, __) => _RandomSelectionFallback(
            teamName: teamName,
            selected: selected,
            onChanged: onChanged,
          ),
          data: (team) {
            if (team == null || teamId == 0) {
              return _RandomSelectionFallback(
                teamName: teamName,
                selected: selected,
                onChanged: onChanged,
              );
            }
            final squad = team.squad;
            if (squad.isEmpty) {
              return _RandomSelectionFallback(
                teamName: teamName,
                selected: selected,
                onChanged: onChanged,
              );
            }
            return _SquadGrid(
              squad: squad,
              selected: selected,
              onChanged: onChanged,
            );
          },
        ),
      ],
    );
  }
}

class _SquadGrid extends StatelessWidget {
  final List<Athlete> squad;
  final List<Athlete> selected;
  final void Function(List<Athlete>) onChanged;

  const _SquadGrid({
    required this.squad,
    required this.selected,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final selectedIds = selected.map((a) => a.id).toSet();

    return Column(
      children: [
        Row(
          children: [
            Expanded(
              child: OutlinedButton.icon(
                onPressed: () => _randomEleven(),
                icon: const Icon(Icons.shuffle, size: 16),
                label: const Text('Zufällig'),
                style: OutlinedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 8),
                ),
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: OutlinedButton.icon(
                onPressed: selected.isEmpty ? null : () => onChanged([]),
                icon: const Icon(Icons.clear_all, size: 16),
                label: const Text('Leeren'),
                style: OutlinedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 8),
                ),
              ),
            ),
          ],
        ),
        const SizedBox(height: 8),
        ...squad.map((athlete) {
          final isSelected = selectedIds.contains(athlete.id);
          return Padding(
            padding: const EdgeInsets.only(bottom: 6),
            child: AthleteCard(
              athlete: athlete,
              isSelected: isSelected,
              compact: true,
              onTap: () {
                final newSelected = List<Athlete>.from(selected);
                if (isSelected) {
                  newSelected.removeWhere((a) => a.id == athlete.id);
                } else {
                  if (newSelected.length >= 11) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(
                        content: const Text('Maximal 11 Spieler auswählen'),
                        backgroundColor: theme.colorScheme.error,
                        duration: const Duration(seconds: 1),
                      ),
                    );
                    return;
                  }
                  newSelected.add(athlete);
                }
                onChanged(newSelected);
              },
            ),
          );
        }),
      ],
    );
  }

  void _randomEleven() {
    final shuffled = List<Athlete>.from(squad)..shuffle();
    onChanged(shuffled.take(11).toList());
  }
}

class _RandomSelectionFallback extends StatelessWidget {
  final String teamName;
  final List<Athlete> selected;
  final void Function(List<Athlete>) onChanged;

  const _RandomSelectionFallback({
    required this.teamName,
    required this.selected,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: theme.colorScheme.outline),
      ),
      child: Column(
        children: [
          Text(
            'Kein Kader verfügbar — Backend lädt Spielerdaten beim Erstellen',
            style: theme.textTheme.bodySmall?.copyWith(
              color: theme.colorScheme.onSurfaceVariant,
            ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 8),
          Text(
            selected.isEmpty
                ? '11 Spieler werden automatisch zugewiesen'
                : '${selected.length} Spieler ausgewählt',
            style: theme.textTheme.bodySmall?.copyWith(
              color: theme.colorScheme.primary,
            ),
          ),
        ],
      ),
    );
  }
}

class _StartButton extends StatelessWidget {
  final bool isLoading;
  final VoidCallback onPressed;

  const _StartButton({required this.isLoading, required this.onPressed});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 24),
      child: FilledButton.icon(
        onPressed: isLoading ? null : onPressed,
        icon: isLoading
            ? const SizedBox(
                width: 20,
                height: 20,
                child: CircularProgressIndicator(
                  strokeWidth: 2,
                  color: Colors.white,
                ),
              )
            : const Icon(Icons.rocket_launch),
        label: Text(isLoading ? 'Spiel wird erstellt...' : 'Spiel starten!'),
        style: FilledButton.styleFrom(
          minimumSize: const Size.fromHeight(54),
        ),
      ),
    );
  }
}
