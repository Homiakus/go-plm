# Markdown-PLM v2.0

**Local Engineering Document Management** — a desktop PLM/PDM system where every engineering object is a Markdown file with YAML frontmatter, tracked by Git, validated by finite state machines.

```
┌──────────────────────────────────────────────────────┐
│  Editor (Live Preview)  │  Preview Panel             │
│  # Part 0001            │  ┌─────────────────────┐   │
│  **Material:** Alu      │  │ Part 0001            │   │
│  See [datasheet]        │  │ **Material:** Alu    │   │
│                         │  │ See datasheet        │   │
├──────────────────────────┤  └─────────────────────┘   │
│  Project Tree            │  Properties                 │
│  ● demo-asm-0001         │  ● draft  [PRT]             │
│    ● demo-prt-0001       │  Class: prt                 │
│      📐 Drawing          │  Title: Bracket Motor       │
│    ● demo-prt-0002       │  Material: Aluminum 6061    │
│  [Filter...]   (5)       │  [Submit] [Block]           │
└──────────────────────────────────────────────────────┘
```

## Philosophy

- **Files are source of truth** — no database, no hidden state. Everything is `.md` + `.yaml` + `.json` on disk.
- **Git-native** — every change is a commit, every release is a tag.
- **FSM-driven** — object lifecycle, document states, releases, and shop-floor builds are finite state machines with guard validation.
- **Desktop-first** — Wails 2 + React + Vditor, runs as a single binary.
- **Markdown-everywhere** — YAML frontmatter for metadata, Markdown body for documentation.

## Quick Start

### Prerequisites
- [Go](https://go.dev/dl/) 1.26+
- [Node.js](https://nodejs.org/) 20+
- Windows 10+ / macOS 13+ / Linux (WebView2/WebKitGTK)

### Build & Run

```powershell
# Clone
git clone https://github.com/your-org/markdown-plm.git
cd markdown-plm

# Install frontend dependencies
cd frontend && npm install && cd ..

# Build & run (interactive menu)
.\run.ps1
# [1] Install dependencies
# [3] Build all
# [5] Run desktop
```

### First Project

1. Launch the app
2. Enter a folder path (e.g. `C:\projects\MyProject`) — it will be auto-created
3. Click **+ New** to create your first part
4. Edit the Markdown body with live preview
5. Use the FSM triggers to progress through draft → review → approved → released

## Architecture

```
cmd/plm/                  # Wails 3 desktop entry point
├── desktop.go            # Core: OpenProject, FSM, API dispatcher
├── services_*.go         # Domain services (10 files by concern)
├── assets.go             # Embedded frontend

internal/
├── runtime/              # Project session, snapshots, CRUD
├── domain/               # ObjectRecord, ProjectConfig, validation types
├── fsm/                  # 6 finite state machines
├── naming/               # Naming Standard v10.0 parser/validator
├── parser/               # YAML frontmatter, Markdown blocks
├── validation/           # Object validation, release readiness
├── manufacturing/        # Completeness matrix, process plans
├── instructions/         # Work instructions, tech cards
├── procurement/          # Buy lists, substitutions
├── standardparts/        # Standard parts library
├── bom/                  # Bill of Materials
├── shopfloor/            # Build records, QR tokens
├── export/               # Release export, Typst PDF
├── gitops/               # Git status, diff, commit
├── history/              # Append-only JSONL event log
├── files/                # Checksums, safe copy, routing
├── api/                  # HTTP JSON-RPC bridge
├── recent/               # Recent projects list
├── folderpicker/         # Cross-platform folder picker
└── migration/            # v1 → v2 project layout

frontend/src/
├── app/                  # App shell, LandingPage, hooks
├── features/             # 12 feature components
├── editor/               # Vditor, Live Preview, markdown renderer
└── bindings/             # API types, HTTP bridge
```

## Key Features

| Feature | Description |
|---------|-------------|
| **Live Preview Editor** | Vditor with WYSIWYG / instant render / split preview modes. |
| **Project Tree** | Search/filter, keyboard nav (↑↓←→), sort (tree/alpha/class/status), status dots, indent guides |
| **3-Step Create Wizard** | Identity → Properties → Documents & Thumbnail, with live preview card |
| **Collapsible Properties** | Sections: Identity, Relations, Attachments, Metadata. Inline editing on click. |
| **FSM Lifecycle** | Draft → Review → Approved → Released. Guard-validated transitions with history log. |
| **Manufacturing Matrix** | Object/operation/artifact readiness with color-coded status and blockers |
| **Work Instructions** | Step-by-step instructions with checks, evidence capture, safety warnings |
| **Release Package** | "Why Can't I Release" assistant, manifest generation, export |
| **Shop Floor Player** | QR token resolution, step-by-step build with checks/evidence |
| **Git Integration** | Status, diff, commit within the app |
| **Standard Parts Center** | Library with duplicate detection |
| **Procurement Buy List** | BOM-to-buy-list with substitution workflow |
| **Change Impact Radar** | What breaks if this object changes |

## Naming Standard v10.0

All objects follow the naming convention:

```
[project]-[class]-[identifier]-v[major].[minor][-status].[ext]
```

Examples:
- `demo-prt-0001-v1.0.md` — Part 0001, version 1.0
- `a320-asm-0100-v2.0.step` — Assembly 0100, version 2.0
- `std-fst-iso4017-m8x30-v1.0.md` — Standard part

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

Each transition validates guard conditions: `ValidName`, `RequiredMetadataFilled`, `NoBlockingIssues`, `ChecksumsActual`, `NoReleaseBlockers`.

## Project Structure on Disk

```
MyProject/
├── project.yaml          # Project config (code, naming, object types, lifecycle)
├── objects/
│   └── demo-prt-0001-v1.0/
│       ├── item.json     # Full object record
│       ├── demo-prt-0001-v1.0.md  # YAML frontmatter + Markdown body
│       ├── history.log   # Append-only event log (JSONL)
│       ├── files/        # Attachments (STEP, DXF, PDF, images)
│       └── generated/    # Auto-generated documents (drawings, BOM)
├── config/               # Naming policy, file routing rules
├── templates/            # Markdown templates for generated docs
└── exports/              # Release package exports
```

## Development

```powershell
.\run.ps1          # Interactive menu
.\run.ps1          # [1] Install deps  [3] Build all  [5] Run desktop
.\run.ps1          # [7] Watch frontend (auto-rebuild on change)
.\run.ps1          # [8] Run tests
```

### Tech Stack

| Layer | Technology |
|-------|-----------|
| Desktop Shell | [Wails v2](https://wails.io/) |
| Frontend | React 19, TypeScript, Vite, Tailwind CSS |
| Editor | Vditor with WYSIWYG + instant render + split preview |
| Backend | Go 1.26, go-git/v5, yaml.v3 |
| State Machines | Custom FSM engine (7 state machines) |
| IPC | HTTP JSON-RPC on localhost |
| Versioning | Git (embedded via go-git/v5) |

## License

Apache 2.0
