import 'package:flutter/material.dart';

class LegalNoticeScreen extends StatelessWidget {
  const LegalNoticeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(title: const Text('Impressum')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Text(
            'Impressum',
            style: theme.textTheme.titleLarge?.copyWith(
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 16),
          const Text('GoalBooze – Das Party-Trinkspiel für Sport-Freunde'),
          const SizedBox(height: 24),
          Text(
            'Haftungsausschluss',
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 8),
          const Text(
            'GoalBooze ist ein reines Freizeit-Spielzeug. '
            'Die App soll zu verantwortungsvollem Trinken anregen. '
            'Bitte konsumiere Alkohol nur in Maßen und verantworte '
            'deinen Konsum selbst. Das Spiel ist für Personen ab 18 Jahren.',
          ),
          const SizedBox(height: 24),
          Text(
            'Daten & API',
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 8),
          const Text(
            'Die Spieldaten werden über die API-Football-Schnittstelle '
            '(api-football.com) bezogen. Es werden keine personenbezogenen '
            'Daten gespeichert oder an Dritte weitergegeben.',
          ),
          const SizedBox(height: 24),
          Text(
            'Quellcode',
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 8),
          const Text(
            'Diese App ist ein Open-Source-Projekt. '
            'Der vollständige Quellcode ist auf GitHub verfügbar.',
          ),
        ],
      ),
    );
  }
}
