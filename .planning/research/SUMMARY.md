# Project Research Summary

**Project:** FootballIncentiveLab
**Domain:** Sports analytics / football contract management
**Researched:** 2026-04-20
**Confidence:** HIGH

## Executive Summary

FootballIncentiveLab is a Go monolith that calculates annual football player salaries from a base contract plus variable KPI-based bonuses, applying fairness rules and bonus caps. It's a deterministic, auditable back-office tool for a single club — not a SaaS platform. The recommended stack is deliberately minimal: Go 1.26.2 with pure-Go SQLite (no CGo), the chi HTTP router, server-side rendered templates with HTMX for interactivity, and Tailwind CSS via CDN for styling. This produces a single binary deployable on a low-cost VPS with no build pipeline, no JavaScript framework, and no external database server.

The core technical risk is the calculation engine — it handles money and must be deterministic, auditable, and free of floating-point errors. Research identifies critical pitfalls: using `float64` for monetary values, KPI evaluation order producing non-deterministic results, and SQLite WAL-mode corruption in older driver versions. The architecture mitigates these by isolating the calculation engine as a pure Go package (no DB, no HTTP imports), using integer cents (`int64`) for all monetary values, and pinning to a modern SQLite driver version (>= 3.51.3). The project structure follows established Go conventions with clear layer separation: domain types → store interface → pure calc engine → SQLite implementation → services → HTTP handlers.

Feature prioritization is clear: 10 table-stake features (Player CRUD, Contract CRUD, KPI definitions, CSV import, calculation engine, REST API, dashboard, history) form the MVP, with differentiators like simulation mode, fairness rules engine, and multi-season comparison deferred to Phase 2. No direct competitor product with this exact scope was found — existing tools (Capology, Transfermarkt) focus on market values and transfers, not internal club compensation calculation with KPI-based bonuses. This validates the niche but also means user expectations must be shaped through the product itself.

## Key Findings

### Recommended Stack

See [STACK.md](./STACK.md) for full details.

**Core technologies:**
- **Go 1.26.2**: Runtime and language — latest stable, standard library covers 90% of needs for this domain
- **modernc.org/sqlite v1.49.1**: Pure Go SQLite driver (no CGo) — enables single-binary cross-compilation without gcc on target
- **go-chi/chi/v5 v5.2.5**: HTTP router — zero external deps, 100% net/http compatible, built-in middleware stack
- **html/template (stdlib)**: Server-side rendering — built-in, auto-escaping against XSS, perfect for back-office dashboards
- **htmx.org 2.0.8**: AJAX without JS — HTML-first, partial page updates, no SPA build step
- **goose v3.27.0**: Schema migrations — Go library + CLI, SQLite support, embedded migrations via `//go:embed`
- **log/slog (stdlib)**: Structured logging — built-in since Go 1.21, no third-party dependency

**Key version constraint:** `modernc.org/sqlite` v1.49.1 requires `modernc.org/libc` v1.49.1 (exact match). Pin carefully.

### Expected Features

See [FEATURES.md](./FEATURES.md) for full details.

**Must have (table stakes):**
- Player CRUD — foundational entity, no system without it
- Contract CRUD with base salary — defines fixed compensation component
- KPI definition per role — different positions track different metrics
- KPI scaling formula per KPI — raw stats map to bonus percentages via curves
- Match statistics CSV import — external data ingestion without API providers
- Deterministic calculation engine — core value proposition, pure computation
- Bonus cap enforcement — prevents runaway payouts
- Calculation history / audit trail — clubs need to see past runs
- REST API — all operations must be programmable
- Basic web dashboard — visualize roster and compensation results

**Should have (differentiators — Phase 2):**
- Simulation mode ("what-if") — test scenarios before committing
- Fairness rules engine — configurable constraints beyond simple caps
- Role-based KPI templates — pre-built KPI sets per position (GK, DEF, MID, FWD)
- CSV export of calculation breakdowns — finance team needs reports
- KPI contribution breakdown — show how each KPI contributed to final bonus
- Multi-season comparison — track year-over-year compensation changes
- Batch run with diff view — run entire squad, show what changed
- API key authentication — simple auth for internal club use

**Defer to v2+:**
- Real-time match data feeds — CSV is sufficient for MVP
- Multi-club/tenant support — single club is the validated need
- Team-level KPIs — add after individual KPIs work
- Pro-rata calculation for mid-season signings — edge case, not critical for first runs

### Architecture Approach

See [ARCHITECTURE.md](./ARCHITECTURE.md) for full details.

A three-layer Go monolith with strict dependency boundaries: HTTP handlers → Service layer → Repository (Store interface) → SQLite. The calculation engine is a fourth isolated component — pure functions with no database or HTTP imports. This separation is non-negotiable because the engine handles money and must be maximally testable.

**Major components:**
1. **Domain (`internal/domain/`)** — Pure data types (Player, Contract, KPI, Performance, CompensationResult). Zero internal imports; everything depends on this.
2. **Calculation Engine (`internal/calc/`)** — Pure Go, no I/O. Takes domain types in, returns results. Table-driven tests with golden files for regression safety.
3. **Store (`internal/store/`)** — Interface-based data access. SQLite implementation in `sqlite/`, mock in `mock/` for testing services without a database.
4. **Services (`internal/service/`)** — Orchestration layer. Wires Store + CalcEngine + CSV. Business logic validation lives here.
5. **HTTP Layer (`internal/http/`)** — Thin handlers: parse requests, call services, format responses. REST JSON + SSR dashboard.

**Build order (dependency-driven):** domain → store interface → calc engine → store/sqlite → csv → services → http handlers → cmd/server/main.go

### Critical Pitfalls

See [PITFALLS.md](./PITFALLS.md) for full details.

1. **Float64 for money calculations** — IEEE 754 rounding errors compound across squad calculations. Use `int64` cents from day 1. *Address in Phase 1.*
2. **KPI calculation order dependence** — Go maps iterate randomly; same inputs produce different totals if KPIs are stored in a map. Always sort KPIs before evaluation, apply caps as post-processing. *Address in Phase 4.*
3. **SQLite WAL-reset bug (versions < 3.51.3)** — Data corruption when two connections write/checkpoint simultaneously. Verify driver version before writing code. *Address in Phase 0.*
4. **Connection pool misconfiguration** — Default unlimited pool causes SQLITE_BUSY errors. Set `MaxOpenConns(1)`. *Address in Phase 2.*
5. **CSV import encoding chaos** — Italian/European football data uses semicolons, comma decimals, BOM headers. Detect encoding and delimiter before parsing. *Address in Phase 3.*
6. **Unbounded WAL file growth** — Long-running read transactions block checkpointing. Keep read transactions short, close rows immediately. *Address in Phase 2.*
7. **Template XSS via `template.HTML()`** — Bypassing auto-escaping creates XSS vectors. Never use unsafe template types. *Address in Phase 5.*

## Implications for Roadmap

Based on research, suggested phase structure:

### Phase 0: Project Setup & Foundation
**Rationale:** Establish project skeleton, verify SQLite driver version (>= 3.51.3 for WAL-reset fix), configure connection pool, set up build toolchain before any feature code.
**Delivers:** `go mod init`, dependency installation, Makefile, basic `cmd/server/main.go` that starts and stops, CI-ready project structure.
**Uses:** Go 1.26.2, modernc.org/sqlite v1.49.1, chi v5, goose v3
**Avoids:** Pitfall 3 (SQLite WAL-reset bug) — verify driver before proceeding
**Research flag:** Standard patterns — no deep research needed

### Phase 1: Domain Models & Schema
**Rationale:** Define all domain types and database schema first. Everything else depends on these. Money types must be `int64` cents from the start — retrofitting is painful.
**Delivers:** `internal/domain/` package with all types (Player, Contract, KPI, Performance, CompensationResult), SQLite migrations (players, contracts, kpis, performances tables), Store interface definition.
**Addresses:** Player CRUD (entity), Contract CRUD (entity), KPI definition (entity)
**Avoids:** Pitfall 1 (float64 for money) — define int64 cents in domain types; Pitfall 8 (backup without WAL) — establish backup strategy early
**Research flag:** Standard patterns — well-documented Go project layout and SQLite schema design

### Phase 2: Data Layer & Store Implementation
**Rationale:** Implement the Store interface with SQLite. Establish query patterns that avoid WAL growth and connection pool issues before adding business logic.
**Delivers:** `internal/store/sqlite/` with all CRUD queries, `internal/store/mock/` for testing, WAL configuration, connection pool setup.
**Uses:** modernc.org/sqlite with WAL PRAGMAs, database/sql with MaxOpenConns(1)
**Avoids:** Pitfall 2 (unbounded WAL growth) — short read transactions; Pitfall 4 (connection pool misconfiguration) — configure pool at init
**Research flag:** Standard patterns — SQLite CRUD with Go is well-documented

### Phase 3: CSV Import & Data Pipeline
**Rationale:** Data ingestion must work before the calculation engine can be tested with real data. Build robust CSV parsing that handles European formats.
**Delivers:** `internal/csv/` package (parser + validator), `internal/service/import.go`, `/api/import` endpoint, dry-run validation mode.
**Addresses:** Match statistics CSV import (table stakes)
**Avoids:** Pitfall 5 (CSV encoding chaos) — detect encoding/delimiter, handle European formats, validate before persisting
**Research flag:** May need research — European CSV format handling patterns (semicolons, comma decimals, BOM)

### Phase 4: Calculation Engine
**Rationale:** This is the highest-risk component (money calculations). Build it in isolation with exhaustive tests before wiring it to anything else. Pure Go, no dependencies outside domain.
**Delivers:** `internal/calc/` package (engine.go, scaling.go, fairness.go), `internal/calc/engine_test.go` with golden-file tests, deterministic pipeline: scale KPIs → sum weighted bonuses → apply caps.
**Addresses:** Deterministic calculation engine (core value), Bonus cap enforcement, KPI scaling formulas
**Avoids:** Pitfall 1 (float64 for money) — int64 cents throughout; Pitfall 6 (KPI order dependence) — sort KPIs, pure pipeline, post-processing caps
**Research flag:** Needs deeper research — KPI scaling formula design (linear, threshold, step curves), golden-file test strategy

### Phase 5: Services & REST API
**Rationale:** Wire Store + CalcEngine + CSV into services, then expose via REST API. Services are thin orchestration; handlers are thinner request/response formatting.
**Delivers:** `internal/service/` (PlayerSvc, KPISvc, CompensationSvc, ImportSvc), `internal/http/handlers/` for all endpoints, API key auth middleware.
**Addresses:** REST API (table stakes), Calculation history / audit trail, all CRUD operations
**Implements:** Service layer pattern from ARCHITECTURE.md
**Research flag:** Standard patterns — Go REST API with chi is well-documented

### Phase 6: SSR Dashboard
**Rationale:** Build last because it depends on all other components. HTMX provides interactivity without a JS build step.
**Delivers:** `internal/http/templates/` (layout, dashboard, partials), `internal/http/handlers/dashboard.go`, HTMX partials for live updates, Tailwind CSS styling.
**Addresses:** Basic web dashboard (table stakes)
**Avoids:** Pitfall 7 (template XSS) — never use `template.HTML()`, trust auto-escaping; Pitfall 2 (WAL growth) — fetch all data, close transaction, then render
**Research flag:** Standard patterns — Go html/template + HTMX is well-documented

### Phase 7: Simulation & Differentiators (Phase 2)
**Rationale:** Only after core engine is stable and tested. Simulation mode reuses the CalcEngine with hypothetical inputs. Fairness rules extend the engine's post-processing.
**Delivers:** `/api/simulate` endpoint, fairness rules configuration, KPI contribution breakdown, CSV export, role-based KPI templates.
**Addresses:** Simulation mode, Fairness rules engine, KPI breakdown, CSV export, KPI templates
**Research flag:** Needs deeper research — simulation UX patterns, fairness rule expression language design

### Phase Ordering Rationale

- **Domain-first:** All components depend on domain types. No code can be written without them.
- **Calc engine isolated:** The highest-risk component (money) is built in Phase 4 with zero dependencies on infrastructure. This enables exhaustive testing before integration.
- **CSV before services:** Data ingestion must work so the calc engine can be tested with real data, not just hand-crafted test fixtures.
- **Services before UI:** REST API enables programmatic testing. Dashboard is a thin layer on top — building it last avoids premature optimization of templates.
- **Simulation deferred:** Reuses the stable calc engine. Building it early would mean debugging two things at once (engine + simulation).
- **Pitfall-driven ordering:** Each phase addresses specific pitfalls identified in research, preventing costly retrofitting.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 3:** CSV import — European format handling patterns (semicolons, comma decimals, BOM detection). May need `/gsd-research-phase`.
- **Phase 4:** Calculation engine — KPI scaling formula design (linear, threshold, step curves), golden-file test methodology. May need `/gsd-research-phase`.
- **Phase 7:** Simulation & fairness — rule expression language, simulation UX patterns, fairness constraint algorithms. Needs `/gsd-research-phase`.

Phases with standard patterns (skip research-phase):
- **Phase 0:** Project setup — standard `go mod init` + dependency installation
- **Phase 1:** Domain models — standard Go struct + SQL schema design
- **Phase 2:** Store implementation — standard SQLite CRUD with Go
- **Phase 5:** REST API — standard chi router + JSON handlers
- **Phase 6:** SSR dashboard — standard html/template + HTMX

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All technologies verified via pkg.go.dev, official docs, and import counts. Version numbers confirmed as of April 2026. |
| Features | MEDIUM | No direct competitor product found with identical scope. Feature list derived from domain analysis (Capology, Transfermarkt feature observation) and PROJECT.md requirements. User validation needed. |
| Architecture | HIGH | Based on established Go project layout conventions (golang-standards/project-layout, 55.8k stars). Patterns are well-documented and battle-tested. |
| Pitfalls | HIGH | All pitfalls sourced from official SQLite docs, Go database/sql docs, and real-world experience reports. WAL-reset bug specifically documented by SQLite team (March 2026). |

**Overall confidence:** HIGH

The technology stack and architecture are well-grounded in established Go and SQLite best practices. The main uncertainty is feature prioritization — no exact comparable product exists, so the MVP feature set should be validated with real users (club staff) as early as possible. The calculation engine is the highest-risk component and warrants the most rigorous testing strategy.

### Gaps to Address

- **User validation of MVP features:** FEATURES.md confidence is MEDIUM because no competitor with this exact scope was found. Validate the 10 table-stake features with actual club staff during early phases.
- **KPI scaling formula specifics:** The research identifies that scaling formulas (linear, threshold, step, capped) are needed but doesn't specify exact mathematical models. This needs domain expert input during Phase 4 planning.
- **Fairness rules complexity:** The fairness rules engine is flagged as a differentiator but the rule expression language and constraint algorithms are not researched in depth. Needs dedicated research before Phase 7.
- **Italian football data format specifics:** CSV import needs to handle European formats but the exact column schemas from Italian football data providers (Opta, manual Excel exports) are not confirmed. Test with real data samples early.
- **SQLite driver version verification:** PITFALLS.md flags the WAL-reset bug for versions < 3.51.3. The bundled SQLite version in modernc.org/sqlite v1.49.1 must be explicitly verified before proceeding past Phase 0.

## Sources

### Primary (HIGH confidence)
- `modernc.org/sqlite` v1.49.1 — [pkg.go.dev](https://pkg.go.dev/modernc.org/sqlite) — Pure Go SQLite driver
- `go-chi/chi` v5.2.5 — [pkg.go.dev](https://pkg.go.dev/github.com/go-chi/chi/v5) — HTTP router
- `pressly/goose` v3.27.0 — [pkg.go.dev](https://pkg.go.dev/github.com/pressly/goose/v3) — Migrations
- Go 1.26.2 stdlib — `log/slog`, `html/template`, `database/sql`, `encoding/csv`, `embed`
- SQLite WAL documentation — [sqlite.org/wal.html](https://www.sqlite.org/wal.html)
- SQLite WAL-Reset Bug — [sqlite.org/wal.html#the_wal_reset_bug](https://www.sqlite.org/wal.html#the_wal_reset_bug) (2026-03-03)
- golang-standards/project-layout — [github.com/golang-standards/project-layout](https://github.com/golang-standards/project-layout) — 55.8k stars

### Secondary (MEDIUM confidence)
- Capology (capology.com) — observed feature set for football salary management
- Transfermarkt (transfermarkt.com) — observed feature set for player contracts and statistics
- Football Benchmark (footballbenchmark.com) — football business intelligence confirming demand for financial transparency tools
- GitHub football-analytics topic — 506 repos; mostly analytics/prediction, none focused on internal club compensation calculation

### Tertiary (LOW confidence)
- `htmx.org` 2.0.8 — [htmx.org/docs](https://htmx.org/docs/) — Frontend enhancement; patterns assumed from documentation, not production experience
- Tailwind CSS via CDN — styling approach for MVP; production would require build step

---

*Research completed: 2026-04-20*
*Ready for roadmap: yes*
