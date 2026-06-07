# Tools Stack Specification

# Markdown-PLM / go-plm

## Полное описание используемых инструментов: что, зачем и где применяется

Версия: `0.1`  
Статус: `draft`  
Назначение: описать инструменты, библиотеки и подсистемы, используемые в `go-plm`, а также объяснить, за какую часть продукта отвечает каждый инструмент.

---

# 1. Общая архитектурная идея

`go-plm` — это локальная desktop PLM/PDM-система, где:

```text
Markdown/YAML = источник истины
Go = backend, доменная логика, надежность, файловые транзакции
SQLite = пересобираемый индекс и быстрые проекции
Git = история изменений и релизы
React/TypeScript = пользовательский интерфейс
Vditor = Markdown-редактор/просмотрщик
TanStack Table = BOM, standard parts, procurement и другие таблицы
React Flow = графы связей, where-used, change impact
Zustand = состояние frontend
```

Система не должна зависеть от тяжелой инфраструктуры:

```text
без PostgreSQL в MVP
без Docker requirement
без Kubernetes
без серверного backend
без обязательного облака
без Electron, если используется Wails
```

Главный принцип:

```text
Файлы являются источником истины.
Индекс можно удалить и пересобрать.
Git фиксирует историю.
UI работает через backend commands/queries.
```

---

# 2. Итоговая карта инструментов

|Слой|Инструмент|Назначение|
|---|---|---|
|Desktop shell|Wails|Desktop-приложение с Go backend и web frontend|
|Backend language|Go|Доменная логика, файловые операции, Git, индекс|
|Frontend|React + TypeScript|Сложный инженерный интерфейс|
|Build frontend|Vite|Быстрая сборка frontend|
|UI-kit|Tailwind + shadcn/ui + Radix UI|Кнопки, диалоги, формы, меню, layout|
|Markdown editor|Vditor|Редактирование и просмотр Markdown-документов|
|Tables|TanStack Table|BOM, standard parts, buy list, diagnostics|
|Graphs|React Flow / xyflow|Графы связей, BOM graph, change impact|
|UI state|Zustand|Локальное состояние frontend|
|Server state/cache|TanStack Query или свой query layer|Кэширование ответов backend|
|YAML|gopkg.in/yaml.v3|Чтение/запись YAML|
|Markdown backend|goldmark|Парсинг Markdown на Go|
|Index|SQLite|Локальный быстрый индекс|
|SQLite driver|modernc.org/sqlite|SQLite без CGO|
|Search MVP|SQLite FTS5|Быстрый полнотекстовый поиск|
|Search Advanced|Bleve|Продвинутый поиск, fuzzy, facets, highlighting|
|Git|go-git|Git status, commit, tag, log, diff|
|Diagrams|Mermaid|Диаграммы в Markdown|
|UML optional|PlantUML|UML/sequence/component diagrams|
|Sketches optional|Excalidraw|Ручные инженерные схемы|
|Tests backend|Go test + testify|Unit/integration tests backend|
|Tests frontend|Vitest|Unit tests frontend|
|E2E|Playwright|Проверка сценариев UI|
|UI catalog|Storybook|Каталог компонентов|
|CI|GitHub Actions|Сборка, тесты, проверки|
|Packaging|Wails build|Desktop packaging|

---

# 3. Wails

## 3.1. Назначение

Wails используется как desktop shell для приложения.

Он связывает:

```text
Go backend
+
React/TypeScript frontend
+
desktop window
+
локальный доступ к файловой системе
```

## 3.2. Где используется

```text
запуск desktop-приложения
отображение React UI
вызовы Go backend из frontend
open folder dialogs
save dialogs
system notifications
packaging под Windows/macOS/Linux
```

## 3.3. Почему Wails

Для `go-plm` нужен desktop-first подход:

```text
локальные файлы;
локальный Git;
локальный индекс;
работа без сервера;
один установщик;
нативный доступ к папкам проекта.
```

Wails подходит лучше, чем Electron, потому что backend уже на Go, и не нужно держать отдельный Node/Electron runtime как основной слой приложения.

## 3.4. Что не должен делать Wails

Wails не должен содержать бизнес-логику PLM.

Правильно:

```text
Wails = оболочка и мост frontend/backend
```

Неправильно:

```text
Wails = место, где реализована логика BOM, lifecycle, validation
```

Бизнес-логика должна быть в `internal/app`, `internal/core`, `internal/modules`.

---

# 4. Go

## 4.1. Назначение

Go — основной backend-язык проекта.

Он отвечает за:

```text
доменную модель;
создание объектов;
чтение и запись Markdown/YAML;
файловые транзакции;
атомарные операции;
историю событий;
индекс SQLite;
Git-интеграцию;
validation engine;
FSM lifecycle;
BOM engine;
standard parts engine;
release engine;
procurement export.
```

## 4.2. Почему Go

Go хорошо подходит для такой системы, потому что:

```text
быстрый;
компилируется в один бинарник;
хорошо работает с файловой системой;
удобен для CLI/desktop backend;
имеет сильную стандартную библиотеку;
прост в сопровождении;
хорош для надежных backend-сервисов.
```

## 4.3. Go-пакеты верхнего уровня

Рекомендуемая структура:

```text
cmd/plm
internal/core
internal/app
internal/store
internal/parser
internal/naming
internal/validation
internal/process
internal/gitops
internal/modules
internal/api
```

## 4.4. Принцип backend

Frontend никогда не должен напрямую писать `.md`, `.yaml`, `.db`, `.git`.

Все изменения идут через команды backend:

```text
CreateObject
UpdateObjectDocument
UpdateObjectMetadata
AttachArtifact
AddRelation
RunTransition
CreateReleasePackage
CreateCheckpoint
```

---

# 5. React + TypeScript

## 5.1. Назначение

React + TypeScript используются для frontend UI.

UI системы сложный:

```text
Project Tree
Object Detail Screen
Markdown Editor
Properties Panel
BOM Table
Standard Parts Center
Procurement List
Manufacturing Matrix
Release Dashboard
Change Impact Graph
Git Changes Screen
Settings Screens
```

## 5.2. Почему React

React хорошо подходит для:

```text
сложных интерактивных интерфейсов;
компонентной архитектуры;
интеграции Vditor, TanStack Table, React Flow;
создания reusable UI-компонентов;
работы с TypeScript.
```

## 5.3. Почему TypeScript обязателен

PLM-интерфейс работает со сложными сущностями:

```text
ObjectSnapshot
Relation
Artifact
ValidationDiagnostic
BOMRow
StandardPart
ReleasePackage
GitStatus
TaskStatus
```

Без TypeScript frontend быстро станет хрупким.

---

# 6. Vite

## 6.1. Назначение

Vite используется для сборки frontend.

## 6.2. Где применяется

```text
development server
hot reload
React build
production bundle для Wails
```

## 6.3. Почему Vite

```text
быстрый dev startup;
простая настройка;
хорошая поддержка React/TypeScript;
удобная интеграция с Wails.
```

---

# 7. Tailwind CSS

## 7.1. Назначение

Tailwind используется для стилизации UI.

## 7.2. Где применяется

```text
layout;
панели;
отступы;
цвета статусов;
таблицы;
кнопки;
карточки diagnostics;
responsive behavior;
compact engineering UI.
```

## 7.3. Почему Tailwind

Для инженерного приложения нужен плотный и кастомный UI. Tailwind позволяет быстро делать:

```text
compact mode;
dense tables;
status badges;
split panels;
toolbars;
cards;
dark theme.
```

---

# 8. shadcn/ui

## 8.1. Назначение

shadcn/ui используется как набор готовых React-компонентов поверх Radix UI и Tailwind.

## 8.2. Где применяется

```text
Button
Dialog
DropdownMenu
ContextMenu
Tabs
Command
Popover
Tooltip
Input
Select
Checkbox
Switch
Resizable Panels
Sheet
Toast
```

## 8.3. Почему shadcn/ui

Это не тяжелый UI framework. Компоненты копируются в проект и становятся частью кода. Их можно адаптировать под PLM UX.

## 8.4. Примеры использования

```text
Command Palette → shadcn Command
Create Object Wizard → Dialog + Form
Right-click object menu → ContextMenu
Lifecycle actions → DropdownMenu
Settings → Tabs
Diagnostics → Alert / Card
```

---

# 9. Radix UI

## 9.1. Назначение

Radix UI — набор low-level accessible primitives.

## 9.2. Где применяется

Radix обычно используется через shadcn/ui, но напрямую может понадобиться для:

```text
сложных dropdown;
контекстных меню;
resizable panels;
modal/dialog;
tooltip;
popover;
tabs;
accordion.
```

## 9.3. Почему важен

PLM-система должна быть удобной и предсказуемой:

```text
клавиатурная навигация;
фокус;
доступность;
правильные dropdown/dialog behavior.
```

---

# 10. Lucide React

## 10.1. Назначение

Lucide React используется для иконок.

## 10.2. Где применяется

```text
object class icons;
toolbar icons;
file type icons;
lifecycle state icons;
validation severity icons;
Git status icons;
BOM row actions;
settings navigation.
```

## 10.3. Примеры

```text
Part → Box / Component
Assembly → Boxes
Drawing → FileText
Warning → TriangleAlert
Released → BadgeCheck
Git changes → GitBranch
Search → Search
Settings → Settings
```

---

# 11. Vditor (primary editor)

## 11.1. Назначение

Vditor — основной Markdown editor/viewer.

Он используется для:

```text
редактирования Markdown body;
preview документа;
WYSIWYG режима;
instant render режима;
split preview режима;
таблиц;
task lists;
Mermaid/PlantUML/Graphviz блоков;
математических формул;
работы с изображениями;
просмотра Markdown-документов.
```

## 11.2. Где применяется

```text
Object Detail → Document Tab
Work Instruction Editor
Release Notes Editor
Change Request Editor
Technical Process Card Editor
Inspection Plan Editor
Markdown Preview Panel
```

## 11.3. Как связан с backend

Vditor не пишет файлы напрямую.

Поток сохранения:

```text
Vditor content
→ React SaveController
→ backend UpdateObjectDocument
→ parse Markdown/frontmatter
→ validate
→ atomic write
→ append history.yaml.log
→ update SQLite index
→ return ObjectSnapshot
```

## 11.4. Как работать с YAML frontmatter

Рекомендуемая схема:

```text
YAML frontmatter редактируется через Properties Panel
Markdown body редактируется через Vditor
```

YAML frontmatter редактируется через structured Properties Panel. Config-файлы — через Settings UI.

## 11.5. Зачем не отдавать YAML полностью в Vditor

Пользователь может случайно сломать:

```text
id
class
state
relations
artifacts
revision
```

Поэтому системные поля лучше редактировать формами с валидацией.

## 11.6. Что добавить поверх Vditor

Для PLM нужны кастомные вставки:

```text
Insert Object Link
Insert Artifact Link
Insert BOM Table
Insert Inspection Checklist
Insert Safety Warning
Insert Release Block
Insert Mermaid Diagram
```

## 11.7. MVP-настройки Vditor

```text
mode: split preview или instant render
theme: light/dark
toolbar: ограниченный инженерный набор
autosave: controlled by app
upload: routed through backend AttachArtifact
preview: enabled
frontmatter: enabled
```

---

# 12. CodeMirror 6 (REMOVED)

**Decision**: CodeMirror 6 исключён из стека. Vditor — единственный редактор.

YAML frontmatter редактируется через structured Properties Panel (metadata editing),
а не как raw text. YAML config-файлы редактируются через dedicated Settings UI
с form-based редактированием, что исключает необходимость в raw YAML editor.

---

# 13. TanStack Table

## 13.1. Назначение

TanStack Table используется для всех сложных таблиц.

## 13.2. Где применяется

```text
BOM Table
Flat BOM
Standard Parts Center
Procurement Buy List
Release Package contents
Manufacturing Matrix
Validation Diagnostics
Git Changed Objects
Where-used list
Artifact list
Task list
```

## 13.3. Почему TanStack Table

PLM-таблицы требуют:

```text
сортировку;
фильтры;
группировку;
выбор строк;
массовые операции;
виртуализацию;
кастомные ячейки;
inline editing;
закрепленные колонки;
server-side queries;
экспорт.
```

## 13.4. Пример использования в BOM

Колонки:

```text
Position
Object ID
Class
Title
Revision
State
Quantity
Unit
Make/Buy
Supplier
Validation
Actions
```

Действия:

```text
Open
Edit Quantity
Replace
Remove
Show Where Used
Show Impact
```

Backend-запросы:

```text
GetBOM
UpdateBOMItem
RemoveBOMItem
ReplaceBOMItem
ValidateBOM
ExportBOM
```

---

# 14. React Flow / xyflow

## 14.1. Назначение

React Flow используется для интерактивных графов.

## 14.2. Где применяется

```text
Where-used graph
Change Impact Radar
BOM graph
eBOM to mBOM graph
Release dependency graph
Lifecycle/FSM visual editor
Process plan graph
Object relation graph
```

## 14.3. Почему нужен

В PLM связи важнее папок. Графы позволяют понять:

```text
где используется объект;
что сломается при изменении;
какие сборки затронуты;
какие релизы затронуты;
какие инструкции зависят;
какие стандартные детали можно заменить.
```

## 14.4. Связь с backend

Backend строит graph projection:

```text
nodes:
  object nodes
  artifact nodes
  release nodes
  process nodes

edges:
  contains
  has_drawing
  manufacturing_file_for
  substitutes
  affects
  includes
```

Frontend отображает через React Flow.

Команды:

```text
GetRelationGraph
GetWhereUsedGraph
GetChangeImpactGraph
ExpandGraphNode
CollapseGraphNode
OpenGraphNode
CreateRelationFromGraph
```

---

# 15. Zustand

## 15.1. Назначение

Zustand хранит локальное состояние frontend.

## 15.2. Что хранить в Zustand

```text
currentProject
selectedObjectId
openTabs
activeTab
activeView
leftPanelState
rightInspectorState
expandedTreeNodes
editorDirtyState
selectedBOMRows
selectedGraphNodes
UI preferences
```

## 15.3. Что не хранить в Zustand

Нельзя считать Zustand источником истины для PLM-данных.

Не хранить как truth:

```text
object source data
relations source data
artifact source data
lifecycle state as canonical
BOM canonical data
```

Эти данные приходят из backend.

## 15.4. Рекомендуемые stores

```text
useAppStore
useProjectStore
useWorkspaceStore
useEditorStore
useSelectionStore
useTaskStore
useSettingsStore
```

---

# 16. TanStack Query или свой query layer

## 16.1. Назначение

Нужен слой для кэширования backend-запросов.

## 16.2. Где применяется

```text
GetObject
ListObjects
SearchObjects
GetBOM
GetWhereUsed
GetValidationSummary
GetGitStatus
GetTasks
```

## 16.3. Варианты

Вариант A:

```text
TanStack Query
```

Вариант B:

```text
свой thin query cache поверх Wails bindings
```

## 16.4. Рекомендация

Для MVP можно сделать свой простой слой:

```text
api.ts
commands.ts
queries.ts
cache invalidation by refresh hints
```

Позже перейти на TanStack Query, если появится много async-запросов и фоновых обновлений.

---

# 17. gopkg.in/yaml.v3

## 17.1. Назначение

Основной YAML parser/encoder в Go.

## 17.2. Где используется

```text
project.yaml
classes.yaml
lifecycle.yaml
validation.yaml
file-routing.yaml
exports.yaml
Markdown frontmatter
manifest.yaml
bom.yaml
history.yaml.log
```

## 17.3. Почему YAML

Философия проекта:

```text
YAML = человекочитаемые структурные данные
Markdown = инженерное описание
Git = история
SQLite = индекс
```

## 17.4. Правила использования

Backend должен:

```text
парсить YAML строго;
сохранять понятный порядок ключей;
валидировать unknown fields;
давать diagnostics с line/field path;
не терять комментарии там, где возможно.
```

---

# 18. goldmark

## 18.1. Назначение

Markdown parser на Go.

## 18.2. Где используется

```text
парсинг Markdown body;
извлечение заголовков;
генерация outline;
валидация структуры документа;
рендер preview/export;
поиск по Markdown;
извлечение object links;
извлечение embedded directives.
```

## 18.3. Почему нужен backend Markdown parser

Frontend preview недостаточно. Backend должен понимать документы для:

```text
validation;
indexing;
release package;
export;
document outline;
link validation;
traceability.
```

## 18.4. Пример backend-задач

```text
Validate required sections:
  ## Назначение
  ## Материал
  ## Контроль качества

Extract links:
  [[a320-prt-0001-v1.0]]

Extract directives:
  @bom[a320-asm-0100-v1.0]
```

---

# 19. SQLite

## 19.1. Назначение

SQLite используется как локальный пересобираемый индекс.

## 19.2. Что хранить в SQLite

```text
objects
relations
artifacts
metadata projections
search index
BOM projections
where-used projections
standard parts normalized keys
procurement projections
validation summaries
release readiness
task status
```

## 19.3. Что нельзя хранить только в SQLite

SQLite не должен быть единственным источником для:

```text
object identity
canonical metadata
relations source of truth
artifact list source of truth
history
release manifest source of truth
```

Это должно быть в Markdown/YAML.

## 19.4. Почему SQLite

```text
локальный;
быстрый;
транзакционный;
не требует сервера;
хорошо подходит для desktop app;
можно удалить и пересобрать.
```

## 19.5. Файл индекса

```text
.plm/index.db
```

Если индекс поврежден:

```text
RebuildIndex
```

должен пересоздать его из Markdown/YAML.

---

# 20. modernc.org/sqlite

## 20.1. Назначение

SQLite driver для Go без CGO.

## 20.2. Почему важно без CGO

Для desktop-приложения проще сборка:

```text
Windows
macOS
Linux
GitHub Actions
Wails build
cross-platform distribution
```

Без CGO меньше проблем с компиляторами C и системными библиотеками.

---

# 21. SQLite FTS5

## 21.1. Назначение

FTS5 используется для MVP-поиска.

## 21.2. Что индексировать

```text
object id
title
class
state
metadata text
Markdown headings
Markdown body
artifact names
standard part normalized key
supplier part number
manufacturer part number
```

## 21.3. Где используется

```text
Global Search
Search Objects
Search Standard Parts
Search Documents
Search Artifacts
```

## 21.4. Когда заменить или дополнить Bleve

Если понадобятся:

```text
fuzzy search;
faceted search;
highlighting;
advanced scoring;
synonyms;
сложный поиск по стандартным деталям;
поиск по большим проектам.
```

---

# 22. Bleve

## 22.1. Назначение

Bleve — продвинутый поисковый движок на Go.

## 22.2. Где использовать

```text
advanced global search;
standard parts search;
duplicate detection support;
fuzzy search;
search highlighting;
facets by class/state/material/supplier;
search inside Markdown documents.
```

## 22.3. Когда добавлять

Не в первый MVP, а после появления:

```text
тысяч объектов;
большой библиотеки стандартных деталей;
сложных фильтров;
необходимости fuzzy-поиска.
```

## 22.4. Связь с SQLite

SQLite остается основным индексом проекций.

Bleve может стать отдельным search index:

```text
.plm/search.bleve
```

Он тоже должен быть пересобираемым.

---

# 23. go-git

## 23.1. Назначение

go-git используется для Git-интеграции в backend.

## 23.2. Где применяется

```text
Git status
Create checkpoint
Show Git log
Create release tag
Compare with last release
Detect changed files
Detect untracked files
```

## 23.3. UX-правило

Пользователь не должен видеть Git как набор команд:

```text
git add
git commit
git tag
git status
```

Он должен видеть инженерные действия:

```text
Create Checkpoint
Show Changes
Compare with Release
Create Release Tag
Restore Checkpoint
```

## 23.4. Backend commands

```text
GetGitStatus
GetDomainDiff
CreateCheckpoint
CreateReleaseTag
GetGitLog
CompareWithRelease
```

## 23.5. Что оставить на будущее

```text
merge;
conflict resolution;
remote push/pull;
branch management;
multi-user Git workflows.
```

Для MVP достаточно локального Git.

---

# 24. Mermaid

## 24.1. Назначение

Mermaid используется для диаграмм внутри Markdown.

## 24.2. Где применяется

```text
lifecycle diagrams;
FSM;
release workflow;
change request workflow;
manufacturing process;
architecture diagrams;
sequence diagrams;
BOM overview.
```

## 24.3. Пример

````markdown
```mermaid
stateDiagram-v2
  draft --> in_review: submit_review
  in_review --> approved: approve
  approved --> released: release
````

````

## 24.4. Почему Mermaid

```text
пишется текстом;
хранится в Markdown;
отлично diff-ится Git;
подходит к file-based философии проекта.
````

---

# 25. PlantUML

## 25.1. Назначение

PlantUML — optional-инструмент для строгих UML/architecture diagrams.

## 25.2. Где полезен

```text
developer documentation;
architecture diagrams;
component diagrams;
sequence diagrams;
state diagrams;
ER diagrams;
deployment diagrams.
```

## 25.3. Почему optional

PlantUML мощный, но сложнее в интеграции. Для MVP достаточно Mermaid.

---

# 26. Excalidraw

## 26.1. Назначение

Excalidraw можно использовать для sketch-документов.

## 26.2. Где полезен

```text
design review sketches;
ручные схемы;
объяснение проблемы;
layout sketches;
concept notes;
assembly notes.
```

## 26.3. Как хранить

```yaml
artifacts:
  - kind: sketch
    role: design_note
    path: files/sketches/a320-prt-0001-v1.0.excalidraw
```

## 26.4. Когда добавлять

Не в MVP. Хорошо для этапа 2 или 3.

---

# 27. MarkText

## 27.1. Назначение

MarkText не нужно встраивать напрямую. Его стоит использовать как UX-референс.

## 27.2. Что изучить

```text
clean Markdown UX;
focus mode;
typewriter mode;
source mode;
frontmatter editing;
export HTML/PDF;
темы;
простота интерфейса.
```

## 27.3. Почему не встраивать

MarkText — полноценное desktop-приложение. `go-plm` строится как Wails + React + Go, поэтому проще использовать идеи MarkText, а не сам MarkText как зависимость.

---

# 28. React Hook Form

## 28.1. Назначение

React Hook Form используется для форм metadata и настроек.

## 28.2. Где применяется

```text
Create Project Wizard
Create Object Wizard
Metadata Panel
Standard Part Form
Supplier Form
Lifecycle Config Form
Validation Rule Form
Export Settings
```

## 28.3. Почему нужен

Форм в PLM будет много. Нужны:

```text
валидация;
default values;
dirty state;
controlled/uncontrolled inputs;
быстрая работа;
интеграция со схемами.
```

---

# 29. Zod

## 29.1. Назначение

Zod используется для frontend-схем и валидации DTO.

## 29.2. Где применяется

```text
validation of forms;
backend response parsing;
command payload validation;
frontend type inference;
settings forms.
```

## 29.3. Важно

Zod не заменяет backend validation.

Правильно:

```text
Zod = быстрые frontend checks
Go validation = источник истины
```

---

# 30. Go validation engine

## 30.1. Назначение

Свой validation engine нужен для проверки PLM-правил.

## 30.2. Почему не достаточно Zod/YAML schema

PLM-валидация зависит от графа связей:

```text
все дочерние детали released;
нет циклов BOM;
чертеж существует;
checksum актуален;
standard part approved;
replacement allowed;
release blockers отсутствуют.
```

Это нельзя полноценно проверить только схемой одного YAML-файла.

## 30.3. Типы проверок

```text
syntax validation;
schema validation;
naming validation;
relation validation;
artifact validation;
BOM validation;
lifecycle validation;
release validation;
standard parts validation;
procurement validation.
```

---

# 31. Custom FSM Engine

## 31.1. Назначение

FSM engine управляет жизненными циклами.

## 31.2. Где применяется

```text
object lifecycle;
document lifecycle;
standard part lifecycle;
release lifecycle;
change request lifecycle;
substitution workflow;
build record workflow.
```

## 31.3. Почему свой FSM

Процессы достаточно простые, но должны быть объяснимыми:

```text
state
transition
guards
effects
reason
history event
```

Большой workflow engine для MVP будет лишним.

## 31.4. Конфигурация

```yaml
processes:
  object_lifecycle:
    initial: draft
    states:
      - draft
      - in_review
      - approved
      - released
    transitions:
      submit_review:
        from: draft
        to: in_review
        guards:
          - required_metadata
          - valid_name
```

---

# 32. File Transaction Engine

## 32.1. Назначение

Собственный file transaction engine нужен для надежной записи файлов.

## 32.2. Почему это важно

Источник истины — файлы. Значит, нельзя делать небезопасные операции:

```text
write directly;
overwrite without backup;
partial writes;
silent failure.
```

## 32.3. Правильная схема

```text
prepare transaction
write temp files
fsync
atomic rename
append history
update index
commit transaction marker
cleanup
```

## 32.4. Где используется

```text
CreateObject
UpdateObjectDocument
UpdateMetadata
AttachArtifact
AddRelation
ReviseObject
CreateReleasePackage
UpdateConfig
```

---

# 33. History YAML Log

## 33.1. Назначение

`history.yaml.log` хранит историю событий объекта.

## 33.2. Почему YAML log

```text
читаемый человеком;
хорошо подходит философии проекта;
diff-friendly;
можно восстановить события.
```

## 33.3. Пример

```yaml
---
event_id: evt-20260606-000001
time: 2026-06-06T10:00:00Z
actor: local-user
type: object.created
object_id: a320-prt-0001-v1.0
payload:
  title: Bracket Motor
```

## 33.4. Где используется

```text
Object History tab
Audit trail
Release notes generation
Change impact
Debugging
Recovery
```

---

# 34. GitHub Actions

## 34.1. Назначение

GitHub Actions используется для CI.

## 34.2. Что проверять

```text
go test ./...
go vet
frontend typecheck
frontend build
unit tests
lint
Wails build smoke
docs validation
sample project validation
```

## 34.3. Когда добавить packaging

На этапе MVP достаточно CI. Packaging можно добавить позже:

```text
Windows installer
macOS dmg
Linux AppImage/deb
```

---

# 35. Go test

## 35.1. Назначение

Основная система тестирования backend.

## 35.2. Что тестировать

```text
naming engine;
YAML parser;
frontmatter parser;
filesystem store;
file transactions;
FSM transitions;
guards;
relations;
BOM cycle detection;
standard part normalization;
duplicate detection;
release readiness;
Git wrapper.
```

---

# 36. Testify

## 36.1. Назначение

Testify можно использовать для удобных assertions в Go tests.

## 36.2. Где полезен

```text
unit tests;
integration tests;
fixtures;
assert.Equal;
require.NoError;
mock interfaces.
```

## 36.3. Рекомендация

Можно использовать, но не обязательно. Стандартного `testing` тоже достаточно. Для читаемости большого проекта Testify полезен.

---

# 37. Vitest

## 37.1. Назначение

Vitest используется для frontend unit tests.

## 37.2. Что тестировать

```text
utility functions;
stores;
formatters;
table transforms;
validation display logic;
command availability logic;
component behavior.
```

---

# 38. Playwright

## 38.1. Назначение

Playwright используется для E2E-тестов.

## 38.2. Какие сценарии тестировать

```text
create project;
create part;
edit markdown;
save object;
attach file;
add BOM row;
run validation;
open blockers dialog;
submit review;
create checkpoint;
search object;
open standard parts center.
```

## 38.3. Как тестировать Wails

Для MVP проще тестировать React frontend в dev/web mode с тестовым backend API или mock API. Позже можно тестировать packaged desktop.

---

# 39. Storybook

## 39.1. Назначение

Storybook используется как каталог UI-компонентов.

## 39.2. Какие компоненты вынести

```text
ObjectBadge
LifecycleBadge
ClassIcon
ValidationCard
DiagnosticList
BOMTable
StandardPartCard
RelationGraph
ArtifactList
ReleaseBlockerDialog
ReadinessMatrix
GitStatusBadge
```

## 39.3. Когда добавлять

Не обязательно для самого первого MVP, но очень полезно, если UI будет расти.

---

# 40. Инструменты для стандартных деталей

## 40.1. Основные инструменты

```text
TanStack Table → таблица стандартных деталей
SQLite → normalized keys и быстрые фильтры
Bleve → fuzzy search и duplicate candidates
React Hook Form → форма создания std
Zod → frontend validation
Go validation engine → окончательная проверка
React Flow → граф замен и equivalents
```

## 40.2. Для чего каждый нужен

### TanStack Table

```text
список стандартных деталей;
фильтры по std_class;
фильтры по standard/material/supplier;
выделение строк;
массовая обработка;
merge duplicates;
export.
```

### SQLite

```text
normalized_key;
std_class;
supplier_part_number;
manufacturer_part_number;
state;
where-used count.
```

### Bleve

```text
поиск похожих деталей;
fuzzy match;
поиск по описаниям;
подсветка найденных совпадений.
```

### React Flow

```text
substitutes graph;
equivalent_to graph;
replaces graph;
affected BOM graph.
```

---

# 41. Инструменты для BOM

## 41.1. Основные инструменты

```text
TanStack Table
SQLite
React Flow
Go BOM engine
YAML frontmatter relations
```

## 41.2. Роли

```text
YAML relations → источник BOM;
Go BOM engine → строит structured/flat BOM;
SQLite → хранит projection;
TanStack Table → показывает BOM;
React Flow → показывает BOM graph.
```

---

# 42. Инструменты для релизов

## 42.1. Основные инструменты

```text
Go release engine
Validation engine
go-git
YAML manifest
TanStack Table
Vditor
```

## 42.2. Роли

```text
Validation engine → проверяет готовность;
Go release engine → собирает release package;
YAML manifest → фиксирует состав релиза;
go-git → tag release;
TanStack Table → показывает состав;
Vditor → release notes editor.
```

---

# 43. Инструменты для change impact

## 43.1. Основные инструменты

```text
Relation graph
SQLite projections
React Flow
TanStack Table
Go impact engine
```

## 43.2. Роли

```text
Go impact engine → находит затронутые объекты;
SQLite → быстро отвечает where-used;
React Flow → показывает граф;
TanStack Table → показывает affected objects list.
```

---

# 44. Инструменты для manufacturing

## 44.1. Основные инструменты

```text
TanStack Table
Vditor
Validation engine
YAML process plans
React Flow
```

## 44.2. Где применяются

```text
Manufacturing Matrix → TanStack Table
Work Instructions → Vditor
Process routes → YAML + tables
Operation graph → React Flow
Readiness checks → Validation engine
```

---

# 45. Инструменты для procurement

## 45.1. Основные инструменты

```text
TanStack Table
SQLite
YAML export
CSV export
Standard Parts engine
```

## 45.2. Где применяются

```text
Buy List
Supplier list
Approved vendor list
Substitution workflow
ERP export
```

---

# 46. Инструменты для экспорта

## 46.1. MVP exports

```text
YAML
CSV
Markdown
ZIP folder
```

## 46.2. Инструменты

```text
Go stdlib archive/zip
encoding/csv
yaml.v3
goldmark для Markdown rendering
```

## 46.3. Что экспортировать

```text
BOM
Flat BOM
Buy List
Release Manifest
Validation Report
Standard Parts List
Change Impact Report
```

---

# 47. Что не брать в MVP

## 47.1. Не брать сразу

```text
PostgreSQL
MongoDB
Redis
Kafka
NATS
Temporal
Kubernetes
GraphQL
Electron
full plugin VM
cloud sync
complex permissions engine
server-side multi-user collaboration
```

## 47.2. Почему

Потому что MVP должен доказать основную концепцию:

```text
Markdown/YAML source of truth
Object/Relation/Artifact model
BOM
Lifecycle
Validation
Git checkpoint
Desktop UX
```

Лишняя инфраструктура усложнит проект раньше времени.

---

# 48. MVP toolset

Для первой рабочей версии достаточно:

```text
Go
Wails
React
TypeScript
Vite
Tailwind
shadcn/ui
Vditor
TanStack Table
Zustand
yaml.v3
goldmark
SQLite
modernc.org/sqlite
SQLite FTS5
go-git
Mermaid
Go test
Vitest
GitHub Actions
```

---

# 49. Phase 2 toolset

Добавить после появления стабильного ядра:

```text
React Flow / xyflow
Bleve
Playwright
Storybook
Excalidraw
PlantUML
```

---

# 50. Phase 3 toolset

Добавить для расширенной системы:

```text
D2 / Graphviz export
advanced CAD previewers
ERP adapters
standard parts importers
PDF generation
digital signature module
multi-user sync
```

---

# 51. Рекомендуемая структура frontend по инструментам

```text
frontend/src/
├── app/
│   ├── App.tsx
│   ├── routes.tsx
│   └── providers.tsx
├── api/
│   ├── commands.ts
│   ├── queries.ts
│   └── types.ts
├── stores/
│   ├── useAppStore.ts
│   ├── useWorkspaceStore.ts
│   ├── useEditorStore.ts
│   └── useTaskStore.ts
├── components/
│   ├── ui/
│   ├── ObjectBadge.tsx
│   ├── LifecycleBadge.tsx
│   └── ValidationCard.tsx
├── features/
│   ├── objects/
│   ├── bom/
│   ├── standard-parts/
│   ├── release/
│   ├── manufacturing/
│   ├── procurement/
│   ├── change-impact/
│   └── git/
├── editor/
│   ├── VditorEditor.tsx
│   ├── MarkdownPreview.tsx
├── tables/
│   ├── BOMTable.tsx
│   ├── StandardPartsTable.tsx
│   └── DiagnosticsTable.tsx
└── graphs/
    ├── RelationGraph.tsx
    ├── ImpactGraph.tsx
    └── BomGraph.tsx
```

---

# 52. Рекомендуемая структура backend по инструментам

```text
internal/
├── core/
│   ├── object/
│   ├── relation/
│   ├── artifact/
│   ├── revision/
│   └── event/
├── app/
│   ├── commands/
│   ├── queries/
│   └── services/
├── store/
│   ├── filesystem/
│   ├── transaction/
│   ├── index/
│   └── recovery/
├── parser/
│   ├── markdown/
│   ├── frontmatter/
│   └── yaml/
├── naming/
├── validation/
├── process/
│   ├── fsm/
│   ├── guards/
│   └── effects/
├── gitops/
├── search/
│   ├── sqlitefts/
│   └── bleve/
└── modules/
    ├── bom/
    ├── release/
    ├── manufacturing/
    ├── procurement/
    ├── shopfloor/
    ├── standardparts/
    └── change/
```

---

# 53. Главный принцип выбора инструментов

Каждый инструмент должен выполнять одну понятную роль:

```text
Vditor не управляет PLM-данными — он редактирует Markdown.
TanStack Table не хранит BOM — он показывает BOM.
React Flow не хранит связи — он показывает граф.
Zustand не источник истины — он хранит состояние UI.
SQLite не источник истины — он индекс.
Git не бизнес-логика — он история.
Go backend — единственный слой, который пишет source of truth.
```

---

# 54. Финальная рекомендация

Оптимальный стек для `go-plm`:

```text
Go + Wails + React + TypeScript
Markdown/YAML source of truth
Vditor for Markdown. YAML/config — через structured Properties Panel и Settings UI.
TanStack Table for tables
React Flow for graphs
Zustand for UI state
SQLite for index
go-git for history
Bleve later for advanced search
Mermaid for diagrams
```

Такой стек дает баланс:

```text
быстрый MVP;
читаемые инженерные данные;
мощный frontend;
надежный Go backend;
простая локальная установка;
хорошая расширяемость;
минимальная инфраструктурная сложность.
```