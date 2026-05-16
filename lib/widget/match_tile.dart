import 'package:flutter/material.dart';
import 'package:goalbooze/model/event.dart';

enum LineupStatus { notLoaded, loading, squadReady, liveLineup }

class MatchTile extends StatelessWidget {
  final SportEvent event;
  final bool isSelected;
  final LineupStatus lineupStatus;
  final VoidCallback onTap;

  const MatchTile({
    super.key,
    required this.event,
    required this.isSelected,
    required this.onTap,
    this.lineupStatus = LineupStatus.notLoaded,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final primary = theme.colorScheme.primary;

    return GestureDetector(
      onTap: onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        margin: const EdgeInsets.only(bottom: 8),
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: isSelected
              ? primary.withValues(alpha: 0.1)
              : theme.colorScheme.surface,
          borderRadius: BorderRadius.circular(14),
          border: Border.all(
            color: isSelected ? primary : theme.colorScheme.outline,
            width: isSelected ? 1.5 : 1,
          ),
        ),
        child: Row(
          children: [
            _SelectionIndicator(isSelected: isSelected, primary: primary),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    '${event.homeTeam}  🆚  ${event.awayTeam}',
                    style: theme.textTheme.bodyMedium?.copyWith(
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    _formatDate(event.date),
                    style: theme.textTheme.bodySmall,
                  ),
                ],
              ),
            ),
            _LineupBadge(status: lineupStatus),
          ],
        ),
      ),
    );
  }

  String _formatDate(DateTime dt) {
    final local = dt.toLocal();
    final weekdays = ['Mo', 'Di', 'Mi', 'Do', 'Fr', 'Sa', 'So'];
    final day = weekdays[local.weekday - 1];
    final time = '${local.hour.toString().padLeft(2, '0')}:${local.minute.toString().padLeft(2, '0')}';
    return '$day ${local.day}.${local.month}. · $time Uhr';
  }
}

class _SelectionIndicator extends StatelessWidget {
  final bool isSelected;
  final Color primary;

  const _SelectionIndicator({required this.isSelected, required this.primary});

  @override
  Widget build(BuildContext context) {
    return AnimatedContainer(
      duration: const Duration(milliseconds: 200),
      width: 22,
      height: 22,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        color: isSelected ? primary : Colors.transparent,
        border: Border.all(
          color: isSelected ? primary : const Color(0xFF3A3A5C),
          width: 2,
        ),
      ),
      child: isSelected
          ? const Icon(Icons.check, size: 14, color: Color(0xFF0D0D1A))
          : null,
    );
  }
}

class _LineupBadge extends StatelessWidget {
  final LineupStatus status;

  const _LineupBadge({required this.status});

  @override
  Widget build(BuildContext context) {
    switch (status) {
      case LineupStatus.liveLineup:
        return _badge('Live', const Color(0xFF00D4AA));
      case LineupStatus.squadReady:
        return _badge('Kader', const Color(0xFF4FC3F7));
      case LineupStatus.loading:
        return const SizedBox(
          width: 16,
          height: 16,
          child: CircularProgressIndicator(strokeWidth: 2),
        );
      case LineupStatus.notLoaded:
        return const SizedBox.shrink();
    }
  }

  Widget _badge(String label, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.15),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: color.withValues(alpha: 0.4)),
      ),
      child: Text(
        label,
        style: TextStyle(color: color, fontSize: 11, fontWeight: FontWeight.w600),
      ),
    );
  }
}
