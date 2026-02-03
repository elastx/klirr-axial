# Axial BBS — AI Coding Agent Instructions

These instructions help AI agents work productively in this repo. Keep changes minimal and consistent with existing patterns.

## Big Picture
- Decentralized node: discovery via multicast, sync via HTTP API, eventual consistency across nodes.
- Backend (Go): API handlers, sync engine, models (hashing, DB), discovery multicast, remote client wrappers.
- Frontend (React/Vite/TS): SPA served by backend; all API routes under `/v1/*`.
- Data flow: multicast announces `nodeID|fullHash|:APIPort|localIP`, API compares hashes, drills down by time/user ranges, exchanges messages/users/files.

## Key Files & Boundaries
- API routes and SPA serving: [src/api/router.go](src/api/router.go)
- Sync API contracts and engine: [src/api/sync.go](src/api/sync.go), [src/api/sync_engine.go](src/api/sync_engine.go)
- Sync client/process utilities and tests: [src/synchronization](src/synchronization), e.g. [src/synchronization/sync_test.go](src/synchronization/sync_test.go)
- Models + hashing: [src/models](src/models), esp. [src/models/sync.go](src/models/sync.go), [src/models/hashing_database.go](src/models/hashing_database.go)
- Discovery multicast/broadcast: [src/discovery/multicast.go](src/discovery/multicast.go)
- Remote API wrappers: [src/remote/api.go](src/remote/api.go), [src/remote/endpoints.go](src/remote/endpoints.go)
- Frontend API usage: [web/src/services/api.ts](web/src/services/api.ts)
- Project overview and goals: [README.md](README.md)

## Architectural Highlights
- SPA serving: `spaFileSystem` serves `frontend/dist` and falls back to `index.html` for non-API paths; API paths start with `/v1/`.
- Sync algorithm: server returns either `MessagesPeriod` for small mismatches or hashed subranges to drill down; client iterates via `NextRequestFromHashes()`.
- Hashing conventions:
  - Messages hash: ordered by `created_at`, hashing IDs (prefers `id`, legacy fallback `message_id`). See [src/models/hashing_database.go](src/models/hashing_database.go).
  - Time ranges: `nil` start → 2025-01-01, `nil` end → now. See `RealizeStart/End` in [src/models/sync.go](src/models/sync.go).
  - Users hash: by `fingerprint` (alphabetical) and by fingerprint ranges.
- Users in sync: `SyncResponse` reports `user_range_hashes`; actual users transfer via `/v1/sync/users` (separate call), not in `SyncResponse`.
- Remote client: typed `Endpoint[T,R,G]` encapsulates `POST`/`GET` with validation and JSON encoding.

## Developer Workflows
- Build backend + bundle frontend:
  - `make src/axial` builds `src/axial` (depends on `src/frontend/dist`).
  - `cd web && npm install && npm run build` writes to `src/frontend/dist` (also via Make).
- Run locally (macOS needs sudo for multicast):
  - `make run` or `cd src && sudo ./axial`
- Dev mode and deps:
  - `make dev` starts services and app; `make deps` installs `air` (Go live-reload).
- Clean:
  - `make clean` removes build artifacts and Docker resources.
- Tests (sync engine focus):
  - From `src/`: `go test ./synchronization -run 'Test_BulletinConversationMerge|Test_SyncFromEmptyNode'`
  - VS Code task available: “Run synchronization tests”.

## Patterns & Conventions
- API route shape: use `http.HandleFunc` with `corsMiddleware`, method switching, status correctness (e.g., `201 Created` on POST). See [src/api/router.go](src/api/router.go).
- Upserts: `ModelSyncStore.UpsertMessage()` ignores duplicate key errors by substring match; favor idempotency in sync flows.
- Hash drill-down: split oversized mismatches via `SplitTimeRange()`; rank ranges by `CountMessagesByPeriod()` to fit `maxBatchSize`.
- Frontend API base: axios defaults to `http://localhost:8080/v1`; keep `/v1` prefix stable.

## Integration & Config
- Multicast/broadcast: requires appropriate socket options; macOS typically needs `sudo`. See [src/discovery/multicast.go](src/discovery/multicast.go) and [README.md](README.md).
- Docker Compose (Postgres + pgAdmin): started by `make run/dev`. Frontend assets served statically from `src/frontend/dist`.

## Examples to Follow
- Add a new route: mirror existing patterns in [src/api/router.go](src/api/router.go) with method gating and CORS.
- Extend sync: update `SyncRequest/SyncResponse` in [src/api/sync.go](src/api/sync.go) and pure logic in [src/api/sync_engine.go](src/api/sync_engine.go); add focused tests in [src/synchronization](src/synchronization).
- Use remote client: define an `Endpoint` with path under `v1/*` and response validators in [src/remote/endpoints.go](src/remote/endpoints.go).

Notes:
- Some tests document expected behavior that may currently fail (e.g., full convergence loop transferring users/messages). Treat them as guidance when evolving the sync.
