import 'package:flutter/material.dart';

class RulesScreen extends StatelessWidget {
  const RulesScreen({super.key});

  static const _rules = [
    (
      icon: Icons.sports_soccer,
      title: 'Tor!',
      desc: 'Erzielt einer deiner Athleten ein Tor → 2 Schlücke trinken.',
      color: Color(0xFF00D4AA),
    ),
    (
      icon: Icons.block,
      title: 'Rote Karte!',
      desc: 'Fliegt einer deiner Athleten vom Platz → 5 Schlücke trinken.',
      color: Color(0xFFCF6679),
    ),
    (
      icon: Icons.swap_horiz,
      title: 'Auswechslung',
      desc: 'Wird dein Athlet ausgewechselt, übernimmst du den eingewechselten Spieler.',
      color: Color(0xFF4FC3F7),
    ),
    (
      icon: Icons.shield,
      title: 'Torwart-Gegentor',
      desc: 'Ist dein Athlet ein Torwart und kassiert ein Gegentor → 1 Schluck trinken.',
      color: Color(0xFFFFD700),
    ),
    (
      icon: Icons.shuffle,
      title: 'Zuweisung',
      desc: 'Jeder Spieler bekommt pro Spiel einen Athleten aus Heim- und Gastmannschaft.',
      color: Color(0xFFFF6B35),
    ),
  ];

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(title: const Text('Spielregeln')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          ..._rules.asMap().entries.map((entry) {
            final i = entry.key;
            final rule = entry.value;
            return _RuleCard(
              number: i + 1,
              icon: rule.icon,
              title: rule.title,
              description: rule.desc,
              accentColor: rule.color,
            );
          }),
          const SizedBox(height: 8),
          Container(
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
              color: theme.colorScheme.surface,
              borderRadius: BorderRadius.circular(14),
              border: Border.all(color: theme.colorScheme.outline),
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Icon(Icons.info_outline,
                    size: 18,
                    color: theme.colorScheme.onSurfaceVariant),
                const SizedBox(width: 10),
                Expanded(
                  child: Text(
                    'Eigene Regeln sind natürlich erlaubt — '
                    'dann ist es aber nicht mehr das originale GoalBooze!',
                    style: theme.textTheme.bodySmall?.copyWith(
                      fontStyle: FontStyle.italic,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _RuleCard extends StatelessWidget {
  final int number;
  final IconData icon;
  final String title;
  final String description;
  final Color accentColor;

  const _RuleCard({
    required this.number,
    required this.icon,
    required this.title,
    required this.description,
    required this.accentColor,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                color: accentColor.withValues(alpha: 0.15),
                shape: BoxShape.circle,
              ),
              child: Icon(icon, color: accentColor, size: 20),
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Container(
                        width: 20,
                        height: 20,
                        decoration: BoxDecoration(
                          color: accentColor,
                          shape: BoxShape.circle,
                        ),
                        alignment: Alignment.center,
                        child: Text(
                          '$number',
                          style: const TextStyle(
                            color: Color(0xFF0D0D1A),
                            fontSize: 11,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                      const SizedBox(width: 8),
                      Text(
                        title,
                        style: theme.textTheme.titleSmall
                            ?.copyWith(fontWeight: FontWeight.bold),
                      ),
                    ],
                  ),
                  const SizedBox(height: 6),
                  Text(
                    description,
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: theme.colorScheme.onSurface,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
