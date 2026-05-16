import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:goalbooze/model/player.dart';
import 'package:goalbooze/provider/game_provider.dart';
import 'package:goalbooze/provider/leagues_provider.dart';
import 'package:goalbooze/widget/match_tile.dart';

class SetupScreen extends ConsumerStatefulWidget {
  const SetupScreen({super.key});

  @override
  ConsumerState<SetupScreen> createState() => _SetupScreenState();
}

class _SetupScreenState extends ConsumerState<SetupScreen>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;
  final _nameController = TextEditingController();
  int _selectedLeagueId = 1;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    // Reset any previous game state
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(gameProvider.notifier).reset();
    });
  }

  @override
  void dispose() {
    _tabController.dispose();
    _nameController.dispose();
    super.dispose();
  }

  void _addPlayer() {
    final name = _nameController.text.trim();
    if (name.isEmpty) return;
    ref.read(gameProvider.notifier).addPlayer(name);
    _nameController.clear();
  }

  bool get _canProceed {
    final state = ref.read(gameProvider);
    return state.players.isNotEmpty && state.selectedEvents.isNotEmpty;
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final gameState = ref.watch(gameProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Neues Spiel'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => context.pop(),
        ),
        actions: [
          Padding(
            padding: const EdgeInsets.only(right: 12),
            child: FilledButton(
              onPressed: _canProceed
                  ? () => context.push('/new-game/lineup')
                  : null,
              style: FilledButton.styleFrom(
                minimumSize: const Size(90, 36),
                padding: const EdgeInsets.symmetric(horizontal: 16),
              ),
              child: const Text('Weiter'),
            ),
          ),
        ],
        bottom: TabBar(
          controller: _tabController,
          labelColor: theme.colorScheme.primary,
          unselectedLabelColor: theme.colorScheme.onSurfaceVariant,
          indicatorColor: theme.colorScheme.primary,
          indicatorSize: TabBarIndicatorSize.tab,
          tabs: [
            Tab(
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(Icons.group, size: 18),
                  const SizedBox(width: 6),
                  Text('Spieler (${gameState.players.length})'),
                ],
              ),
            ),
            Tab(
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(Icons.sports_soccer, size: 18),
                  const SizedBox(width: 6),
                  Text('Spiele (${gameState.selectedEvents.length})'),
                ],
              ),
            ),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          _PlayersTab(
            controller: _nameController,
            onAdd: _addPlayer,
            players: gameState.players,
            onRemove: (name) => ref.read(gameProvider.notifier).removePlayer(name),
          ),
          _EventsTab(
            selectedLeagueId: _selectedLeagueId,
            onLeagueChanged: (id) => setState(() => _selectedLeagueId = id),
          ),
        ],
      ),
    );
  }
}

class _PlayersTab extends StatelessWidget {
  final TextEditingController controller;
  final VoidCallback onAdd;
  final List<Player> players;
  final void Function(String) onRemove;

  const _PlayersTab({
    required this.controller,
    required this.onAdd,
    required this.players,
    required this.onRemove,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        children: [
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: controller,
                  decoration: const InputDecoration(
                    hintText: 'Name eingeben...',
                    prefixIcon: Icon(Icons.person_add_outlined),
                  ),
                  onSubmitted: (_) => onAdd(),
                  textCapitalization: TextCapitalization.words,
                ),
              ),
              const SizedBox(width: 10),
              FilledButton(
                onPressed: onAdd,
                style: FilledButton.styleFrom(
                  minimumSize: const Size(48, 54),
                  padding: EdgeInsets.zero,
                ),
                child: const Icon(Icons.add),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            'Maximal 10 Spieler',
            style: theme.textTheme.bodySmall,
          ),
          const SizedBox(height: 16),
          Expanded(
            child: players.isEmpty
                ? Center(
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.group_outlined,
                            size: 48,
                            color: theme.colorScheme.onSurfaceVariant),
                        const SizedBox(height: 12),
                        Text(
                          'Noch keine Spieler',
                          style: theme.textTheme.bodyMedium?.copyWith(
                            color: theme.colorScheme.onSurfaceVariant,
                          ),
                        ),
                      ],
                    ),
                  )
                : ListView.separated(
                    itemCount: players.length,
                    separatorBuilder: (_, __) => const SizedBox(height: 8),
                    itemBuilder: (_, i) {
                      final player = players[i];
                      return Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 14, vertical: 10),
                        decoration: BoxDecoration(
                          color: theme.colorScheme.surface,
                          borderRadius: BorderRadius.circular(12),
                          border:
                              Border.all(color: theme.colorScheme.outline),
                        ),
                        child: Row(
                          children: [
                            CircleAvatar(
                              radius: 16,
                              backgroundColor: theme.colorScheme.primary
                                  .withValues(alpha: 0.2),
                              child: Text(
                                player.name[0].toUpperCase(),
                                style: TextStyle(
                                  color: theme.colorScheme.primary,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ),
                            const SizedBox(width: 12),
                            Expanded(
                              child: Text(
                                player.name,
                                style: theme.textTheme.bodyMedium,
                              ),
                            ),
                            IconButton(
                              icon: const Icon(Icons.close, size: 18),
                              color: theme.colorScheme.onSurfaceVariant,
                              onPressed: () => onRemove(player.name),
                              padding: EdgeInsets.zero,
                              constraints: const BoxConstraints(),
                            ),
                          ],
                        ),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }
}

class _EventsTab extends ConsumerWidget {
  final int selectedLeagueId;
  final void Function(int) onLeagueChanged;

  const _EventsTab({
    required this.selectedLeagueId,
    required this.onLeagueChanged,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final theme = Theme.of(context);
    final leaguesAsync = ref.watch(leaguesProvider);
    final eventsAsync = ref.watch(eventsByLeagueProvider(selectedLeagueId));
    final gameState = ref.watch(gameProvider);

    return Column(
      children: [
        // League selector
        leaguesAsync.when(
          loading: () => const LinearProgressIndicator(),
          error: (_, __) => const SizedBox.shrink(),
          data: (leagues) => SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
            child: Row(
              children: leagues.map((league) {
                final isSelected = league.id == selectedLeagueId;
                return Padding(
                  padding: const EdgeInsets.only(right: 8),
                  child: FilterChip(
                    label: Text(league.name),
                    selected: isSelected,
                    onSelected: (_) => onLeagueChanged(league.id),
                  ),
                );
              }).toList(),
            ),
          ),
        ),
        Expanded(
          child: eventsAsync.when(
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (e, _) => Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(Icons.cloud_off,
                      size: 48,
                      color: theme.colorScheme.onSurfaceVariant),
                  const SizedBox(height: 12),
                  Text('Spiele nicht verfügbar',
                      style: theme.textTheme.bodyMedium),
                  const SizedBox(height: 4),
                  Text(e.toString(),
                      style: theme.textTheme.bodySmall,
                      textAlign: TextAlign.center),
                ],
              ),
            ),
            data: (events) {
              final upcoming = events
                  .where((e) =>
                      e.status == 'scheduled' ||
                      e.date.isAfter(DateTime.now().subtract(
                          const Duration(hours: 3))))
                  .toList()
                ..sort((a, b) => a.date.compareTo(b.date));

              if (upcoming.isEmpty) {
                return Center(
                  child: Text(
                    'Keine bevorstehenden Spiele',
                    style: theme.textTheme.bodyMedium?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                );
              }

              return ListView.builder(
                padding: const EdgeInsets.all(16),
                itemCount: upcoming.length,
                itemBuilder: (_, i) {
                  final event = upcoming[i];
                  final isSelected =
                      gameState.selectedEvents.any((e) => e.id == event.id);
                  return MatchTile(
                    event: event,
                    isSelected: isSelected,
                    onTap: () =>
                        ref.read(gameProvider.notifier).toggleEvent(event),
                  );
                },
              );
            },
          ),
        ),
      ],
    );
  }
}
