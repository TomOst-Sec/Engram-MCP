# Colony Roadmap

> Maintained by ATLAS. Updated every 30-minute cycle.
> Last updated: 2026-03-15 18:00

## Current Milestones
- **M2:** Nearly complete (12/14 done, TASK-023 active with merge conflict, TASK-035 active)
- **M3:** In progress (4/8 done, 1 active, 3 queued)
- **M4:** Starting (1 task queued)

## Milestone Status

| Milestone | Status | Tasks | Done | Active | Queued |
|-----------|--------|-------|------|--------|--------|
| M1: MVP | **COMPLETE** | 17 | 17 | 0 | 0 |
| M2: Core Features | Nearly Complete | 16 | 14 | 2 | 0 |
| M3: Polish & Growth | In Progress | 8 | 4 | 1 | 3 |
| M4: Ecosystem | Starting | 1 | 0 | 0 | 1 |

Note: M2 includes TASK-036/037 (bug fixes by beta-tester) and TASK-038 (export/import).

## Active Tasks
- TASK-023: --watch mode (alpha) — BUG-023: go.mod merge conflict, needs rebase
- TASK-035: Multi-repo support (alpha) — in progress

## All M2+ Tasks

| Task | Title | Milestone | Assigned | Status |
|------|-------|-----------|----------|--------|
| TASK-018..022, 024..031 | M2 tasks | M2 | mixed | done |
| TASK-036 | FTS5 Build Tag Fix | M2 | alpha | done |
| TASK-037 | Register Missing Parsers | M2 | bravo | done |
| TASK-023 | --watch Mode | M2 | alpha | active (BUG-023) |
| TASK-038 | Export + Import CLI | M2 | bravo | queued |
| TASK-032..034 | M3 first batch | M3 | mixed | done |
| TASK-035 | Multi-Repo Support | M3 | alpha | active |
| TASK-039 | Community Convention Modules | M3 | bravo | queued |
| TASK-040 | Benchmark Suite | M3 | alpha | queued |
| TASK-041 | TUI Conventions+Architecture | M3 | alpha | queued |
| TASK-042 | GoReleaser + Packaging | M3 | alpha | queued |
| TASK-043 | Docker Image | M3 | bravo | queued |
| TASK-044 | CI/CD Memory Hook | M4 | alpha | queued |

## Bugs
- BUG-023: TASK-023 go.mod merge conflict — coder needs to rebase on main

## Queue (7 tasks ready)
- TASK-038: Export/Import CLI (bravo) — deps met
- TASK-039: Community Conventions (bravo) — deps met
- TASK-040: Benchmark Suite (alpha) — deps met
- TASK-041: TUI Panels (alpha) — deps met
- TASK-042: GoReleaser (alpha) — deps met
- TASK-043: Docker Image (bravo) — deps met
- TASK-044: CI/CD Memory Hook (alpha) — deps met

## Engram Feature Coverage

| Feature | GOALS.md | Status |
|---------|----------|--------|
| 1. MCP Server Core | M1 | ✅ done |
| 2. search_code Tool | M1 | ✅ done |
| 3. Tree-Sitter Indexer | M1 | ✅ 15 languages |
| 4. ONNX Embeddings | M1 | ✅ done |
| 5. get_architecture Tool | M1 | ✅ done |
| 6. remember Tool | M1 | ✅ done |
| 7. recall Tool | M1 | ✅ done |
| 8. get_conventions Tool | M2 | ✅ done |
| 9. SQLite Storage | M1 | ✅ done |
| 10. Git History Analyzer | M2 | ✅ done |
| 11. --watch Mode | M2 | 🔧 active (bug) |
| 12. npx engram init | M2 | ✅ done |
| 13. Full CLI | M2 | ✅ done (lipgloss + export/import queued) |
| 14. HTTP/SSE Transport | M3 | ✅ done |
| 15. Convention Prompts | M2 | ✅ done |
| 16. TUI Dashboard | M3 | ✅ foundation done, panels queued |
| 17. Multi-Repo Support | M3 | 🔧 active |
| 18. Ollama Integration | M3 | ✅ done |
| 19. Community Conventions | M3 | queued |
| 20. CI/CD Memory Hook | M4 | queued |

## Velocity
- Total: 35 tasks done in ~3 hours
- Average: ~5 min/task (including review)
- Colony running at exceptional velocity
