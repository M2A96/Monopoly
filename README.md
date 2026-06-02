# Monopoly — Multiplayer Backend

A server-authoritative, event-driven multiplayer Monopoly backend written in Go.

---

## Architecture

The system follows a **hexagonal (ports-and-adapters)** design:

```
cmd/server/main.go                   → dependency wiring & boot
internal/api/                        → HTTP handlers (transport layer)
internal/core/ports/input/           → input port interfaces (service contracts)
internal/core/ports/output/          → output port interfaces (repository contracts)
internal/core/domain/services/       → domain service implementations
internal/core/domain/bo/             → business objects (domain entities)
internal/infrastructure/repository/  → repository implementations (GORM)
internal/game/                       → game engine (commands, reducers, runtime, rules...)
object/dao/                          → data access objects (DB-level structs)
config/                              → centralized configuration (Viper)
log/                                 → structured logging (Zap)
util/                                → shared utilities
pkg/board/, pkg/dice/, pkg/monopoly/ → static game definitions
```

### Core Design Rules

- **Server authoritative** — clients send commands, receive events; they never calculate outcomes.
- **Deterministic reducers** — same state + same commands always produce the same result (enables replay, debugging, anti-cheat).
- **Single writer per game** — one goroutine owns each active game room; no concurrent state mutation.
- **Commands vs Events** — commands are intentions (can fail), events are facts (immutable after persistence).
- **Immutable event history** — events are never modified after being written.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.23 |
| HTTP framework | Echo v4 |
| Database | CockroachDB (Postgres-compatible) |
| ORM | GORM |
| Configuration | Viper |
| Logging | Zap |
| Tracing | OpenTelemetry (OTLP/gRPC export) |
| Profiling | Grafana Pyroscope |
| Validation | go-playground/validator |
| Auth tokens | lestrrat-go/jwx v2 |
| Mocking | uber/mock |
| Migrations | golang-migrate |
| Container | Docker Compose |

---

## REST API

### Health

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` `/healthz` | Liveness — process is alive |
| `GET` | `/ready` `/readyz` | Readiness — database reachable |

### Games

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/games` | Create a game |
| `GET` | `/api/v1/games/:id` | Get a game (UUID or base62 ID) |
| `POST` | `/api/v1/games/:id/start` | Start a game |

### Players

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/players` | Create a player |
| `GET` | `/api/v1/players/:id` | Get a player |

### Properties

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/properties/:id` | Get a property |

### Trade Requests

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/trade-requests` | Create a trade request |
| `GET` | `/api/v1/trade-requests/:id` | Get a trade request |

### Game State

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/games/:id/state/current` | Get the current game state |

### Game Logs

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/games/:id/logs` | List game logs |

---

## Infrastructure

Services are orchestrated with Docker Compose (`compose.yaml`):

| Service | Port | Purpose |
|---|---|---|
| CockroachDB | `26257` (SQL), `8080` (admin) | Primary database |
| pgAdmin4 | `5050` | Database UI |

### Start infrastructure

```bash
docker compose up -d
```

### Run migrations

```bash
make migrate-up
```

### Drop migrations

```bash
make migrate-down
```

---

## Development

### Requirements

- Go 1.23+
- Docker & Docker Compose
- `make`

### Run

```bash
go run ./cmd/server
```

### Test

```bash
make test
```

### Coverage

```bash
make coverage
```

### Format

```bash
make fmt
```

### Lint

```bash
make golangci-lint
```

### Build

```bash
make build
```

---

## Middleware

Every request is processed by:

- **RequestIDMiddleware** — generates/propagates `X-Request-ID`
- **CorrelationIDMiddleware** — generates/propagates `X-Correlation-ID`
- **StructuredLogMiddleware** — logs method, path, status, latency, request_id, correlation_id

---

## Roadmap

| Phase | Description | Status |
|---|---|---|
| 0 | Foundation stabilization (config, DI, health, logging, tests) | Done |
| 1 | Domain modeling (GameState, Player, Turn, Board, Events, Commands) | In progress |
| 2 | Deterministic game engine (reducers, replay, pure dice provider) | Planned |
| 3 | Rule enforcement (turn order, property, jail, auction, bankruptcy) | Planned |
| 4 | In-memory runtime engine (GameRoom, goroutine per room, snapshots) | Planned |
| 5 | WebSocket realtime layer (transport protocol, reconnect, presence) | Planned |
| 6 | Persistence & recovery (event store, snapshot store, replay) | Planned |
| 7 | Redis integration (room registry, presence cache, pub/sub, locks) | Planned |
| 8 | Multi-node scalability (room ownership, command routing, failover) | Planned |
| 9 | Production hardening (metrics, idempotency, anti-cheat, soak tests) | Planned |

See [Phases.md](Phases.md) for the full engineering roadmap and [AUDIT.md](AUDIT.md) for the Phase 0 architecture audit.

---

## Final System Goals

- Deterministic and replayable game engine
- Realtime multiplayer over WebSocket
- Server authoritative with no client-side state mutation
- Horizontally scalable across multiple nodes
- Race-free via single-writer-per-room model
- Crash-recoverable via event sourcing and snapshots
