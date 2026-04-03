# API Reference

All endpoints are prefixed with `/api`. Auth-required endpoints need `Authorization: Bearer <token>` header.

## Authentication

### POST /api/auth/login
**Public.** Creates or retrieves a user and issues a session token.

Request:
```json
{ "username": "alice", "password": "" }
```
- Players: any username, password ignored.
- Host: username must be `"host"`, password must match `HOST_PASSWORD` env var (default: `"host123"`).

Response `200`:
```json
{ "token": "...", "username": "alice", "role": "player", "credits": 1000.00 }
```
Response `429`: session cap reached (`max 9`).
Response `401`: invalid host password.

---

### POST /api/auth/logout
**Auth required.** Removes the session.

---

## Player Endpoints

### GET /api/round
Returns current round state. `target_value` is omitted unless state is `ROUND_CALLED` or `SETTLING`.

Response when round is open:
```json
{ "id": 1, "question_text": "...", "unit": "goals", "mode": "sniper", "state": "ROUND_OPEN", "opened_at": "..." }
```

### POST /api/round/bet
**Auth required. State must be `ROUND_OPEN`.** Places a bet for the authenticated player.

Request:
```json
{ "guess": 3.0, "wager": 100.0 }
```
- Blocked if player has unsettled debts.
- Blocked if player already bet this round.
- Wager must not exceed current credits.

---

### GET /api/debts
Returns the authenticated player's pending debts.

### POST /api/debts/:id/confirm-paid
Payer confirms they have paid. Sets `payer_confirmed = true`. If `payee_confirmed` already true, debt becomes `SETTLED`.

### POST /api/debts/:id/confirm-received
Payee confirms they received payment. Sets `payee_confirmed = true`. If `payer_confirmed` already true, debt becomes `SETTLED`.

---

## Admin (Host-only) Endpoints

All require `role = host`.

### GET /api/admin/questions
Returns the full 50-question bank.

### POST /api/admin/question/random
Returns a randomly selected question from the bank.

### POST /api/admin/round/open
Opens a new betting round. Blocked if any debts are pending.

Request:
```json
{ "question_id": 7, "mode": "sniper" }
```
Mode values: `sniper`, `social`, `volatile`.

### POST /api/admin/round/call
Closes betting, computes payouts, resolves debts, transitions to SETTLING.

### POST /api/admin/credits/replenish
Adds credits to a player (unilateral host action).

Request:
```json
{ "username": "alice", "amount": 500.0 }
```

### GET /api/admin/users
Returns all registered users with current credit balances.

---

## WebSocket

Connect to `ws://{host}/ws`. No auth on the WS handshake (frontend connects immediately after login).

### Events pushed by backend

| Event | Payload |
|---|---|
| `session_update` | `{ active_users: [{username, role}] }` |
| `round_opened` | `{ question_text, mode, unit }` |
| `bet_ticker` | `{ bets: [{username, guess, wager}] }` |
| `distribution_update` | `{ curve_points: [{x,y}], bets_on_numberline: [{username,guess,wager}] }` |
| `round_called` | `{ target_value, unit, rankings: [...], debts: [...], distribution: {...} }` |
| `debt_update` | full Debt object |
| `round_settled` | `{}` |

All events are JSON with schema: `{ "type": "event_name", "payload": {...} }`.
