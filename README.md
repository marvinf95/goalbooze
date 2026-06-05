# GoalBooze

Party drinking game for football match nights. Every player is randomly assigned athletes from the starting lineups — goal, red card, substitution: time to drink.

---

## Rules

| Event | Rule |
|-------|------|
| Goal | 2 sips |
| Red card | 5 sips |
| Substitution | You take over the substituted-in player |
| Goalkeeper concedes | 1 sip |
| Assignment | Everyone gets 1 athlete per team per match |

---

## Tech stack

| Component | Technology |
|-----------|------------|
| Frontend | Flutter (Dart) — mobile & web |
| State management | Riverpod |
| Routing | go_router |
| Backend | Go 1.22 + chi v5 |
| Database | SQLite (modernc.org/sqlite) |
| Fixtures & squads | [football-data.org](https://www.football-data.org) |
| Live lineups | Google Gemini (free tier) → Anthropic Claude fallback, both with web search |

---

## How lineups are resolved

Clubs publish their starting eleven only ~90 minutes before kickoff, and no free API delivers them earlier. GoalBooze resolves a lineup in this order:

1. **Permanent cache** — once fetched, a match lineup is stored and never re-fetched.
2. **AI provider chain** — **Gemini** (free tier, tried first to save tokens) then **Claude** as fallback, each using web search to find the official lineup. Configure either, both, or neither.
3. **Squad fallback** — a random eleven from the cached season squad.
4. **Manual mode** — create a match yourself and type in both teams and their players.

Set `LINEUP_MOCK=true` to serve deterministic lineups without any external API calls (useful for local development).

---

## Local development

### Prerequisites

- Go 1.22+
- Flutter 3.x

### Run the backend

```bash
cd backend
cp .env.example .env
# Optional: add API keys to .env (see below)
go run ./cmd/server
```

The backend runs on `http://localhost:8080`.

### Run the frontend

```bash
# In the browser (recommended)
flutter run -d chrome

# As a desktop app (Linux/macOS/Windows)
flutter run -d linux
```

The frontend connects to `http://localhost:8080` by default.
For other backends: `flutter run --dart-define=API_BASE_URL=https://your-domain.com`

> On WSL2, `flutter run -d chrome` may render a blank page (the in-WSL browser has no WebGL). Use `flutter run -d web-server --web-port=5000` and open the URL in your Windows browser instead.

### Project structure

```
├── backend/
│   ├── cmd/server/       # Entry point
│   ├── config/           # Configuration & league definitions
│   ├── internal/
│   │   ├── client/       # football-data.org HTTP client
│   │   ├── handler/      # HTTP handlers
│   │   ├── model/        # Data models
│   │   ├── repository/   # SQLite persistence
│   │   └── service/      # Business logic + AI lineup providers
│   └── .env.example
├── lib/
│   ├── model/            # Dart models
│   ├── provider/         # Riverpod providers
│   ├── router/           # App navigation
│   ├── screen/           # Screens
│   ├── service/          # API service
│   └── widget/           # Reusable widgets
└── assets/
```

---

## Configuration

All backend configuration is via environment variables (see [`backend/.env.example`](backend/.env.example)).

### football-data.org (recommended)

Register a free key at [football-data.org/client/register](https://www.football-data.org/client/register).
Without a key, the app serves a few built-in test fixtures.

```env
FOOTBALL_DATA_API_KEY=your-key
```

### Gemini (recommended — tried first)

Enables AI-powered live lineups via Google Gemini's free tier with built-in Google Search grounding. Tried before Claude to save tokens. Get a key at [aistudio.google.com](https://aistudio.google.com).

```env
GEMINI_API_KEY=...
GEMINI_MODEL=gemini-2.5-flash   # default
```

### Anthropic / Claude (optional fallback)

Used only if Gemini is unconfigured or fails. Claude searches for the official lineup ~90 minutes before kickoff and returns it structured; the result is cached permanently. ~$0.001 per lineup fetch (Claude Haiku).

```env
ANTHROPIC_API_KEY=sk-ant-...
ANTHROPIC_MODEL=claude-haiku-4-5-20251001   # default
```

### Other

```env
LINEUP_MOCK=true                # serve mock lineups, no external API calls
CORS_ALLOWED_ORIGINS=https://your-domain.com   # empty = *
PORT=8080
```

---

## Supported leagues

| League | Code |
|--------|------|
| 1. Bundesliga | BL1 |
| 2. Bundesliga | BL2 |
| Champions League | CL |
| World Cup 2026 | WC |

For the World Cup, full 26-man squads for all 48 teams are bundled and seeded on startup (`backend/data/wc2026_squads.json`). Add more leagues in `backend/config/config.go`.

---

## Game modes

- **From the schedule** — pick upcoming matches from a league; lineups are resolved via the provider chain above.
- **Manual** — create a match yourself: enter both team names and at least one player per team. No external data needed; everyone still gets one home and one away athlete assigned.

---

## Deployment (Docker)

```bash
cd backend
docker build -t goalbooze-backend .
docker run -p 8080:8080 \
  -e FOOTBALL_DATA_API_KEY=... \
  -e GEMINI_API_KEY=... \
  -e CORS_ALLOWED_ORIGINS=https://your-domain.com \
  -v $(pwd)/data:/data \
  goalbooze-backend
```

Flutter web build for production:

```bash
flutter build web --dart-define=API_BASE_URL=https://your-domain.com
```

---

## Background

The idea was born watching the Saturday Bundesliga conference — who gets which player, and when do you have to drink? The tedious manual prep (looking up lineups, generating random numbers, keeping track) was the motivation for this app.

---

## License

[MIT](LICENSE)
