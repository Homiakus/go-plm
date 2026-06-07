# CODEBASE MAP — go-plm v2.0

## Project Type

Desktop-first локальная PLM/PDM система. Один Go-бинарник со встроенным React-фронтендом.
Инженерная документация хранится в Markdown+YAML файлах на диске.
Git — native история. SQLite — пересобираемый поисковый индекс. FSM — жизненные циклы.

**Стек:** Go 1.26 + React 19 + TypeScript + Vite + Tailwind CSS + SQLite (FTS5) + go-git v5
**Бинарник:** 26MB (Go binary + embedded frontend dist/)
**Тесты:** 33/33 пакетов OK

---

## Directory Map

```
go-plm/                              # Module root (github.com/Homiakus/go-plm)
├── .github/workflows/               # [supporting] CI/CD (GitHub Actions)
├── .serena/                         # [generated] IDE tool cache (gitignored)
├── build/                           # [generated] Build output (plm.exe)
├── cmd/plm/                         # [critical] Entry point
│   ├── main.go                      #   CLI + server orchestration
│   └── server.go                    #   HTTP server, JSON-RPC dispatcher, embedded frontend
├── demo/                            # [generated] Test project (gitignored)
├── docs/                            # [supporting] Architecture docs, source specs
├── examples/                        # [supporting] Demo project template
├── frontend/                        # [critical] React SPA (15 components)
│   ├── embed.go                     #   go:embed all:dist → embedded into Go binary
│   ├── index.html                   #   Vite entry HTML
│   ├── package.json                 #   npm dependencies
│   ├── vite.config.ts               #   Vite configuration
│   ├── tsconfig.json                #   TypeScript config (strict)
│   ├── tailwind.config.js           #   Tailwind with PLM color palette
│   ├── postcss.config.js            #   PostCSS pipeline
│   ├── dist/                        #   [generated] Vite production build
│   └── src/                         #   Source
│       ├── main.tsx                 #     React entry
│       ├── App.tsx                  #     Root component (layout, routing, state)
│       ├── types.ts                 #     TypeScript types (DTOs, enums, constants)
│       ├── utils.ts                 #     Tailwind-safe class name helpers
│       ├── index.css                #     Tailwind imports + custom styles
│       ├── bindings/api.ts          #     JSON-RPC 2.0 client ↔ Go backend
│       └── components/              #     15 UI components
│           ├── LandingPage.tsx      #       Open/Create/Recent projects
│           ├── Sidebar.tsx          #       Project tree + right-click menu
│           ├── ContextMenu.tsx      #       Right-click context menu builder
│           ├── Toolbar.tsx          #       Nav, search, dark mode toggle
│           ├── ObjectEditor.tsx     #       Object detail: Info, Metadata, Lifecycle, WhereUsed
│           ├── MarkdownEditor.tsx   #       YAML+Markdown editor, autosave 5s, Ctrl+S
│           ├── CreateWizard.tsx     #       3-step Create Object wizard
│           ├── BOMViewer.tsx        #       Structured/flat BOM + CSV export
│           ├── SearchPanel.tsx      #       FTS5 full-text search
│           ├── ValidationPanel.tsx  #       Diagnostics by severity + suggested actions
│           ├── ReleaseDashboard.tsx #       Readiness check + Explain Blockers + Create Release
│           ├── CommandPalette.tsx   #       Ctrl+K command interface
│           ├── SafeDeleteDialog.tsx  #       Delete with where-used impact analysis
│           ├── StandardPartsCenter.tsx #    Browse/search standard parts by class
│           ├── GitChangesPanel.tsx  #       Git status + create checkpoint
│           └── StatusBar.tsx        #       Git status, index stats, file path
├── internal/                        # [critical] Go backend — Clean Architecture
│   ├── api/                         #   [important] Frontend contract
│   │   ├── dto/dto.go               #     Data Transfer Objects (13 types)
│   │   ├── mapper/mapper.go         #     Domain ↔ DTO conversion
│   │   ├── tree/tree.go             #     Icon, StatusColor, ThumbnailURL helpers
│   │   └── wails/api.go             #     33 JSON-RPC 2.0 methods
│   ├── app/                         #   [critical] Application layer (CQS)
│   │   ├── command/command.go       #     All write operations
│   │   ├── query/query.go           #     All read operations
│   │   ├── ports/ports.go           #     Canonical interfaces
│   │   └── service/service.go       #     DI container: Open(), InitProject()
│   ├── core/                        #   [critical] Domain model — zero deps
│   │   ├── object/object.go         #     Object, ID, Class (16), State (7), RelationRef, ArtifactRef
│   │   ├── relation/relation.go     #     Relation, Type (13), ValidateShape
│   │   ├── artifact/artifact.go     #     Artifact, Kind (8), Status (3)
│   │   ├── event/event.go           #     Event, Type (11), New()
│   │   ├── diagnostic/diagnostic.go #     Diagnostic, Severity (4)
│   │   ├── revision/revision.go     #     Version, ParseVersion, Bump
│   │   └── project/project.go       #     Config, DefaultConfig
│   ├── store/                       #   [critical] Persistence
│   │   ├── transaction/transaction.go #  Atomic file ops (Plan, Execute, Recover)
│   │   ├── fsrepo/fsrepo.go         #   Filesystem repository (CRUD + frontmatter)
│   │   └── index/index.go           #   SQLite + FTS5 index
│   ├── parser/                      #   [important] Format parsing
│   │   ├── frontmatter/frontmatter.go # YAML ↔ Markdown split/join
│   │   ├── yamlx/yamlx.go           #   Typed YAML decode/encode
│   │   └── markdown/markdown.go     #   Outline extraction, link detection
│   ├── naming/naming.go             #   [important] Naming Standard v10
│   ├── validation/                  #   [important] Rule-based validation
│   │   ├── engine/engine.go         #     Rule interface, Registry, Engine
│   │   └── rules/rules.go           #     6 rules: name, metadata, relations, artifacts, checksums, lifecycle
│   ├── process/                     #   [important] FSM engine
│   │   ├── fsm/fsm.go               #     YAML-configurable state machine
│   │   ├── guards/guards.go         #     8 guard implementations
│   │   └── effects/effects.go       #     4 effect implementations
│   ├── gitops/gitops.go             #   [important] Git integration (go-git v5)
│   └── modules/                     #   [important] Domain modules
│       ├── bom/bom.go               #     BOM construction (structured/flat), cycle detection
│       ├── release/release.go       #     Release scope, readiness, manifest
│       └── standardparts/standardparts.go # Normalized key, duplicate detection (5 levels)
├── pkg/plmclient/                   # [legacy] Empty — future public client library
├── templates/                       # [supporting] Object/project/release templates
├── go.mod                           # [critical] Go module definition
├── go.sum                           # [generated] Go dependency checksums
├── Makefile                         # [important] Build, test, lint, CI targets
├── README.md                        # [important] Project documentation
├── LICENSE                          # [supporting] Apache 2.0
├── .gitignore                       # [important] Git exclusion rules
├── SRS - go-plm - Complete.md       # [important] Software Requirements (39 sections)
├── GitHub Epics and Issues.md       # [supporting] 13 epics, 51 issues
├── AUDIT - go-plm.md                # [supporting] Initial documentation audit
├── AUDIT - Parallelization.md       # [supporting] Concurrency optimization audit
├── AUDIT - Full Codebase 2026-06-07.md # [supporting] Code quality/security audit
└── AUDIT - SRS Compliance 2026-06-07.md # [supporting] SRS compliance cross-reference
```

## Entry Points

| File | Type | Purpose | Priority |
|------|------|---------|----------|
| `cmd/plm/main.go:main()` | CLI | `plm init/stats/rebuild/help` or starts HTTP server | Critical |
| `cmd/plm/server.go:NewServer().Start()` | HTTP Server | Embedded frontend + JSON-RPC API on `127.0.0.1:8470` | Critical |
| `frontend/src/main.tsx` | Web UI | React 19 root render into `#root` | Critical |
| `frontend/embed.go` | Build | `//go:embed all:dist` — embeds Vite output into Go binary | Critical |
| `Makefile:build` | Build | `frontend-build` → `go build` → single binary | Important |

## Core Execution Flow

```
plm demo/
  └─ main() → runServer("demo")
       ├─ NewServer("demo")
       │    ├─ service.Open("demo")
       │    │    ├─ loadConfig("demo/project.md") → project.Config
       │    │    ├─ transaction.NewManager() + recover incomplete txs
       │    │    ├─ fsrepo.New() + index.Open("demo/.plm/index.db")
       │    │    ├─ naming.NewGenerator() + fsm.NewFromYAML("config/lifecycle.md")
       │    │    ├─ gitops.Open("demo/.git") + bom.New() + release.New()
       │    │    └─ return App{...}
       │    └─ wailsapi.NewAPI(app)
       ├─ Start(8470)
       │    ├─ http.Handle("/api", handleJSONRPC)  ← POST JSON-RPC 2.0
       │    ├─ http.Handle("/", embedded frontend)  ← GET → React SPA
       │    ├─ net.Listen("127.0.0.1:8470")
       │    └─ openBrowser("http://127.0.0.1:8470")
       └─ select {} // block forever

Browser → http://127.0.0.1:8470
  └─ GET / → index.html → React SPA loads
       ├─ App.tsx: loadProject()
       │    ├─ POST /api → JSON-RPC "ProjectInfo" → Go: project.Config
       │    ├─ POST /api → JSON-RPC "GetTree" → Go: query.GetTree → fsrepo.ListObjects → SQLite index
       │    └─ POST /api → JSON-RPC "ListObjects" → Go: fsrepo.ListObjects
       └─ Render: Sidebar + Toolbar + WelcomeScreen/Editor/BOM/Search
```

## Architecture Layers

| Layer | Location | Responsibility | Depends on | Should NOT depend on |
|-------|----------|----------------|------------|---------------------|
| **Entry Point** | `cmd/plm/` | CLI parsing, HTTP server, browser launch | `internal/app/service`, `internal/api/wails` | — |
| **API Contract** | `internal/api/` | JSON-RPC 2.0, DTOs, domain↔DTO mapping | `internal/app/`, `internal/core/` | UI, store |
| **Application** | `internal/app/` | Command/Query handlers, DI wiring | `internal/core/`, `internal/store/`, `internal/process/`, `internal/modules/` | UI |
| **Domain Modules** | `internal/modules/` | BOM, Release, Standard Parts logic | `internal/core/` | Store, UI, API |
| **Process** | `internal/process/` | FSM engine, guards, effects | `internal/core/`, `internal/naming/` | Store, API |
| **Validation** | `internal/validation/` | Rule engine, 6 rules | `internal/core/`, `internal/naming/` | Store, API |
| **Store** | `internal/store/` | File I/O, SQLite index, transactions | `internal/core/` | UI, API |
| **Parser** | `internal/parser/` | YAML↔Markdown, outline extraction | — | — |
| **Naming** | `internal/naming/` | ID parse/generate/validate v10 | — | — |
| **Core Domain** | `internal/core/` | Pure domain types, zero deps | stdlib only | Everything else |
| **GitOps** | `internal/gitops/` | Git status/commit/tag (go-git v5) | — | — |

**Dependency direction:** Strictly inward — `cmd` → `api` → `app` → `{core, modules, store, process}`. Чистая Clean Architecture без нарушений.

## Key Modules

| Module | Location | Responsibility | Change Safety |
|--------|----------|----------------|---------------|
| **Object** | `core/object/` | Core entity: ID, Class (16), State (7), RelationRef, ArtifactRef | **Low** — everything depends on this |
| **Relation** | `core/relation/` | Relation, Type (13), inverse, BOM membership | **Low** — BOM, release depend on this |
| **Artifact** | `core/artifact/` | Artifact file metadata, Kind (8), Status (3) | **Medium** |
| **Event** | `core/event/` | Immutable event record, Type (11) | **Low** — history integrity |
| **FSRepo** | `store/fsrepo/` | Filesystem CRUD, YAML frontmatter split/join | **Medium** — format changes break data |
| **Index** | `store/index/` | SQLite+FTS5, Upsert/Delete/Search/Rebuild | **Medium** — schema migrations needed |
| **Transaction** | `store/transaction/` | Atomic file operations, temp+rename, recovery | **Low** — data integrity critical |
| **FSM** | `process/fsm/` | YAML-configurable state machine | **Medium** — transition definitions |
| **Guards** | `process/guards/` | 8 precondition checks for transitions | **Medium** — lifecycle integrity |
| **Naming** | `naming/` | ID parse/generate/validate v10 | **Low** — all IDs depend on this |
| **Validation** | `validation/` | 6 validation rules, registry, engine | **Medium** |
| **BOM** | `modules/bom/` | Structured/flat BOM, cycle detection | **Medium** — core engineering data |
| **Release** | `modules/release/` | Scope collection, readiness check, manifest | **Medium** |
| **StandardParts** | `modules/standardparts/` | Normalize, duplicate detection (5 levels) | **Medium** |
| **GitOps** | `gitops/` | Git status, checkpoint, tag | **High** — wraps go-git v5 |
| **Command** | `app/command/` | All write operations (11 commands) | **Medium** — orchestration layer |
| **Query** | `app/query/` | All read operations (8 queries) | **High** — read-only, safe |
| **Service** | `app/service/` | DI container, Open/InitProject | **Medium** — wiring |
| **Wails API** | `api/wails/` | 33 JSON-RPC 2.0 methods | **Medium** — frontend contract |

## Data Model

### Domain Entities

| Entity | Go Type | Storage | Fields |
|--------|---------|---------|--------|
| **Project** | `project.Config` | `project.md` (YAML frontmatter) | code, title, naming, sequences, modules, storage, git |
| **Object** | `object.Object` | `objects/[id]/[id].md` (YAML frontmatter) | ID, Project, Class, Sequence, Version, Revision, State, Title, Metadata, Relations, Artifacts |
| **Relation** | `object.RelationRef` | Stored in object frontmatter | ToID, Type, Quantity, Unit |
| **Artifact** | `object.ArtifactRef` | Stored in object frontmatter | ID, Kind, Role, Path, OriginalName, Checksum, SizeBytes, Generated, Required, Status |
| **Event** | `event.Event` | `objects/[id]/history.jsonl` (append-only JSONL) | ID, Time, Actor, Type, ObjectID, Payload |
| **Diagnostic** | `diagnostic.Diagnostic` | Transient (generated by validation engine) | ID, Severity, Code, ObjectID, Path, Message, SuggestedActions |
| **Lifecycle** | `fsm.Definition` | `config/lifecycle.md` (YAML frontmatter) | Name, Initial, States, Transitions |

### Data Flows

```
CREATE OBJECT:
  Frontend (CreateWizard) → JSON-RPC "CreateObject"
    → API.CreateObject → command.CreateObject
      → naming.NextID() → new ID
      → fsrepo.SaveObject() → write .md file (transaction layer)
      → index.UpsertObject() → SQLite INSERT (parallel)
      → fsrepo.AppendHistory() → JSONL append (parallel)
      → [if parent] → command.AddRelation()

RUN TRANSITION:
  Frontend (ObjectEditor) → JSON-RPC "RunTransition"
    → API.RunTransition → command.RunTransition
      → fsrepo.GetObject() → read object
      → fsm.Can() → validate transition
      → fsm.Next() → new state
      → fsrepo.SaveObject() → write updated .md
      → index.UpsertObject() + AppendHistory() (parallel)
      → [if release] → gitops.Checkpoint() + CreateTag()

BOM QUERY:
  Frontend (BOMViewer) → JSON-RPC "GetBOM"
    → API.GetBOM → query.GetBOM
      → bom.GetStructured() → walkBOM()
        → RelationReader.ListRelations() → "contains" links
        → ObjectReader.GetObject() → child metadata
        → recurse for assemblies
      → return []BOMRow with levels, positions, quantities

SEARCH:
  Frontend (SearchPanel) → JSON-RPC "SearchObjects"
    → API.SearchObjects → query.SearchObjects
      → index.SearchObjects() → FTS5 MATCH (sanitized)
      → return []object.ID → GetObject() each → SearchResult[]
```

## Configuration

| Config Key | Defined In | Used In | Required | Default |
|-----------|-----------|---------|----------|---------|
| `project.code` | `project.md` (YAML) | `naming.Generator`, all IDs | Yes | (from init) |
| `project.title` | `project.md` (YAML) | UI title | Yes | (from init) |
| `naming.standard` | `project.md` (YAML) | Validation | Yes | "v10" |
| `naming.pattern` | `project.md` (YAML) | ID generation | Yes | `[project]-[class]-[sequence]-v[major].[minor]` |
| `naming.project_code` | `project.md` (YAML) | ID prefix | Yes | (from init) |
| `sequences.*` | `project.md` (YAML) | `naming.NextID()` | Yes | prt:0, asm:99, etc. |
| `storage.source_of_truth` | `project.md` (YAML) | Validation | Yes | "markdown_yaml" |
| `storage.index` | `project.md` (YAML) | `index.Open()` | Yes | ".plm/index.db" |
| `git.enabled` | `project.md` (YAML) | `service.InitProject()` | No | true |
| `git.release_tags` | `project.md` (YAML) | `command.RunTransition` | No | true |
| `lifecycle.*` | `config/lifecycle.md` (YAML) | `fsm.NewFromYAML()` | No | StandardObjectLifecycle |
| `classes.*` | `config/classes.md` (YAML) | Reference only | No | (generated) |
| `validation.rules` | `config/validation.md` (YAML) | Reference only | No | (generated) |
| Port | CLI flag / hardcoded | `cmd/plm/server.go` | No | 8470 |
| Dark mode | `localStorage` | `Toolbar.tsx` | No | false |

## External Dependencies

### Go (go.mod)
| Package | Purpose | Critical? |
|---------|---------|-----------|
| `gopkg.in/yaml.v3` | YAML marshal/unmarshal | Yes |
| `modernc.org/sqlite` | SQLite (pure Go, no CGO) | Yes |
| `github.com/go-git/go-git/v5` | Git operations | Yes |
| `golang.org/x/sync` | errgroup for fan-out | Yes |
| `github.com/google/uuid` | (indirect) UUID generation | No |

### npm (package.json)
| Package | Purpose |
|---------|---------|
| `react`, `react-dom` ^19 | UI framework |
| `lucide-react` | Icons (16 engineering icons) |
| `tailwindcss` ^3.4 | Utility CSS |
| `vite` ^6 | Build tool |
| `typescript` ^5.7 | Type checking |

## Test Coverage

| Status | Packages |
|--------|----------|
| **100%** | artifact, diagnostic, event, object, project, naming, markdown, rules |
| **90%+** | revision (95%), frontmatter (96%), bom (91%), release (91%) |
| **80%+** | fsm (84%), yamlx (86%), fsrepo (82%) |
| **70%+** | standardparts (70%), index (73%), transaction (71%) |
| **65%+** | engine (65%), relation (67%) |
| **Integration only** | api/dto, api/mapper, api/tree, api/wails, app/command, app/query, app/ports, app/service |
| **0% (no test files)** | cmd/plm, frontend (embed.go) |

**Test maturity: Medium.** Core domain is well-covered. Application and API layers rely on integration tests in `internal/app/`. No dedicated tests for HTTP server, JSON-RPC handler, or CLI commands.
