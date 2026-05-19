# phases.md

# Monopoly Multiplayer Backend — Full Engineering Roadmap

Version: 1.0  
Architecture Style:
- Event-driven
- Deterministic reducer engine
- Server authoritative
- Realtime multiplayer
- Horizontally scalable
- Replayable game state

---

# TABLE OF CONTENTS

1. Global Architecture Rules
2. Global Coding Standards
3. Project Structure
4. Phase 0 — Stabilize Foundation
5. Phase 1 — Domain Modeling
6. Phase 2 — Deterministic Game Engine
7. Phase 3 — Rule Enforcement
8. Phase 4 — In-Memory Runtime Engine
9. Phase 5 — WebSocket Realtime Layer
10. Phase 6 — Persistence & Recovery
11. Phase 7 — Redis Integration
12. Phase 8 — Multi-Node Scalability
13. Phase 9 — Production Hardening
14. Final Expected System Behavior

---

# 1. GLOBAL ARCHITECTURE RULES

These rules are mandatory for ALL phases.

---

## 1.1 Server Authoritative Model

The server is the ONLY authority.

Clients:
- send commands
- receive updates

Clients NEVER:
- mutate state
- calculate final outcomes
- decide rule validity

GOOD:

Client:
```json
{
  "type": "ROLL_DICE"
}
```

BAD:

```json
{
  "player_position": 17
}
```

---

## 1.2 Deterministic State Transitions

The engine MUST always produce the same result given:
- same initial state
- same command sequence

This enables:
- replay
- debugging
- anti-cheat
- recovery
- synchronization

Reducers MUST NEVER:
- use current time
- generate randomness
- access DB
- access Redis
- access websocket connections

---

## 1.3 Single Writer Principle

Each game room MUST have:
- exactly one active state owner
- exactly one command processing loop

Commands MUST be processed sequentially.

Never allow:
- concurrent reducer mutation
- parallel state writes

Target implementation:
- one goroutine per active game

---

## 1.4 Commands vs Events

Commands are INTENTIONS.

Examples:
- RollDice
- BuyProperty
- EndTurn

Events are FACTS.

Examples:
- DiceRolled
- PropertyPurchased
- TurnEnded

Commands can fail.
Events cannot fail.

---

## 1.5 Immutable Event History

Events MUST NEVER be modified after persistence.

Event history becomes:
- replay source
- audit log
- debugging source
- recovery source

---

## 1.6 Static Data vs Runtime Data

Static Data:
- board layout
- property definitions
- card definitions
- game constants

Runtime Data:
- player money
- ownership
- positions
- active turn
- jail state

Static data MUST NOT mutate during gameplay.

---

# 2. GLOBAL CODING STANDARDS

---

## 2.1 No Business Logic In Handlers

Handlers may:
- parse requests
- authenticate
- serialize responses

Handlers MUST NOT:
- calculate rent
- mutate game state
- process turn logic

---

## 2.2 Repositories Only Persist

Repositories:
- save
- load
- query

Repositories MUST NOT:
- apply game rules
- calculate outcomes
- orchestrate workflows

---

## 2.3 Reducers Must Be Pure

Reducers:
- deterministic
- side-effect free
- replayable

Reducer signature:

```go
func Reduce(
    state GameState,
    command Command,
) (
    GameState,
    []Event,
    error,
)
```

---

## 2.4 Runtime Owns Live State

Only runtime rooms may mutate active state.

Database persistence is SECONDARY.

Database is not the active execution engine.

---

# 3. RECOMMENDED PROJECT STRUCTURE

```text
/cmd
    /server

/internal
    /game
        /commands
        /events
        /reducers
        /runtime
        /validation
        /state
        /models
        /rules
        /services
        /transport
        /websocket
        /persistence
        /snapshots
        /replay
        /presence

/pkg
    /board
    /dice
    /monopoly

/config

/tests
```

---

# PHASE 0 — STABILIZE FOUNDATION

# OBJECTIVE

Stabilize the existing backend before introducing runtime architecture.

At the end of this phase:
- application boots reliably
- dependency graph is clean
- services/repositories aligned
- testing foundation exists
- no CRUD-style game mutations remain

---

# PHASE 0 — IMPLEMENTATION STEPS

---

## STEP 0.1 — Audit Existing Codebase

Review:
- handlers
- services
- repositories
- DAOs
- DTOs
- schemas
- BOs

Identify:
- duplicated responsibilities
- inconsistent naming
- circular dependencies
- business logic inside handlers
- direct DB access from handlers

Deliverable:
- architecture audit document

---

## STEP 0.2 — Normalize Naming

Standardize:
- ID types
- enum values
- timestamps
- entity naming
- package naming

BAD:
```go
playerId
PlayerID
player_id
```

GOOD:
```go
PlayerID
```

---

## STEP 0.3 — Remove CRUD Mutation APIs

DELETE patterns like:

```go
UpdatePlayerMoney()
UpdatePlayerPosition()
UpdateTurn()
```

These violate authoritative runtime ownership.

Future mutations MUST happen ONLY through:
- commands
- reducers
- runtime processing

---

## STEP 0.4 — Centralize Config

Introduce:
```text
/config
```

Support:
- environment variables
- local development
- docker deployment
- production overrides

---

## STEP 0.5 — Wire Dependency Injection

main.go MUST:
- initialize repositories
- initialize services
- initialize transport
- initialize runtime manager

Avoid:
- global state
- package singletons

---

## STEP 0.6 — Add Health Endpoints

Required endpoints:

```text
GET /health
GET /ready
```

Health:
- process alive

Ready:
- DB reachable
- migrations complete
- Redis reachable (future)

---

## STEP 0.7 — Add Structured Logging

Introduce logger abstraction.

Every log MUST contain:
- timestamp
- request ID
- correlation ID
- service name

---

## STEP 0.8 — Add Integration Tests

Required:
- boot test
- DB connectivity test
- migration test
- handler test

CI MUST fail if application cannot boot.

---

# PHASE 0 — EXPECTED RESULT

After Phase 0:
- application architecture is stable
- dependencies are clean
- runtime migration is possible
- codebase is safe to extend

---

# PHASE 1 — DOMAIN MODELING

# OBJECTIVE

Create formal Monopoly domain definitions.

At end:
- all concepts modeled explicitly
- runtime state centralized
- static data separated

---

# PHASE 1 — REQUIRED DOMAIN OBJECTS

Implement:

```text
Game
GameState
Player
PlayerState
Board
BoardSpace
Property
Railroad
Utility
CardDeck
Card
Turn
DiceRoll
JailState
Auction
Trade
Mortgage
Ownership
Bank
ActionRequest
```

---

# PHASE 1 — IMPLEMENTATION STEPS

---

## STEP 1.1 — Create Static Board Definitions

Static board data should exist in:
- JSON
- YAML
- embedded Go structs

NOT runtime DB tables.

Example:

```json
{
  "index": 1,
  "type": "PROPERTY",
  "name": "Baltic Avenue",
  "price": 60,
  "rent": [2,10,30,90,160,250]
}
```

---

## STEP 1.2 — Create GameState Aggregate

GameState becomes the authoritative runtime model.

Example:

```go
type GameState struct {
    GameID string
    Players []PlayerState
    CurrentTurn int
    Phase TurnPhase
}
```

ALL gameplay mutations happen through GameState.

---

## STEP 1.3 — Model Runtime Player State

Must include:
- money
- position
- owned properties
- jail status
- bankruptcy status
- cards
- pending actions
- mortgage state

---

## STEP 1.4 — Model Turn State

Must track:
- active player
- turn phase
- dice state
- doubles count
- action availability

---

## STEP 1.5 — Define Action Windows

Examples:
- waiting for roll
- waiting for purchase
- waiting for trade response
- waiting for auction bid

This is critical for deterministic validation.

---

## STEP 1.6 — Define Typed Events

Every meaningful action MUST emit events.

Example:

```go
type DiceRolled struct {
    PlayerID string
    Die1 int
    Die2 int
}
```

---

## STEP 1.7 — Define Command Contracts

Commands are player intentions.

Example:

```go
type RollDiceCommand struct {
    PlayerID string
}
```

---

# PHASE 1 — EXPECTED RESULT

After Phase 1:
- game domain is explicit
- GameState is canonical
- reducers can now be implemented

---

# PHASE 2 — DETERMINISTIC GAME ENGINE

# OBJECTIVE

Implement pure reducer-based gameplay engine.

At end:
- commands become deterministic state transitions
- reducers emit events
- replay becomes possible

---

# PHASE 2 — IMPLEMENTATION STEPS

---

## STEP 2.1 — Create Command Package

Commands MUST be explicit structs.

Examples:

```text
CreateGame
JoinGame
StartGame
RollDice
EndTurn
BuyProperty
PayRent
BuildHouse
MortgageProperty
TradePropose
TradeAccept
DeclareBankruptcy
```

---

## STEP 2.2 — Create Reducer Interface

```go
type Reducer interface {
    Reduce(
        state GameState,
        command Command,
    ) (
        GameState,
        []Event,
        error,
    )
}
```

---

## STEP 2.3 — Implement Validation Layer

Validation happens BEFORE mutation.

Example:

```go
ValidateRollDice()
ValidateBuyProperty()
ValidateEndTurn()
```

Validation MUST reject:
- invalid turn
- insufficient money
- illegal actions
- invalid phase

---

## STEP 2.4 — Emit Domain Events

Reducers MUST emit events.

Example flow:

Command:
```text
RollDice
```

Events:
```text
DiceRolled
PlayerMoved
PassedGo
RentCharged
```

---

## STEP 2.5 — Implement Replay System

Support:

```go
Replay(events []Event) GameState
```

Replay MUST reconstruct identical state.

---

## STEP 2.6 — Introduce Deterministic Dice Provider

Reducers MUST NOT generate randomness.

Inject randomness externally.

Example:

```go
type DiceProvider interface {
    Roll() (int, int)
}
```

---

## STEP 2.7 — Add Reducer Tests

Every reducer MUST have:
- happy path test
- invalid command test
- replay test
- deterministic consistency test

---

# PHASE 2 — EXPECTED RESULT

After Phase 2:
- deterministic engine operational
- reducers replayable
- gameplay logic isolated

---

# PHASE 3 — RULE ENFORCEMENT

# OBJECTIVE

Implement Monopoly gameplay rules.

At end:
- illegal gameplay impossible
- turn sequencing enforced
- server authoritative gameplay operational

---

# PHASE 3 — IMPLEMENTATION STEPS

---

## STEP 3.1 — Turn Enforcement

Validate:
- active player only
- valid turn phase
- legal action windows

---

## STEP 3.2 — Dice Rules

Implement:
- doubles
- triple doubles -> jail
- movement
- passing GO

---

## STEP 3.3 — Property Rules

Implement:
- purchasing
- ownership
- rent
- monopoly bonuses
- utility logic
- railroad logic

---

## STEP 3.4 — Building Rules

Implement:
- even building
- hotel upgrades
- house limits
- monopoly ownership checks

---

## STEP 3.5 — Jail Rules

Implement:
- entering jail
- leaving by doubles
- leaving by payment
- leaving by card

---

## STEP 3.6 — Auction Rules

If purchase declined:
- auction starts
- bids accepted
- highest bidder wins

---

## STEP 3.7 — Bankruptcy Rules

Implement:
- debt handling
- asset transfer
- elimination

---

## STEP 3.8 — Win Conditions

Game ends when:
- one player remains
- resignation timeout occurs

---

# PHASE 3 — EXPECTED RESULT

After Phase 3:
- Monopoly fully playable
- deterministic rule enforcement operational

---

# PHASE 4 — IN-MEMORY RUNTIME ENGINE

# OBJECTIVE

Create low-latency authoritative runtime execution.

At end:
- active games run entirely in memory
- commands processed sequentially
- runtime owns live state

---

# PHASE 4 — RUNTIME ARCHITECTURE

Each active game owns:
- GameState
- command queue
- event stream
- subscriber list
- snapshot timer

---

# PHASE 4 — IMPLEMENTATION STEPS

---

## STEP 4.1 — Create GameRoom

Example:

```go
type GameRoom struct {
    GameID string

    State GameState

    CommandQueue chan Command

    Subscribers map[string]Subscriber

    Reducer Reducer
}
```

---

## STEP 4.2 — Add Processing Loop

Each room MUST own one goroutine.

Flow:

1. receive command
2. validate
3. reduce
4. update state
5. emit events
6. notify subscribers
7. checkpoint snapshot

---

## STEP 4.3 — Create Runtime Manager

Responsibilities:
- activate rooms
- lookup rooms
- shutdown rooms
- manage lifecycle

Example:

```go
type RuntimeManager struct {
    Rooms map[string]*GameRoom
}
```

---

## STEP 4.4 — Add Sequential Command Processing

Commands MUST execute in strict order.

Never:
- parallel reducer execution
- concurrent state mutation

---

## STEP 4.5 — Add Subscriber System

Subscribers:
- websocket clients
- spectators
- bots

Receive:
- events
- state updates
- snapshots

---

## STEP 4.6 — Add Runtime Snapshots

Checkpoint:
- every N events
- every N seconds

Persist:
- GameState
- event offset

---

## STEP 4.7 — Add Runtime Recovery

On room activation:
1. load snapshot
2. replay trailing events
3. rebuild GameState

---

## STEP 4.8 — Add Idle Room Cleanup

If room inactive:
- persist final snapshot
- terminate goroutine
- release memory

---

# PHASE 4 — EXPECTED APPLICATION BEHAVIOR

After Phase 4:
- games execute in memory
- latency is low
- race conditions eliminated
- runtime authoritative
- recovery possible

---

# PHASE 5 — WEBSOCKET REALTIME LAYER

# OBJECTIVE

Implement realtime multiplayer transport.

---

# PHASE 5 — IMPLEMENTATION STEPS

---

## STEP 5.1 — Add WebSocket Endpoint

Example:

```text
/ws/game/{gameId}
```

---

## STEP 5.2 — Define Transport Protocol

Client -> server:

```json
{
  "type": "command",
  "payload": {}
}
```

Server -> client:

```json
{
  "type": "event",
  "payload": {}
}
```

---

## STEP 5.3 — Define Message Types

Required:
- command
- event
- state_patch
- presence
- reconnect
- heartbeat
- error

---

## STEP 5.4 — Bind Player Sessions

Associate:
- player
- websocket
- runtime subscriber

---

## STEP 5.5 — Add Reconnect Support

Support:
- reconnect tokens
- session restoration
- missed event replay

---

## STEP 5.6 — Add Presence Tracking

Track:
- online
- offline
- reconnecting

---

## STEP 5.7 — Add Backpressure Handling

Protect runtime from:
- slow clients
- stuck websocket writes

Disconnect lagging clients safely.

---

# PHASE 5 — EXPECTED RESULT

After Phase 5:
- realtime multiplayer operational
- reconnect supported
- live synchronization stable

---

# PHASE 6 — PERSISTENCE & RECOVERY

# OBJECTIVE

Make gameplay durable.

---

# PHASE 6 — IMPLEMENTATION STEPS

---

## STEP 6.1 — Implement Event Store

Persist:
- sequence number
- event type
- payload
- timestamp

---

## STEP 6.2 — Implement Snapshot Store

Persist:
- GameState snapshot
- event offset

---

## STEP 6.3 — Implement Recovery Pipeline

Recovery:
1. load snapshot
2. load trailing events
3. replay events

---

## STEP 6.4 — Add Replay Tooling

Support:
- full game replay
- debugging
- corruption inspection

---

## STEP 6.5 — Add Event Versioning

Support future schema evolution.

---

# PHASE 6 — EXPECTED RESULT

After Phase 6:
- crash recovery operational
- replay available
- durable game persistence exists

---

# PHASE 7 — REDIS INTEGRATION

# OBJECTIVE

Add distributed coordination infrastructure.

---

# PHASE 7 — IMPLEMENTATION STEPS

---

## STEP 7.1 — Active Room Registry

Track:
- room owner node
- heartbeat
- runtime status

---

## STEP 7.2 — Presence Cache

Track:
- online players
- websocket ownership

---

## STEP 7.3 — Add Pub/Sub

Use for:
- node communication
- notifications
- room routing

NOT for:
- authoritative state evaluation

---

## STEP 7.4 — Add Distributed Locks

Use for:
- ownership election
- recovery coordination

---

# PHASE 7 — EXPECTED RESULT

After Phase 7:
- distributed coordination operational
- multi-node preparation complete

---

# PHASE 8 — MULTI-NODE SCALABILITY

# OBJECTIVE

Scale horizontally across multiple nodes.

---

# PHASE 8 — IMPLEMENTATION STEPS

---

## STEP 8.1 — Add Room Ownership

Exactly one node owns each room.

Only owner:
- mutates state
- processes commands

---

## STEP 8.2 — Add Command Routing

If incorrect node receives command:
- proxy to owner node

---

## STEP 8.3 — Add Ownership Rebalancing

Support:
- failover
- shutdown migration
- load balancing

---

## STEP 8.4 — Add Distributed Recovery

If node crashes:
- another node restores runtime

---

# PHASE 8 — EXPECTED RESULT

After Phase 8:
- horizontally scalable runtime operational

---

# PHASE 9 — PRODUCTION HARDENING

# OBJECTIVE

Production-grade reliability.

---

# PHASE 9 — IMPLEMENTATION STEPS

---

## STEP 9.1 — Add Metrics

Track:
- reducer latency
- active games
- reconnect success
- command throughput

---

## STEP 9.2 — Add Structured Logging

Every command log:
- player ID
- game ID
- command ID
- trace ID

---

## STEP 9.3 — Add Idempotency

Prevent duplicate command execution.

---

## STEP 9.4 — Add Anti-Cheat Validation

Validate:
- impossible actions
- replay consistency
- illegal state transitions

---

## STEP 9.5 — Add Soak Testing

Test:
- thousands of rooms
- reconnect storms
- recovery scenarios

---

## STEP 9.6 — Add Replay Consistency Tests

Replay historical events and verify:
- deterministic state hashes
- consistency across deployments

---

# PHASE 9 — EXPECTED RESULT

After Phase 9:
- production-grade backend complete

---

# FINAL EXPECTED SYSTEM BEHAVIOR

Final system characteristics:

- deterministic
- replayable
- realtime
- horizontally scalable
- websocket-driven
- event-driven
- server authoritative
- race-free
- recoverable
- low latency
- production resilient

Final runtime flow:

Client
-> Command
-> Runtime Queue
-> Reducer
-> Events
-> State Update
-> Snapshot
-> WebSocket Broadcast
-> Persistence