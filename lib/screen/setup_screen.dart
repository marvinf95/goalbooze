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

  // Manual mode: self-entered match + athletes (no football-data).
  bool _manualMode = false;
  final _homeTeamController = TextEditingController();
  final _awayTeamController = TextEditingController();
  final List<String> _homeAthletes = [];
  final List<String> _awayAthletes = [];

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    // Re-evaluate the proceed button as team names are typed.
    _homeTeamController.addListener(_onChanged);
    _awayTeamController.addListener(_onChanged);
    // Reset any previous game state
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(gameProvider.notifier).reset();
    });
  }

  @override
  void dispose() {
    _tabController.dispose();
    _nameController.dispose();
    _homeTeamController.dispose();
    _awayTeamController.dispose();
    super.dispose();
  }

  void _onChanged() => setState(() {});

  void _addPlayer() {
    final name = _nameController.text.trim();
    if (name.isEmpty) return;
    ref.read(gameProvider.notifier).addPlayer(name);
    _nameController.clear();
  }

  bool get _canProceed {
    final state = ref.read(gameProvider);
    if (state.players.isEmpty) return false;
    if (_manualMode) {
      return _homeTeamController.text.trim().isNotEmpty &&
          _awayTeamController.text.trim().isNotEmpty &&
          _homeAthletes.isNotEmpty &&
          _awayAthletes.isNotEmpty;
    }
    return state.selectedEvents.isNotEmpty;
  }

  Future<void> _startManualGame() async {
    final notifier = ref.read(gameProvider.notifier);
    notifier.setManualMatch(
      homeTeam: _homeTeamController.text,
      awayTeam: _awayTeamController.text,
      homeAthletes: _homeAthletes,
      awayAthletes: _awayAthletes,
    );
    await notifier.createGame(manual: true);
    if (!mounted) return;
    final state = ref.read(gameProvider);
    if (state.currentGame != null) {
      context.go('/game/${state.currentGame!.id}');
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(state.error ?? 'Spiel konnte nicht erstellt werden')),
      );
    }
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
              onPressed: !_canProceed || gameState.isLoading
                  ? null
                  : (_manualMode
                      ? _startManualGame
                      : () => context.push('/new-game/lineup')),
              style: FilledButton.styleFrom(
                minimumSize: const Size(90, 36),
                padding: const EdgeInsets.symmetric(horizontal: 16),
              ),
              child: gameState.isLoading
                  ? const SizedBox(
                      width: 18,
                      height: 18,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : Text(_manualMode ? 'Spiel starten' : 'Weiter'),
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
                  Icon(_manualMode ? Icons.edit : Icons.sports_soccer, size: 18),
                  const SizedBox(width: 6),
                  Text(_manualMode
                      ? 'Match'
                      : 'Spiele (${gameState.selectedEvents.length})'),
                ],
              ),
            ),
          ],
        ),
      ),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
            child: SegmentedButton<bool>(
              segments: const [
                ButtonSegment(
                  value: false,
                  label: Text('Aus Spielplan'),
                  icon: Icon(Icons.calendar_month, size: 18),
                ),
                ButtonSegment(
                  value: true,
                  label: Text('Manuell'),
                  icon: Icon(Icons.edit, size: 18),
                ),
              ],
              selected: {_manualMode},
              onSelectionChanged: (s) => setState(() => _manualMode = s.first),
            ),
          ),
          Expanded(
            child: TabBarView(
              controller: _tabController,
              children: [
                _PlayersTab(
                  controller: _nameController,
                  onAdd: _addPlayer,
                  players: gameState.players,
                  onRemove: (name) =>
                      ref.read(gameProvider.notifier).removePlayer(name),
                ),
                _manualMode
                    ? _ManualMatchTab(
                        homeTeamController: _homeTeamController,
                        awayTeamController: _awayTeamController,
                        homeAthletes: _homeAthletes,
                        awayAthletes: _awayAthletes,
                        onAddHome: (n) => setState(() => _homeAthletes.add(n)),
                        onRemoveHome: (i) =>
                            setState(() => _homeAthletes.removeAt(i)),
                        onAddAway: (n) => setState(() => _awayAthletes.add(n)),
                        onRemoveAway: (i) =>
                            setState(() => _awayAthletes.removeAt(i)),
                      )
                    : _EventsTab(
                        selectedLeagueId: _selectedLeagueId,
                        onLeagueChanged: (id) =>
                            setState(() => _selectedLeagueId = id),
                      ),
              ],
            ),
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

class _ManualMatchTab extends StatefulWidget {
  final TextEditingController homeTeamController;
  final TextEditingController awayTeamController;
  final List<String> homeAthletes;
  final List<String> awayAthletes;
  final void Function(String) onAddHome;
  final void Function(int) onRemoveHome;
  final void Function(String) onAddAway;
  final void Function(int) onRemoveAway;

  const _ManualMatchTab({
    required this.homeTeamController,
    required this.awayTeamController,
    required this.homeAthletes,
    required this.awayAthletes,
    required this.onAddHome,
    required this.onRemoveHome,
    required this.onAddAway,
    required this.onRemoveAway,
  });

  @override
  State<_ManualMatchTab> createState() => _ManualMatchTabState();
}

class _ManualMatchTabState extends State<_ManualMatchTab> {
  final _homeAddController = TextEditingController();
  final _awayAddController = TextEditingController();

  @override
  void dispose() {
    _homeAddController.dispose();
    _awayAddController.dispose();
    super.dispose();
  }

  void _addHome() {
    final name = _homeAddController.text.trim();
    if (name.isEmpty) return;
    widget.onAddHome(name);
    _homeAddController.clear();
  }

  void _addAway() {
    final name = _awayAddController.text.trim();
    if (name.isEmpty) return;
    widget.onAddAway(name);
    _awayAddController.clear();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Text(
          'Trage beide Mannschaften und ihre Spieler ein. Jeder Gast bekommt '
          'beim Start zufällig je einen Heim- und einen Gast-Spieler.',
          style: theme.textTheme.bodySmall,
        ),
        const SizedBox(height: 16),
        _TeamCard(
          accent: const Color(0xFF4C8DFF),
          label: 'Heimmannschaft',
          teamController: widget.homeTeamController,
          addController: _homeAddController,
          athletes: widget.homeAthletes,
          onAdd: _addHome,
          onRemove: widget.onRemoveHome,
        ),
        const SizedBox(height: 16),
        _TeamCard(
          accent: const Color(0xFFFF6B35),
          label: 'Gastmannschaft',
          teamController: widget.awayTeamController,
          addController: _awayAddController,
          athletes: widget.awayAthletes,
          onAdd: _addAway,
          onRemove: widget.onRemoveAway,
        ),
      ],
    );
  }
}

class _TeamCard extends StatelessWidget {
  final Color accent;
  final String label;
  final TextEditingController teamController;
  final TextEditingController addController;
  final List<String> athletes;
  final VoidCallback onAdd;
  final void Function(int) onRemove;

  const _TeamCard({
    required this.accent,
    required this.label,
    required this.teamController,
    required this.addController,
    required this.athletes,
    required this.onAdd,
    required this.onRemove,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: theme.colorScheme.outline),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(width: 10, height: 10,
                  decoration: BoxDecoration(color: accent, shape: BoxShape.circle)),
              const SizedBox(width: 8),
              Text(label,
                  style: theme.textTheme.labelSmall?.copyWith(color: accent)),
            ],
          ),
          const SizedBox(height: 8),
          TextField(
            controller: teamController,
            decoration: const InputDecoration(hintText: 'Mannschaftsname'),
            textCapitalization: TextCapitalization.words,
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: addController,
                  decoration: const InputDecoration(
                    hintText: 'Spielername...',
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
          const SizedBox(height: 12),
          if (athletes.isEmpty)
            Text('Noch keine Spieler', style: theme.textTheme.bodySmall)
          else
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                for (var i = 0; i < athletes.length; i++)
                  Chip(
                    label: Text(athletes[i]),
                    onDeleted: () => onRemove(i),
                    deleteIcon: const Icon(Icons.close, size: 16),
                  ),
              ],
            ),
        ],
      ),
    );
  }
}

class _EventsTab extends ConsumerStatefulWidget {
  final int selectedLeagueId;
  final void Function(int) onLeagueChanged;

  const _EventsTab({
    required this.selectedLeagueId,
    required this.onLeagueChanged,
  });

  @override
  ConsumerState<_EventsTab> createState() => _EventsTabState();
}

class _EventsTabState extends ConsumerState<_EventsTab> {
  // When false, only games of the current day are shown.
  bool _showAll = false;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final leaguesAsync = ref.watch(leaguesProvider);
    final eventsAsync =
        ref.watch(eventsByLeagueProvider(widget.selectedLeagueId));
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
                final isSelected = league.id == widget.selectedLeagueId;
                return Padding(
                  padding: const EdgeInsets.only(right: 8),
                  child: FilterChip(
                    label: Text(league.name),
                    selected: isSelected,
                    onSelected: (_) => widget.onLeagueChanged(league.id),
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

              final now = DateTime.now();
              final visible = _showAll
                  ? upcoming
                  : upcoming.where((e) => e.isOnSameLocalDay(now)).toList();

              if (visible.isEmpty) {
                return Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        _showAll
                            ? 'Keine bevorstehenden Spiele'
                            : 'Heute keine Spiele',
                        style: theme.textTheme.bodyMedium?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        ),
                      ),
                      if (!_showAll && upcoming.isNotEmpty) ...[
                        const SizedBox(height: 8),
                        TextButton.icon(
                          onPressed: () => setState(() => _showAll = true),
                          icon: const Icon(Icons.calendar_month, size: 18),
                          label: const Text('Alle Spiele anzeigen'),
                        ),
                      ],
                    ],
                  ),
                );
              }

              return Column(
                children: [
                  Align(
                    alignment: Alignment.centerRight,
                    child: Padding(
                      padding:
                          const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
                      child: TextButton.icon(
                        onPressed: () => setState(() => _showAll = !_showAll),
                        icon: Icon(
                          _showAll ? Icons.today : Icons.calendar_month,
                          size: 18,
                        ),
                        label: Text(
                          _showAll ? 'Nur heutige Spiele' : 'Alle Spiele anzeigen',
                        ),
                      ),
                    ),
                  ),
                  Expanded(
                    child: ListView.builder(
                      padding: const EdgeInsets.fromLTRB(16, 4, 16, 16),
                      itemCount: visible.length,
                      itemBuilder: (_, i) {
                        final event = visible[i];
                        final isSelected = gameState.selectedEvents
                            .any((e) => e.id == event.id);
                        return MatchTile(
                          event: event,
                          isSelected: isSelected,
                          onTap: () => ref
                              .read(gameProvider.notifier)
                              .toggleEvent(event),
                        );
                      },
                    ),
                  ),
                ],
              );
            },
          ),
        ),
      ],
    );
  }
}
