# Architecture Reference

## System Overview

```
Browser (Next.js — pure view layer)
    │
    │  REST  → /api/auth/*, /api/round/*, /api/debts/*, /api/admin/*
    │  WS    ← ws://{host}/ws  (all live state pushed by backend)
    ▼
Go Backend (Gin)
    │  All business logic:
    │  - session management (in-memory SessionStore)
    │  - game state machine (in-memory StateMachine)
    │  - Gaussian payout math
    │  - P2P debt graph (greedy algorithm)
    │  - WebSocket hub (broadcast to all clients)
    ▼
Postgres
    │  Persistent tables: users, rounds, bets, debts
```

## Backend Package Layout

| Package | Responsibility |
|---|---|
| `cmd/server` | Entrypoint: DB init, migrations, router wiring |
| `internal/model` | Plain Go structs: User, Session, Round, Bet, Debt |
| `internal/repository` | SQL queries (no ORM) |
| `internal/game` | Pure-function business logic (payout, debt, distribution, state machine) |
| `internal/handler` | Gin handlers + WebSocket hub + session store + middleware |
| `internal/questions` | Hardcoded 50-question bank (NHL + Champions League) |

## State Machine

```
IDLE
 └─► ROUND_OPEN   (host calls POST /api/admin/round/open)
       └─► ROUND_CALLED  (host calls POST /api/admin/round/call)
             └─► SETTLING  (automatic after payout + debt generation)
                   └─► IDLE  (automatic when all debts confirmed)
```

State lives in `game.StateMachine` (sync.RWMutex-protected). Round ID is co-located with state so handlers always know which round is active without a DB query.

## Session Management

- `handler.SessionStore`: in-memory `map[token]*model.Session`
- Token: 32-byte cryptographically random hex string
- Max 9 concurrent sessions (1 host + 8 players)
- Sessions expire only on explicit logout; server restart clears them (acceptable for demo)

## WebSocket Hub

- `handler.Hub`: fan-out broadcaster
- Each client has a 64-message buffered send channel; slow clients drop messages rather than blocking broadcasts
- Only the backend pushes events; clients cannot send game commands over WS (all game input is REST)

## Concurrency Safety

- `StateMachine`: `sync.RWMutex`
- `SessionStore`: `sync.RWMutex`
- `Hub.clients`: `sync.RWMutex`
- DB writes serialize via Postgres
