# Frontend Reference

## Stack

- Next.js 16 (App Router, `output: standalone`)
- React 19
- Tailwind CSS 4
- TypeScript 5
- `recharts` installed (available for future chart enhancements)

## Principle

The frontend is a **pure view layer**. It:
- Renders whatever the backend sends over WebSocket or REST
- Fires REST calls for user actions (login, bet, debt confirmation)
- Computes nothing — no payout math, no state derivation

## Routes

| Route | Component | Who |
|---|---|---|
| `/` | `app/page.tsx` | Public — login gate |
| `/player` | `app/player/page.tsx` | Players |
| `/host` | `app/host/page.tsx` | Host |

Role-based redirect: after login, players go to `/player`, host goes to `/host`. If the wrong role accesses a page, they are redirected to `/`.

## WebSocket Hook

`src/hooks/useWebSocket.ts`

Connects to `ws(s)://{host}/ws` on mount. Automatically reconnects after 2 seconds if disconnected. Calls `onEvent(WSEvent)` for every inbound message. `onEvent` is stored in a ref to avoid re-registering the socket when the callback changes.

Usage:
```tsx
useWebSocket(useCallback((event) => {
  switch (event.type) {
    case 'round_opened': ...
    case 'bet_ticker': ...
  }
}, []));
```

## Components

### `BetTicker`
Props: `bets: {username, guess, wager}[]`, `unit: string`

Renders the live stream of bets in reverse chronological order. Pure display — no interaction.

### `BubbleChart`
Props: `bets`, `targetValue?`, `unit`

SVG number line with sized bubbles (radius ∝ wager). Target shown as red vertical line when `targetValue` is provided. All coordinates derived from the `bets` prop data — no backend math.

### `GaussianCurve`
Props: `curvePoints: {x,y}[]`, `betsOnNumberline`, `targetValue?`, `unit`, `showTarget: boolean`

Pure SVG renderer. Plots the 100-point array from `distribution_update` / `round_called` payloads. Target shown as dashed red line when `showTarget=true` (host view only during betting).

### `RankingsTable`
Props: `rankings: {rank, username, guess, wager, payout, net}[]`, `unit`, `targetValue`

Sortable by rank (already sorted descending by payout from backend). Net shown in green/red.

### `DebtChecklist`
Props: `debts`, `myUsername`, `token`, `onUpdate`

Shows only `PENDING` debts. Calls `POST /api/debts/:id/confirm-paid` or `confirm-received` directly. Passes updated debt object to `onUpdate` callback for local state reconciliation. WS `debt_update` events also update state independently.

## Auth Flow

1. POST `/api/auth/login` → store `{token, username, role, credits}` in `localStorage`
2. All subsequent API calls include `Authorization: Bearer <token>`
3. Logout: POST `/api/auth/logout` + clear localStorage + redirect to `/`

## Local State vs WebSocket State

| Source | Data |
|---|---|
| REST (initial fetch) | Current round state, my debts |
| WebSocket | All live updates: bets, distribution, rankings, debts, session list |
| localStorage | Auth token, username, role, initial credits |
| React state | Everything else — derived from WS events |
