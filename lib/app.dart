import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'router/app_router.dart';

class GoalBoozeApp extends StatelessWidget {
  const GoalBoozeApp({super.key});

  @override
  Widget build(BuildContext context) {
    final router = AppRouter().router;

    return MaterialApp.router(
      title: 'GoalBooze',
      debugShowCheckedModeBanner: false,
      routerConfig: router,
      localizationsDelegates: const [
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: const [
        Locale('de'),
        Locale('en'),
      ],
      locale: const Locale('de'),
      theme: _buildTheme(),
    );
  }

  ThemeData _buildTheme() {
    const background = Color(0xFF0D0D1A);
    const surface = Color(0xFF1A1A2E);
    const primary = Color(0xFF00D4AA);
    const secondary = Color(0xFFFF6B35);
    const onPrimary = Color(0xFF0D0D1A);
    const onBackground = Color(0xFFF0F0F0);
    const muted = Color(0xFF888899);

    final colorScheme = ColorScheme.dark(
      brightness: Brightness.dark,
      primary: primary,
      onPrimary: onPrimary,
      secondary: secondary,
      onSecondary: Colors.white,
      surface: surface,
      onSurface: onBackground,
      surfaceContainerHighest: const Color(0xFF252540),
      onSurfaceVariant: muted,
      error: const Color(0xFFCF6679),
      outline: const Color(0xFF3A3A5C),
    );

    return ThemeData(
      useMaterial3: true,
      colorScheme: colorScheme,
      scaffoldBackgroundColor: background,
      cardTheme: CardThemeData(
        color: surface,
        elevation: 0,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
          side: const BorderSide(color: Color(0xFF2A2A45), width: 1),
        ),
        margin: const EdgeInsets.only(bottom: 12),
      ),
      appBarTheme: const AppBarTheme(
        backgroundColor: background,
        foregroundColor: onBackground,
        elevation: 0,
        surfaceTintColor: Colors.transparent,
      ),
      bottomNavigationBarTheme: BottomNavigationBarThemeData(
        backgroundColor: surface,
        selectedItemColor: primary,
        unselectedItemColor: muted,
        type: BottomNavigationBarType.fixed,
        elevation: 0,
      ),
      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(
          backgroundColor: primary,
          foregroundColor: onPrimary,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 14),
          textStyle: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: primary,
          side: const BorderSide(color: primary),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: surface,
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: Color(0xFF3A3A5C)),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: Color(0xFF3A3A5C)),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: primary, width: 2),
        ),
        labelStyle: const TextStyle(color: muted),
        hintStyle: const TextStyle(color: muted),
      ),
      chipTheme: ChipThemeData(
        backgroundColor: surface,
        selectedColor: primary.withValues(alpha: 0.2),
        labelStyle: const TextStyle(color: onBackground),
        side: const BorderSide(color: Color(0xFF3A3A5C)),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      ),
      dividerTheme: const DividerThemeData(
        color: Color(0xFF2A2A45),
        thickness: 1,
      ),
      textTheme: const TextTheme(
        displaySmall: TextStyle(color: onBackground, fontWeight: FontWeight.bold),
        headlineMedium: TextStyle(color: onBackground, fontWeight: FontWeight.bold),
        titleLarge: TextStyle(color: onBackground, fontWeight: FontWeight.w600),
        titleMedium: TextStyle(color: onBackground),
        bodyLarge: TextStyle(color: onBackground),
        bodyMedium: TextStyle(color: onBackground),
        bodySmall: TextStyle(color: muted),
        labelSmall: TextStyle(color: muted),
      ),
    );
  }
}
