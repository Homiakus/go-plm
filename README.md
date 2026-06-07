# go-plm / Markdown-PLM v2.0

**Local Git-native PLM/PDM system for engineering documentation.**

Every engineering object (Part, Assembly, Drawing, Work Instruction, Release) is a **Markdown file with YAML frontmatter** on disk. Git tracks history. Finite state machines manage lifecycles. SQLite provides a rebuildable search index.

```
No server. No database. No cloud.
Just files. Just Git. Just Markdown.
```

---

## Philosophy

| Principle | Description |
|-----------|-------------|
| **Files are Source of Truth** | All data in `.md` files with YAML frontmatter. No hidden state. No proprietary formats. |
| **Git-native** | Every change is a commit. Every release is a tag. No `git` knowledge required. |
| **FSM-driven** | Object lifecycles managed by configurable finite state machines with guard validation. |
| **Index is Rebuildable** | SQLite index can be deleted and rebuilt from Markdown files at any time. |
| **Desktop-first** | Wails v2 + React + Vditor. Single binary. No Docker, no server. |

---

## Quick Start

### Prerequisites

- [Go](https://go.dev/dl/) 1.26+
- [Node.js](https://nodejs.org/) 20+ (for frontend)
- Windows 10+ / macOS 13+ / Linux (WebKitGTK)

### Build backend

```bash
git clone https://github.com/Homiakus/go-plm.git
cd go-plm

# Build
make build

# Run tests (22 packages, all passing)
make test

# Lint
make lint
```

### Build with frontend (requires Wails + npm)

```bash
make frontend-install
make frontend-build
make build
```

---

## Architecture

```
cmd/plm/              # Wails v2 entry point

internal/
├── core/             # Domain model (zero deps)
│   ├── object/       Object, ID, Class, State
│   ├── relation/     Relation, RelationType (13 types)
│   ├── artifact/     Artifact, Kind (8 kinds), Role
│   ├── event/        Event, EventType (11 types)
│   ├── diagnostic/   Diagnostic, Severity (4 levels)
│   ├── revision/     Version, ParseVersion, Bump
│   ├── lifecycle/    Machine, Definition, Transition
│   └── project/      Config, DefaultConfig
│
├── store/            # Persistence
│   ├── transaction/  Atomic file operations (Plan, Execute, Recover)
│   ├── fsrepo/       Filesystem repository (CRUD + frontmatter)
│   └── index/        SQLite index + FTS5 search
│
├── parser/           # Format parsing
│   ├── frontmatter/  YAML ↔ Markdown split/join
│   ├── yamlx/        Typed YAML decode/encode
│   └── markdown/     Outline extraction, link detection
│
├── naming/           # Naming Standard v10
│   Parse, Generate, Validate, SequenceStore
│
├── validation/       # Rule-based validation
│   ├── engine/       Rule, Registry, Engine
│   └── rules/        ValidName, RequiredMetadata
│
├── process/          # FSM engine
│   ├── fsm/          YAML-configurable state machine
│   ├── guards/       Guard implementations (6 guards)
│   └── effects/      Effect implementations (4 effects)
│
├── gitops/           # Git integration (go-git v5)
│   Status, Checkpoint, Tag, Log, Diff
│
├── modules/          # Domain modules
│   ├── bom/          BOM construction (structured/flat), cycle detection
│   ├── release/      Release scope, readiness check, manifest
│   └── standardparts/ Normalized key, duplicate detection (5 levels)
│
├── app/              # Application layer
│   ├── command/      CreateObject, RunTransition, AttachArtifact
│   ├── query/        GetObject, SearchObjects, GetBOM, GetTree
│   └── ports/        Repository interfaces
│
└── api/              # Frontend contract
    ├── dto/           Data Transfer Objects (13 types)
    ├── mapper/        Domain ↔ DTO conversion
    └── tree/          Icon, StatusColor, ThumbnailURL helpers

frontend/src/         # React + TypeScript + Vditor
```



## Project Structure on Disk

```
MyProject/
├── project.md                    # Project config (YAML frontmatter)
├── objects/
│   └── demo-prt-0001-v1.0/
│       ├── demo-prt-0001-v1.0.md # Object: YAML frontmatter + Markdown body
│       ├── history.jsonl         # Append-only event log
│       └── files/                # Attachments
│           ├── cad/              # .step, .stp
│           ├── drawings/         # .pdf, .dxf
│           ├── manufacturing/    # .dxf, .nc, .gcode
│           ├── images/           # .png, .jpg
│           ├── certificates/     # .pdf, .md
│           └── evidence/         # .png, .jpg, .pdf
├── config/                       # Project configuration (.md with YAML frontmatter)
│   ├── lifecycle.md
│   ├── classes.md
│   └── validation.md
├── templates/                    # Object templates (.md)
├── .plm/                         # Internal (derived data)
│   └── index.db                  # SQLite index (rebuildable)
└── .git/
```



## Object Lifecycle

```
draft ──submit_review──▶ in_review ──approve──▶ approved ──release──▶ released
  ▲                         │                                      │
  │────── reject ───────────┘                                      │
  │                                                                 │
  └──────────────────── revise ─────────────────────────────────────┘

blocked ←── block ── (any state)
blocked ── unblock ──▶ draft
released ── obsolete ──▶ obsolete ── archive ──▶ archived
```

Each transition validates guard conditions: `ValidName`, `RequiredMetadata`, `NoBlockingIssues`, `ChecksumsActual`, `NoReleaseBlockers`.

## Object Classes

| Class | Name | Type |
|-------|------|------|
| `prt` | Part | Mechanical component |
| `asm` | Assembly | Contains parts/sub-assemblies |
| `drw` | Drawing | Technical drawing |
| `doc` | Document | Specification, report |
| `std` | Standard Part | Fastener, bearing, resistor... |
| `mat` | Material | Raw material |
| `cut` | Cutting File | Laser/plasma cutting |
| `bnd` | Bending File | Press brake bending |
| `nc`  | NC Program | CNC machining |
| `ins` | Inspection | Quality control plan |
| `tpc` | Tech Process Card | Manufacturing route |
| `wi`  | Work Instruction | Assembly steps |
| `bom` | Formal BOM | Bill of Materials |
| `rel` | Release Package | Frozen release |
| `cr`  | Change Request | Engineering change |

## Naming Standard v10

```
[project]-[class]-[sequence]-v[major].[minor]
```

Examples:
- `demo-prt-0001-v1.0` — Part 0001
- `a320-asm-0100-v1.0` — Assembly 0100
- `a320-std-fst-iso4017-m6x30-a2-v1.0` — Standard part

## Key Features

| Feature | Status | Description |
|---------|--------|-------------|
| **Domain Model** | ✅ Done | 12 entity types with invariants |
| **Atomic Transactions** | ✅ Done | Write/Copy/Delete/Rename with rollback |
| **SQLite Index** | ✅ Done | FTS5 search, rebuildable |
| **YAML Frontmatter** | ✅ Done | Split/join, validation |
| **Naming Engine** | ✅ Done | Parse/Generate/Validate v10 |
| **FSM Lifecycle** | ✅ Done | YAML-configurable, guards + effects |
| **BOM Engine** | ✅ Done | Structured/flat, cycle detection |
| **Standard Parts** | ✅ Done | 11 classes, 5 duplicate levels |
| **Release Package** | ✅ Done | Scope, readiness, manifest |
| **Git Integration** | ✅ Done | Status, checkpoint, tag |
| **Media Attachments** | ✅ Done | 6 artifact kinds, auto-routing |
| **Project Tree** | ✅ Spec | Icons, status dots, thumbnails |
| **Desktop UI** | 🔜 Next | Wails v2 + React + Vditor |
| **Shop Floor** | 📋 Planned | Build records, QR tokens |
| **Procurement** | 📋 Planned | Buy list, ERP export |

## Development

```bash
# Build
make build

# Run all tests
make test

# Test with coverage
make test-cover

# Lint
make lint

# Benchmark
make bench

# Frontend
make frontend-install
make frontend-dev
make frontend-build
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.26 |
| Desktop Shell | Wails v2 |
| Frontend | React 19 + TypeScript + Vite + Tailwind |
| Editor | Vditor (WYSIWYG + instant render) |
| Database | SQLite (modernc.org/sqlite, no CGO) |
| Search | FTS5 |
| Git | go-git v5 |
| YAML | gopkg.in/yaml.v3 |
| IPC | HTTP JSON-RPC on localhost |
| Tests | Go test + testify |
| CI | GitHub Actions |

## Concurrency

Production-grade parallelization applied:
- **errgroup fan-out**: SaveObject → Index+History in parallel
- **Batch INSERT**: 500 rows/batch in RebuildIndex
- **Worker pools**: 8 goroutines for ListObjects, CheckReadiness, GenerateManifest
- **Ordered validation**: Deterministic rule execution order

See `AUDIT - Parallelization.md` for full analysis.

## Documentation

| Document | Description |
|----------|-------------|
| `SRS - go-plm - Complete.md` | Full Software Requirements Specification (39 sections) |
| `GitHub Epics and Issues.md` | 13 epics, 51 issues |
| `AUDIT - go-plm.md` | Initial documentation audit |
| `AUDIT - Parallelization.md` | Concurrency optimization audit |
| `docs/source-specs/` | Original specification drafts |

## License

Apache 2.0 — see [LICENSE](LICENSE)
