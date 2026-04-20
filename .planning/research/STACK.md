# Technology Stack

**Project:** FootballIncentiveLab
**Researched:** 2026-04-20
**Confidence:** HIGH

## Recommended Stack

### Core Framework

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go | 1.26.2 | Runtime & language | Latest stable (April 2026), standard library covers 90% of needs for this domain |
| `modernc.org/sqlite` | v1.49.1 | SQLite driver (database/sql) | Pure Go, no CGo — single binary cross-compilation without gcc on target. SQLite 3.53.0. Imported by 3,500+ packages. |
| `github.com/go-chi/chi/v5` | v5.2.5 | HTTP router | 17,000+ importers, zero external deps, 100% net/http compatible, built-in middleware stack. Production at Cloudflare, Heroku. |
| `html/template` | stdlib | Server-side rendering | Built-in, battle-tested, auto-escaping against XSS. Perfect for back-office dashboards. |

### Database

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| `modernc.org/sqlite` | v1.49.1 | SQLite via database/sql | Pure Go (no CGo), WAL mode support, DSN-based PRAGMA config. Single-file DB, no server overhead. |
| `github.com/pressly/goose/v3` | v3.27.0 | Schema migrations | Go library + CLI, supports SQLite, embedded migrations via `//go:embed`, 1,700+ importers. |

### Frontend Enhancement

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| `htmx.org` | 2.0.8 | AJAX without JS | HTML-first, works perfectly with Go SSR. Partial page updates, no SPA build step. 60KB gzipped. |
| Tailwind CSS (CDN) | v4.x | Styling | Single `<script>` tag via CDN for MVP. No build pipeline. Use Play CDN for prototyping, compile for production. |

### Logging & Observability

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| `log/slog` | stdlib (Go 1.21+) | Structured logging | Built-in since Go 1.21. JSON/Text handlers, levels, context-aware. 77,000+ importers. No third-party dependency. |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `encoding/csv` | stdlib | CSV parsing | CSV import for match statistics |
| `database/sql` | stdlib | SQL interface | All database operations via standard interface |
| `net/http` | stdlib | HTTP server | Core HTTP server, compatible with chi |
| `text/template` / `html/template` | stdlib | Templating | SSR dashboard pages |
| `embed` | stdlib | File embedding | Embed static assets (CSS, JS, templates) in binary |
| `context` | stdlib | Request context | Timeout, cancellation, request-scoped values |
| `testing` | stdlib | Test framework | Table-driven tests, subtests |
| `strconv` | stdlib | String conversion | CSV parsing, form handling |
| `math` | stdlib | Math operations | Salary/bonus calculations |
| `time` | stdlib | Date handling | Contract periods, calculation timestamps |
| `github.com/go-chi/chi/v5/middleware` | v5.2.5 | HTTP middleware | Logger, Recoverer, RequestID, RealIP, Timeout |
| `github.com/go-chi/cors` | latest | CORS | If API needs cross-origin access |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| Go toolchain (`go build`, `go test`) | Build, test, format | Built-in, `go build -o footballincentivelab .` produces single binary |
| `go vet` | Static analysis | Built-in, catches common mistakes |
| `golangci-lint` | Linter suite | Optional but recommended: `golangci-lint run` |
| `goose` CLI | Migration management | `goose -dir migrations sqlite3 ./football.db up` |
| `air` or `gow` | Hot reload (dev) | Auto-restart on file changes during development |

## Installation

```bash
# Initialize module
go mod init github.com/nicolasgentilucci/FootballIncentiveLab

# Core dependencies
go get github.com/go-chi/chi/v5@v5.2.5
go get github.com/go-chi/chi/v5/middleware
go get modernc.org/sqlite@v1.49.1
go get github.com/pressly/goose/v3@v3.27.0

# HTMX (CDN — no Go dependency needed)
# In HTML: <script src="https://cdn.jsdelivr.net/npm/htmx.org@2.0.8/dist/htmx.min.js"></script>

# Dev tools
go install github.com/air-verse/air@latest
go install github.com/pressly/goose/v3/cmd/goose@latest
```

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| `github.com/mattn/go-sqlite3` | Requires CGo + gcc on target system, complicates cross-compilation and single-binary deployment | `modernc.org/sqlite` — pure Go, no CGo |
| Gin / Echo / Fiber | Heavy framework overhead, less idiomatic, custom handler signatures break stdlib compatibility | `chi` — lightweight, stdlib-compatible |
| GORM / sqlx | ORM adds complexity, hides SQL, harder to debug, unnecessary for this scale | `database/sql` directly — full control, simple, no magic |
| React / Vue / Next.js SPA | Massive complexity for a back-office tool, build step, API layer, separate deployment | `html/template` + HTMX — same binary, SSR, no JS build |
| logrus / zerolog / zap | Third-party dependencies for what `log/slog` now does natively | `log/slog` — stdlib, structured, context-aware |
| golang-migrate | Less SQLite support, more complex API | `goose` — better SQLite support, simpler API, embed-friendly |
| Docker Compose (for MVP) | Unnecessary complexity for a single binary on a VPS | Direct binary + systemd service + Nginx/Caddy reverse proxy |

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| SQLite driver | `modernc.org/sqlite` | `mattn/go-sqlite3` | CGo dependency — harder to cross-compile, needs gcc on target |
| HTTP router | `chi` v5 | `net/http.ServeMux` (stdlib) | Chi's middleware stack, route groups, and URL params are worth the single dependency |
| HTTP router | `chi` v5 | `gin` | Gin uses custom handler signatures, breaks net/http compatibility, heavier |
| Migrations | `goose` v3 | `golang-migrate` | Goose has better SQLite support and simpler embedded migration API |
| Logging | `log/slog` | `zerolog` | Slog is stdlib, no dependency, and sufficient for this load |
| Frontend | `html/template` + HTMX | `templ` (type-safe templates) | Templ adds build step and complexity; html/template is sufficient for MVP |
| CSS | Tailwind CDN | Bootstrap | Tailwind utility-first is more flexible, lighter via CDN for MVP |

## Version Compatibility

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| `modernc.org/sqlite` v1.49.1 | `modernc.org/libc` v1.49.1 (exact match required) | Pin libc to exact same version — fragile dependency per upstream docs |
| `chi` v5.2.5 | Go 1.22+ | Supports four most recent Go major versions |
| `goose` v3.27.0 | Go 1.21+ | Provider API recommended over legacy global functions |
| `htmx` 2.0.8 | All modern browsers | No IE11 support (use htmx 1.x if needed) |
| `go` 1.26.2 | All listed dependencies | Latest stable as of April 2026 |

## SQLite WAL Configuration

Use DSN-based pragma configuration via `modernc.org/sqlite`:

```go
db, err := sql.Open("sqlite", "file:football.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=synchronous(NORMAL)")
```

| PRAGMA | Value | Why |
|--------|-------|-----|
| `journal_mode(WAL)` | WAL | Concurrent reads with writes, no readers-writer lock contention |
| `busy_timeout(5000)` | 5000ms | Wait up to 5s on lock before returning SQLITE_BUSY |
| `foreign_keys(1)` | ON | Enforce referential integrity |
| `synchronous(NORMAL)` | NORMAL | Balance between durability and performance for WAL mode |

Also set in application code:
```go
db.SetMaxOpenConns(1) // SQLite write serialization — only one writer at a time
```

## Sources

- `modernc.org/sqlite` v1.49.1 — [pkg.go.dev](https://pkg.go.dev/modernc.org/sqlite) — Published Apr 17, 2026 — HIGH confidence
- `mattn/go-sqlite3` v1.14.42 — [pkg.go.dev](https://pkg.go.dev/github.com/mattn/go-sqlite3) — Published Apr 8, 2026 — HIGH confidence (comparison)
- `go-chi/chi` v5.2.5 — [pkg.go.dev](https://pkg.go.dev/github.com/go-chi/chi/v5) — Published Feb 5, 2026 — HIGH confidence
- `pressly/goose` v3.27.0 — [pkg.go.dev](https://pkg.go.dev/github.com/pressly/goose/v3) — Published Feb 22, 2026 — HIGH confidence
- `log/slog` — [pkg.go.dev](https://pkg.go.dev/log/slog) — Go 1.26.2 stdlib — HIGH confidence
- `htmx.org` 2.0.8 — [htmx.org/docs](https://htmx.org/docs/) — HIGH confidence
- Go 1.26.2 — [go.dev/dl](https://go.dev/dl/) — Released Apr 2026 — HIGH confidence

---

*Stack research for: Go + SQLite WAL backend with html/template frontend*
*Researched: 2026-04-20*
