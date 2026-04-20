# Pitfalls Research

**Domain:** Football player compensation calculation (Go + SQLite WAL)
**Researched:** 2026-04-20
**Confidence:** HIGH

## Critical Pitfalls

### Pitfall 1: Float64 for Money Calculations

**What goes wrong:**
Using `float64` to represent salaries, bonuses, or any monetary value introduces rounding errors that compound across calculations. A KPI bonus calculation that multiplies a base salary by a percentage can produce results like `100000 * 0.15 = 14999.999999999998` instead of `15000.00`. Over an entire squad of 25+ players with multiple KPIs, these micro-errors accumulate into significant discrepancies — potentially hundreds of euros off per season.

**Why it happens:**
Go's `float64` uses IEEE 754 binary floating point, which cannot exactly represent most decimal fractions. Developers default to `float64` because it's the standard numeric type and "seems precise enough." Financial APIs and databases often use it too, normalizing the mistake.

**How to avoid:**
Use integer cents (store `int64` = value * 100) or a fixed-point decimal library like [`shopspring/decimal`](https://github.com/shopspring/decimal). For this project, integer cents is simplest — store all monetary values as `int64` cents in SQLite as `INTEGER` columns. All calculations operate on integers, division happens only at display time. This guarantees deterministic, reproducible results across runs.

**Warning signs:**
- Any `float64` field in a struct representing money
- Test assertions using approximate equality (`math.Abs(a-b) < 0.01`)
- Different results for the same calculation when run in different order

**Phase to address:**
Phase 1 (Schema & Domain Models) — define money types before writing any calculation code.

---

### Pitfall 2: Unbounded WAL File Growth

**What goes wrong:**
SQLite WAL mode works by appending changes to a `-wal` file before checkpointing them back to the main database. If a long-running read transaction holds an "end mark," checkpointing cannot complete past that point. In this app, the SSR dashboard rendering a page runs a read transaction; if the page takes time or multiple dashboard requests overlap, the WAL file grows unbounded. This causes disk space consumption to spike and read performance to degrade proportionally to WAL size.

**Why it happens:**
Developers assume WAL "just works" without understanding checkpointing. Long-running reads (dashboard queries, report generation) block the checkpointer from resetting the WAL. With concurrent dashboard users, the WAL can grow to hundreds of MB.

**How to avoid:**
- Keep read transactions short — fetch data, close rows, then render templates. Never hold a read transaction open while doing slow template rendering.
- Use `rows.Close()` immediately after scanning, not deferred to function exit if the function does more work.
- Consider running `PRAGMA wal_checkpoint(PASSIVE)` periodically during idle moments (e.g., a background goroutine).
- Set `PRAGMA journal_size_limit=N` to cap WAL file size.

**Warning signs:**
- `-wal` file growing beyond a few MB
- Dashboard queries getting progressively slower over time
- Disk usage increasing without proportional data growth

**Phase to address:**
Phase 2 (Data Layer & Repositories) — establish query patterns that close transactions promptly.

---

### Pitfall 3: SQLite WAL-Reset Bug (Versions < 3.51.3)

**What goes wrong:**
A data race bug present in SQLite versions 3.7.0 through 3.51.2 (fixed in 3.51.3, March 2026) can cause database corruption when two connections write or checkpoint simultaneously in WAL mode. The bug is rare but catastrophic — it silently corrupts data by skipping committed transactions during checkpointing.

**Why it happens:**
The Go `modernc.org/sqlite` driver (pure Go, CGo-free) may lag behind the latest C SQLite release. If the driver embeds a vulnerable version, the corruption risk exists even though "it almost never happens." For a financial calculation system, "almost never" is unacceptable.

**How to avoid:**
- Verify that the embedded SQLite version in your Go driver is >= 3.51.3 or has the backported fix.
- If using `modernc.org/sqlite`, check release notes for the bundled SQLite version.
- Alternative: use `github.com/mattn/go-sqlite3` (CGo) with a system SQLite >= 3.51.3.
- As a safety net: periodically run `PRAGMA integrity_check` and keep database backups.

**Warning signs:**
- Using a SQLite driver version older than March 2026 without verified backport
- No automated integrity checks
- No backup strategy

**Phase to address:**
Phase 0 (Project Setup) — choose driver version wisely; verify before writing code.

---

### Pitfall 4: Connection Pool Misconfiguration with SQLite

**What goes wrong:**
Go's `database/sql` defaults to a connection pool with up to `MaxOpenConns=0` (unlimited). With SQLite, each connection is a separate OS process opening the same file. Too many connections cause `SQLITE_BUSY` errors on writes (only one writer at a time in WAL mode) and excessive WAL index shared memory contention. Developers coming from PostgreSQL backgrounds configure large pools and hit mysterious lock errors.

**Why it happens:**
Documentation for `database/sql` is written for client-server databases where more connections = more throughput. SQLite is embedded — the pool size should match the actual concurrency the file can handle (1 writer, multiple readers).

**How to avoid:**
- Set `SetMaxOpenConns(1)` for write-heavy paths, or at most a small number (e.g., 4) for mixed read/write.
- Use `SetMaxIdleConns(2)` to keep a couple connections warm.
- For write operations, consider using a single dedicated connection via `db.Conn(ctx)`.
- Always set `SetConnMaxLifetime(0)` — SQLite connections don't benefit from recycling.

**Warning signs:**
- Intermittent `SQLITE_BUSY` errors under load
- Default `db.SetMaxOpenConns` not being called
- Pool stats showing high wait counts

**Phase to address:**
Phase 2 (Data Layer & Repositories) — configure pool at database initialization.

---

### Pitfall 5: CSV Import Encoding and Format Chaos

**What goes wrong:**
Football statistics CSVs come from various sources: Opta feeds, manual Excel exports, third-party data providers. Each uses different encodings (UTF-8, Windows-1252, ISO-8859-1), delimiters (comma, semicolon, tab), decimal separators (period vs comma for European data), and date formats. A CSV parser that assumes standard RFC 4180 format silently corrupts data or crashes on the first non-conforming file.

**Why it happens:**
CSV seems trivial — "just split by comma." But Italian football data likely uses semicolons as delimiters and commas as decimal separators (European convention). Excel exports often include BOM bytes. Some providers quote inconsistently.

**How to avoid:**
- Detect encoding (handle BOM) and delimiter before parsing.
- Normalize decimal commas to periods during import.
- Validate column count per row — reject files with inconsistent column counts.
- Define a strict schema per CSV type (match stats, player data) and reject files that don't conform.
- Log the first N import errors and report them, don't just fail silently.
- Provide a dry-run mode that validates without persisting.

**Warning signs:**
- Hardcoded delimiter or encoding assumptions
- No validation step before database insertion
- Import "succeeds" but data looks wrong

**Phase to address:**
Phase 3 (CSV Import & Data Pipeline) — build robust import before connecting to calculation engine.

---

### Pitfall 6: KPI Calculation Order Dependence

**What goes wrong:**
When calculating bonuses for multiple KPIs, the order of evaluation can affect the final result if there are interdependencies (e.g., a "total bonus" cap, or KPIs that reference each other's results). Running the same player's calculation in a different KPI order produces different totals. This violates the core "deterministic" requirement.

**Why it happens:**
Go maps iterate in random order. If KPIs are stored in a map and evaluated in iteration order, results are non-deterministic. Even with sorted KPIs, cascading effects (bonus cap applies after each KPI vs. after all KPIs) produce different results.

**How to avoid:**
- Always sort KPIs by a defined key (e.g., KPI ID or priority) before evaluation.
- Define a clear calculation pipeline: (1) evaluate each KPI independently, (2) sum bonuses, (3) apply caps and fairness rules as a post-processing step — never inline.
- Make the calculation engine a pure function: input → output with no side effects, no randomness, no time dependency.
- Write golden-file tests: fixed input → expected exact output (in cents).

**Warning signs:**
- Using `map[string]KPI` without sorting keys before iteration
- Caps or fairness rules applied inside the KPI loop rather than after
- Test results flaky (sometimes pass, sometimes fail with different totals)

**Phase to address:**
Phase 4 (Calculation Engine) — design the calculation pipeline with determinism as a first-class constraint.

---

### Pitfall 7: html/template XSS via User Data

**What goes wrong:**
Player names, KPI descriptions, and contract notes are user-supplied data rendered in HTML templates. Go's `html/template` auto-escapes by default, but only when using `{{.}}` or `{{.Field}}` — not when using `template.HTML()` or injecting raw strings. A KPI description containing `<script>alert(1)</script>` stored in the database becomes an XSS vector if rendered without escaping.

**Why it happens:**
Developers use `template.HTML()` to bypass escaping for "trusted" formatted content, or construct HTML strings in Go code and pass them to templates. In a back-office tool with internal users, the risk seems low — but a compromised or careless user can still exploit it.

**How to avoid:**
- Never use `template.HTML()`, `template.JS()`, or related unsafe types.
- Let `html/template` handle all escaping — it's context-aware (HTML, JS, CSS, URL).
- If rich text is needed for descriptions, sanitize server-side with a library like `bluemonday`.
- Set `Content-Security-Policy` headers to inline-script restrictions.

**Warning signs:**
- Any use of `template.HTML()` in code
- String concatenation building HTML in Go handlers
- No CSP headers in HTTP responses

**Phase to address:**
Phase 5 (SSR Dashboard) — establish template patterns that never bypass escaping.

---

### Pitfall 8: SQLite Backup Without WAL Files

**What goes wrong:**
When backing up a SQLite database in WAL mode, copying only the `.db` file loses uncommitted WAL transactions. The backup appears valid but is missing recent data. Restoring from such a backup loses the most recent player contracts, KPI updates, or calculation runs.

**Why it happens:**
Developers treat SQLite like a regular file — "just copy it." But WAL mode creates three files: `.db`, `.db-wal`, `.db-shm`. All three must be copied atomically (or the database must be checkpointed first).

**How to avoid:**
- For file-level backups: copy all three files (`.db`, `-wal`, `-shm`) together.
- Better: use the SQLite Online Backup API (`VACUUM INTO` or the backup API) which handles WAL correctly.
- Or: run `PRAGMA wal_checkpoint(TRUNCATE)` before file-copy backups to flush WAL to the main file.
- Automate backups and test restore procedures.

**Warning signs:**
- Backup script copies only the `.db` file
- No restore test has ever been run
- Backups taken while the application is running without checkpointing

**Phase to address:**
Phase 1 (Schema & Domain Models) — implement backup strategy early, not as an afterthought.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Using float64 for money | Faster to code, no extra dependencies | Rounding errors compound, audit failures | Never — use int64 cents from day 1 |
| Skipping input validation on CSV import | Faster to prototype | Silent data corruption, wrong calculations | Never — validation is the import's job |
| Hardcoding KPI formulas in Go | Simple implementation | Every formula change requires recompile + redeploy | MVP only — migrate to config-driven ASAP |
| No schema migration tool | Zero setup | Manual ALTER TABLE on production, drift risk | First 2-3 iterations only |
| Single `main.go` with everything | Fastest possible start | Unmaintainable beyond ~500 lines | First prototype only |
| Skipping integration tests | Faster feature delivery | Calculation bugs found by users, not devs | Never for calculation engine |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| SQLite via `modernc.org/sqlite` | Assuming version parity with C SQLite | Check bundled version, verify critical fixes |
| `database/sql` with SQLite | Default pool size too large | Set `MaxOpenConns` to 1-4 |
| CSV parsing | Assuming UTF-8 and comma delimiter | Detect encoding, handle European formats |
| html/template | Bypassing auto-escaping for "trusted" data | Never use `template.HTML()`, trust the escaper |
| Nginx/Caddy reverse proxy | Not configuring request timeouts for long calculations | Set appropriate read/write timeouts |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| WAL file growing unbounded | Disk usage spikes, reads slow down | Short read transactions, periodic checkpointing | Concurrent dashboard users > 3-5 |
| N+1 queries in dashboard | Page load > 2s with 25 players | Use JOINs or batch queries | Player roster > 15 |
| Re-parsing CSV on every request | Import endpoint slow, CPU spikes | Parse once, validate, bulk insert in transaction | CSV files > 1000 rows |
| Full table scans on KPI lookup | Queries slow as data grows | Index on `(player_id, season, kpi_id)` | > 10K performance records |
| Template rendering holding DB transactions | WAL growth, slow commits | Fetch all data, close transaction, then render | Any multi-second template |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| No auth on API endpoints | Anyone on the network can modify contracts | Token-based auth from Phase 1, even if simple |
| SQL injection via string concatenation | Data exfiltration, modification | Always use parameterized queries (`?` placeholders) |
| Storing API tokens in source code | Credential leak via git history | Environment variables or config file excluded from VCS |
| No HTTPS behind reverse proxy | Credentials and salary data in transit as plaintext | TLS termination at Nginx/Caddy, force HTTPS |
| Backup files in web-accessible directory | Database download via URL | Store backups outside document root |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Showing raw calculation numbers without formatting | "1523497" instead of "€1,523,497.00" | Format with locale-aware number formatting in templates |
| No loading indicator during calculation run | User thinks app is frozen | Show progress or "calculating..." state |
| CSV import with no feedback | User doesn't know if import succeeded or failed | Report rows imported, rows rejected, errors per row |
| Error messages that say "internal server error" | User has no idea what went wrong | Show actionable messages: "CSV column 'goals' contains non-numeric value at row 14" |
| Dashboard requires full page reload after calculation | Poor experience for iterative simulations | Consider HTMX partial refresh for results |

## "Looks Done But Isn't" Checklist

- [ ] **Money calculations:** Verified all values are int64 cents — search for `float64` in calculation paths
- [ ] **CSV import:** Tested with European-format CSVs (semicolons, comma decimals, BOM header)
- [ ] **Calculation determinism:** Same input produces byte-identical output across 100 runs
- [ ] **WAL behavior:** Monitored `-wal` file size during typical usage patterns
- [ ] **Template safety:** Searched codebase for `template.HTML` usage, verified CSP headers
- [ ] **Backup/restore:** Actually restored a backup and verified data integrity
- [ ] **Connection pool:** Verified `MaxOpenConns` setting, tested under concurrent load
- [ ] **SQLite version:** Confirmed driver version >= 3.51.3 or has WAL-reset fix

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Float64 rounding errors in production | HIGH | Rewrite all money fields to int64 cents, re-import data, re-run all calculations |
| WAL corruption (pre-3.51.3 bug) | HIGH | Restore from last known-good backup, upgrade driver, run integrity_check |
| CSV import with bad data | MEDIUM | Identify corrupted records via audit log, re-import affected data |
| Unbounded WAL growth | LOW | Checkpoint manually (`PRAGMA wal_checkpoint`), fix long-running read transactions |
| XSS in templates | MEDIUM | Audit all template output, fix escaping, rotate any exposed tokens |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|-------------|
| Float64 for money | Phase 1 — Schema & Domain Models | Grep for float64 in money-related structs; golden-file tests |
| Unbounded WAL growth | Phase 2 — Data Layer | Monitor -wal file size; test with concurrent reads |
| SQLite WAL-reset bug | Phase 0 — Project Setup | Check driver version; add integrity_check to health endpoint |
| Connection pool misconfiguration | Phase 2 — Data Layer | Verify MaxOpenConns=1-4; load test with concurrent writes |
| CSV import chaos | Phase 3 — CSV Import | Test with real European-format CSVs; dry-run mode |
| KPI calculation order dependence | Phase 4 — Calculation Engine | Golden-file tests; run same calculation 100x, compare outputs |
| Template XSS | Phase 5 — SSR Dashboard | Search for template.HTML; verify CSP headers |
| Backup without WAL | Phase 1 — Schema & Domain Models | Restore test; verify all 3 files copied |

## Sources

- SQLite WAL documentation: https://www.sqlite.org/wal.html (retrieved 2026-04-20)
- SQLite WAL-Reset Bug disclosure: https://www.sqlite.org/wal.html#the_wal_reset_bug (2026-03-03)
- Go database/sql documentation: https://pkg.go.dev/database/sql (go1.26.2)
- Go prepared statements guide: https://go.dev/doc/database/prepared-statements
- Effective Go: https://go.dev/doc/effective_go
- modernc.org/sqlite driver: pure Go SQLite implementation (verify bundled version)
- shopspring/decimal: https://github.com/shopspring/decimal (Go arbitrary-precision decimal)

---

*Pitfalls research for: Football player compensation calculation (Go + SQLite WAL)*
*Researched: 2026-04-20*
