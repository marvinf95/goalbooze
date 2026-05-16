import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

class ScaffoldWithBottomNav extends StatelessWidget {
  final Widget child;

  const ScaffoldWithBottomNav({super.key, required this.child});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: child,
      bottomNavigationBar: _BottomNav(),
    );
  }
}

class _BottomNav extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    final location = GoRouterState.of(context).uri.toString();
    final theme = Theme.of(context);

    int currentIndex = 0;
    if (location.startsWith('/rules')) {
      currentIndex = 1;
    } else if (location.startsWith('/history')) {
      currentIndex = 2;
    }

    return Container(
      decoration: BoxDecoration(
        border: Border(
          top: BorderSide(color: theme.colorScheme.outline, width: 1),
        ),
      ),
      child: BottomNavigationBar(
        currentIndex: currentIndex,
        onTap: (index) {
          switch (index) {
            case 0:
              context.go('/');
            case 1:
              context.go('/rules');
            case 2:
              context.go('/history');
          }
        },
        items: const [
          BottomNavigationBarItem(
            icon: Icon(Icons.sports_soccer),
            label: 'Spiel',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.menu_book),
            label: 'Regeln',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.history),
            label: 'Verlauf',
          ),
        ],
      ),
    );
  }
}
