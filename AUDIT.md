# Phase 0 Architecture Audit

## Architecture Overview

The project follows a hexagonal (ports-and-adapters) architecture:

```
cmd/server/main.go              → wires all dependencies
internal/api/                   → HTTP handlers (transport layer)
internal/core/ports/input/      → input port interfaces (service contracts)
internal/core/ports/output/     → output port interfaces (repository contracts)
internal/core/domain/services/  → domain service implementations
internal/core/domain/bo/        → business objects (domain entities)
internal/infrastructure/repository/ → repository implementations (GORM/Postgres)
object/dao/                     → data access objects (DB-level structs)
config/                         → centralized configuration
log/                            → structured logging (zap)
util/                           → shared utilities
```

---

## Findings

### Critical Bugs Fixed

**`bo.Player` did not embed `bo` struct (Step 0.2)**
- `NewPlayer(id, ...)` accepted an `id` parameter but the struct had no `id` field — the ID was silently discarded on every player creation.
- Fix: embedded `bo` struct into `player`; extended `Player` interface with `BOer`.

**`playerService.List()` pre-allocation bug (Step 0.2)**
- `make([]bo.Player, len(daoPlayers))` created `n` zero-value slots, then `append` added `n` more real entries, returning `2n` items (first half nil/zero).
- Fix: changed to `make([]bo.Player, 0, len(daoPlayers))`.

---

### Duplicated Responsibilities Removed

**`GameServicer.GetGameState` removed (Step 0.3)**
- `GameServicer.GetGameState` was a direct alias for `GameServicer.Get` — it returned a `Gamer` (not a `GameStater`) and just delegated to `Get`.
- The route `GET /api/v1/games/:id/state` duplicated functionality already in `GameStateHandler` (`GET /api/v1/games/:id/state/current`).
- Fix: removed `GetGameState` from interface, service, handler, and route table.

---

### CRUD Mutation API Surface (Step 0.3)

The input port interfaces (`ports/input/`) do NOT expose raw `Update` or `Delete` methods — this is correct. The service layer communicates through domain-oriented actions (`CreateGame`, `StartGame`, `CreatePlayer`).

`StartGame` internally calls `repository.Update` — acceptable for Phase 0 since the full command/reducer pattern arrives in Phase 2. No direct mutation HTTP endpoints (`PUT /games/:id`, `DELETE /games/:id`) exist at the handler level.

---

### Naming Issues Fixed (Step 0.2)

- Wrong comment `GetAddressRepositorier` → `GetGameRepositorier` in `game_service.go`.
- Indentation normalized in `game_handlers.go`.
- `Player` interface now consistently extends `BOer` (matching `Gamer`).

---

### Config (Step 0.4) — Already Done ✓

- `/config` package exists with environment variable support via Viper.
- All config sub-structs (`DatabaseConfig`, `LogConfig`, `OtelConfig`, `RuntimeConfig`, `ServerConfig`) are independently optionable.
- Defaults are set in `main.go` via `viper.SetDefault(...)`.

---

### Dependency Injection (Step 0.5) — Already Done ✓

- `main.go` explicitly constructs and wires all repositories, services, and handlers.
- No global state or package-level singletons exist.
- Handler → Service → Repository dependency graph is clean.

---

### Health Endpoints (Step 0.6) — Already Done ✓

- `GET /health` → process alive check (always returns 200).
- `GET /ready` → database ping check (returns 503 if DB is unreachable).
- Aliases `/healthz` and `/readyz` also registered.

---

### Structured Logging (Step 0.7)

- Zap-based logger already existed with `log.RuntimeLogger` abstraction.
- **Added**: `service_name` field to the base logger in `main.go`.
- **Added**: `RequestIDMiddleware` — generates/propagates `X-Request-ID` per request.
- **Added**: `CorrelationIDMiddleware` — generates/propagates `X-Correlation-ID` per request.
- **Added**: `StructuredLogMiddleware` — logs every request with: timestamp (zap), request_id, correlation_id, service_name, method, path, status, latency_ms.
- All three middlewares are wired in `main.go` before route registration.

---

### Integration Tests (Step 0.8)

Existing tests:
- `internal/api/boot_test.go` — route registration smoke tests.
- `internal/api/health_handlers_test.go` — health endpoint unit tests.

Added tests:
- `internal/api/game_handlers_test.go` — game handler unit tests with stub service (create, get, start, error paths).
- `internal/api/player_handlers_test.go` — player handler unit tests with stub service (get, error paths).
- `internal/api/middleware_test.go` — request ID, correlation ID, and structured log middleware tests.

---

## Remaining Risks for Phase 1

1. `bo.GameStater` carries `[]bo.Player` — once Player has `GetID()`, downstream callers building player lists should verify ID is populated correctly from DB reads.
2. `StartGame` uses `repository.Update` directly — acceptable for Phase 0, MUST be replaced by a command/reducer in Phase 2.
3. No migration runner in the boot path — migrations are applied manually via the `Makefile`. Phase 0.8 DB migration test requires a running database; current tests are pure unit tests.
4. No authentication/authorization middleware — out of scope for Phase 0 but needed before Phase 5 (WebSocket layer).

---

# Phase 1 Domain Modeling Audit

## New Package Structure

```
pkg/board/
    spaces.go           → 40-space static board (BoardSpace, SpaceByIndex map)
    data/spaces.json    → JSON board data (all 40 spaces)
    data/spaces.yaml    → YAML board data (all 40 spaces)

internal/game/
    models/
        game_state.go   → GameState aggregate (canonical runtime model)
        player_state.go → PlayerState (money, position, jail, ownership cache)
        turn.go         → Turn, TurnPhase, DiceRoll interface, NewDiceRoll()
        action_window.go → ActionWindow with 7 typed waiting states
        auction.go      → Auction (space, bids map, highest bid, winner)
        bank.go         → Bank (houses/hotels supply), NewBank()
        card.go         → Card, CardDecks, DefaultChanceDeck(), DefaultCommunityChestDeck()
        jail_state.go   → JailState (turns-in-jail counter)
        owenership.go   → Ownership (space, owner, mortgaged, buildings 0–5)
        trade.go        → Trade, TradeStatus
    events/
        events.go       → 30 typed events, Event interface, BaseEvent
    commands/
        commands.go     → 20 typed commands, Command interface, BaseCommand
```

---

## Step 1.1 — Static Board Definitions ✓

All 40 Monopoly spaces defined in `pkg/board/spaces.go` with verified rent tables:

| Type | Count | Notes |
|------|-------|-------|
| Properties | 22 | Rent[6] encodes [no house, 1H, 2H, 3H, 4H, hotel] |
| Railroads | 4 | Rent[6] encodes [1 owned, 2 owned, 3 owned, 4 owned, 0, 0] |
| Utilities | 2 | Rent all zeros — computed at runtime as 4× or 10× dice |
| Tax | 2 | TaxAmount field: Income Tax $200, Luxury Tax $100 |
| Special | 10 | GO, Jail, Free Parking, Go To Jail, 3× Chance, 3× Community Chest |

`SpaceByIndex map[int]BoardSpace` initialized via `init`-style `var` for O(1) position lookups in the reducer.

**Decision:** Static data lives in `pkg/board/` (not a DB table). The board never changes at runtime.

---

## Step 1.2 — GameState Aggregate ✓

```go
type GameState struct {
    GameID     uuid.UUID
    Status     GameStatus            // WAITING | ACTIVE | FINISHED
    Players    []PlayerState
    Turn       Turn
    Bank       Bank
    Ownerships map[int]Ownership     // spaceIndex → Ownership
    CardDecks  CardDecks
    Auction    *Auction              // nil when no active auction
    Trades     []Trade
}
```

**Decision:** `Ownerships` is the single source of truth for property ownership and development state. `PlayerState.OwnedSpaceIndexes` is a convenience cache — reducers must keep both in sync.

**Decision:** `Auction *Auction` is a pointer; `nil` means no auction is active. Only one auction can be active at a time.

---

## Step 1.3 — Runtime PlayerState ✓

```go
type PlayerState struct {
    PlayerID          uuid.UUID
    Name              string
    Money             int
    Position          int         // 0–39
    IsInJail          bool
    JailState         JailState   // TurnsInJail counter (max 3)
    IsBankrupt        bool
    OwnedSpaceIndexes []int       // cache; authoritative source: GameState.Ownerships
    GetOutOfJailCards int
    PendingActions    []ActionWindow
}
```

**Decision:** Mortgage state is NOT stored in `PlayerState` — it lives in `GameState.Ownerships[idx].Mortgaged`. This avoids duplication and the consistency bugs that come with it.

---

## Step 1.4 — Turn State ✓

```go
type Turn struct {
    Number         int
    ActivePlayerID uuid.UUID
    Phase          TurnPhase     // PRE_ROLL | LAND | ACTION | END
    DiceRoll       DiceRoll      // interface; nil until player has rolled
    DoublesCount   int           // 0–2; hitting 3 sends player to jail
    ActionWindow   *ActionWindow // nil when no action is pending
}
```

`DiceRoll` is an **interface**, not a concrete struct. The runtime constructs a `diceRoll` via `NewDiceRoll(die1, die2)` and injects it. Reducers call `DiceRoll.Total()` and `DiceRoll.IsDoubles()` — they never generate the values.

**Risk:** `Turn.DiceRoll` is `nil` in the PRE_ROLL phase. Phase 2 reducers must guard all `DiceRoll` access with a nil check.

---

## Step 1.5 — Action Windows ✓

```go
const (
    ActionWaitingForRoll              = "WAITING_FOR_ROLL"
    ActionWaitingForPurchaseDecision  = "WAITING_FOR_PURCHASE_DECISION"
    ActionWaitingForTradeResponse     = "WAITING_FOR_TRADE_RESPONSE"
    ActionWaitingForAuctionBid        = "WAITING_FOR_AUCTION_BID"
    ActionWaitingForJailDecision      = "WAITING_FOR_JAIL_DECISION"
    ActionWaitingForBankruptcyResolution = "WAITING_FOR_BANKRUPTCY_RESOLUTION"
    ActionWaitingForCardEffect        = "WAITING_FOR_CARD_EFFECT"
)
```

`Turn.ActionWindow` holds the game-level pending action. `PlayerState.PendingActions []ActionWindow` holds per-player pending responses (e.g., incoming trade proposals).

---

## Step 1.6 — Typed Events ✓

30 domain events in `internal/game/events/events.go`. All embed `BaseEvent`:

```go
type BaseEvent struct {
    Type           EventType
    GameID         uuid.UUID
    SequenceNumber int64     // strictly increasing; enables replay ordering
    OccurredAt     time.Time
}
```

Event categories: game lifecycle (4), turn flow (2), dice/movement (4), property (7), auction (3), jail (2), trade (3), bankruptcy (1), cards (2), tax (1) = 29 concrete event types + 1 base.

**Decision:** `SequenceNumber int64` is included at domain model level, not just persistence level. This ensures replay works correctly even when events come from different storage backends.

**Decision:** `BankruptcyDeclaredEvent.CreditorID` uses `uuid.Nil` to represent the bank (no real UUID). Phase 2 reducers must check for `uuid.Nil` when distributing assets.

---

## Step 1.7 — Command Contracts ✓

20 typed commands in `internal/game/commands/commands.go`. All embed `BaseCommand`:

```go
type BaseCommand struct {
    GameID   uuid.UUID
    PlayerID uuid.UUID
}
```

**Key design — RollDiceCommand:**

```go
type RollDiceCommand struct {
    BaseCommand
    Die1 int  // injected by runtime
    Die2 int  // injected by runtime
}
```

The runtime generates randomness (via `pkg/dice` in Phase 2) and injects the values into the command before dispatching. Reducers remain pure — they read `Die1`/`Die2` but never call `rand`.

---

## Phase 1 Risks for Phase 2

| Risk | Mitigation in Phase 2 |
|------|-----------------------|
| `GameState.Ownerships` is `nil` in zero value | `NewGameState()` constructor must initialize the map |
| `PlayerState.OwnedSpaceIndexes` is a cache | Reducer must update both Ownerships and OwnedSpaceIndexes atomically |
| Card deck shuffling must be deterministic | Use a seeded `rand.Source`; seed is stored in event log for replay |
| `Turn.DiceRoll` is `nil` before roll | All reducers touching DiceRoll must guard with nil check or enforce PRE_ROLL phase |
| `bo/game_state.go`, `object/dao/game_state.go` have pre-existing `go vet` warnings (unexported fields with json tags) | Pre-existing; not introduced in Phase 1. Fix separately when migrating away from the legacy layer |
| `StartGame` still uses `repository.Update` directly | MUST be replaced by `StartGameCommand` → reducer → events in Phase 2 |
