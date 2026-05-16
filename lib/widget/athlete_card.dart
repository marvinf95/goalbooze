import 'package:flutter/material.dart';
import 'package:goalbooze/model/athlete.dart';

class AthleteCard extends StatelessWidget {
  final Athlete athlete;
  final bool isSelected;
  final VoidCallback? onTap;
  final bool compact;

  const AthleteCard({
    super.key,
    required this.athlete,
    this.isSelected = true,
    this.onTap,
    this.compact = false,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final primary = theme.colorScheme.primary;
    final surface = theme.colorScheme.surface;

    final positionColor = _positionColor(athlete.position);

    return GestureDetector(
      onTap: onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: compact
            ? const EdgeInsets.symmetric(horizontal: 10, vertical: 6)
            : const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: isSelected
              ? primary.withValues(alpha: 0.12)
              : surface,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: isSelected ? primary : theme.colorScheme.outline,
            width: isSelected ? 1.5 : 1,
          ),
        ),
        child: compact
            ? _CompactContent(athlete: athlete, positionColor: positionColor, theme: theme)
            : _FullContent(athlete: athlete, positionColor: positionColor, theme: theme),
      ),
    );
  }

  static Color _positionColor(String position) {
    switch (position.toUpperCase()) {
      case 'GK':
        return const Color(0xFFFFD700);
      case 'CB':
      case 'RB':
      case 'LB':
      case 'DEF':
        return const Color(0xFF4FC3F7);
      case 'CM':
      case 'DM':
      case 'RM':
      case 'LM':
      case 'AM':
      case 'MF':
        return const Color(0xFF81C784);
      default:
        return const Color(0xFFFF6B35);
    }
  }
}

class _FullContent extends StatelessWidget {
  final Athlete athlete;
  final Color positionColor;
  final ThemeData theme;

  const _FullContent({required this.athlete, required this.positionColor, required this.theme});

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: 36,
          height: 36,
          decoration: BoxDecoration(
            color: positionColor.withValues(alpha: 0.2),
            shape: BoxShape.circle,
            border: Border.all(color: positionColor, width: 1.5),
          ),
          alignment: Alignment.center,
          child: Text(
            athlete.number > 0 ? '${athlete.number}' : athlete.position,
            style: TextStyle(
              color: positionColor,
              fontWeight: FontWeight.bold,
              fontSize: 12,
            ),
          ),
        ),
        const SizedBox(height: 6),
        Text(
          _shortName(athlete.name),
          style: theme.textTheme.bodySmall?.copyWith(fontWeight: FontWeight.w600),
          textAlign: TextAlign.center,
          maxLines: 2,
          overflow: TextOverflow.ellipsis,
        ),
        if (athlete.position.isNotEmpty)
          Text(
            athlete.position,
            style: TextStyle(color: positionColor, fontSize: 10),
          ),
      ],
    );
  }

  String _shortName(String name) {
    final parts = name.split(' ');
    if (parts.length > 1) return parts.last;
    return name;
  }
}

class _CompactContent extends StatelessWidget {
  final Athlete athlete;
  final Color positionColor;
  final ThemeData theme;

  const _CompactContent({required this.athlete, required this.positionColor, required this.theme});

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Container(
          width: 28,
          height: 28,
          decoration: BoxDecoration(
            color: positionColor.withValues(alpha: 0.2),
            shape: BoxShape.circle,
          ),
          alignment: Alignment.center,
          child: Text(
            athlete.position.isNotEmpty ? athlete.position[0] : '?',
            style: TextStyle(color: positionColor, fontWeight: FontWeight.bold, fontSize: 11),
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: Text(
            athlete.name,
            style: theme.textTheme.bodySmall?.copyWith(fontWeight: FontWeight.w500),
            overflow: TextOverflow.ellipsis,
          ),
        ),
        if (athlete.number > 0)
          Text(
            '#${athlete.number}',
            style: theme.textTheme.labelSmall,
          ),
      ],
    );
  }
}
