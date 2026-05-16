import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:goalbooze/model/game.dart';
import 'package:goalbooze/provider/game_provider.dart';
import 'package:goalbooze/service/api_service.dart';

class GameOverviewScreen extends ConsumerWidget {
  final int gameId;
  final bool readOnly;

  const GameOverviewScreen({
    super.key,
    required this.gameId,
    this.readOnly = false,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final gameAsync = ref.watch(_gameProvider(gameId));
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: Text(readOnly ? 'Vergangenes Spiel' : 'GoalBooze'),
        actions: [
          if (!readOnly)
            Padding(
              padding: const EdgeInsets.only(right: 12),
              child: OutlinedButton.icon(
                onPressed: () {
                  ref.read(gameProvider.notifier).reset();
                  context.go('/');
                },
                icon: const Icon(Icons.stop, size: 18),
                label: const Text('Beenden'),
                style: OutlinedButton.styleFrom(
                  foregroundColor: theme.colorScheme.error,
                  side: BorderSide(color: theme.colorScheme.error),
                ),
              ),
            ),
        ],
      ),
      body: gameAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.error_outline,
                  size: 48, color: theme.colorScheme.error),
              const SizedBox(height: 12),
              Text('Fehler beim Laden', style: theme.textTheme.titleMedium),
              const SizedBox(height: 4),
              Text(e.toString(),
                  style: theme.textTheme.bodySmall,
                  textAlign: TextAlign.center),
            ],
          ),
        ),
        data: (game) => _GameView(game: game, readOnly: readOnly),
      ),
    );
  }
}

class _GameView extends StatelessWidget {
  final Game game;
  final bool readOnly;

  const _GameView({required this.game, required this.readOnly});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final grouped = <String, List<Assignment>>{};
    for (final a in game.assignments) {
      grouped.putIfAbsent(a.playerName, () => []).add(a);
    }

    if (grouped.isEmpty) {
      return Center(
        child: Text(
          'Keine Zuweisungen vorhanden',
          style: theme.textTheme.bodyLarge?.copyWith(
            color: theme.colorScheme.onSurfaceVariant,
          ),
        ),
      );
    }

    return ListView(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
      children: [
        // Meta info
        Padding(
          padding: const EdgeInsets.only(bottom: 16),
          child: Row(
            children: [
              Icon(Icons.calendar_today,
                  size: 14, color: theme.colorScheme.onSurfaceVariant),
              const SizedBox(width: 6),
              Text(
                _formatDate(game.createdAt),
                style: theme.textTheme.bodySmall,
              ),
              const SizedBox(width: 16),
              Icon(Icons.group,
                  size: 14, color: theme.colorScheme.onSurfaceVariant),
              const SizedBox(width: 6),
              Text(
                '${game.players.length} Spieler',
                style: theme.textTheme.bodySmall,
              ),
            ],
          ),
        ),
        // Player assignment cards
        ...grouped.entries.toList().asMap().entries.map((entry) {
          final idx = entry.key;
          final playerName = entry.value.key;
          final assignments = entry.value.value;
          return _PlayerCard(
            playerName: playerName,
            assignments: assignments,
            animationDelay: Duration(milliseconds: idx * 80),
          );
        }),
        // Rules quick-reference
        if (!readOnly) ...[
          const SizedBox(height: 8),
          _RulesQuickRef(),
        ],
      ],
    );
  }

  String _formatDate(DateTime dt) {
    return '${dt.day}.${dt.month}.${dt.year} · ${dt.hour}:${dt.minute.toString().padLeft(2, '0')} Uhr';
  }
}

class _PlayerCard extends StatelessWidget {
  final String playerName;
  final List<Assignment> assignments;
  final Duration animationDelay;

  const _PlayerCard({
    required this.playerName,
    required this.assignments,
    required this.animationDelay,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Player header
            Row(
              children: [
                Container(
                  width: 36,
                  height: 36,
                  decoration: BoxDecoration(
                    color: theme.colorScheme.primary.withValues(alpha: 0.2),
                    shape: BoxShape.circle,
                  ),
                  alignment: Alignment.center,
                  child: Text(
                    playerName[0].toUpperCase(),
                    style: TextStyle(
                      color: theme.colorScheme.primary,
                      fontWeight: FontWeight.bold,
                      fontSize: 16,
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                Text(
                  playerName,
                  style: theme.textTheme.titleMedium
                      ?.copyWith(fontWeight: FontWeight.bold),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Divider(color: theme.colorScheme.outline, height: 1),
            const SizedBox(height: 12),
            // Athlete assignments
            ...assignments.map((a) => _AssignmentRow(assignment: a)),
          ],
        ),
      ),
    )
        .animate(delay: animationDelay)
        .fadeIn(duration: 300.ms)
        .slideY(begin: 0.1, end: 0, duration: 300.ms, curve: Curves.easeOut);
  }
}

class _AssignmentRow extends StatelessWidget {
  final Assignment assignment;

  const _AssignmentRow({required this.assignment});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isHome = !assignment.teamName.contains('(away)');

    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: Row(
        children: [
          Container(
            width: 8,
            height: 8,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: isHome
                  ? const Color(0xFF4FC3F7)
                  : const Color(0xFFFF6B35),
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  assignment.athleteName,
                  style: theme.textTheme.bodyMedium
                      ?.copyWith(fontWeight: FontWeight.w600),
                ),
                Text(
                  '${assignment.teamName}${assignment.position.isNotEmpty ? ' · ${assignment.position}' : ''}',
                  style: theme.textTheme.bodySmall,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _RulesQuickRef extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: theme.colorScheme.outline),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Trinkregeln',
            style: theme.textTheme.bodySmall?.copyWith(
              letterSpacing: 0.8,
            ),
          ),
          const SizedBox(height: 8),
          Wrap(
            spacing: 12,
            runSpacing: 6,
            children: const [
              _RuleChip('⚽ Tor', '2 Schlücke'),
              _RuleChip('🟥 Rote Karte', '5 Schlücke'),
              _RuleChip('🔄 Auswechslung', 'Spieler wechselt'),
              _RuleChip('🧤 TW-Gegentor', '1 Schluck'),
            ],
          ),
        ],
      ),
    );
  }
}

class _RuleChip extends StatelessWidget {
  final String event;
  final String rule;

  const _RuleChip(this.event, this.rule);

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return RichText(
      text: TextSpan(
        style: theme.textTheme.bodySmall,
        children: [
          TextSpan(
            text: '$event: ',
            style: const TextStyle(fontWeight: FontWeight.w600),
          ),
          TextSpan(text: rule),
        ],
      ),
    );
  }
}

final _gameProvider = FutureProvider.family<Game, int>((ref, id) async {
  final api = ref.read(apiServiceProvider);
  return api.getGame(id);
});
