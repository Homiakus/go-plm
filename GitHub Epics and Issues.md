# GitHub Epics & Issues Breakdown

# go-plm / Markdown-PLM v2.0

---

## Epic 0: Repository Skeleton

**Labels**: `epic`, `phase-0`, `p0`
**Goal**: Инициализация проекта, CI, структура директорий

### Issue 0.1: Initialize Go module and project structure
**Labels**: `backend`, `p0`
**Description**:
Создать go.mod, структуру директорий согласно SRS section 14.2, .gitignore, README.md.
**Acceptance criteria**:
- go.mod с module github.com/Homiakus/go-plm
- Директории cmd/plm, internal/{core,app,parser,naming,store,validation,process,gitops,search,modules,api}
- .gitignore включает: .env, .plm/transactions/, .plm/cache/, *.token, node_modules/

### Issue 0.2: Set up CI with GitHub Actions
**Labels**: `devops`, `p0`
**Description**:
GitHub Actions workflow: lint (golangci-lint), test (go test ./...), build (go build ./...).
**Acceptance criteria**:
- .github/workflows/ci.yml
- Запускается на push/PR
- Статус отображается в PR

### Issue 0.3: Set up frontend skeleton
**Labels**: `frontend`, `p0`
**Description**:
Инициализировать frontend/: package.json, vite.config.ts, React 19 + TypeScript + Tailwind + shadcn/ui.
**Acceptance criteria**:
- npm install проходит без ошибок
- vite dev запускается
- Базовая структура: src/{app,features,editor,components,bindings}/

---

## Epic 1: Core Domain

**Labels**: `epic`, `phase-1`, `p0`
**Goal**: Чистая доменная модель (zero dependencies)
**Depends on**: Epic 0

### Issue 1.1: Implement core/object
**Labels**: `backend`, `core`, `p0`
**Description**:
Реализовать типы Object, ObjectID, Class, State согласно SRS section 9.3.
```go
type Object struct { ID, Project, Class, Sequence, Version, Revision, State, Title string; Metadata map[string]any; Relations []RelationRef; Artifacts []ArtifactRef }
```
**Acceptance criteria**:
- Object.IsZero(), IsReleased(), IsDraft(), WithState(), ValidateIdentity()
- Не импортирует store, api, gitops, modules
- Покрытие тестами 100%

### Issue 1.2: Implement core/relation
**Labels**: `backend`, `core`, `p0`
**Description**:
Relation, RelationType (contains, has_drawing, has_cad, substitutes, equivalent_to, affects, includes, ...).
**Acceptance criteria**:
- Константы типов связей
- ValidateShape(), IsBOMRelation(), HasQuantity()
- Покрытие тестами 100%

### Issue 1.3: Implement core/artifact
**Labels**: `backend`, `core`, `p0`
**Description**:
Artifact, Kind, Role, Status.
**Acceptance criteria**:
- Константы Kind и Role
- Placeholder artifact (status=missing)
- Покрытие тестами 100%

### Issue 1.4: Implement core/event
**Labels**: `backend`, `core`, `p0`
**Description**:
Event, EventType (object.created, metadata.updated, relation.added, lifecycle.transitioned, artifact.attached, ...).
**Acceptance criteria**:
- Все константы EventType
- Покрытие тестами 100%

### Issue 1.5: Implement core/diagnostic
**Labels**: `backend`, `core`, `p0`
**Description**:
Diagnostic, Severity (info, warning, blocker, error).
**Acceptance criteria**:
- Поля: ID, Severity, Code, ObjectID, Path, Message, SuggestedActions
- IsBlocker(), IsWarning()
- Покрытие тестами 100%

### Issue 1.6: Implement core/lifecycle and core/revision
**Labels**: `backend`, `core`, `p0`
**Description**:
LifecycleState, Revision типы.
**Acceptance criteria**:
- Типы определены, тесты

---

## Epic 2: Filesystem Store

**Labels**: `epic`, `phase-2`, `p0`
**Goal**: Надёжное файловое хранение с атомарными транзакциями
**Depends on**: Epic 1

### Issue 2.1: Implement store/transaction
**Labels**: `backend`, `store`, `p0`
**Description**:
Атомарные файловые операции: Plan, temp-файлы, fsync, atomic rename, rollback markers, idempotency.
**Acceptance criteria**:
- Manager.Execute(ctx, Plan) — атомарно
- Manager.Recover(ctx) — находит незавершённые транзакции
- Тесты: success, interrupted, rollback, idempotent retry

### Issue 2.2: Implement store/fsrepo
**Labels**: `backend`, `store`, `p0`
**Description**:
Чтение/запись объектов через transaction layer. Чтение project.yaml, objects/*, history.yaml.log.
**Acceptance criteria**:
- GetObject(ctx, id) — читает .md, парсит frontmatter
- SaveObject(ctx, obj) — пишет через transaction
- ListObjects(ctx) — сканирует objects/
- Тесты с реальными файлами (t.TempDir)

### Issue 2.3: Implement store/recovery
**Labels**: `backend`, `store`, `p1`
**Description**:
Recovery mode: поиск незавершённых транзакций, repair actions, rollback/complete.
**Acceptance criteria**:
- Находит битые/незавершённые транзакции
- Предлагает complete/rollback
- Тесты с fault injection

---

## Epic 3: Parser / Naming / Validation

**Labels**: `epic`, `phase-3`, `p0`
**Goal**: Парсинг, генерация ID, валидация
**Depends on**: Epic 1

### Issue 3.1: Implement parser/frontmatter
**Labels**: `backend`, `parser`, `p0`
**Description**:
Split/join YAML frontmatter + Markdown body.
**Acceptance criteria**:
- Split([]byte) → (frontmatter, body, error)
- Join(frontmatter, body) → []byte
- Обработка edge cases (no frontmatter, empty, large files)
- Golden tests

### Issue 3.2: Implement parser/yamlx
**Labels**: `backend`, `parser`, `p0`
**Description**:
Typed YAML decode/encode с line/column errors.
**Acceptance criteria**:
- Decode[T](data) → (T, error) с позицией ошибки
- Encode(v) → []byte
- DecodeStrict для строгого режима

### Issue 3.3: Implement parser/markdown
**Labels**: `backend`, `parser`, `p0`
**Description**:
Backend Markdown: outline extraction, object links, section validation.
**Acceptance criteria**:
- ExtractOutline(md) → []OutlineItem
- ExtractObjectLinks(md) → []string
- ValidateRequiredSections(md, sections) → []Diagnostic

### Issue 3.4: Implement naming engine
**Labels**: `backend`, `naming`, `p0`
**Description**:
Parse/generate/validate ID по Naming Standard v10. Sequence store.
**Acceptance criteria**:
- Parse("a320-prt-0001-v1.0") → ParsedID
- NextID(ctx, "a320", "prt") → "a320-prt-0008-v1.0"
- Validate(id) — reject invalid
- Sequence store: Next, Current, восстановление из objects/
- Тесты: valid IDs, invalid, sequence increment, std IDs

### Issue 3.5: Implement validation engine
**Labels**: `backend`, `validation`, `p0`
**Description**:
Rule registry, Rule interface, Engine (ValidateObject, ValidateProject, ValidateRelease).
**Acceptance criteria**:
- Rule interface: ID() + Check(ctx, target) → []Diagnostic
- Registry: Register/Get
- Engine с базовыми правилами (valid_name, required_metadata)
- Тесты: fixture project

### Issue 3.6: Implement validation rules
**Labels**: `backend`, `validation`, `p1`
**Description**:
Все правила: valid_name, required_metadata, no_broken_relations, checksums_actual, bom_valid, children_released, no_release_blockers.
**Acceptance criteria**:
- Каждое правило — отдельный тип с интерфейсом Rule
- Тесты для каждого правила с готовыми fixture-объектами

---

## Epic 4: SQLite Index

**Labels**: `epic`, `phase-4`, `p0`
**Goal**: Пересобираемый индекс на SQLite
**Depends on**: Epic 1, Epic 2

### Issue 4.1: Implement SQLite schema and migrations
**Labels**: `backend`, `store`, `p0`
**Description**:
Таблицы: objects, relations, artifacts, metadata_projection, bom_projection, fts_objects. Миграции.
**Acceptance criteria**:
- Schema creation при первом запуске
- Миграции: version-based, forward-only
- modernc.org/sqlite (без CGO)

### Issue 4.2: Implement index read/write operations
**Labels**: `backend`, `store`, `p0`
**Description**:
UpsertObject, DeleteObject, поиск, проекции.
**Acceptance criteria**:
- UpsertObject — создаёт/обновляет запись во всех таблицах
- DeleteObject — удаляет и перестраивает связанные проекции
- Интеграционные тесты: create → index → verify

### Issue 4.3: Implement RebuildIndex
**Labels**: `backend`, `store`, `p0`
**Description**:
Полная пересборка индекса из Source of Truth.
**Acceptance criteria**:
- Удаление index.db → RebuildIndex → индекс идентичен исходному
- Тест: 100 объектов, удалить индекс, пересобрать, сравнить

### Issue 4.4: Implement FTS5 search
**Labels**: `backend`, `search`, `p1`
**Description**:
Полнотекстовый поиск по object id, title, class, state, metadata, Markdown body, artifact names.
**Acceptance criteria**:
- Search(ctx, query) → []SearchResult
- Ранжирование по релевантности
- ≤ 200 мс для 10K объектов

---

## Epic 5: FSM Lifecycle

**Labels**: `epic`, `phase-5`, `p0`
**Goal**: Конфигурируемый FSM engine
**Depends on**: Epic 1

### Issue 5.1: Implement process/fsm — pure FSM engine
**Labels**: `backend`, `process`, `p0`
**Description**:
Definition, State, Transition, Machine. Загрузка из YAML.
**Acceptance criteria**:
- Machine.Can(state, transition) → bool
- Machine.Next(state, transition) → (new_state, error)
- Machine.Available(state) → []Transition
- YAML → Definition парсинг
- Не читает файлы, не пишет историю, не вызывает Git

### Issue 5.2: Implement process/guards
**Labels**: `backend`, `process`, `p0`
**Description**:
Guard interface + implementations: valid_name, required_metadata, no_broken_relations, checksums_actual, bom_valid, children_released, no_release_blockers.
**Acceptance criteria**:
- Guard interface: ID() + Check(ctx, GuardInput) → []Diagnostic
- Каждый guard — отдельная реализация
- Тесты с fixture-объектами

### Issue 5.3: Implement process/effects
**Labels**: `backend`, `process`, `p0`
**Description**:
Effect interface + implementations: append_history_event, update_index, create_new_revision, generate_manifest, create_git_tag.
**Acceptance criteria**:
- Effect interface: ID() + Apply(ctx, EffectInput) → error
- Тесты с mock-зависимостями

### Issue 5.4: Implement RunTransition command
**Labels**: `backend`, `app`, `p0`
**Description**:
Application command RunTransition: check guards, execute transition, apply effects, update state, write history.
**Acceptance criteria**:
- Успешный переход: draft → in_review с прохождением guards
- Блокировка: не прошли guards → диагностика в ответе
- Интеграционный тест: draft → in_review → approved → released

---

## Epic 6: Desktop MVP UI

**Labels**: `epic`, `phase-6`, `p1`
**Goal**: Работающее desktop-приложение
**Depends on**: Epics 0-5

### Issue 6.1: Wails 3 setup with Go backend
**Labels**: `backend`, `frontend`, `p1`
**Description**:
Настроить Wails 3, cmd/plm/main.go, app.go, bindings.go. Связать Go backend с React frontend.
**Acceptance criteria**:
- `wails dev` запускает desktop-окно с React UI
- Frontend может вызывать Go backend методы

### Issue 6.2: Landing Page + Create/Open Project
**Labels**: `frontend`, `p0`
**Description**:
Landing Page с выбором Open Project / Create New Project / Recent Projects.
**Acceptance criteria**:
- Folder picker для Open Project
- Create Project Wizard (code, title, path, naming)
- Recent Projects список

### Issue 6.3: Project Tree component
**Labels**: `frontend`, `p0`
**Description**:
Дерево объектов с фильтрацией, поиском, сортировкой, иконками классов, status dots.
**Acceptance criteria**:
- Отображение всех объектов из индекса
- Search/filter (по ID, title, class)
- Keyboard nav (↑↓←→)
- Context menu (Open, Rename, Delete, Validate)

### Issue 6.4: Object Detail Screen + Markdown Editor
**Labels**: `frontend`, `p0`
**Description**:
Vditor (WYSIWYG) + Live Preview Markdown body.
**Acceptance criteria**:
- Vditor с WYSIWYG/instant render/split preview режимами
- Live Preview panel (синхронное обновление)
- Вкладки: Document, Metadata, Relations, BOM, Files, Validation, History

### Issue 6.5: Properties Panel (Right Inspector)
**Labels**: `frontend`, `p0`
**Description**:
Панель свойств: Identity, Lifecycle, Relations summary, Artifacts, Actions.
**Acceptance criteria**:
- Inline редактирование metadata
- Lifecycle actions (доступные переходы)
- Blockers список с suggested actions

### Issue 6.6: Create Object Wizard
**Labels**: `frontend`, `p0`
**Description**:
3-шаговый wizard: Identity → Properties → Documents & Attachments.
**Acceptance criteria**:
- Preview сгенерированного ID
- Выбор класса из classes.yaml
- Прикрепление файлов
- Preview создаваемой структуры перед подтверждением

### Issue 6.7: Command Palette (Ctrl+K)
**Labels**: `frontend`, `p0`
**Description**:
shadcn/ui Command component. Список доступных команд с состоянием enabled/disabled.
**Acceptance criteria**:
- Ctrl+K открывает палитру
- Команды отфильтрованы по контексту
- Disabled команды показывают причину

### Issue 6.8: Git Status / Checkpoint UI
**Labels**: `frontend`, `p1`
**Description**:
Git status в Top Bar, Checkpoint creation dialog.
**Acceptance criteria**:
- Статус: Clean / N changes / Error
- Create Checkpoint dialog с выбором объектов
- Domain-level diff (объекты, metadata, BOM, а не строки)

---

## Epic 7: BOM

**Labels**: `epic`, `phase-7`, `p1`
**Goal**: BOM engine + UI
**Depends on**: Epics 1, 4

### Issue 7.1: Implement modules/bom — BOM service
**Labels**: `backend`, `modules`, `p1`
**Description**:
Structured BOM, flat BOM, quantity rollup, cycle detection, validation.
**Acceptance criteria**:
- GetStructured(ctx, root) → tree BOM
- GetFlat(ctx, root) → flat rows с quantity rollup
- Validate(ctx, root) → diagnostics (cycles, missing quantities)
- Не пишет YAML, не создаёт объекты

### Issue 7.2: BOM View in UI
**Labels**: `frontend`, `p1`
**Description**:
TanStack Table: BOM для выбранной сборки. Колонки: позиция, объект, класс, ревизия, статус, кол-во, ед.изм., make/buy, замены, проблемы.

### Issue 7.3: Where-used View
**Labels**: `frontend`, `p1`
**Description**:
Отображение where-used для выбранного объекта: дерево/таблица всех parents.

---

## Epic 8: Standard Parts

**Labels**: `epic`, `phase-8`, `p1`
**Goal**: Стандартные детали и duplicate detection
**Depends on**: Epics 1-5

### Issue 8.1: Implement modules/standardparts
**Labels**: `backend`, `modules`, `p1`
**Description**:
Normalized key generation, duplicate detection, std validation.
**Acceptance criteria**:
- Normalize(input) → NormalizedPart
- FindDuplicates(input) → []DuplicateCandidate с уровнями
- Validate(id) → diagnostics
- SuggestSubstitutes(id) → []ID

### Issue 8.2: Standard Parts Center UI
**Labels**: `frontend`, `p1`
**Description**:
Каталог стандартных деталей: таблица, поиск, фильтры (std_class, standard, material, supplier), duplicate check при создании.

### Issue 8.3: Duplicate detection UI flow
**Labels**: `frontend`, `p1`
**Description**:
При создании std part — проверка дублей. Exact duplicate → BLOCK + ссылка. Probable duplicate → CONFIRM dialog. Equivalent → предложение создать relation.

---

## Epic 9: Git UX

**Labels**: `epic`, `phase-9`, `p1`
**Goal**: Инженерный Git-интерфейс
**Depends on**: Epic 1

### Issue 9.1: Implement gitops service
**Labels**: `backend`, `gitops`, `p1`
**Description**:
Git status, checkpoint (commit), tag, log, diff через go-git.
**Acceptance criteria**:
- Status() → modified/added/deleted files
- Checkpoint(req) → commit hash
- CreateTag(tag, message)
- Diff(base, head) → domain-aware diff
- Не делает BOM, release readiness, manifest

### Issue 9.2: Git Changes Screen (UI)
**Labels**: `frontend`, `p1`
**Description**:
Инженерное представление git изменений: какие объекты изменились, какие metadata поля, какие файлы.

---

## Epic 10: Release Package

**Labels**: `epic`, `phase-10`, `p1`
**Goal**: Релизный пакет и readiness check
**Depends on**: Epics 1, 4, 5, 7

### Issue 10.1: Implement modules/release
**Labels**: `backend`, `modules`, `p1`
**Description**:
Release scope, readiness check, manifest generation.
**Acceptance criteria**:
- BuildScope(ctx, root) → scope объектов
- CheckReadiness(ctx, scope) → readiness + blockers
- GenerateManifest(ctx, scope) → manifest YAML

### Issue 10.2: Release Dashboard UI
**Labels**: `frontend`, `p1`
**Description**:
Dashboard готовности к релизу: readiness summary, список blockers, кнопка Release.

---

## Epic 11: Manufacturing / Procurement (Post-MVP)

**Labels**: `epic`, `phase-11`, `p2`
**Goal**: Производственная готовность и закупки
**Depends on**: Epics 1, 4, 7

### Issue 11.1: Manufacturing matrix
### Issue 11.2: Procurement buy list
### Issue 11.3: Substitution workflow
### Issue 11.4: ERP export

---

## Epic 12: Change Impact (Post-MVP)

**Labels**: `epic`, `phase-12`, `p2`

### Issue 12.1: Change impact radar
### Issue 12.2: Change request FSM

---

## Epic 13: Hardening

**Labels**: `epic`, `phase-13`, `p2`
**Goal**: Стабильность, производительность, документация

### Issue 13.1: Performance optimization
**Labels**: `backend`, `p2`
**Description**:
Оптимизация под 10K объектов: lazy loading tree, index query tuning, pagination.

### Issue 13.2: E2E tests (Playwright)
**Labels**: `testing`, `p2`
**Description**:
Сценарии: create project → create part → edit → submit review → approve → release → verify Git tag.

### Issue 13.3: User and developer documentation
**Labels**: `docs`, `p2`
**Description**:
User guide, developer guide, API reference (godoc).

### Issue 13.4: Bug fixes and stabilization
**Labels**: `bug`, `p2`
