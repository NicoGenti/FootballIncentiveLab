# Architecture Research

**Domain:** Sports analytics / football contract management (Go monolith + SQLite)
**Researched:** 2026-04-20
**Confidence:** HIGH

## System Overview

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Layer (net/http)                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐   │
│  │ REST Handlers│  │ SSR Handlers │  │ Middleware Chain  │   │
│  │  (JSON API)  │  │(html/template)│  │ (auth, logging)  │   │
│  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘   │
├─────────┴────────────────┴────────────────────┴─────────────┤
│                    Service Layer                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐   │
│  │   PlayerSvc  │  │   KPISvc     │  │  CalcEngine      │   │
│  │  (CRUD ops)  │  │  (definitions)│  │ (pure Go, no DB) │   │
│  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘   │
├─────────┴────────────────┴────────────────────┴─────────────┤
│                   Repository Layer                           │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Store (interface-based)                   │   │
│  │  players │ contracts │ kpis │ performances │ results  │   │
│  └──────────────────────┬───────────────────────────────┘   │
├─────────────────────────┴───────────────────────────────────┤
│                    SQLite (WAL mode)                         │
│         database/sql + modernc.org/sqlite (pure Go)          │
└─────────────────────────────────────────────────────────────┘
```

## Component Responsibilities

| Component | Responsibility | Communicates With |
|-----------|---------------|-------------------|
| REST Handlers | HTTP JSON endpoints for API consumers | Service layer |
| SSR Handlers | Server-rendered HTML pages (dashboard) | Service layer |
| Middleware Chain | Auth, logging, request ID, error recovery | All handlers |
| PlayerSvc | Player & contract CRUD operations | PlayerRepo |
| KPISvc | KPI definition CRUD, weight/target management | KPIRepo |
| CalcEngine | Deterministic salary calculation (pure functions) | KPISvc (reads), no DB |
| CSVImporter | Parse & validate CSV match statistics | PerformanceRepo |
| Store (interface) | Data access abstraction for all entities | SQLite via database/sql |
| SQLite (WAL) | Persistent storage, concurrent reads | Store implementation |

## Recommended Project Structure

```
footballincentivelab/
├── cmd/
│   └── server/
│       └── main.go              # Entry point: config, DI, server start
├── internal/
│   ├── domain/                  # Core domain types (no imports from other internal pkgs)
│   │   ├── player.go            # Player, Contract structs
│   │   ├── kpi.go               # KPIDefinition, ScalingFormula types
│   │   ├── performance.go       # PerformanceRecord, MatchStat types
│   │   └── compensation.go      # CalculationResult, SalaryBreakdown types
│   ├── calc/                    # Pure calculation engine (no DB, no HTTP)
│   │   ├── engine.go            # Calculate() — deterministic salary computation
│   │   ├── fairness.go          # Fairness rules (team budget caps, per-player caps)
│   │   ├── scaling.go           # KPI scaling formulas (linear, threshold, etc.)
│   │   └── engine_test.go       # Table-driven tests with golden files
│   ├── store/                   # Data access layer (interface + SQLite impl)
│   │   ├── store.go             # Store interface definition
│   │   ├── sqlite/              # SQLite implementation
│   │   │   ├── db.go            # Open, migrations, connection setup
│   │   │   ├── players.go       # Player/contract queries
│   │   │   ├── kpis.go          # KPI definition queries
│   │   │   ├── performances.go  # Performance record queries
│   │   │   ├── results.go       # Calculation result storage
│   │   │   └── migrations/      # SQL migration files (embedded via go:embed)
│   │   │       ├── 001_players.sql
│   │   │       ├── 002_kpis.sql
│   │   │       └── 003_performances.sql
│   │   └── mock/                # In-memory store for testing
│   │       └── store.go
│   ├── service/                 # Business logic orchestration
│   │   ├── player.go            # PlayerSvc (CRUD + validation)
│   │   ├── kpi.go               # KPISvc (CRUD + weight validation)
│   │   ├── compensation.go      # CompensationSvc (trigger calc, retrieve results)
│   │   └── import.go            # ImportSvc (CSV parse, validate, store)
│   ├── csv/                     # CSV parsing logic
│   │   ├── parser.go            # CSV reader with column mapping
│   │   └── validator.go         # Data validation rules
│   └── http/                    # HTTP layer
│       ├── server.go            # Router setup, middleware chain
│       ├── handlers/
│       │   ├── players.go       # REST: /api/players
│       │   ├── contracts.go     # REST: /api/contracts
│       │   ├── kpis.go          # REST: /api/kpis
│       │   ├── simulate.go      # REST: /api/simulate
│       │   ├── run.go           # REST: /api/run
│       │   ├── history.go       # REST: /api/history
│       │   ├── import.go        # REST: /api/import
│       │   └── dashboard.go     # SSR: / (dashboard pages)
│       ├── middleware/
│       │   ├── auth.go          # API key / token auth
│       │   ├── logging.go       # Request logging
│       │   └── recovery.go      # Panic recovery
│       └── templates/           # html/template files
│           ├── layout.html
│           ├── dashboard.html
│           └── partials/
├── web/
│   └── static/                  # CSS, JS, images
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Structure Rationale

- **`cmd/server/`:** Thin main function — just wires dependencies and starts the HTTP server. Follows the golang-standards/project-layout convention. Makes `go install ./cmd/server` work.
- **`internal/domain/`:** Pure data types with no imports from other internal packages. This is the dependency root — everything else depends on domain, domain depends on nothing.
- **`internal/calc/`:** Isolated calculation engine. Pure Go, no database imports, no HTTP imports. This is the highest-risk component (money calculations) and must be maximally testable. Table-driven tests with golden files for regression safety.
- **`internal/store/`:** Interface-based data access. The `Store` interface is defined once; `sqlite/` implements it. This decouples the service layer from SQLite specifics and enables a `mock/` for testing.
- **`internal/service/`:** Orchestration layer. Each service holds a Store reference and business logic. Services call the CalcEngine when needed but never touch SQL directly.
- **`internal/http/`:** Handlers are thin — they parse requests, call services, format responses. No business logic here.
- **`internal/csv/`:** Isolated CSV parsing. No knowledge of HTTP or database — just "CSV → domain types + validation errors."

## Architectural Patterns

### Pattern 1: Interface-Based Store (Dependency Inversion)

**What:** Define a `Store` interface in `internal/store/` with methods like `GetPlayer(id)`, `ListContracts()`, `SaveResult()`. Implement it in `sqlite/`. Inject into services via constructor.

**When to use:** Always for this project. Enables mock testing of services, possible future database swap, and clean separation of concerns.

**Trade-offs:** Slight boilerplate for the interface definition, but the testing benefits and decoupling far outweigh this for a financial calculation system.

**Example:**
```go
// internal/store/store.go
type Store interface {
    GetPlayer(ctx context.Context, id int64) (*domain.Player, error)
    ListPlayers(ctx context.Context) ([]domain.Player, error)
    CreatePlayer(ctx context.Context, p domain.Player) (int64, error)
    // ... contract, kpi, performance, result methods
}

// internal/service/player.go
type PlayerSvc struct {
    store store.Store
}

func NewPlayerSvc(s store.Store) *PlayerSvc {
    return &PlayerSvc{store: s}
}
```

### Pattern 2: Pure Calculation Engine

**What:** The `internal/calc/` package accepts domain types as input and returns `CalculationResult` as output. It imports ONLY from `internal/domain/`. No database, no HTTP, no I/O of any kind.

**When to use:** For the core salary calculation. This is the "money code" — it must be deterministic, auditable, and testable without any infrastructure.

**Trade-offs:** Requires fetching all data before calling the engine (slight memory overhead), but enables perfect testability and auditability.

**Example:**
```go
// internal/calc/engine.go
func Calculate(
    player domain.Player,
    contract domain.Contract,
    kpis []domain.KPIDefinition,
    performances []domain.PerformanceRecord,
) domain.CompensationResult {
    // Pure computation — no side effects
}
```

### Pattern 3: Embedded Migrations

**What:** SQL migration files in `internal/store/sqlite/migrations/` are embedded at compile time via `//go:embed`. On startup, the application runs pending migrations against SQLite before accepting requests.

**When to use:** For any Go + SQLite project that needs schema evolution without external tools.

**Trade-offs:** Less flexible than dedicated migration tools (goose, atlas), but eliminates external dependencies and keeps the single-binary promise.

**Example:**
```go
// internal/store/sqlite/db.go
//go:embed migrations/*.sql
var migrationsFS embed.FS

func Open(dbPath string) (*sql.DB, error) {
    db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
    // ... run migrations from migrationsFS
}
```

### Pattern 4: StatsProvider Interface (Future-Proofing)

**What:** Define a `StatsProvider` interface for data ingestion. MVP implements it as CSV. Future implementations can be Opta, Wyscout, or any API.

**When to use:** When the data source is known to change. CSV is a placeholder — the club will eventually want automated feeds.

**Trade-offs:** Adds an abstraction layer that isn't strictly needed for MVP, but prevents ripping out CSV code later.

**Example:**
```go
// internal/domain/provider.go
type StatsProvider interface {
    Name() string
    Parse(ctx context.Context, source io.Reader) ([]domain.PerformanceRecord, error)
    Validate(records []domain.PerformanceRecord) []domain.ValidationError
}
```

## Data Flow

### Request Flow (API)

```
HTTP Request
    ↓
Middleware (auth → logging → recovery)
    ↓
Handler (parse JSON → validate)
    ↓
Service (business logic, orchestration)
    ↓
Store Interface → SQLite Implementation
    ↓
SQLite WAL (read/write)
    ↑
Service (transform domain → response)
    ↑
Handler (format JSON)
    ↑
HTTP Response
```

### Calculation Flow

```
User triggers /api/run (or /api/simulate)
    ↓
CompensationSvc
    ↓ fetches: players, contracts, kpis, performances (via Store)
    ↓ validates: all required data present
    ↓
CalcEngine.Calculate() per player
    ├── scale KPIs to [0, 1] based on targets
    ├── compute weighted bonus = Σ(weight × scaled_value) × base_salary
    ├── apply bonus cap (max_bonus_ratio × base_salary)
    ├── apply fairness rules (team budget, per-player caps)
    └── return CompensationResult
    ↓
Store.SaveResults() (persist for history/audit)
    ↓
Return results to caller
```

### CSV Import Flow

```
CSV file upload → /api/import
    ↓
CSVImporter.Parse()
    ├── read CSV rows
    ├── map columns to PerformanceRecord fields
    └── return records + parse errors
    ↓
CSVImporter.Validate()
    ├── check player references exist
    ├── check date ranges
    ├── check numeric ranges (goals ≥ 0, minutes 0-90+)
    └── return validation errors
    ↓
If errors → return 400 with error list
If valid → ImportSvc.Store(records) via Store
    ↓
Return 201 with import summary
```

### SSR Dashboard Flow

```
Browser GET /
    ↓
Dashboard Handler
    ↓ fetches: players, latest results, summary stats (via Store)
    ↓ renders: html/template with data
    ↓
HTML Response
```

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 1 club, 25-30 players | Monolith is sufficient. SQLite handles concurrent reads easily. |
| 1 club, 100+ players | Pre-calc aggregates. Cache dashboard queries. SQLite still fine. |
| 5-10 clubs | Add multi-tenant (club_id column). Consider PostgreSQL. |
| 100+ clubs | Split calculation to background workers. PostgreSQL required. |

### Scaling Priorities

1. **First bottleneck:** Calculation time for large squads (100+ players × 10+ KPIs). Fix: pre-compute in background, cache results.
2. **Second bottleneck:** SQLite write contention under heavy concurrent imports. Fix: WAL mode + single-writer pattern (queue CSV imports).
3. **Third bottleneck:** Dashboard rendering time with historical data. Fix: materialized summary tables.

## Anti-Patterns

### Anti-Pattern 1: Putting Business Logic in Handlers

**What people do:** Calculate salaries, validate contracts, apply fairness rules directly in HTTP handler functions.

**Why it's wrong:** Handlers become untestable without HTTP infrastructure. Business logic can't be reused (CLI, background jobs). Code duplication across REST and SSR handlers.

**Do this instead:** Handlers only parse requests and format responses. All logic lives in services. The CalcEngine is a separate pure function.

### Anti-Pattern 2: Direct SQLite Calls in Services

**What people do:** Services import `database/sql` and write raw SQL queries inline.

**Why it's wrong:** Tightly couples business logic to SQLite. Impossible to mock for tests. SQL scattered across codebase.

**Do this instead:** Services call `store.GetPlayer()`, `store.SaveResult()`. All SQL lives in `internal/store/sqlite/`. Tests use `internal/store/mock/`.

### Anti-Pattern 3: Using ORM for SQLite

**What people do:** Use GORM or Ent with SQLite.

**Why it's wrong:** Adds a heavy dependency for a simple schema. SQLite's strengths (direct SQL, simplicity, single-file) are undermined. ORMs generate suboptimal queries for SQLite's query planner.

**Do this instead:** Use `database/sql` directly with prepared statements. Write SQL by hand — it's readable, optimizable, and matches SQLite's philosophy. For a schema this small (< 10 tables), raw SQL is simpler than any ORM.

### Anti-Pattern 4: Non-Deterministic Calculation

**What people do:** Use floating-point operations, time-dependent functions, or random values in the calculation engine.

**Why it's wrong:** Same inputs produce different outputs. Audit trail becomes unreliable. Players and clubs lose trust in the system.

**Do this instead:** Use `decimal` or integer arithmetic (cents) for all money calculations. The CalcEngine must produce identical results given identical inputs, always.

### Anti-Pattern 5: Monolithic main.go

**What people do:** All configuration, dependency wiring, HTTP setup, and server start in a single 500-line `main()` function.

**Why it's wrong:** Hard to test individual components. Configuration can't be tested separately. Startup failures are opaque.

**Do this instead:** `cmd/server/main.go` is thin — it calls a `NewServer()` constructor that wires dependencies, then `server.Start()`. Configuration is a separate struct loaded from env/flags.

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| CSV files | File upload via API endpoint | Manual import, no external dependency |
| Reverse proxy (Nginx/Caddy) | Standard HTTP proxy | TLS termination, static file serving |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| HTTP ↔ Service | Direct function calls | Handlers import services |
| Service ↔ Store | Interface method calls | Services only know the Store interface |
| Service ↔ CalcEngine | Direct function calls | CalcEngine is stateless, pure functions |
| CalcEngine ↔ Store | None (strict boundary) | CalcEngine never touches the database |
| CSV ↔ Service | Via StatsProvider interface | Service receives parsed records, stores them |

## Build Order Implications

Based on dependency analysis, suggested component build order:

1. **`internal/domain/`** — Zero dependencies. Define all types first.
2. **`internal/store/` (interface)** — Define Store interface based on domain types.
3. **`internal/calc/`** — Pure calculation engine. Build and test thoroughly (this is the money code).
4. **`internal/store/sqlite/`** — Implement Store with SQLite. Migrations + queries.
5. **`internal/csv/`** — CSV parser and validator (independent of HTTP).
6. **`internal/service/`** — Wire Store + CalcEngine + CSV together.
7. **`internal/http/`** — Handlers that call services. Build last (depends on everything).
8. **`cmd/server/main.go`** — Wire everything together.

## Sources

- golang-standards/project-layout (github.com/golang-standards/project-layout) — 55.8k stars, community standard
- Go official docs: Organizing a Go module (go.dev/doc/modules/layout)
- Go standard library: database/sql (go1.26.2) — connection pooling, transactions, prepared statements
- modernc.org/sqlite — Pure Go SQLite driver, no CGo required
- Effective Go (go.dev/doc/effective_go) — Naming, interfaces, error handling conventions

---
*Architecture research for: FootballIncentiveLab*
*Researched: 2026-04-20*
