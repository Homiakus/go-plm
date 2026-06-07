# Module Decomposition Specification
# go-plm / Markdown-PLM

Версия: `1.0-draft`  
Статус: `основной архитектурный документ`  
Назначение: определить разбиение Go-backend на изолированные атомарные модули, где каждый пакет выполняет строго одну функцию и имеет идиоматичное Go-поведение.

---

## 1. Главная цель

`go-plm` должен быть построен как набор маленьких Go-пакетов с явными границами ответственности.

Каждый модуль обязан:

- выполнять одну понятную функцию;
- быть тестируемым отдельно;
- не иметь циклических импортов;
- зависеть от интерфейсов, а не от конкретной инфраструктуры;
- не писать файлы напрямую, если это не `store/transaction`;
- не обращаться к UI;
- не хранить глобальное изменяемое состояние;
- явно документировать, что он делает и чего не делает.

Главная доменная формула:

```text
Object + Relation + Artifact + Process + Validation + Event + Projection
```

Главная техническая формула:

```text
Markdown/YAML = source of truth
Go modules = domain logic
SQLite = projections/index
Git = history/checkpoints
Frontend = UI only
```

---

## 2. Архитектурный стиль

Использовать:

```text
Clean Architecture lite
+ Hexagonal Architecture
+ Go package-oriented design
+ Command / Query separation
```

Смысл:

- `core` не знает о файлах, SQLite, Git, Wails, React;
- `app` координирует пользовательские сценарии;
- `store` реализует хранение;
- `modules` реализуют предметные функции;
- `api` только адаптирует frontend-вызовы к application layer.

---

## 3. Правила зависимостей

Разрешенное направление:

```text
cmd
 ↓
api
 ↓
app
 ↓
core

app → modules
app → store interfaces
app → validation
app → process
app → gitops interfaces

modules → core
modules → local interfaces
modules → validation contracts

store → core
parser → core
validation → core
process → core
```

Запрещено:

```text
core → app
core → store
core → api
core → gitops
core → sqlite
core → filesystem
core → frontend

modules → api
modules → Wails
modules → React
modules → concrete SQLite implementation
modules → concrete filesystem transaction implementation

parser → app
validation → app
gitops → app
```

Пример правильного импорта:

```go
package bom

import (
    "context"

    "github.com/homiakus/go-plm/internal/core/object"
    "github.com/homiakus/go-plm/internal/core/relation"
)
```

Пример плохого импорта:

```go
package bom

import (
    "github.com/homiakus/go-plm/internal/api"
    "github.com/homiakus/go-plm/internal/store/sqlite"
    "github.com/wailsapp/wails/v2/pkg/runtime"
)
```

---

## 4. Целевая структура репозитория

```text
go-plm/
├── cmd/
│   └── plm/
│       ├── main.go
│       ├── app.go
│       └── bindings.go
├── internal/
│   ├── core/
│   │   ├── object/
│   │   ├── relation/
│   │   ├── artifact/
│   │   ├── revision/
│   │   ├── event/
│   │   ├── diagnostic/
│   │   ├── lifecycle/
│   │   └── project/
│   ├── app/
│   │   ├── command/
│   │   ├── query/
│   │   ├── service/
│   │   ├── bus/
│   │   └── ports/
│   ├── parser/
│   │   ├── frontmatter/
│   │   ├── markdown/
│   │   └── yamlx/
│   ├── naming/
│   ├── store/
│   │   ├── fsrepo/
│   │   ├── transaction/
│   │   ├── index/
│   │   └── recovery/
│   ├── validation/
│   │   ├── engine/
│   │   ├── rules/
│   │   └── registry/
│   ├── process/
│   │   ├── fsm/
│   │   ├── guards/
│   │   └── effects/
│   ├── gitops/
│   ├── search/
│   │   ├── sqlitefts/
│   │   └── bleve/
│   ├── modules/
│   │   ├── bom/
│   │   ├── release/
│   │   ├── standardparts/
│   │   ├── procurement/
│   │   ├── manufacturing/
│   │   ├── change/
│   │   └── shopfloor/
│   └── api/
│       ├── dto/
│       ├── mapper/
│       └── wails/
├── pkg/
│   └── plmclient/
├── frontend/
├── examples/
├── templates/
└── docs/
```

---

## 5. `cmd/plm` — entrypoint

### Назначение

`cmd/plm` только запускает приложение.

Отвечает за:

```text
инициализацию зависимостей;
создание application container;
подключение Wails;
регистрацию bindings;
запуск приложения.
```

Не отвечает за:

```text
создание объектов;
валидацию;
BOM;
релизы;
Git-логику;
парсинг YAML;
файловые транзакции.
```

Идиоматичное поведение:

```go
func main() {
    container, err := bootstrap.NewContainer()
    if err != nil {
        log.Fatal(err)
    }

    if err := desktop.Run(container); err != nil {
        log.Fatal(err)
    }
}
```

Запреты:

```text
не реализовывать бизнес-логику в main.go;
не писать файлы из cmd;
не импортировать modules напрямую из UI, минуя app layer.
```

---

## 6. `internal/core` — доменное ядро

### Назначение

`core` содержит чистую доменную модель.

Содержит:

```text
Object
Relation
Artifact
Revision
Event
Diagnostic
LifecycleState
Project
ObjectClass
```

`core` не должен знать:

```text
как объект хранится на диске;
что такое SQLite;
что такое Git;
что такое Wails;
что такое React;
как выглядит UI;
как устроен индекс.
```

---

### 6.1. `core/object`

Ответственность: инженерный объект.

```go
package object

type ID string
type Class string
type State string

type Object struct {
    ID        ID
    Project   string
    Class     Class
    Sequence  string
    Version   string
    Revision  string
    State     State
    Title     string
    Metadata  map[string]any
    Relations []RelationRef
    Artifacts []ArtifactRef
}
```

Разрешенные методы:

```go
func (o Object) IsZero() bool
func (o Object) IsReleased() bool
func (o Object) IsDraft() bool
func (o Object) WithState(state State) Object
func (o Object) ValidateIdentity() error
```

Запрещено:

```text
читать Markdown;
писать YAML;
обращаться к SQLite;
вызывать Git;
строить BOM;
валидировать весь проект.
```

---

### 6.2. `core/relation`

Ответственность: связь между объектами.

```go
package relation

type Type string

const (
    Contains             Type = "contains"
    HasDrawing           Type = "has_drawing"
    ManufacturingFileFor Type = "manufacturing_file_for"
    Substitutes          Type = "substitutes"
    Affects              Type = "affects"
    Includes             Type = "includes"
)

type Relation struct {
    FromID   string
    ToID     string
    Type     Type
    Quantity *float64
    Unit     string
    Metadata map[string]any
}
```

Разрешено:

```go
func (r Relation) IsBOMRelation() bool
func (r Relation) HasQuantity() bool
func (r Relation) ValidateShape() error
```

Запрещено:

```text
проверять наличие target object в файловой системе;
строить where-used;
обновлять индекс.
```

---

### 6.3. `core/artifact`

Ответственность: описание файла.

```go
package artifact

type Kind string
type Role string
type Status string

type Artifact struct {
    ID           string
    Kind         Kind
    Role         Role
    Path         string
    OriginalName string
    Checksum     string
    SizeBytes    int64
    Generated    bool
    Required     bool
    Status       Status
}
```

Запрещено:

```text
копировать файл;
рассчитывать checksum;
открывать файл;
валидировать путь на диске.
```

---

### 6.4. `core/event`

Ответственность: событие истории.

```go
package event

type Type string

const (
    ObjectCreated       Type = "object.created"
    MetadataUpdated     Type = "metadata.updated"
    RelationAdded       Type = "relation.added"
    ArtifactAttached    Type = "artifact.attached"
    LifecycleTransition Type = "lifecycle.transitioned"
)

type Event struct {
    ID       string
    Time     time.Time
    Actor    string
    Type     Type
    ObjectID string
    Payload  map[string]any
}
```

Event-модуль не пишет `history.yaml.log`. Он только описывает событие.

---

### 6.5. `core/diagnostic`

Ответственность: диагностика и blockers.

```go
package diagnostic

type Severity string

const (
    Info    Severity = "info"
    Warning Severity = "warning"
    Blocker Severity = "blocker"
    Error   Severity = "error"
)

type Diagnostic struct {
    ID               string
    Severity         Severity
    Code             string
    ObjectID         string
    Path             string
    Message          string
    SuggestedActions []string
}
```

---

## 7. `internal/app` — application layer

### Назначение

`app` реализует пользовательские сценарии.

Отвечает за:

```text
commands;
queries;
оркестрацию модулей;
транзакционные сценарии;
сборку ответов для UI;
координацию store, validation, FSM, Git.
```

---

### 7.1. `app/command`

Команды меняют состояние.

Примеры:

```text
CreateProject
OpenProject
CreateObject
UpdateObjectDocument
UpdateObjectMetadata
AttachArtifact
AddRelation
RemoveRelation
RunTransition
CreateReleasePackage
CreateCheckpoint
RebuildIndex
```

Идиоматичный интерфейс:

```go
type Handler[Req any, Res any] interface {
    Handle(ctx context.Context, req Req) (Res, error)
}
```

Пример request/response:

```go
type CreateObjectRequest struct {
    ProjectPath string
    Class       string
    Title       string
    Metadata    map[string]any
    Context     CreateObjectContext
}

type CreateObjectResponse struct {
    ObjectID    string
    Diagnostics []diagnostic.Diagnostic
    Refresh     RefreshHints
}
```

Правила:

```text
команда принимает context.Context;
команда возвращает typed response;
fatal errors идут через error;
предметные проблемы идут через diagnostics;
команда не обращается к frontend.
```

---

### 7.2. `app/query`

Queries читают, но не меняют source of truth.

Примеры:

```text
GetObject
ListObjects
SearchObjects
GetBOM
GetWhereUsed
GetChangeImpact
ValidateObject
GetGitStatus
GetIndexStatus
GetReleaseReadiness
```

Правила:

```text
query не пишет Markdown/YAML;
query не создает Git commit;
query не меняет lifecycle;
query может читать индекс;
query может запускать transient validation без записи.
```

---

### 7.3. `app/service`

Application services координируют несколько модулей.

```go
type ObjectService struct {
    Repo       ObjectRepository
    Tx         TransactionManager
    Naming     NamingService
    Templates  TemplateRenderer
    Validator  ValidationService
    Index      IndexUpdater
    Events     EventWriter
}
```

Правило: service зависит от интерфейсов, а не от конкретных реализаций.

---

### 7.4. `app/ports`

Порты — интерфейсы инфраструктуры.

```go
type ObjectRepository interface {
    Get(ctx context.Context, id object.ID) (object.Object, error)
    Save(ctx context.Context, obj object.Object) error
    List(ctx context.Context, filter ObjectFilter) ([]object.Object, error)
}

type TransactionManager interface {
    Run(ctx context.Context, name string, fn func(ctx context.Context) error) error
}

type Indexer interface {
    UpdateObject(ctx context.Context, obj object.Object) error
    Rebuild(ctx context.Context) error
}
```

---

## 8. `internal/parser`

### 8.1. `parser/frontmatter`

Ответственность: разделить Markdown на YAML frontmatter и body.

```go
type Document struct {
    Frontmatter []byte
    Body        []byte
}

func Split(src []byte) (Document, error)
func Join(frontmatter []byte, body []byte) []byte
```

Не делает:

```text
business validation;
поиск объекта;
запись файла.
```

---

### 8.2. `parser/yamlx`

Ответственность: YAML encode/decode.

```go
func Decode[T any](data []byte) (T, error)
func Encode(v any) ([]byte, error)
func DecodeStrict[T any](data []byte) (T, error)
```

Правила:

```text
возвращать line/column, если возможно;
не делать graph validation;
не смешивать parsing и business rules.
```

---

### 8.3. `parser/markdown`

Ответственность: backend-парсинг Markdown.

Используется для:

```text
outline;
поиска object links;
валидации обязательных секций;
рендера preview/export;
извлечения директив @bom, @object, @artifact.
```

API:

```go
type OutlineItem struct {
    Level int
    Text  string
    Slug  string
}

func ExtractOutline(md []byte) ([]OutlineItem, error)
func ExtractObjectLinks(md []byte) ([]string, error)
func ValidateRequiredSections(md []byte, sections []string) []diagnostic.Diagnostic
```

---

## 9. `internal/naming`

### Назначение

Генерация и проверка ID.

Отвечает за:

```text
project-class-sequence-vmajor.minor;
std naming;
kebab-case;
sequence allocation;
ID parsing;
validation against naming standard.
```

API:

```go
type ParsedID struct {
    Project  string
    Class    string
    Sequence string
    Version  string
}

type Generator interface {
    NextID(ctx context.Context, project string, class string) (string, error)
    Parse(id string) (ParsedID, error)
    Validate(id string) error
}
```

Sequence store:

```go
type SequenceStore interface {
    Next(ctx context.Context, class string) (string, error)
    Current(ctx context.Context, class string) (string, error)
}
```

Запрещено:

```text
создавать файлы;
обновлять project.yaml напрямую;
обновлять index.db;
читать весь проект без интерфейса.
```

---

## 10. `internal/store`

### 10.1. `store/fsrepo`

Ответственность: файловый репозиторий.

Делает:

```text
читает project.yaml;
читает объекты;
читает Markdown;
читает history.yaml.log;
находит object folders;
сохраняет объекты через transaction layer.
```

API:

```go
type Repository struct {
    Root string
    Tx   *transaction.Manager
}

func (r *Repository) GetObject(ctx context.Context, id object.ID) (object.Object, error)
func (r *Repository) SaveObject(ctx context.Context, obj object.Object) error
func (r *Repository) ListObjects(ctx context.Context) ([]object.Object, error)
```

Не делает:

```text
BOM;
FSM transitions;
Git commit;
complex validation;
UI DTO.
```

---

### 10.2. `store/transaction`

Ответственность: атомарные файловые операции.

Делает:

```text
transaction plan;
temp files;
atomic rename;
fsync;
rollback markers;
recovery metadata;
idempotency.
```

API:

```go
type Manager struct {
    Root string
}

type Plan struct {
    ID      string
    Writes  []WriteOp
    Copies  []CopyOp
    Deletes []DeleteOp
    Renames []RenameOp
}

func (m *Manager) Execute(ctx context.Context, plan Plan) error
func (m *Manager) Recover(ctx context.Context) ([]RecoveryItem, error)
```

Правила:

```text
никакой другой модуль не делает опасный overwrite;
все записи source of truth идут через transaction;
operation_id обеспечивает idempotency.
```

---

### 10.3. `store/index`

Ответственность: SQLite index/projections.

Делает:

```text
objects table;
relations table;
artifacts table;
metadata projection;
BOM projection;
where-used projection;
standard parts normalized keys;
search FTS;
validation summaries.
```

API:

```go
type Index interface {
    UpsertObject(ctx context.Context, obj object.Object) error
    DeleteObject(ctx context.Context, id object.ID) error
    Rebuild(ctx context.Context, source SourceReader) error
    Search(ctx context.Context, query string) ([]SearchResult, error)
}
```

Запрещено:

```text
считать SQLite источником истины;
хранить непересобираемые данные;
менять Markdown/YAML.
```

---

### 10.4. `store/recovery`

Ответственность: восстановление.

Делает:

```text
находит незавершенные транзакции;
диагностирует поврежденные объекты;
предлагает repair actions;
делает rollback/complete.
```

---

## 11. `internal/validation`

### Назначение

Единая система проверок.

Виды:

```text
syntax;
schema;
naming;
metadata;
relations;
artifacts;
BOM;
lifecycle;
release;
standard parts;
project integrity.
```

Интерфейс правила:

```go
type Rule interface {
    ID() string
    Check(ctx context.Context, target Target) ([]diagnostic.Diagnostic, error)
}
```

Registry:

```go
type Registry struct {
    rules map[string]Rule
}

func (r *Registry) Register(rule Rule)
func (r *Registry) Get(id string) (Rule, bool)
```

Engine:

```go
type Engine struct {
    Registry *Registry
}

func (e *Engine) ValidateObject(ctx context.Context, id object.ID) ([]diagnostic.Diagnostic, error)
func (e *Engine) ValidateProject(ctx context.Context) ([]diagnostic.Diagnostic, error)
func (e *Engine) ValidateRelease(ctx context.Context, id object.ID) ([]diagnostic.Diagnostic, error)
```

Запрещено:

```text
validation не исправляет данные сама;
validation не пишет файлы;
validation не создает Git commit.
```

Quick fix — отдельная application command.

---

## 12. `internal/process`

### 12.1. `process/fsm`

Ответственность: чистый FSM.

```go
type Machine struct {
    Definition Definition
}

type Definition struct {
    Name        string
    Initial     string
    States      []State
    Transitions []Transition
}

type Transition struct {
    Name    string
    From    []string
    To      string
    Guards  []string
    Effects []string
}
```

API:

```go
func (m Machine) Can(state string, transition string) bool
func (m Machine) Next(state string, transition string) (string, error)
func (m Machine) Available(state string) []Transition
```

FSM не читает файлы, не пишет history и не вызывает Git.

---

### 12.2. `process/guards`

Ответственность: guard adapters.

```go
type Guard interface {
    ID() string
    Check(ctx context.Context, input GuardInput) ([]diagnostic.Diagnostic, error)
}
```

Примеры:

```text
required_metadata;
valid_name;
bom_valid;
children_released;
checksums_actual;
standard_parts_approved;
no_release_blockers.
```

---

### 12.3. `process/effects`

Ответственность: эффекты переходов.

```go
type Effect interface {
    ID() string
    Apply(ctx context.Context, input EffectInput) error
}
```

Примеры:

```text
append_history_event;
update_index;
create_new_revision;
generate_manifest;
create_git_tag.
```

Эффекты вызываются application layer внутри transaction boundary.

---

## 13. `internal/gitops`

### Назначение

Изолирует Git.

Делает:

```text
status;
changed files;
checkpoint;
log;
tag;
diff;
compare refs.
```

API:

```go
type Service interface {
    Status(ctx context.Context) (Status, error)
    Checkpoint(ctx context.Context, req CheckpointRequest) (CheckpointResult, error)
    CreateTag(ctx context.Context, tag string, message string) error
    Log(ctx context.Context, limit int) ([]Commit, error)
    Diff(ctx context.Context, base string, head string) (Diff, error)
}
```

Не делает:

```text
BOM;
release readiness;
manifest generation;
Markdown parsing;
material validation.
```

---

## 14. `internal/search`

### 14.1. `search/sqlitefts`

MVP full-text search.

Индексирует:

```text
object id;
title;
class;
state;
metadata;
Markdown headings;
Markdown body;
artifact names;
standard part normalized key;
supplier/manufacturer numbers.
```

### 14.2. `search/bleve`

Будущий advanced search.

Используется для:

```text
fuzzy search;
facets;
highlighting;
duplicate candidates;
advanced standard parts search.
```

Индекс должен быть пересобираемым.

---

## 15. `internal/modules/bom`

### Назначение

BOM-модуль строит eBOM, flat BOM и проверяет BOM.

Делает:

```text
читает contains relations;
строит structured BOM;
строит flat BOM;
quantity rollup;
cycle detection;
BOM row validation;
BOM export model.
```

Не делает:

```text
создание объекта;
запись YAML;
release package;
procurement rules;
UI table rendering.
```

API:

```go
type Service struct {
    Objects   ObjectReader
    Relations RelationReader
}

func (s *Service) GetStructured(ctx context.Context, root object.ID) (StructuredBOM, error)
func (s *Service) GetFlat(ctx context.Context, root object.ID) (FlatBOM, error)
func (s *Service) Validate(ctx context.Context, root object.ID) ([]diagnostic.Diagnostic, error)
```

Правило: изменение BOM делается через relation commands, а не через `bom.Service`.

---

## 16. `internal/modules/standardparts`

### Назначение

Стандартные детали.

Делает:

```text
std ID support;
normalized key;
duplicate detection;
classification;
supplier data validation;
substitution relations;
where-used для std;
procurement projection support.
```

Не делает:

```text
UI table;
низкоуровневое сохранение файла;
Git commit;
общий BOM traversal, кроме std-specific logic.
```

API:

```go
type Service struct {
    Objects ObjectReader
    Index   StandardPartIndex
}

func (s *Service) Normalize(ctx context.Context, input StandardPartInput) (NormalizedPart, error)
func (s *Service) FindDuplicates(ctx context.Context, input StandardPartInput) ([]DuplicateCandidate, error)
func (s *Service) Validate(ctx context.Context, id object.ID) ([]diagnostic.Diagnostic, error)
func (s *Service) SuggestSubstitutes(ctx context.Context, id object.ID) ([]object.ID, error)
```

Duplicate levels:

```go
type DuplicateLevel string

const (
    ExactDuplicate     DuplicateLevel = "exact_duplicate"
    ProbableDuplicate  DuplicateLevel = "probable_duplicate"
    PossibleEquivalent DuplicateLevel = "possible_equivalent"
    Substitutable      DuplicateLevel = "substitutable"
    NotDuplicate       DuplicateLevel = "not_duplicate"
)
```

Правила:

```text
exact duplicate блокирует создание;
probable duplicate требует подтверждения;
possible equivalent не объединяется автоматически;
substitution требует relation/workflow.
```

---

## 17. `internal/modules/release`

### Назначение

Release package.

Делает:

```text
release scope;
release readiness;
manifest generation;
included objects list;
included files list;
checksums list;
release export model.
```

Не делает:

```text
Git tag напрямую;
низкоуровневую запись файлов;
все validation rules самостоятельно;
UI rendering.
```

API:

```go
type Service struct {
    Objects   ObjectReader
    Relations RelationReader
    Artifacts ArtifactReader
    Validator ReleaseValidator
}

func (s *Service) BuildScope(ctx context.Context, root object.ID) (Scope, error)
func (s *Service) CheckReadiness(ctx context.Context, scope Scope) (Readiness, error)
func (s *Service) GenerateManifest(ctx context.Context, scope Scope) (Manifest, error)
```

Git tag вызывается application layer после успешного release.

---

## 18. `internal/modules/procurement`

### Назначение

Закупки и buy list.

Делает:

```text
make/buy filtering;
buy list generation;
supplier selection;
order quantity calculation;
ERP export DTO;
standard parts procurement rollup.
```

API:

```go
func (s *Service) GenerateBuyList(ctx context.Context, root object.ID) (BuyList, error)
func (s *Service) ValidateProcurement(ctx context.Context, root object.ID) ([]diagnostic.Diagnostic, error)
func (s *Service) ExportERP(ctx context.Context, list BuyList) ([]byte, error)
```

Не делает:

```text
создание standard part;
supplier lifecycle approval;
низкоуровневый CSV writer, если он вынесен в exporter.
```

---

## 19. `internal/modules/manufacturing`

### Назначение

Производственная готовность.

Делает:

```text
manufacturing matrix;
process plan references;
work instruction readiness;
cut/bnd/nc/ins readiness;
operation readiness;
manufacturing blockers.
```

API:

```go
func (s *Service) Matrix(ctx context.Context, root object.ID) (Matrix, error)
func (s *Service) ValidateReadiness(ctx context.Context, id object.ID) ([]diagnostic.Diagnostic, error)
```

Не делает:

```text
shop-floor execution;
evidence capture;
build record mutation.
```

---

## 20. `internal/modules/change`

### Назначение

Change impact.

Делает:

```text
where-used expansion;
affected assemblies;
affected releases;
affected instructions;
affected procurement;
impact report model.
```

API:

```go
func (s *Service) AnalyzeImpact(ctx context.Context, id object.ID) (ImpactReport, error)
func (s *Service) AffectedObjects(ctx context.Context, id object.ID) ([]object.ID, error)
```

Не делает:

```text
создание CR объекта напрямую;
изменение released объекта;
Git operations.
```

CR создается application command на основе impact report.

---

## 21. `internal/modules/shopfloor`

### Назначение

Будущий модуль выполнения производственных инструкций.

Делает:

```text
build record model;
step execution;
evidence references;
deviation model;
completion report.
```

Не входит в MVP.

Разделение:

```text
manufacturing = определяет готовность и инструкции;
shopfloor = фиксирует факт выполнения.
```

---

## 22. `internal/api`

### Назначение

Адаптация backend к frontend.

Содержит:

```text
DTO;
mappers;
Wails bindings;
error mapping;
response envelopes.
```

Не содержит:

```text
domain logic;
BOM algorithms;
file writes;
validation rules;
Git operations.
```

DTO пример:

```go
type ObjectDTO struct {
    ID       string         `json:"id"`
    Class    string         `json:"class"`
    Title    string         `json:"title"`
    State    string         `json:"state"`
    Metadata map[string]any `json:"metadata"`
}
```

Mapper:

```go
func ObjectToDTO(obj object.Object) ObjectDTO
func DTOToCreateObjectRequest(dto CreateObjectDTO) command.CreateObjectRequest
```

Правило:

```text
DTO не проникает в core/modules.
```

---

## 23. Ошибки и diagnostics

### Error vs Diagnostic

`error` — технический/fatal failure:

```text
file not found;
permission denied;
invalid YAML syntax preventing parsing;
SQLite unavailable;
transaction failed.
```

`Diagnostic` — предметная проблема:

```text
missing material;
draft child object;
checksum outdated;
standard part not approved;
BOM cycle.
```

Идиоматично:

```go
if err != nil {
    return Response{}, fmt.Errorf("create object: %w", err)
}
```

Sentinel errors:

```go
var (
    ErrObjectNotFound = errors.New("object not found")
    ErrInvalidID      = errors.New("invalid object id")
    ErrConflict       = errors.New("conflict")
)
```

Проверка:

```go
if errors.Is(err, object.ErrInvalidID) {
    // map to API response
}
```

---

## 24. Context usage

Все публичные операции backend принимают `context.Context`.

```go
func (h *CreateObjectHandler) Handle(ctx context.Context, req CreateObjectRequest) (CreateObjectResponse, error)
```

Используется для:

```text
cancel;
timeout;
progress;
user cancellation;
background task cancellation.
```

Запрещено:

```text
долгие операции без context;
игнорировать ctx.Done() при обходе больших проектов.
```

---

## 25. Логирование

Правила:

```text
логировать технические события;
не логировать секреты;
не засорять domain code логгером;
передавать logger через service/container;
не использовать fmt.Println в библиотеках.
```

Уровни:

```text
debug;
info;
warn;
error.
```

---

## 26. Конфигурация

Файлы:

```text
config/classes.yaml
config/naming.yaml
config/lifecycle.yaml
config/validation.yaml
config/relation-types.yaml
config/file-routing.yaml
config/exports.yaml
```

Каждый config loader:

```text
парсит YAML;
валидирует схему;
возвращает typed config;
дает diagnostics с путями.
```

---

## 27. Общие интерфейсы между модулями

```go
type ObjectReader interface {
    GetObject(ctx context.Context, id object.ID) (object.Object, error)
}

type RelationReader interface {
    Outgoing(ctx context.Context, id object.ID) ([]relation.Relation, error)
    Incoming(ctx context.Context, id object.ID) ([]relation.Relation, error)
}

type ArtifactReader interface {
    ListArtifacts(ctx context.Context, id object.ID) ([]artifact.Artifact, error)
}

type EventWriter interface {
    Append(ctx context.Context, id object.ID, evt event.Event) error
}
```

Интерфейсы объявляются у потребителя.

---

## 28. Правило атомарности модулей

Правильно:

```text
bom строит BOM, но не пишет файлы;
release генерирует manifest model, но не делает Git tag;
gitops делает Git, но не знает о release readiness;
standardparts ищет дубли, но не рисует таблицу;
validation выдает diagnostics, но не исправляет файлы;
transaction пишет файлы, но не знает о BOM.
```

Неправильно:

```text
bom.Service сам меняет YAML;
standardparts.Service делает Git commit;
validation.Rule создает новый объект;
api.Binding строит flat BOM вручную;
fsrepo вызывает Wails runtime;
gitops проверяет material metadata.
```

---

## 29. Module registration

Модули могут регистрировать capabilities.

```go
type Module interface {
    Name() string
    Register(reg Registry) error
}
```

Registry принимает:

```text
commands;
queries;
validation rules;
guards;
effects;
index projections;
templates;
exporters.
```

Пример:

```go
type BOMModule struct{}

func (m BOMModule) Name() string { return "bom" }

func (m BOMModule) Register(reg Registry) error {
    reg.RegisterQuery("bom.get", NewGetBOMHandler(...))
    reg.RegisterRule(bom.NewNoCyclesRule(...))
    return nil
}
```

---

## 30. Идиоматичные Go-правила проекта

### Маленькие предметные пакеты

Хорошо:

```text
internal/modules/bom
internal/modules/release
internal/naming
internal/parser/frontmatter
```

Плохо:

```text
internal/utils
internal/common
internal/helpers
internal/manager
```

### Имена интерфейсов

Хорошо:

```go
type Reader interface {}
type Writer interface {}
type Validator interface {}
type Renderer interface {}
type Repository interface {}
```

Плохо:

```go
type IObjectService interface {}
type ObjectServiceInterface interface {}
```

### Интерфейсы у потребителя

```go
package bom

type ObjectReader interface {
    GetObject(ctx context.Context, id object.ID) (object.Object, error)
}
```

### Не создавать преждевременные generic-абстракции

Хорошо:

```go
type ObjectRepository interface {
    GetObject(...)
    SaveObject(...)
}
```

Плохо:

```go
type Repository[T any] interface {
    Get(id string) (T, error)
}
```

если generic не дает пользы.

### Ошибки оборачивать

```go
return fmt.Errorf("read object %s: %w", id, err)
```

### Не паниковать

Запрещено в библиотечном коде:

```go
panic(err)
```

### Не использовать глобальное изменяемое состояние

Запрещено:

```go
var CurrentProject string
var GlobalRepo *Repository
```

### Context первым параметром

```go
func (s *Service) GetBOM(ctx context.Context, root object.ID) (BOM, error)
```

### Не смешивать DTO и domain

Плохо:

```go
func (s *BOMService) Build(dto api.ObjectDTO) {}
```

Хорошо:

```go
func (s *BOMService) Build(ctx context.Context, root object.ID) {}
```

---

## 31. Backend response envelope

Для UI команды возвращают единый формат.

```go
type Response[T any] struct {
    OK          bool                    `json:"ok"`
    Result      T                       `json:"result,omitempty"`
    Diagnostics []diagnostic.Diagnostic `json:"diagnostics,omitempty"`
    Refresh     RefreshHints            `json:"refresh,omitempty"`
    Error       *APIError               `json:"error,omitempty"`
}
```

Refresh hints:

```go
type RefreshHints struct {
    Objects     []string `json:"objects,omitempty"`
    Projections []string `json:"projections,omitempty"`
    GitStatus   bool     `json:"git_status,omitempty"`
    IndexStatus bool     `json:"index_status,omitempty"`
}
```

---

## 32. Модульная карта MVP

### P0

```text
core/object
core/relation
core/artifact
core/event
core/diagnostic
parser/frontmatter
parser/yamlx
naming
store/transaction
store/fsrepo
app/command
app/query
api/dto
```

### P1

```text
validation
process/fsm
process/guards
store/index
search/sqlitefts
gitops
modules/bom
```

### P2

```text
modules/standardparts
modules/release
modules/procurement
modules/manufacturing
modules/change
```

### P3

```text
modules/shopfloor
search/bleve
advanced exporters
plugin registry
```

---

## 33. Тестирование модулей

### Unit tests

```text
core/object → identity validation;
naming → parse/generate;
bom → flat BOM, cycle detection;
standardparts → normalized key, duplicate detection;
validation → rules;
fsm → transitions;
transaction → atomic write plan;
parser → frontmatter split/join.
```

### Integration tests

```text
CreateObject → fsrepo → parser → index;
AddRelation → BOM projection;
RunTransition → FSM → guards → history;
CreateRelease → release → validation → git tag.
```

### Fixtures

```text
examples/fixtures/simple-project
examples/fixtures/broken-yaml
examples/fixtures/bom-cycle
examples/fixtures/standard-parts-duplicates
examples/fixtures/release-ready
```

---

## 34. Package documentation

Каждый пакет должен иметь `doc.go`.

Пример:

```go
// Package bom builds engineering and flat BOM projections from relation graphs.
//
// The package is read-only: it does not mutate project source files.
// BOM mutations are performed by application commands that add/update/remove
// contains relations.
package bom
```

---

## 35. Acceptance criteria for modularity

Архитектура принята, если:

```text
нет циклических импортов;
core не импортирует инфраструктуру;
api не содержит domain algorithms;
каждый модуль имеет package doc;
каждый модуль имеет unit tests;
каждая команда имеет typed request/response;
каждый query не мутирует source of truth;
filesystem writes идут только через transaction layer;
SQLite можно удалить и пересобрать;
Git вызывается только через gitops;
UI работает только через api/app layer.
```

---

## 36. Финальное правило

Каждый модуль должен отвечать на три вопроса:

```text
1. Что я делаю?
2. Чего я принципиально не делаю?
3. Через какие интерфейсы я общаюсь с остальной системой?
```

Если модуль не может четко ответить на эти вопросы, его нужно разделить.

Финальная карта ответственности:

```text
core = что такое PLM-данные
app = что пользователь хочет сделать
modules = предметные возможности
store = как данные лежат
parser = как данные читаются
validation = почему данные корректны/некорректны
process = как меняются состояния
gitops = как фиксируется история
api = как frontend вызывает backend
```
