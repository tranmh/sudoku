# Sudoku (Go) — Web UI

This repository contains a Sudoku web application written in Go using a clean architecture. It includes a solver (DLX planned), generator, validator, and hint system (to be implemented).

## Quick Start
```bash
make tidy
make build
make run # starts on :8080
```
Open http://localhost:8080

## Cross Compilation
```bash
make cross
# outputs in ./bin for Windows/Linux (amd64), CGO disabled
```

## Project Layout
```
/cmd/sudoku-web        # entrypoint (web server)
/internal              # domain, ports, solver, generator, usecases, adapters
/web/templates         # Go HTML templates
/web/static            # JS/CSS (embedded)
/docs                  # design docs
```

## Design Document
See `docs/high-level-design.md` for goals, architecture, algorithms, and acceptance criteria.

## Status
Scaffold created; HTTP endpoints and core algorithms are stubbed and will be implemented next.
