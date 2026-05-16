# GoalBooze

Party-Trinkspiel für Bundesliga-Abende. Jeder Spieler bekommt zufällig Athleten aus den Startaufstellungen zugewiesen — Tor, Rote Karte, Auswechslung: Es wird getrunken.

---

## Spielregeln

| Ereignis | Regel |
|----------|-------|
| Tor | 2 Schlücke |
| Rote Karte | 5 Schlücke |
| Auswechslung | Du übernimmst den eingewechselten Spieler |
| Torwart kassiert Gegentor | 1 Schluck |
| Zuweisung | Jeder bekommt 1 Athlet pro Team pro Spiel |

---

## Tech-Stack

| Komponente | Technologie |
|------------|-------------|
| Frontend | Flutter (Dart) — Mobile & Web |
| State Management | Riverpod |
| Routing | go_router |
| Backend | Go 1.22 + chi v5 |
| Datenbank | SQLite (modernc.org/sqlite) |
| Spielpläne & Kader | [football-data.org](https://www.football-data.org) |
| Live-Lineups | Claude API (Haiku) + WebSearch |

---

## Lokale Entwicklung

### Voraussetzungen

- Go 1.22+
- Flutter 3.x

### Backend starten

```bash
cd backend
cp .env.example .env
# Optional: API-Keys in .env eintragen (siehe unten)
go run ./cmd/server
```

Backend läuft auf `http://localhost:8080`.

### Frontend starten

```bash
# Im Browser (empfohlen)
flutter run -d chrome

# Als Desktop-App (Linux/macOS/Windows)
flutter run -d linux
```

Das Frontend verbindet sich automatisch mit `http://localhost:8080`.  
Für andere Backends: `flutter run --dart-define=API_BASE_URL=https://deine-domain.com`

### Projektstruktur

```
├── backend/
│   ├── cmd/server/       # Einstiegspunkt
│   ├── config/           # Konfiguration & Liga-Definitionen
│   ├── internal/
│   │   ├── client/       # football-data.org HTTP-Client
│   │   ├── handler/      # HTTP-Handler
│   │   ├── model/        # Datenmodelle
│   │   ├── repository/   # SQLite-Datenbankschicht
│   │   └── service/      # Geschäftslogik + KI-Lineup-Service
│   └── .env.example
├── lib/
│   ├── model/            # Dart-Modelle
│   ├── provider/         # Riverpod-Provider
│   ├── router/           # App-Navigation
│   ├── screen/           # Screens
│   ├── service/          # API-Service
│   └── widget/           # Wiederverwendbare Widgets
└── assets/
```

---

## API-Keys

### football-data.org (empfohlen)

Kostenlosen Key unter [football-data.org/client/register](https://www.football-data.org/client/register) registrieren.  
Ohne Key werden einige eingebaute Testspiele angezeigt.

```env
FOOTBALL_DATA_API_KEY=dein-key
```

### Anthropic API (optional)

Aktiviert KI-gestützte Live-Startaufstellungen: Claude sucht ~90 Minuten vor Kickoff die offizielle Aufstellung auf Sportseiten und gibt sie strukturiert zurück. Das Ergebnis wird permanent gecacht (einmaliger Abruf pro Spiel).

Ohne Key werden 11 zufällige Spieler aus dem Saisonkader gewählt.

```env
ANTHROPIC_API_KEY=sk-ant-...
```

Kosten: ca. $0,001 pro Lineup-Abruf (Claude Haiku).

---

## Unterstützte Ligen

| Liga | Code |
|------|------|
| 1. Bundesliga | BL1 |
| 2. Bundesliga | BL2 |
| Champions League | CL |

Weitere Ligen lassen sich in `backend/config/config.go` ergänzen.

---

## Deployment (Docker)

```bash
cd backend
docker build -t goalbooze-backend .
docker run -p 8080:8080 \
  -e FOOTBALL_DATA_API_KEY=... \
  -e ANTHROPIC_API_KEY=... \
  -e CORS_ALLOWED_ORIGINS=https://deine-domain.com \
  -v $(pwd)/data:/data \
  goalbooze-backend
```

Flutter-Web-Build für Produktion:

```bash
flutter build web --dart-define=API_BASE_URL=https://deine-domain.com
```

---

## Geschichte

Die Idee entstand beim Gucken der Samstagskonferenz — wer bekommt welchen Spieler, und wann muss getrunken werden? Das lästige manuelle Vorbereiten (Aufstellungen suchen, Zufallszahlen generieren, mitschreiben) war der Anlass für diese App.
