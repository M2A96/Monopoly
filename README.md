Phased Plan

- Phase 0: Stabilize the foundation
  
  - Finish module wiring in main.go for all existing repositories/services/handlers.
  - Align every DAO, schema, BO, and repository contract.
  - Remove CRUD assumptions that conflict with authoritative game-state processing.
  - Add health/readiness endpoints and basic integration tests.
- Phase 1: Design the domain model
  
  - Define explicit game concepts: Game , Player , Turn , Board , Space , CardDeck , Ownership , Bank , Trade , JailState , DiceRoll , ActionRequest .
  - Separate static game data from mutable runtime state .
  - Introduce a canonical GameState aggregate instead of spreading logic across unrelated tables.
  - Decide what must be persisted vs derived.
- Phase 2: Build the deterministic game engine
  
  - Create a pure reducer-style engine: nextState = Reduce(state, command) .
  - Define commands like CreateGame , JoinGame , StartGame , RollDice , EndTurn , BuyProperty , PayRent , DrawCard , BuildHouse , TradePropose , TradeAccept , Mortgage , DeclareBankruptcy .
  - Emit typed domain events for every accepted command.
  - Keep the reducer side-effect free so it can be replayed and tested.
- Phase 3: Implement turn system and rule enforcement
  
  - Add turn order, action windows, dice rules, doubles, jail handling, passing Go, rent calculation, auctions, house/hotel constraints, bankruptcy, and win conditions.
  - Enforce server-authoritative sequencing so clients cannot mutate state directly.
  - Add validation guards so illegal actions fail deterministically.
- Phase 4: In-memory runtime engine
  
  - Introduce a runtime GameRoom or Session manager that owns active game state in memory for low-latency play.
  - Load game snapshot on activation, process commands sequentially, publish state/event updates, and checkpoint periodically.
  - Use one goroutine or serialized command queue per game to avoid race conditions.
- Phase 5: WebSocket real-time layer
  
  - Keep REST for admin/lobby/bootstrap flows.
  - Add WebSocket endpoints for live gameplay, presence, reconnect, and push updates.
  - Define transport messages for command , event , state_patch , presence , and error .
  - Add player authentication/session binding and reconnect tokens.
- Phase 6: Persistence and recovery
  
  - Persist commands/events and periodic snapshots.
  - On restart, rebuild a game from latest snapshot + trailing events.
  - Store append-only game history for auditing/debugging.
  - Decide whether current GAME_LOG becomes a real event store or remains a human-readable audit table.
- Phase 7: Redis integration
  
  - Use Redis for ephemeral coordination: active room registry, connection presence, pub/sub fanout, idempotency keys, locks, and short-lived session data.
  - Do not use Redis as the source of truth for game correctness; keep authoritative persistence in DB/event storage.
  - Use Redis streams or pub/sub only for distribution, not for rule evaluation.
- Phase 8: Multi-node scalability
  
  - Add room ownership so exactly one node is authoritative for a live game at a time.
  - Route commands to the owning node through Redis/pubsub or service RPC.
  - Support reconnect to any node with proxying or ownership lookup.
  - Add shard/rebalance strategy for active games and background snapshotting.
- Phase 9: Production hardening
  
  - Add anti-cheat validation, rate limits, reconnect flows, command idempotency, dead-letter handling, and metrics.
  - Add observability around per-game latency, reducer failures, reconnect success rate, and snapshot restore time.
  - Add soak tests and deterministic replay tests.