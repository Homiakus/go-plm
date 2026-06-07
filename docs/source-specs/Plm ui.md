# UI Specification

# Markdown-PLM / go-plm

## Подробное описание пользовательского интерфейса, сценариев и связи с backend

Версия: `0.1`  
Статус: `draft`  
Целевая платформа: Desktop, Wails + React + TypeScript  
Backend: Go application layer + filesystem store + YAML/Markdown source of truth + SQLite index + Git

---

# 1. Общая концепция UI

Интерфейс Markdown-PLM должен быть построен вокруг инженерного рабочего процесса:

```text
Открыть проект
→ найти или создать объект
→ заполнить инженерные данные
→ связать объект с другими объектами
→ проверить готовность
→ провести через lifecycle
→ собрать BOM / release / instruction / buy list
→ зафиксировать checkpoint в Git
```

UI не должен выглядеть как файловый менеджер или Git-клиент. Пользователь работает с инженерными сущностями:

```text
Part
Assembly
Drawing
Document
Work Instruction
Process Plan
Release Package
Change Request
Build Record
Standard Part
```

Файлы, YAML, Git, индекс и транзакции — это backend-механизмы, которые пользователь видит только через понятные действия:

```text
Save
Validate
Submit for Review
Approve
Release
Create Checkpoint
Show Changes
Explain Blockers
Export Package
```

---

# 2. Основной layout приложения

Главное окно делится на 5 зон:

```text
┌────────────────────────────────────────────────────────────────────────────┐
│ 1. Top Bar: Project | Search | Command Palette | Git | User | Settings     │
├───────────────┬──────────────────────────────────────┬─────────────────────┤
│ 2. Left Rail  │ 3. Main Workspace                    │ 4. Right Inspector  │
│ Navigation    │ Editor / BOM / Matrix / Graph / etc. │ Properties / State  │
│ Project Tree  │                                      │ Validation / Files  │
├───────────────┴──────────────────────────────────────┴─────────────────────┤
│ 5. Bottom Bar: Status | Tasks | Diagnostics | Git | Index | Autosave        │
└────────────────────────────────────────────────────────────────────────────┘
```

## 2.1. Top Bar

Назначение: глобальные действия, текущий проект, поиск, командная палитра, состояние Git и системные настройки.

Элементы слева направо:

```text
[App Logo]
[Project Switcher]
[Current Project Code]
[Global Search]
[Command Palette Button]
[Create Button]
[Git Status Button]
[Tasks Button]
[Settings Button]
```

### 2.1.1. App Logo Button

Текст/иконка:

```text
Markdown-PLM
```

Действие:

- открывает стартовое меню приложения;
- показывает версию;
- ссылки на About, Documentation, Diagnostics.

Backend:

```text
Query: GetAppInfo
Query: GetRuntimeStatus
```

---

### 2.1.2. Project Switcher

Вид:

```text
[Demo Project ▼]
```

Открывает dropdown:

```text
Open Project...
Create New Project...
Recent Projects
────────────────
demo — C:\projects\demo
robot-arm — D:\plm\robot-arm
settings-panel — C:\work\settings-panel
────────────────
Project Settings
Close Project
```

Кнопки и пункты:

#### Open Project...

Открывает folder picker.

Backend:

```text
Command: OpenProject(path)
Backend steps:
1. Проверить наличие project.yaml
2. Загрузить конфигурацию
3. Проверить .git
4. Загрузить или пересобрать SQLite index
5. Просканировать objects/
6. Вернуть ProjectSnapshot
```

UI states:

```text
loading: "Opening project..."
success: open workspace
error: show ProjectOpenErrorDialog
```

#### Create New Project...

Открывает wizard создания проекта.

Backend:

```text
Command: CreateProject(CreateProjectRequest)
```

#### Recent Project item

Открывает выбранный проект.

Backend:

```text
Command: OpenProject(path)
Command: AddRecentProject(path)
```

#### Project Settings

Открывает экран настроек проекта.

Backend:

```text
Query: GetProjectConfig
Command: UpdateProjectConfig
```

#### Close Project

Закрывает текущий проект.

Backend:

```text
Command: CloseProject
```

---

### 2.1.3. Current Project Code Badge

Вид:

```text
DEMO
```

Показывает код проекта из `project.yaml`.

Tooltip:

```text
Project: Demo Engineering Project
Path: C:\projects\demo
Objects: 124
Git: clean
Index: up to date
```

Backend:

```text
Query: GetProjectSummary
Query: GetGitStatus
Query: GetIndexStatus
```

---

### 2.1.4. Global Search

Вид:

```text
[Search objects, BOM, files, commands...]
```

Поведение:

- фокус по `Ctrl+P`;
- поиск по ID, названию, классу, статусу, metadata, Markdown body, artifact names;
- результаты группируются.

Dropdown результатов:

```text
Objects
  PRT demo-prt-0001-v1.0 — Bracket Motor
  ASM demo-asm-0100-v1.0 — Motor Assembly

Documents
  DRW demo-drw-0001-v1.0 — Motor Bracket Drawing

Commands
  Create Part
  Validate Project

Files
  bracket.step
  bracket.pdf
```

Backend:

```text
Query: Search(SearchRequest)
Uses:
- SQLite FTS index
- object index
- artifact index
```

Кнопки в результате:

```text
[Open]
[Reveal in Tree]
[Show Where Used]
```

Backend:

```text
Query: GetObject
Query: GetWhereUsed
```

---

### 2.1.5. Command Palette Button

Вид:

```text
[⌘K]
```

Горячая клавиша:

```text
Ctrl+K
```

Открывает command palette.

Команды:

```text
Create Part
Create Assembly
Create Drawing
Create Work Instruction
Create Process Plan
Create Release Package
Create Change Request
Create Build Record
Add Child to BOM
Attach File
Validate Object
Validate Project
Submit for Review
Approve
Release
Block
Revise
Show Where Used
Show Change Impact
Generate BOM
Generate Buy List
Create Checkpoint
Show Git Diff
Rebuild Index
Open Project Settings
```

Каждая команда имеет:

```text
id
title
description
shortcut
enabled/disabled
reason_if_disabled
backend command/query
```

Backend:

```text
Query: ListAvailableCommands(context)
Command: ExecuteCommand(commandID, params)
```

Если команда недоступна, UI показывает:

```text
Release is disabled because object has 4 blockers.
[Explain blockers]
```

Backend:

```text
Query: ExplainCommandAvailability(commandID, context)
```

---

### 2.1.6. Create Button

Вид:

```text
[+ New ▼]
```

Dropdown:

```text
Part
Assembly
Drawing
Document
Work Instruction
Process Plan
Release Package
Change Request
Standard Part
Build Record
Folder / Group
From Template...
```

Каждый пункт открывает Create Wizard с выбранным типом.

Backend:

```text
Query: ListObjectClasses
Query: GetObjectTemplate(class)
Command: CreateObject
```

---

### 2.1.7. Git Status Button

Вид:

```text
[Git: Clean]
```

или

```text
[Git: 7 changes]
```

или

```text
[Git: Error]
```

Dropdown:

```text
Status: 7 changed files
────────────────
Modified objects:
  demo-prt-0001-v1.0
  demo-asm-0100-v1.0

Untracked files:
  files/cad/new-bracket.step

Actions:
[Show Changes]
[Create Checkpoint]
[Discard Selected...]
[Open Git Log]
[Compare with Last Release]
```

Backend:

```text
Query: GetGitStatus
Query: GetDomainDiff
Command: CreateCheckpoint
Command: DiscardChanges
Query: GetGitLog
Query: CompareWithRelease
```

Important UX rule:

Пользователь не видит `git add`, `git commit`, `git tag` как основные команды. Он видит инженерные действия:

```text
Create Checkpoint
Create Release Tag
Compare with Release
```

---

### 2.1.8. Tasks Button

Вид:

```text
[Tasks: 2]
```

Dropdown:

```text
Running
  Rebuilding index... 64%
  Generating release package...

Completed
  Project validation completed
  Checksum recalculation completed

Failed
  Export ERP BOM failed
```

Backend:

```text
Query: ListTasks
Command: CancelTask(taskID)
Query: GetTaskLog(taskID)
```

Задачи используются для:

```text
RebuildIndex
ValidateProject
GeneratePDF
ExportReleasePackage
RecalculateChecksums
ScanDuplicates
```

---

### 2.1.9. Settings Button

Вид:

```text
[⚙]
```

Dropdown:

```text
Project Settings
Application Settings
Templates
Lifecycle Config
Validation Rules
File Routing
Keyboard Shortcuts
Diagnostics
About
```

Backend:

```text
Query: GetProjectConfig
Query: GetAppSettings
Command: UpdateAppSettings
```

---

# 3. Left Rail / Navigation

Левая панель состоит из двух уровней:

```text
Primary Navigation Icons
Project Tree / Current View Content
```

## 3.1. Primary Navigation Icons

Вертикальный список:

```text
[🏠] Home
[🌲] Objects
[📦] BOM
[✅] Release
[🏭] Manufacturing
[🛒] Procurement
[🔁] Changes
[📚] Standard Parts
[🧾] History
[🔧] Admin
```

Каждая кнопка меняет главный режим левой панели и центрального workspace.

Backend:

```text
Query: GetNavigationState
Command: SetActiveView(viewID)
```

---

## 3.2. Objects / Project Tree View

Вид:

```text
Objects
[Filter objects...]
[View: Tree ▼] [Sort: Status ▼] [⚙]

▾ demo-asm-0100-v1.0  ● released
  ├─ demo-prt-0001-v1.0 ● approved
  ├─ demo-prt-0002-v1.0 ● draft
  └─ demo-drw-0001-v1.0 ● released

Ungrouped
  demo-doc-0003-v1.0 ● draft
```

### Кнопки панели Objects

#### Filter input

```text
[Filter objects...]
```

Фильтрует видимое дерево.

Backend:

```text
Query: FilterObjects(filter)
```

Для маленьких проектов может работать на frontend по уже загруженному snapshot.

#### View dropdown

```text
[View: Tree ▼]
```

Пункты:

```text
Tree
Flat List
By Class
By Status
By Release
By Owner
Recently Changed
Invalid Objects
```

Backend:

```text
Query: GetObjectProjection(projectionType)
```

#### Sort dropdown

```text
[Sort: Status ▼]
```

Пункты:

```text
Name
Class
Status
Revision
Last Modified
Validation Status
```

Backend:

```text
Query: ListObjects(sort, filter)
```

#### Settings gear

Открывает настройки отображения дерева:

```text
Show status dots
Show class badges
Show revision
Show validation badges
Show children count
Show file icons
Compact mode
```

Frontend-only, сохраняется в app settings.

Backend:

```text
Command: UpdateUserViewSettings
```

---

### Object Tree Item

Каждая строка содержит:

```text
[Expand/Collapse]
[Class Icon]
[Object ID or Title]
[Revision]
[State Dot]
[Validation Badge]
[Dirty Badge]
[Context Menu Button]
```

Пример:

```text
▾ 📦 Motor Assembly  v1.0  ● released  ✓
  ├─ 🔩 Bracket Motor v1.0 ● approved  ⚠2
```

Клик по строке:

```text
Open object in current workspace tab
```

Backend:

```text
Query: GetObject(objectID)
Query: GetObjectReadiness(objectID)
Query: GetObjectRelations(objectID)
```

Double click:

```text
Open default editor/view
```

Right click / context menu:

```text
Open
Open in New Tab
Reveal Files
Copy Object ID
Copy Markdown Link
Add Child...
Attach File...
Show Where Used
Show Change Impact
Validate
Submit for Review
Approve
Release
Block...
Revise...
Duplicate...
Archive...
Delete...
```

Backend mapping:

```text
Open                         Query: GetObject
Reveal Files                  Command: RevealInFileManager
Copy Object ID                Frontend clipboard
Copy Markdown Link            Frontend clipboard
Add Child                     Command: AddRelation
Attach File                   Command: AttachFile
Show Where Used               Query: GetWhereUsed
Show Change Impact            Query: GetChangeImpact
Validate                      Query/Command: ValidateObject
Submit for Review             Command: RunTransition
Approve                       Command: RunTransition
Release                       Command: RunTransition
Block                         Command: RunTransition
Revise                        Command: ReviseObject
Duplicate                     Command: DuplicateObject
Archive                       Command: ArchiveObject
Delete                        Command: DeleteObject
```

---

# 4. Main Workspace

Центральная область работает как tabbed workspace.

Верхняя строка workspace:

```text
[Object Tab 1] [Object Tab 2] [BOM] [Release] [+]
```

Каждая вкладка имеет:

```text
title
dirty indicator
close button
context menu
```

Tab context menu:

```text
Close
Close Others
Close to Right
Pin Tab
Reveal in Tree
Copy Object ID
```

---

# 5. Object Detail Screen

Это основной экран инженерного объекта.

Layout:

```text
┌────────────────────────────────────────────────────────────┐
│ Object Header                                                │
├────────────────────────────────────────────────────────────┤
│ Tabs: Document | Metadata | BOM | Files | History | Diff     │
├────────────────────────────────────────────────────────────┤
│ Main content                                                 │
└────────────────────────────────────────────────────────────┘
```

---

## 5.1. Object Header

Пример:

```text
[PRT] Bracket Motor
demo-prt-0001-v1.0     State: draft     Validation: 2 blockers
[Submit Review] [Block] [Validate] [Checkpoint] [More ▼]
```

Элементы:

### Class Badge

```text
[PRT]
```

Backend:

```text
Object.class
Query: GetObjectClassDefinition
```

### Title

Editable inline.

Клик:

```text
turns into input
```

Buttons:

```text
[Save]
[Cancel]
```

Backend:

```text
Command: UpdateObjectMetadata(objectID, {title})
```

### Object ID

Read-only.

Buttons:

```text
[Copy]
[Reveal]
```

Backend:

```text
Command: RevealObjectFolder
```

### Revision Badge

```text
v1.0
```

Click opens revision menu:

```text
Create New Revision
Bump Minor
Bump Major
Compare Revisions
Revision History
```

Backend:

```text
Command: ReviseObject
Query: GetRevisionHistory
Query: CompareRevisions
```

### State Badge

```text
draft
in_review
approved
released
blocked
obsolete
archived
```

Click opens lifecycle panel.

Backend:

```text
Query: GetAvailableTransitions(objectID)
```

### Validation Summary

```text
✓ Ready
⚠ 2 warnings
⛔ 4 blockers
```

Click opens validation tab.

Backend:

```text
Query: ValidateObject
Query: GetObjectReadiness
```

### Main Action Buttons

Buttons are context-sensitive.

For `draft`:

```text
[Submit Review]
[Block]
[Validate]
```

For `in_review`:

```text
[Approve]
[Reject]
[Block]
[Validate]
```

For `approved`:

```text
[Release]
[Revise]
[Block]
[Validate]
```

For `released`:

```text
[Revise]
[Obsolete]
[Create Change Request]
[Compare]
```

Backend:

```text
Command: RunTransition(objectID, transition)
Query: ValidateTransition(objectID, transition)
```

If action disabled:

```text
[Release disabled]
Tooltip: "Cannot release: 4 blockers"
Button nearby: [Explain]
```

Backend:

```text
Query: ExplainTransitionBlockers(objectID, transition)
```

### More Dropdown

```text
Duplicate
Archive
Delete
Export Object
Copy Link
Open Folder
Show Raw Markdown
Show YAML Frontmatter
```

Backend:

```text
Command: DuplicateObject
Command: ArchiveObject
Command: DeleteObject
Command: ExportObject
Command: RevealObjectFolder
Query: GetRawMarkdown
Query: GetFrontmatter
```

---

# 6. Object Tabs

## 6.1. Document Tab

Главный редактор Markdown.

Layout:

```text
┌──────────────────────────────┬──────────────────────────────┐
│ Markdown Editor              │ Live Preview                 │
│ CodeMirror                   │ Rendered Markdown            │
└──────────────────────────────┴──────────────────────────────┘
```

Toolbar:

```text
[Save]
[Autosave: On]
[Preview: Split ▼]
[Insert ▼]
[Format ▼]
[Validate]
[Diff]
[Open Raw]
```

### Save Button

Сохраняет Markdown/frontmatter.

Backend:

```text
Command: UpdateObjectDocument(objectID, markdownContent)
Backend steps:
1. Parse frontmatter
2. Validate YAML syntax
3. Validate schema
4. Write temp file
5. Atomic rename
6. Append history event
7. Update index projection
8. Return updated ObjectSnapshot
```

### Autosave Toggle

```text
[Autosave: On/Off]
```

Backend:

```text
Command: UpdateUserPreference(autosave)
```

Autosave uses debounce:

```text
Frontend: wait 800-1500 ms after last edit
Command: UpdateObjectDocument
```

### Preview Mode Dropdown

```text
Split
Editor Only
Preview Only
Live Preview Inline
```

Frontend-only mostly.

Backend used only for server-side Markdown render if needed:

```text
Query: RenderMarkdown
```

### Insert Dropdown

```text
Object Link
BOM Table
Artifact Link
Image
Checklist
Inspection Block
Safety Warning
Frontmatter Field
```

Backend:

```text
Query: SearchObjects
Query: ListArtifacts
Query: GetTemplateSnippet
```

### Format Dropdown

```text
Heading
Table
Checklist
Code Block
Normalize Frontmatter
Sort YAML Keys
```

Backend:

```text
Command: NormalizeDocument
```

### Validate Button

Backend:

```text
Query: ValidateObject(objectID)
```

### Diff Button

Backend:

```text
Query: GetObjectDiff(objectID, base = lastCheckpoint)
```

### Open Raw Button

Shows raw Markdown in modal.

Backend:

```text
Query: GetRawMarkdown(objectID)
```

---

## 6.2. Metadata Tab

Редактирование структурированных данных из YAML frontmatter.

Layout:

```text
Identity
  ID: demo-prt-0001-v1.0 [copy]
  Class: prt
  Title: Bracket Motor
  Revision: 1.0
  State: draft

Engineering
  Material: Aluminum 6061
  Mass: 120 g
  Unit: pcs

Manufacturing
  Make/Buy: make
  Process Plan: demo-pp-0001-v1.0

Procurement
  Supplier:
  Manufacturer PN:

Custom Fields
  ...
```

Buttons:

```text
[Edit]
[Save]
[Cancel]
[Add Field]
[Remove Field]
[Reset to Template]
[Validate Metadata]
```

Backend:

```text
Query: GetObjectMetadataSchema(objectClass)
Command: UpdateObjectMetadata
Query: ValidateMetadata
Command: AddCustomField
Command: RemoveCustomField
```

UX rule:

Metadata fields are generated from `classes.yaml` and validation rules.

---

## 6.3. Relations Tab

Shows all relations of object.

Sections:

```text
Contains
Used In
Drawings
CAD Files
Documents
Process Plans
Work Instructions
Change Requests
Substitutions
Release Packages
```

Buttons:

```text
[Add Relation]
[Remove]
[Edit Quantity]
[Show Graph]
[Validate Relations]
```

### Add Relation Button

Opens relation picker.

Fields:

```text
Relation Type
Target Object
Quantity
Unit
Effectivity
Comment
```

Backend:

```text
Query: ListRelationTypes
Query: SearchObjects
Command: AddRelation
Query: ValidateRelation
```

### Remove Button

Backend:

```text
Command: RemoveRelation
```

Before remove, if relation affects BOM/release:

```text
Confirmation dialog with impact
```

Backend:

```text
Query: GetRelationImpact
```

---

## 6.4. BOM Tab

For assemblies and objects with `contains` relations.

Toolbar:

```text
[Add Item]
[Remove Item]
[Edit Qty]
[Import BOM]
[Export BOM]
[Flatten]
[Where Used]
[Validate BOM]
[Generate Buy List]
```

Table columns:

```text
Item
Object
Title
Class
Revision
State
Qty
Unit
Make/Buy
Validation
Substitutes
Actions
```

Row actions:

```text
Open
Edit Quantity
Replace
Show Where Used
Remove
```

Backend:

```text
Query: GetBOM(objectID, mode)
Command: AddBOMItem
Command: UpdateBOMItem
Command: RemoveBOMItem
Query: ValidateBOM
Command: ExportBOM
Command: GenerateBuyList
```

### Add Item

Dialog:

```text
Search existing object
or
Create new part
```

Backend:

```text
Query: SearchObjects
Command: AddRelation(type=contains)
Command: CreateObjectAndAddToBOM
```

### Replace

Use case: заменить деталь в сборке.

Backend:

```text
Command: ReplaceBOMItem
Backend:
1. Validate replacement
2. Create relation update
3. Append event
4. Update index
5. Return impact
```

---

## 6.5. Files Tab

Shows artifacts.

Toolbar:

```text
[Attach File]
[Attach Folder]
[Generate]
[Recalculate Checksums]
[Open Folder]
[Validate Files]
```

Table columns:

```text
Icon
File
Role
Kind
Path
Checksum
Status
Last Modified
Actions
```

Actions:

```text
Open
Reveal
Rename
Change Role
Recalculate Checksum
Remove
```

Backend:

```text
Command: AttachFile
Command: AttachFolder
Command: OpenArtifact
Command: RevealArtifact
Command: RenameArtifact
Command: UpdateArtifactRole
Command: RecalculateChecksum
Command: RemoveArtifact
Query: ValidateArtifacts
```

### Attach File Button

Flow:

```text
1. User chooses file
2. UI asks role/kind
3. Backend copies file using file-routing.yaml
4. Backend calculates checksum
5. Backend updates frontmatter artifacts
6. Backend appends event
7. Backend updates index
```

---

## 6.6. Validation Tab

Shows diagnostics and readiness.

Sections:

```text
Summary
Blockers
Warnings
Passed Checks
Suggested Actions
```

Toolbar:

```text
[Run Validation]
[Explain]
[Fix Automatically]
[Export Report]
```

Diagnostic item:

```text
⛔ Required metadata missing: material
Location: frontmatter.metadata.material
Action: [Fill Material]
```

Backend:

```text
Query: ValidateObject
Query: ExplainDiagnostic
Command: ApplyQuickFix
Command: ExportValidationReport
```

### Fix Automatically

Only for safe fixes:

```text
normalize YAML
recalculate checksum
rebuild relations index
remove duplicate whitespace
sort frontmatter keys
```

Backend:

```text
Command: ApplyQuickFix(diagnosticID)
```

---

## 6.7. History Tab

Shows object event history.

Toolbar:

```text
[Refresh]
[Filter ▼]
[Compare Selected]
[Export History]
```

Events:

```text
2026-06-06 10:00 object.created by local-user
2026-06-06 10:05 metadata.updated
2026-06-06 10:10 lifecycle.transitioned draft → in_review
```

Row actions:

```text
Show Details
Show Diff
Copy Event ID
```

Backend:

```text
Query: GetObjectHistory
Query: GetEventDetails
Query: GetEventDiff
Command: ExportHistory
```

---

## 6.8. Diff Tab

Shows domain-aware diff.

Modes:

```text
File Diff
Object Diff
Metadata Diff
BOM Diff
Relation Diff
Markdown Diff
```

Toolbar:

```text
[Compare with: Last Checkpoint ▼]
[Refresh]
[Open Git Diff]
[Create Checkpoint]
```

Backend:

```text
Query: GetObjectDiff
Query: GetGitDiff
Command: CreateCheckpoint
```

---

# 7. Right Inspector

Правая панель всегда показывает контекст выбранного объекта, строки BOM, diagnostic, artifact или task.

Sections:

```text
Properties
Lifecycle
Readiness
Relations
Attachments
Actions
```

---

## 7.1. Properties Section

Collapsed/expanded.

Fields:

```text
ID
Class
Title
Revision
State
Material
Unit
Make/Buy
Owner
Last Modified
```

Buttons:

```text
[Edit]
[Copy ID]
[Open Folder]
```

Backend:

```text
Query: GetObjectSummary
Command: UpdateObjectMetadata
Command: RevealObjectFolder
```

---

## 7.2. Lifecycle Section

Shows state machine.

```text
draft → in_review → approved → released
```

Current state highlighted.

Buttons:

```text
[Submit Review]
[Approve]
[Reject]
[Release]
[Block]
[Unblock]
[Revise]
[Obsolete]
[Archive]
```

Visibility depends on current state and available transitions.

Backend:

```text
Query: GetAvailableTransitions
Query: ValidateTransition
Command: RunTransition
```

When user clicks transition:

```text
1. UI calls ValidateTransition
2. If blockers: show blockers dialog
3. If transition requires reason: show reason dialog
4. UI calls RunTransition
5. Backend writes YAML/Markdown, event, index
6. UI refreshes object snapshot
```

---

## 7.3. Readiness Section

Visual checklist:

```text
✓ Valid name
✓ Required metadata
✓ Relations valid
⚠ Drawing missing
⛔ Child object not released
✓ Checksums actual
```

Buttons:

```text
[Explain]
[Run Validation]
[Open Blockers]
```

Backend:

```text
Query: GetObjectReadiness
Query: ExplainReadiness
Query: ValidateObject
```

---

## 7.4. Relations Mini Section

Shows most important links:

```text
Parent assemblies: 2
Children: 5
Drawings: 1
Work instructions: 2
Open change requests: 1
```

Buttons:

```text
[Show Graph]
[Where Used]
[Add Relation]
```

Backend:

```text
Query: GetRelationSummary
Query: GetWhereUsed
Query: GetRelationGraph
Command: AddRelation
```

---

## 7.5. Actions Section

Contextual buttons:

```text
[Create Checkpoint]
[Create Change Request]
[Generate Drawing Placeholder]
[Generate Buy List]
[Export Object]
[Duplicate]
[Archive]
```

Backend:

```text
Command: CreateCheckpoint
Command: CreateChangeRequest
Command: GenerateArtifact
Command: GenerateBuyList
Command: ExportObject
Command: DuplicateObject
Command: ArchiveObject
```

---

# 8. Bottom Bar

Bottom bar shows system state.

Elements:

```text
Project: demo
Objects: 124
Selected: demo-prt-0001-v1.0
Autosave: saved
Validation: 2 blockers
Git: 7 changes
Index: up to date
Tasks: 1 running
```

Clickable elements:

### Autosave Status

States:

```text
Saved
Saving...
Unsaved changes
Save failed
```

Click opens save details.

Backend:

```text
Query: GetSaveState
Command: RetrySave
```

### Validation Status

Click opens diagnostics panel.

Backend:

```text
Query: GetProjectDiagnosticsSummary
```

### Git Status

Click opens Git panel.

Backend:

```text
Query: GetGitStatus
```

### Index Status

States:

```text
up to date
stale
rebuilding
error
```

Click opens index diagnostics.

Backend:

```text
Query: GetIndexStatus
Command: RebuildIndex
```

### Tasks

Click opens tasks panel.

Backend:

```text
Query: ListTasks
```

---

# 9. Landing Page / No Project Open

When no project is open, show:

```text
Markdown-PLM

[Create New Project]
[Open Existing Project]
[Open Recent]

Recent Projects:
- demo
- robot-arm
- settings-panel

[Documentation]
[Diagnostics]
[Settings]
```

## Create New Project Button

Opens Create Project Wizard.

## Open Existing Project Button

Opens folder picker.

## Open Recent

Calls:

```text
Command: OpenProject(path)
```

Backend:

```text
Query: ListRecentProjects
Command: OpenProject
Command: CreateProject
```

---

# 10. Create Project Wizard

Steps:

```text
1. Location
2. Project Identity
3. Template
4. Modules
5. Review & Create
```

---

## Step 1. Location

Fields:

```text
Project Folder
```

Buttons:

```text
[Browse]
[Next]
[Cancel]
```

Backend:

```text
Command: PickFolder
Query: ValidateProjectPath
```

Validation:

```text
folder does not exist → can create
folder exists and empty → ok
folder exists and has project.yaml → offer open
folder not writable → error
```

---

## Step 2. Project Identity

Fields:

```text
Project Code
Project Title
Description
Author
```

Buttons:

```text
[Back]
[Next]
[Cancel]
```

Backend:

```text
Query: ValidateProjectCode
```

---

## Step 3. Template

Cards:

```text
Simple Mechanical Project
Manufacturing Project
Electronics Project
Documentation-only Project
Blank Project
```

Each card:

```text
[Preview]
[Select]
```

Backend:

```text
Query: ListProjectTemplates
Query: GetProjectTemplatePreview
```

---

## Step 4. Modules

Checkboxes:

```text
[x] Objects
[x] BOM
[x] Release
[x] Git Integration
[ ] Manufacturing
[ ] Procurement
[ ] Shop Floor
[ ] Standard Parts
[ ] Change Management
```

Backend:

```text
Query: ListAvailableModules
```

---

## Step 5. Review & Create

Summary:

```text
Folder:
Project code:
Template:
Modules:
```

Buttons:

```text
[Create Project]
[Back]
[Cancel]
```

Backend:

```text
Command: CreateProject
Backend steps:
1. Create folder
2. Create project.yaml
3. Create config/*.yaml
4. Create templates
5. Create .plm/
6. Initialize Git
7. Create initial checkpoint
8. Build index
9. Return ProjectSnapshot
```

---

# 11. Create Object Wizard

Opened from `+ New`.

Steps:

```text
1. Type
2. Identity
3. Metadata
4. Relations
5. Files
6. Review & Create
```

---

## Step 1. Type

Cards:

```text
Part
Assembly
Drawing
Document
Work Instruction
Process Plan
Release Package
Change Request
Standard Part
Build Record
```

Buttons:

```text
[Next]
[Cancel]
```

Backend:

```text
Query: ListObjectClasses
Query: GetObjectClassDefinition
```

---

## Step 2. Identity

Fields:

```text
Title
Class
Identifier
Revision
Generated ID Preview
```

Buttons:

```text
[Generate ID]
[Check Name]
[Next]
[Back]
[Cancel]
```

Backend:

```text
Query: GenerateObjectID
Query: ValidateObjectID
```

---

## Step 3. Metadata

Dynamic form generated from class schema.

For Part:

```text
Material
Unit
Mass
Make/Buy
Description
```

For Assembly:

```text
Unit
Assembly Type
```

For Drawing:

```text
Related Object
Format
Drawing Type
```

Buttons:

```text
[Next]
[Back]
[Cancel]
```

Backend:

```text
Query: GetObjectMetadataSchema
Query: ValidateMetadataDraft
```

---

## Step 4. Relations

Fields:

```text
Parent Assembly
Related Drawing
Related CAD
Process Plan
Work Instruction
```

Buttons:

```text
[Add Relation]
[Remove]
[Next]
[Back]
```

Backend:

```text
Query: SearchObjects
Query: ValidateRelationDraft
```

---

## Step 5. Files

Buttons:

```text
[Attach File]
[Attach Folder]
[Skip]
```

File list shows:

```text
file
role
target path
checksum preview
```

Backend:

```text
Command: PickFiles
Query: PreviewFileRouting
```

---

## Step 6. Review & Create

Shows:

```text
Generated Object ID
Folder path
Markdown template preview
Metadata
Relations
Files
```

Buttons:

```text
[Create]
[Create and Open]
[Create and Add Another]
[Back]
[Cancel]
```

Backend:

```text
Command: CreateObject
Backend steps:
1. Validate object draft
2. Create object folder
3. Render Markdown template with YAML frontmatter
4. Copy files using routing rules
5. Calculate checksums
6. Append history event
7. Update index
8. Return ObjectSnapshot
```

---

# 12. BOM Screen

Global BOM screen opened from left navigation.

Top toolbar:

```text
[BOM Scope ▼]
[View: Structured/Flat ▼]
[Add Item]
[Import]
[Export]
[Validate]
[Generate Buy List]
[Compare BOM]
```

Scope dropdown:

```text
Current Object
Selected Assembly
Entire Project
Release Package
```

Backend:

```text
Query: GetBOM(scope)
```

Table columns:

```text
Level
Item
Object ID
Title
Class
Revision
State
Qty
Unit
Make/Buy
Supplier
Validation
Actions
```

Actions per row:

```text
Open
Edit Qty
Replace
Remove
Where Used
Impact
```

Backend:

```text
Command: UpdateBOMItem
Command: ReplaceBOMItem
Command: RemoveBOMItem
Query: GetWhereUsed
Query: GetChangeImpact
```

---

# 13. Release Screen

Opened from left navigation or object action.

Layout:

```text
Release Dashboard
Release Packages
Readiness
Blockers
Manifest
Exports
```

Toolbar:

```text
[New Release Package]
[Validate Release]
[Explain Blockers]
[Generate Manifest]
[Export Package]
[Create Release Tag]
```

Backend:

```text
Command: CreateReleasePackage
Query: ValidateRelease
Query: ExplainReleaseBlockers
Command: GenerateReleaseManifest
Command: ExportReleasePackage
Command: CreateReleaseTag
```

## Release Readiness Panel

Shows:

```text
Scope: demo-asm-0100-v1.0
Objects: 42
Ready: 36
Warnings: 3
Blockers: 6
```

Blocker cards:

```text
⛔ demo-prt-0004-v1.0 is draft
Action: [Open Object] [Submit Review]

⛔ Drawing missing for demo-prt-0008-v1.0
Action: [Create Drawing] [Attach Drawing]
```

Backend:

```text
Query: GetReleaseReadiness
Command: ApplyReleaseQuickFix
```

---

# 14. Manufacturing Screen

Opened from left navigation.

Tabs:

```text
Readiness Matrix
Process Plans
Work Instructions
Operations
Tools
Inspection
```

## Manufacturing Matrix

Toolbar:

```text
[Scope ▼]
[Refresh]
[Validate]
[Show Blockers Only]
[Export Matrix]
```

Matrix columns:

```text
Object
State
BOM
Drawing
CAD
Process Plan
Work Instruction
Tools
Inspection
Safety
Ready
Actions
```

Cell states:

```text
✓ ready
⚠ warning
⛔ blocker
— not applicable
```

Cell click opens details.

Backend:

```text
Query: GetManufacturingMatrix(scope)
Query: ExplainManufacturingCell(objectID, cellType)
Command: CreateProcessPlan
Command: CreateWorkInstruction
```

---

# 15. Procurement Screen

Tabs:

```text
Buy List
Suppliers
Substitutions
Standard Parts
ERP Export
```

Toolbar:

```text
[Generate Buy List]
[Validate Procurement]
[Find Duplicates]
[Export ERP]
[Create Substitution]
```

Backend:

```text
Command: GenerateBuyList
Query: ValidateProcurement
Query: FindStandardPartDuplicates
Command: ExportERP
Command: CreateSubstitutionRequest
```

Buy List columns:

```text
Part
Title
Qty
Unit
Make/Buy
Supplier
Manufacturer PN
Lead Time
Substitute
Status
Actions
```

Actions:

```text
Open
Set Supplier
Propose Substitute
Mark Purchased
Export Row
```

Backend:

```text
Command: UpdateProcurementMetadata
Command: ProposeSubstitution
```

---

# 16. Change Impact Screen

Opened from object action or left navigation.

Toolbar:

```text
[Select Object]
[Analyze Impact]
[Create Change Request]
[Export Report]
```

Main view:

```text
Impact Graph
Affected Assemblies
Affected Releases
Affected Work Instructions
Affected Buy Lists
Affected Build Records
Risks
```

Backend:

```text
Query: GetChangeImpact(objectID)
Command: CreateChangeRequestFromImpact
Command: ExportImpactReport
```

Graph node actions:

```text
Open
Expand
Collapse
Show Relation
Add to Change Request
```

Backend:

```text
Query: ExpandImpactNode
Command: AddAffectedObjectToChangeRequest
```

---

# 17. Standard Parts Screen

Toolbar:

```text
[New Standard Part]
[Import Library]
[Find Duplicates]
[Validate]
[Export]
```

Table columns:

```text
ID
Title
Standard
Material
Dimensions
Manufacturer
Manufacturer PN
State
Duplicate Score
Actions
```

Actions:

```text
Open
Use in BOM
Propose Substitute
Merge Duplicate
Archive
```

Backend:

```text
Command: CreateStandardPart
Command: ImportStandardParts
Query: FindStandardPartDuplicates
Command: AddRelation(type=substitutes)
Command: MergeStandardPartDuplicates
```

---

# 18. Git / Changes Screen

Opened from Git button or left navigation.

Tabs:

```text
Summary
Domain Diff
File Diff
History
Checkpoints
Releases
```

Toolbar:

```text
[Refresh]
[Create Checkpoint]
[Compare with Last Release]
[Discard Selected]
[Open Git Folder]
```

Summary:

```text
Changed Objects: 4
Changed Files: 9
Untracked Files: 2
Last Checkpoint: 2026-06-06 10:05
```

Backend:

```text
Query: GetGitStatus
Query: GetDomainDiff
Query: GetGitDiff
Command: CreateCheckpoint
Command: DiscardChanges
Query: GetGitLog
```

## Create Checkpoint Dialog

Fields:

```text
Message
Include changed objects
Include untracked files
```

Buttons:

```text
[Create Checkpoint]
[Cancel]
```

Backend:

```text
Command: CreateCheckpoint(message, include)
```

---

# 19. Admin / Settings Screen

Tabs:

```text
Project
Classes
Naming
Lifecycle
Validation
Templates
File Routing
Modules
Index
Diagnostics
Keyboard
```

## Project Tab

Fields:

```text
Project Code
Title
Description
Author
Default Revision
Git Enabled
Autosave
```

Buttons:

```text
[Save]
[Reset]
[Validate Config]
```

Backend:

```text
Query: GetProjectConfig
Command: UpdateProjectConfig
Query: ValidateProjectConfig
```

## Classes Tab

UI for `classes.yaml`.

Buttons:

```text
[Add Class]
[Edit Class]
[Delete Class]
[Duplicate Class]
[Export Classes]
```

Backend:

```text
Query: GetClassDefinitions
Command: UpdateClassDefinitions
```

## Lifecycle Tab

Visual FSM editor.

Buttons:

```text
[Add State]
[Add Transition]
[Edit Guards]
[Validate FSM]
[Save]
```

Backend:

```text
Query: GetLifecycleConfig
Command: UpdateLifecycleConfig
Query: ValidateLifecycleConfig
```

## Validation Tab

Rules list.

Buttons:

```text
[Add Rule]
[Edit Rule]
[Test Rule]
[Run All]
[Save]
```

Backend:

```text
Query: GetValidationConfig
Command: UpdateValidationConfig
Query: TestValidationRule
```

## Index Tab

Shows:

```text
Index status
Object count
Relation count
Last rebuild
Errors
```

Buttons:

```text
[Rebuild Index]
[Verify Index]
[Delete and Rebuild]
[Export Index Diagnostics]
```

Backend:

```text
Query: GetIndexStatus
Command: RebuildIndex
Command: VerifyIndex
Command: ExportIndexDiagnostics
```

---

# 20. Dialogs and Modals

## 20.1. Confirmation Dialog

Used for destructive actions.

Layout:

```text
Title
Description
Impact summary
[Cancel]
[Confirm]
```

Backend usually provides impact:

```text
Query: PreviewDeleteObject
Query: PreviewArchiveObject
Query: PreviewRemoveRelation
```

---

## 20.2. Blockers Dialog

Used when transition/action is not allowed.

Layout:

```text
Cannot Release Object

Blockers:
1. Required metadata missing: material
   [Fill Material]

2. Child object is not released
   [Open Child] [Submit Review]

3. Checksum outdated
   [Recalculate Checksum]

[Close]
[Run Validation Again]
```

Backend:

```text
Query: ExplainTransitionBlockers
Command: ApplyQuickFix
```

---

## 20.3. Reason Dialog

Used for transitions requiring reason:

```text
Block
Reject
Revise
Obsolete
Archive
```

Fields:

```text
Reason
Comment
Affected objects
```

Buttons:

```text
[Confirm]
[Cancel]
```

Backend:

```text
Command: RunTransition(objectID, transition, reason)
```

---

## 20.4. Error Dialog

Should always include:

```text
User-friendly message
Technical details collapsed
Affected operation
Suggested actions
Copy diagnostics
Open logs
```

Backend:

```text
Error response contains:
code
message
details
diagnostics
recoverable
suggested_actions
```

---

# 21. Backend Contract

All UI actions use application commands and queries.

## 21.1. Query pattern

Queries do not change source of truth.

Examples:

```text
GetObject
ListObjects
Search
ValidateObject
GetBOM
GetWhereUsed
GetChangeImpact
GetGitStatus
GetIndexStatus
```

## 21.2. Command pattern

Commands may change files, index, history or Git.

Examples:

```text
CreateObject
UpdateObjectDocument
UpdateObjectMetadata
AttachFile
AddRelation
RunTransition
CreateCheckpoint
ExportReleasePackage
RebuildIndex
```

## 21.3. Command response

Every command returns:

```yaml
ok: true
result:
  object_snapshot: ...
events:
  - event_id: evt_...
diagnostics: []
tasks: []
refresh:
  objects:
    - demo-prt-0001-v1.0
  projections:
    - tree
    - bom
    - readiness
```

On error:

```yaml
ok: false
error:
  code: VALIDATION_FAILED
  message: Cannot release object
  diagnostics:
    - id: diag_001
      severity: blocker
      message: Required metadata missing
      path: metadata.material
      suggested_actions:
        - fill_material
```

---

# 22. Frontend State Model

Frontend state should be split into:

```text
AppState
ProjectState
WorkspaceState
ObjectCache
UISettings
TaskState
DiagnosticsState
```

## AppState

```text
currentProjectPath
recentProjects
appSettings
theme
```

## ProjectState

```text
projectSummary
objectTreeProjection
gitStatus
indexStatus
modules
```

## WorkspaceState

```text
openTabs
activeTab
selectedObjectID
selectedView
rightInspectorContext
```

## ObjectCache

```text
objectID → ObjectSnapshot
objectID → ValidationSummary
objectID → RelationsSummary
```

Important rule:

Frontend cache is not source of truth. Backend filesystem/YAML is source of truth.

---

# 23. Main User Scenarios

## Scenario 1. Create project

Steps:

```text
1. User clicks Create New Project
2. UI opens wizard
3. User selects folder, code, template, modules
4. UI calls CreateProject
5. Backend creates project files, config, Git, index
6. UI opens workspace
```

Backend commands:

```text
CreateProject
BuildIndex
CreateCheckpoint
AddRecentProject
```

---

## Scenario 2. Create part

Steps:

```text
1. User clicks + New → Part
2. UI opens Create Object Wizard
3. User fills title/material/unit
4. UI previews generated ID
5. User attaches STEP file
6. User clicks Create and Open
7. Backend creates Markdown file with YAML frontmatter
8. Backend copies STEP to files/cad
9. Backend calculates checksum
10. Backend appends history event
11. UI opens object detail
```

Backend:

```text
GenerateObjectID
ValidateObjectDraft
CreateObject
AttachFile
UpdateIndex
```

---

## Scenario 3. Add part to assembly BOM

Steps:

```text
1. User opens assembly
2. Opens BOM tab
3. Clicks Add Item
4. Searches part
5. Enters quantity and unit
6. Clicks Add
7. Backend adds contains relation
8. UI refreshes BOM and tree
```

Backend:

```text
SearchObjects
ValidateRelation
AddRelation
GetBOM
GetObjectTreeProjection
```

---

## Scenario 4. Submit object for review

Steps:

```text
1. User opens draft object
2. Clicks Submit Review
3. UI calls ValidateTransition
4. Backend runs guards
5. If valid, UI asks optional comment
6. UI calls RunTransition
7. Backend updates YAML state
8. Backend appends history event
9. UI updates header and tree status
```

Backend:

```text
ValidateTransition
RunTransition
UpdateObjectState
AppendHistoryEvent
UpdateIndex
```

---

## Scenario 5. Release object with blockers

Steps:

```text
1. User opens approved assembly
2. Clicks Release
3. Backend checks release guards
4. Backend returns blockers
5. UI opens Blockers Dialog
6. User clicks Open Draft Children
7. UI opens affected objects
8. User fixes blockers
9. User runs Release again
```

Backend:

```text
ValidateTransition
ExplainTransitionBlockers
GetReleaseReadiness
RunTransition
```

---

## Scenario 6. Create release package

Steps:

```text
1. User opens Release screen
2. Clicks New Release Package
3. Selects root assembly
4. UI shows release scope
5. User clicks Validate Release
6. Backend builds readiness report
7. If no blockers, user clicks Generate Manifest
8. User clicks Export Package
9. Backend creates manifest.yaml, bom.yaml, files
10. User clicks Create Release Tag
```

Backend:

```text
CreateReleasePackage
ValidateRelease
GenerateReleaseManifest
ExportReleasePackage
CreateReleaseTag
```

---

## Scenario 7. Attach drawing

Steps:

```text
1. User opens object Files tab
2. Clicks Attach File
3. Chooses PDF/DXF
4. UI asks role: drawing_pdf / drawing_dxf
5. Backend routes file to files/drawings
6. Backend calculates checksum
7. Backend updates artifact list
8. UI refreshes Files and Readiness
```

Backend:

```text
PickFile
PreviewFileRouting
AttachFile
RecalculateChecksum
ValidateObject
```

---

## Scenario 8. Change impact before revision

Steps:

```text
1. User opens released part
2. Clicks Revise
3. UI warns: released object requires impact analysis
4. User clicks Analyze Impact
5. Backend builds graph
6. UI shows affected assemblies/releases/instructions
7. User creates Change Request
8. Backend creates CR object linked to affected objects
```

Backend:

```text
GetChangeImpact
CreateChangeRequestFromImpact
RunTransition(revise)
```

---

## Scenario 9. Generate buy list

Steps:

```text
1. User opens assembly BOM
2. Clicks Generate Buy List
3. Backend flattens BOM
4. Backend filters make_buy = buy
5. Backend groups quantities
6. UI opens Procurement Buy List
7. User exports ERP CSV/YAML
```

Backend:

```text
GenerateBuyList
GetProcurementList
ExportERP
```

---

## Scenario 10. Shop-floor build

Steps:

```text
1. User opens Shop Floor screen
2. Clicks Create Build Record
3. Selects released assembly
4. Backend creates build object
5. UI shows step-by-step player
6. Operator completes steps
7. Adds evidence
8. Backend appends build events
9. Build is completed
```

Backend:

```text
CreateBuildRecord
GetWorkInstructionSteps
CompleteBuildStep
AttachEvidence
CompleteBuild
```

---

# 24. Keyboard Shortcuts

Global:

```text
Ctrl+K — Command Palette
Ctrl+P — Search Objects
Ctrl+N — New Object
Ctrl+S — Save
Ctrl+Shift+S — Create Checkpoint
Ctrl+B — Toggle Left Sidebar
Ctrl+I — Toggle Inspector
Ctrl+Tab — Next Tab
Ctrl+W — Close Tab
F5 — Refresh Current View
```

Object screen:

```text
Ctrl+Enter — Primary lifecycle action
Ctrl+D — Diff
Ctrl+Shift+V — Validate
Ctrl+L — Copy Object Link
```

Tree:

```text
↑/↓ — Move selection
→ — Expand
← — Collapse
Enter — Open
F2 — Rename title
Delete — Archive/Delete dialog
```

Backend shortcut handling:

Mostly frontend, but actions call same commands.

---

# 25. UI-to-Backend Mapping Summary

```text
Open Project              → OpenProject
Create Project            → CreateProject
Create Object             → CreateObject
Open Object               → GetObject
Save Markdown             → UpdateObjectDocument
Save Metadata             → UpdateObjectMetadata
Attach File               → AttachFile
Add BOM Item              → AddRelation(type=contains)
Edit BOM Qty              → UpdateRelation
Validate Object           → ValidateObject
Run Lifecycle Action      → RunTransition
Explain Blockers          → ExplainTransitionBlockers
Create Release            → CreateReleasePackage
Generate Manifest         → GenerateReleaseManifest
Export Release            → ExportReleasePackage
Create Checkpoint         → CreateCheckpoint
Show Git Diff             → GetGitDiff
Rebuild Index             → RebuildIndex
Search                    → Search
Show Where Used           → GetWhereUsed
Show Change Impact        → GetChangeImpact
Generate Buy List         → GenerateBuyList
Create Build Record       → CreateBuildRecord
```

---

# 26. Design Rules

1. Every button that changes data must map to a backend command.
2. Every readonly screen must map to a backend query or frontend projection.
3. Backend is the only layer allowed to write YAML/Markdown.
4. Frontend never writes files directly.
5. Every successful command that changes source of truth must append event history.
6. Every command that changes object structure must update or invalidate index projections.
7. Destructive actions require preview and confirmation.
8. Disabled actions must explain why.
9. Diagnostics must be actionable.
10. Git must be shown as engineering checkpoints, not raw Git commands.
11. Index can be deleted and rebuilt.
12. YAML/Markdown remain source of truth.

---

# 27. MVP UI Scope

MVP must include:

```text
Landing Page
Create/Open Project
Project Tree
Create Object Wizard
Object Detail Screen
Markdown Editor
Metadata Panel
Relations Panel
Simple BOM Tab
Files Tab
Validation Tab
Lifecycle Buttons
Git Status
Create Checkpoint
Basic Search
Right Inspector
Bottom Status Bar
Project Settings
```

MVP may exclude:

```text
Shop Floor Player
Advanced Manufacturing Matrix
Procurement workflow
Standard Parts duplicate merge
Visual FSM editor
Advanced graph view
ERP export
PDF generation
```

---

# 28. Expected UX Result

The user should be able to work like this:

```text
I create a part.
I describe it in Markdown.
I fill metadata through forms.
I attach CAD/drawing files.
I add it to assembly BOM.
I validate it.
I submit it for review.
I approve and release it.
I create a checkpoint.
I export release package.
```

The system should continuously answer:

```text
What is this object?
Where is it used?
What state is it in?
What blocks the next step?
What changed?
Can it be released?
What will break if I change it?
```

This is the core UX promise of Markdown-PLM.