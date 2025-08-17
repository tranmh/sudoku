# Sudoku (Go) — High-Level Design\n\n**Owner:** Minh Cuong Tran  \n**Location:** `C:\\Users\\tranm\\work\\svw.info\\sudoku`  \n**Doc:** `docs/high-level-design.md`\n\n## 1. Overview\nA fast, testable Sudoku web application written in Go. It provides a Web UI and supports solving, generating (with difficulty levels), validating, and hinting. The design favors clean architecture for long-term maintainability and performance (≤ 1s for solve and generate).\n\n## 2. Goals\n- Web UI playable Sudoku with instant interactions.\n- High-performance solver and generator (≤ 1 second on typical hardware).\n- Validations and hint system integrated with the solver.\n- JSON save/load with seedable RNG for reproducible puzzles.\n- Clean architecture to isolate domain logic from I/O/adapters.\n- Comprehensive testing (unit, property, fuzz, benchmarks).\n- Human-readable logging; cross-compilation for Windows/Linux.\n\n## 3. Non-Goals\n- Multiplayer, accounts, or cloud sync.\n- Mobile native apps (PWA could be future work).\n- Internationalization (initially English-only UI).\n\n## 4. Key Decisions (Summary)\n- **UI:** Web (served by Go); minimal JS; responsive layout.\n- **Architecture:** Clean architecture (domain/usecase/adapters/infrastructure).\n- **Solver:** Algorithm X (Dancing Links) for best performance; fallback to backtracking+MRV for debugging.\n- **Generator:** Generate full solution, then remove clues while enforcing uniqueness via fast solver checks; difficulty graded by solving metrics.\n- **Persistence:** JSON serialization for puzzles, solutions, and metadata; seedable RNG.\n- **Testing:** Unit, property-based (rapid), fuzz (`go test -fuzz`), benchmarks.\n- **Logging:** `log/slog` text handler (human-readable) with levels.\n- **Perf Target:** Solve & generate each ≤ 1s.\n## 5. Architecture &amp; Packages\n**Layers (Clean Architecture):**\n- **domain** (pure logic): board, cell, candidates, constraints, errors\n- **usecase**: SolvePuzzle, GeneratePuzzle, ValidatePuzzle, GetHint, Save, Load\n- **adapters**: http handlers, views/templates, json dto mappers\n- **infrastructure**: storage (fs), rng, logging, config\n\n**Suggested packages:**\n- `/internal/domain` (board, cell, bitset, puzzle, difficulty)\n- `/internal/solver` (dlx, backtracking, heuristics, validator)\n- `/internal/generator` (builder, clue-removal, grader)\n- `/internal/usecase` (orchestrators implementing ports)\n- `/internal/ports` (interfaces for solver/generator/storage/logger)\n- `/internal/adapters/http` (chi router, handlers, middleware)\n- `/internal/adapters/view` (html/template, static assets)\n- `/internal/infrastructure` (fs storage, rng, slog)\n- `/cmd/sudoku-web` (main)\n\n## 6. Domain Model (Core)\n- **Board**: 9×9 grid; cells hold `value` (0..9), `fixed` flag, `candidates` (uint16 bitset)\n- **Puzzle**: `{ id, seed, difficulty, board, createdAt, elapsedNanos }`\n- **Move/Hint**: next action suggestion with rationale and affected cells\n- **JSON schema (sketch):**\n```json\n{\n  \"id\": \"uuid\",\n  \"seed\": 123456,\n  \"difficulty\": \"easy|medium|hard|expert\",\n  \"board\": [[0,0,3,...],[...]],\n  \"fixed\": [[true,false,...],[...]]\n}\n```\n\n## 7. Algorithms (Performance-first)\n- **Solver (default):** Algorithm X with Dancing Links (DLX) for exact cover; typically &lt;1ms for standard puzzles; \n  fallback: backtracking+MRV for debug/validation.\n- **Generator:** create random full solution via DLX; remove clues while ensuring uniqueness by re-solving; \n  grade difficulty using metrics (search nodes, forced moves, strategy tiers); cap attempts to meet ≤1s.\n- **Validator:** fast row/col/box checks; optional uniqueness verify via one extra DLX run.\n- **Hints:** derive next logical step (single candidate/position, naked/hidden pairs; extensible).\n\n## 8. Web UI &amp; API\n- **Server-rendered UI** with `html/template` + light JS (fetch) for actions; responsive CSS (no heavy tooling).\n- **Router:** `github.com/go-chi/chi`.\n- **Endpoints:** `GET /` (UI), `POST /api/solve`, `/api/generate?difficulty=...`, `/api/validate`, `/api/hint`, `/api/save`, `/api/load`.\n- **Static:** embed templates/assets via `embed`.\n## 9. Performance Plan\n- Targets: solve ≤1s, generate ≤1s (single puzzle) on typical desktop; 99th percentile tracked.\n- Use `testing.B` benchmarks for solver/generator; record ops, allocations, ns/op.\n- Instrument critical paths with counters/timers; expose simple `/debug/vars` (expvar) optionally.\n- Concurrency: generator may run checks in limited worker pool; cap by CPU count minus 1.\n- Memory: use bitset candidates (uint16) and pooled nodes to reduce GC churn.\n\n## 10. Testing &amp; QA\n- Unit tests for domain (board ops, constraints), solver, generator, validator, hints.\n- Property-based tests (e.g., pgregory.net/rapid): round-trip generate→solve→validate always true.\n- Fuzz tests: `go test -fuzz=Fuzz` for JSON load/save and API handlers.\n- Golden tests for JSON schema compatibility and hint explanations.\n- Benchmarks: `BenchmarkSolve_*`, `BenchmarkGenerate_*` across difficulties.\n- Optional E2E (chromedp) for basic UI flows.\n\n## 11. Logging &amp; Observability\n- `log/slog` text handler; levels: DEBUG/INFO/WARN/ERROR; request-id middleware.\n- Structured context keys: module, op, duration, nodes, seed, difficulty.\n- Optional Prometheus metrics (future): ops/sec, durations, errors.\n\n## 12. Configuration\n- Hierarchy: flags (highest) &gt; env &gt; defaults; persisted user prefs in JSON (UI only).\n- Key flags: `--addr`, `--log-level`, `--seed`, `--difficulty`, `--persist-path`.\n- Config struct in `/internal/infrastructure/config` with validation.\n\n## 13. Build &amp; Cross-Compilation\n- Go 1.22+; pure Go (cgo off) for static binaries.\n- Windows: `GOOS=windows GOARCH=amd64 go build ./cmd/sudoku-web`.\n- Linux: `GOOS=linux GOARCH=amd64 go build ./cmd/sudoku-web`.\n- Makefile targets: `make build`, `make test`, `make bench`, `make cross`.\n\n## 14. Project Layout (initial)\n```\n/cmd/sudoku-web\n/internal/domain | /solver | /generator | /usecase | /ports\n/internal/adapters/http | /adapters/view | /infrastructure\n/web/static | /web/templates | /docs\n```\n\n## 15. Risks &amp; Mitigations\n- Generator timeouts → cap attempts, degrade difficulty first; fall back to easier puzzle.\n- Solver corner cases → cross-validate with two algorithms (DLX vs backtracking) in CI.\n- UI latency → keep server-rendered HTML + light JS; defer heavy logic to backend.\n\n## 16. Open Questions (Decisions Needed)\n- UI stack: keep server-rendered Go templates + light JS, or prefer SPA/htmx?\n- Difficulty grading thresholds: what node/strategy budgets distinguish levels?\n- Default persist path for saves: `%AppData%/sudoku` or project-relative `./data`?\n- Max CPU usage for generation (percent/cores)?\n- Browser support baseline (Chrome/Edge latest ok?), and A11y priorities?## 17. Finalized Design Decisions (2025-08-16)\n- UI stack: Go templates + small vanilla JS (no HTMX/SPA).\n- Difficulty grading: hybrid (search-node+forced-move AND strategy-tier); label uses stricter result.\n- Default save path: project-relative `./data` next to the binary; ensure directory on startup.\n- CPU usage: `runtime.NumCPU()-1` worker pool cap for generation/validation; opt-out via flag.\n- Browser &amp; a11y: Latest 2 Chrome/Edge/Firefox; keyboard-first controls; ARIA roles; high-contrast &amp; dark modes.\n- UI details: classic 9×9 grid with bold 3×3 separators; inputs via mouse+on-screen numpad AND full keyboard; hints include simple+advanced with “max strategy” setting.\n\n## 18. UI Interaction Flows (brief)\n- New puzzle: user picks difficulty → backend `/api/generate` with seed (optional) → render board.\n- Solve: client sends current grid to `/api/solve` → returns solution + timing + nodes.\n- Validate: `/api/validate` returns conflicts + uniqueness flag (optional).\n- Hint: `/api/hint?maxStrategy=tier` returns next step, affected cells, explanation.\n- Save/Load: JSON round-trip to `./data/{id}.json`; list endpoint to enumerate saves.\n\n## 19. Configuration &amp; Flags (update)\n- `--max-strategy=singles|pairs|advanced|xwing` (default: `advanced`)\n- `--workers` (default: `NumCPU()-1`)\n- `--persist-path` (default: `./data`)\n- `--browser-baseline=stable` (doc-only; impacts UI features and polyfills)\n- `--addr`, `--log-level`, `--seed`, `--difficulty` remain as defined.\n\n## 20. Acceptance Criteria (MVP)\n- Solve and generate each ≤ 1s (p95) on a reference machine; unit and benchmark tests demonstrate.\n- JSON save/load compatible with documented schema; fuzz tests for handler stability.\n- Hints respect `max-strategy`; validator detects all row/col/box conflicts.\n- Cross-compiled binaries for Windows/Linux build in CI; no cgo.\n\n## 21. Next Steps\n- Scaffold repo (folders, go.mod, Makefile, cmd, internal packages, templates, static).\n- Implement DLX solver + backtracking (debug) with common interface; wire difficulty-grading hybrid.\n- Implement generator with uniqueness enforcement + time cap; integrate worker pool.\n- Add initial UI (templates + vanilla JS) with keyboard shortcuts + accessibility roles.\n- Set up tests/benchmarks + CI.## Status &amp; Open Decisions (2025-08-17)

**Current status (MVP):**
- Web UI (vanilla JS + Go templates), endpoints for generate/solve/validate/hint/save/load/list.  
- Backtracking solver + uniqueness checker; generator with seeded RNG and uniqueness preservation; fast validator; singles-only hinter.  
- JSON persistence; human-readable logging (slog text); Makefile incl. cross-compile; initial unit tests (solver + generator).

**Decisions requested (please reply with choices):**
1) **Hint roadmap**  
   A) Singles only (keep) · B) Singles + pairs (naked/hidden) · C) Add pointing/claiming &amp; triples next  
2) **Hint UI behavior**  
   A) Highlight + message (keep) · B) Add “auto-fill singles” action · C) Step-by-step walkthrough mode  
3) **Notes / candidates (pencil marks)**  
   A) Later · B) Manual notes now · C) Auto-candidates toggle  
4) **Solver plan** (perf &amp; robustness)  
   A) Keep Backtracking for now · B) Implement DLX/Exact-Cover next and keep Backtracking as fallback  
5) **Generator quality**  
   A) Keep target-givens heuristic (40/34/28/24) · B) Add strategy-based grading to map puzzles to difficulty  
6) **Request logging**  
   A) Keep app logs only · B) Add HTTP request logging middleware with latency &amp; status (human-readable)  
7) **CI/CD**  
   A) Defer · B) Add GitHub Actions: lint, unit tests, cross-build artifacts  
8) **Persistence UX**  
   A) Keep ID-only · B) Add optional `name`/`notes` metadata to saved puzzles (JSON)

**Next steps once you choose:**
- Implement selected hint strategies + UI wiring.  
- (If 4B) Add DLX solver and route via feature flag.  
- (If 6B) Add lightweight request logger middleware.  
- Expand tests: benchmarks (solve/generate &lt;1s), API handlers, and property tests for validator.## Decisions Locked (2025-08-17)

1) **Hint roadmap**: **A + B + C**  
   - Implement pipeline: Singles → Pairs (naked/hidden) → Pointing/Claiming → Triples.  
   - Exposed via `maxTier` to cap strategies for UX/testing.

2) **Hint UI**: **A + B + C**  
   - Keep highlight + message, add **Auto-fill Singles** action, and a **step-by-step walkthrough** mode (applies one hint at a time).

3) **Candidates / Notes**: **B + C**  
   - Manual notes (user-entered pencil marks) **and** an **Auto-candidates toggle** (computed).

4) **Solver plan (choose better performance)**:  
   - **DLX / Exact-Cover as primary**, Backtracking as fallback (feature flag). Target solve &lt;= 1s.

5) **Generator quality (choose better performance)**:  
   - **Keep target-givens heuristic** (40/34/28/24) + seeded RNG + uniqueness check via DLX.  
   - Strategy-based grading can be added later behind a flag (trades performance for quality).

6) **Request logging**: **A + B**  
   - Keep app logs (slog text) **and** add HTTP request logging middleware (status, latency, duration), human-readable.

7) **CI/CD**: **No CI/CD** for now.

8) **Persistence UX**: **B**  
   - Add optional `name` and `notes` fields to saved puzzles (JSON + API + storage + UI).

9) **1s budget enforcement**: **A (soft)**  
   - Best-effort; use per-operation timeboxing internally, but no hard cancellation at HTTP boundary.

10) **New puzzle controls**: **B**  
   - One “New” button using **last-used difficulty** with a dropdown to change it.

11) **Hint hotkeys**: **B**  
   - `?` = hint, `F` = auto-fill singles.

---

### Implementation Plan (incremental)

**S1. Solver performance**
- Add `internal/solver/dlx.go` (Algorithm X + Dancing Links) with `Solve`/`Unique` conforming to `ports.Solver`.
- Wire selection: default **DLX**, fallback **Backtracking** (flag/env).

**S2. Generator**
- Keep current randomized fill + carve with uniqueness check (seedable). Ensure target givens per difficulty; expose last-used difficulty.

**S3. Hint engine**
- Extend `internal/hint` with pairs, pointing/claiming, triples; compose in a tiered pipeline.
- Add walkthrough mode: apply next hint and persist history for back/forward steps.

**S4. UI/UX**
- Replace four difficulty buttons with **one “New” + dropdown** (stores last-used difficulty).
- Add hotkeys `?` (hint) and `F` (auto-fill singles).
- Add manual pencil marks UI and **Auto-candidates** toggle.

**S5. Persistence & API**
- Extend `domain.Puzzle{ Name, Notes }`; update storage and `/api/save|load|list`; show `name` in list dialog.

**S6. Observability**
- Add HTTP request logging middleware (method, path, status, latency), human-readable.

**S7. Tests & perf**
- Benchmarks: solve/generate under 1s (seeded).  
- Unit tests for new hint strategies and API handlers.  
- Property tests for validator (no duplicates allowed).