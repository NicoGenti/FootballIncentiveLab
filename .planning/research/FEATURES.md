# Feature Landscape

**Domain:** Football contract management & player compensation calculation
**Researched:** 2026-04-20
**Confidence:** MEDIUM (based on domain research + PROJECT.md context; no existing competitor product with identical scope found)

## Table Stakes

Features users expect. Missing = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Player CRUD (create, read, update, deactivate) | Core entity; no system without it | Low | Name, position, date of birth, nationality, status |
| Contract CRUD with base salary | Defines the fixed component of compensation | Low | start_date, end_date, base_annual_salary, currency, max_bonus_ratio |
| KPI definition per role | Different positions track different metrics | Medium | Role (GK, DEF, MID, FWD), KPI name, weight, target value, unit |
| KPI scaling formula per KPI | Raw stats must map to bonus % using curves | Medium | Linear, step, threshold, capped formulas |
| Match statistics CSV import | External data ingestion for KPI values | Medium | Validate columns, map to players, store per-match stats |
| Deterministic calculation engine | Core value proposition — salary = base + KPI bonuses | High | Pure computation, no AI, no randomness, fully auditable |
| Bonus cap enforcement | max_bonus_ratio prevents runaway payouts | Low | Clamp total bonus to contract-defined ceiling |
| Calculation history / audit trail | Clubs need to see past runs and their inputs | Medium | Store each "run" with timestamp, inputs, outputs |
| REST API | All operations must be programmable | Medium | JSON endpoints for all CRUD + calculation triggers |
| Basic web dashboard | Visualize player roster and compensation results | Medium | Go html/template, SSR, tables + summary views |

## Differentiators

Features that set product apart. Not expected, but valued.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Simulation mode ("what-if")** | Clubs can test scenarios before committing (e.g., "what if player scores 15 goals?") | High | Run calculation with hypothetical KPI values without saving; compare scenarios side by side |
| **Fairness rules engine** | Configurable constraints beyond simple caps (e.g., max bonus gap between positions, team-wide bonus budget ceiling) | High | Rule definitions: type, condition, scope (player/position/team), action (cap/warn/block) |
| **Role-based KPI templates** | Pre-built KPI sets per position (e.g., GK template = clean sheets + saves; ST template = goals + assists) | Medium | Reduce setup friction; customizable after applying |
| **CSV export of calculation breakdowns** | Finance teams need reports for payroll/accounting | Low | Per-player and per-team CSV with all KPI contributions itemized |
| **Multi-season comparison** | Track year-over-year compensation changes per player | Medium | Side-by-side view of same player across seasons |
| **KPI contribution breakdown** | Show exactly how each KPI contributed to the final bonus (not just total) | Medium | Transparency: "Goals KPI contributed €12,000 of €45,000 total bonus" |
| **Weighted team KPIs** | Some bonuses depend on team results (league position, cup progression) | Medium | Team-level KPI that distributes to squad based on appearance % |
| **Pro-rata calculation for mid-season signings** | Players who join mid-season should get pro-rated bonuses | Medium | Calculate based on fraction of season played |
| **Batch run with diff view** | Run calculations for entire squad, show what changed vs. previous run | Medium | Highlight increases/decreases per player |
| **API key authentication** | Simple auth for internal club use | Low | Token-based, not full OAuth — appropriate for single-club tool |

## Anti-Features

Features to explicitly NOT build.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| **Real-time match data feeds** | Adds infrastructure complexity (websockets, provider APIs, costs); batch CSV is sufficient for MVP | CSV import now; design StatsProvider interface for future plug-in |
| **AI/ML in calculation core** | Non-deterministic, non-auditable; clubs need to explain exactly how numbers were derived | Pure Go formulas only; AI can live at edges (KPI suggestions post-MVP) |
| **Multi-club/tenant support** | Massive complexity (data isolation, per-club config, billing); single club is the validated need | Single-club schema; evolve only if demand proven |
| **Mobile app** | Back-office tool used at desks; mobile adds platform cost with no user benefit | Responsive web dashboard; works on tablet if needed |
| **Player-facing portal** | Different user persona, different security model, different product | Out of scope; could be separate product later |
| **Transfer fee tracking** | Capology/Transfermarkt already do this well; not part of compensation calculation | Focus solely on salary + bonus computation |
| **Contract document management** | PDF storage/versioning is a different problem domain | Store contract *data* only (dates, amounts, clauses) |
| **Live notifications / alerts** | Batch tool; no real-time events to notify about | Post-MVP if workflow demands it (e.g., "calculation complete") |
| **Microservices architecture** | Premature for single-club, low-traffic tool | Monolith Go binary; separate only if load demands |
| **Full accounting integration** | Different domain; clubs have existing ERP/accounting systems | CSV export that accounting can import; no direct integration |

## Feature Dependencies

```
Player CRUD → Contract CRUD (contract belongs to player)
Player CRUD → Role assignment (determines which KPIs apply)
KPI Definition → KPI Scaling Formula (formula attached to each KPI)
Role assignment → KPI Definition (KPIs scoped to roles)
CSV Import → Player matching (imported stats must map to players)
CSV Import → Match Statistics storage (raw data before aggregation)
Match Statistics → KPI Aggregation (summarize per-player stats over period)
KPI Aggregation → Calculation Engine (aggregated values feed formulas)
Contract + KPI values → Calculation Engine (final salary computation)
Calculation Engine → History (store each run)
Calculation Engine → Dashboard (display results)
Calculation Engine → CSV Export (output reports)
Simulation Mode → Calculation Engine (reuses engine with hypothetical inputs)
Fairness Rules → Calculation Engine (applied as post-processing constraints)
```

## MVP Recommendation

Prioritize:

1. **Player CRUD** — foundational entity
2. **Contract CRUD with base salary** — defines compensation structure
3. **KPI Definition with role mapping** — core differentiation from spreadsheet
4. **KPI Scaling Formula** — makes calculation non-trivial
5. **CSV Import** — data ingestion without API providers
6. **Calculation Engine** — core value: deterministic, auditable salary computation
7. **REST API** — enables frontend and future integrations
8. **Basic Dashboard** — minimum viable visualization
9. **Calculation History** — audit trail for transparency
10. **CSV Export** — finance team needs output

Defer:

- **Simulation mode**: High value but high complexity; add in Phase 2 once core engine is stable
- **Fairness rules engine**: Important but can start with simple max_bonus_ratio cap; expand later
- **Multi-season comparison**: Requires historical data; build after first season runs
- **Team KPIs**: Add after individual KPIs work correctly
- **Pro-rata calculation**: Edge case for mid-season; not critical for first calculation runs

## Sources

- Capology (capology.com) — observed feature set: salaries, extensions, payrolls, transfer windows per club/league/season
- Transfermarkt (transfermarkt.com) — observed feature set: player profiles, contracts, market values, statistics, transfers
- Football Benchmark (footballbenchmark.com) — football business advisory/intelligence; confirms demand for financial transparency tools
- GitHub football-analytics topic — 506 repos; mostly analytics/prediction, none focused on internal club compensation calculation
- PROJECT.md — existing requirements and constraints for FootballIncentiveLab
