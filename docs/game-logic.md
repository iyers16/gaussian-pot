# Game Logic Reference

## Gaussian Payout Model

Source: `backend/internal/game/payout.go`

### Step 1 — Score each bet

For each player bet `(guess_i, wager_i)` and a known `target`:

```
distance_i = |guess_i - target|
score_i    = gaussianKernel(distance_i, sigma, mode)
```

The kernel function per mode:

| Mode | Score formula |
|---|---|
| **Sniper** | `wager_i × exp(−d²/ 2σ²)` where σ shrinks with total wager mass |
| **Social** | `wager_i × exp(−d²/ 2σ²)` where σ grows with total wager mass |
| **Volatile** | `exp(−d²/ 2σ²)` where σ = std-dev of all guesses (wager ignored) |

### Step 2 — Compute sigma per mode

```
Sniper:   σ = 10 / (1 + ln(1 + totalWager/100))    (min 0.5)
Social:   σ = 5 × (1 + ln(1 + totalWager/100))      (max 50)
Volatile: σ = std_dev(all guesses)                   (min 0.5)
```

High σ → wide curve → generous payouts for distant guesses.
Low σ → narrow spike → only the closest guess wins big.

### Step 3 — Normalize into shares

```
totalPot    = Σ wager_i
share_i     = score_i / Σ score_j
payout_i    = share_i × totalPot
net_i       = payout_i − wager_i
```

Conservation law: `Σ payout_i = Σ wager_i`. No house cut.

---

## P2P Debt Resolution (Greedy Algorithm)

Source: `backend/internal/game/debt.go`

After computing `net_i` for each player:
- **Creditors**: `net > 0` (won more than they wagered)
- **Debtors**: `net < 0` (wagered more than they won)

Algorithm:
1. Sort creditors by net descending, debtors by |net| descending.
2. While both lists non-empty:
   - `settle = min(creditor.net, |debtor.net|)` — rounded to cents
   - Emit edge `(debtor → creditor, settle)`
   - Subtract settle from both; advance the exhausted pointer
3. Result: minimal set of payment edges.

This produces at most `max(|creditors|, |debtors|)` edges, versus the naive `|creditors| × |debtors|`.

---

## Distribution Curve Pre-computation

Source: `backend/internal/game/distribution.go`

The backend sends 100 `(x, y)` coordinate pairs for rendering. The frontend only plots them; it computes nothing.

The curve is a unit Gaussian centred on:
- The mean of current guesses during `ROUND_OPEN`
- The `target_value` once the round is called

X-range: `[centre − 3σ, centre + 3σ]` padded to include all guesses.

---

## State Machine Transitions

Source: `backend/internal/game/statemachine.go`

Valid edges:
```
IDLE → ROUND_OPEN → ROUND_CALLED → SETTLING → IDLE
```

All other transitions return `ErrInvalidTransition`. State and active `roundID` are updated atomically under a write lock.

---

## Unit Tests

Run with: `cd backend && go test ./internal/game/... -v`

| File | Tests |
|---|---|
| `payout_test.go` | Conservation, closer-wins, net-zero, empty, single-player, sigma direction for all modes |
| `debt_test.go` | Net sum, payer/payee correctness, all-zero, multiple debtors, empty, edge minimisation |
| `distribution_test.go` | Point count, bets on line, Y in [0,1], empty bets, monotonic X |
| `statemachine_test.go` | Initial state, valid chain, invalid jump, assert pass/fail, roundID tracking |
