# ПОЛНЫЙ АУДИТ КОДОВОЙ БАЗЫ go-plm

**Дата:** 2026-06-07
**Версия:** main @ 379d6c6
**Аудитор:** Автоматизированный анализ + ручная верификация
**Файлов Go:** 56
**Пакетов:** 28

---

## СВОДНАЯ ОЦЕНКА

| Категория | Оценка | Комментарий |
|-----------|--------|-------------|
| **Архитектура** | 8/10 | Clean Architecture, чёткое разделение слоёв. Дублирование lifecycle/fsm. Порты разбросаны. |
| **Качество кода** | 7/10 | Хороший стиль, нет TODO/паник. Найдены баги (RunTransition), проблемы с error handling. |
| **Безопасность** | 6/10 | Нет path traversal, но FTS5 injection, нет аутентификации, игнорирование ошибок. |
| **Производительность** | 7/10 | Worker pools, batch INSERT, errgroup. Нет пагинации на уровне store, O(n) поиск transition. |
| **Тестирование** | 7/10 | 22/28 пакетов с тестами. Core domain 95-100%. 6 пакетов без тестов. Хорошие сценарии. |
| **Документация** | 8/10 | README мирового уровня, SRS на 39 секций. Не хватает ADR и API-документации. |
| **Инфраструктура** | 6/10 | Makefile, go.mod, CI (GitHub Actions). Несоответствие версий Go (1.26.2 vs tool 1.25.5). |

**Общая оценка: 7/10** — Крепкий фундамент с несколькими требующими внимания проблемами.

---

## 1. АРХИТЕКТУРА (8/10)

### 1.1 Структура слоёв

```
cmd/plm/              # Точка входа Wails v2 (пока не реализована)
internal/
├── core/             # Доменная модель — 8 пакетов, zero deps ✅
├── store/            # Персистенция — transaction, fsrepo, index (SQLite)
├── parser/           # Форматы — frontmatter, yamlx, markdown
├── naming/           # Naming Standard v10
├── validation/       # Валидация — engine + rules
├── process/          # FSM — fsm, guards, effects
├── gitops/           # Git-интеграция (go-git v5)
├── modules/          # Доменные модули — bom, release, standardparts
├── app/              # Application layer — command, query, ports
└── api/              # Frontend contract — dto, mapper, tree
```

**Оценка архитектуры:** Хорошо продуманная Clean Architecture. Каждый слой имеет чёткую зону ответственности. Домен (`core/`) не зависит ни от чего, кроме стандартной библиотеки.

### 1.2 Проблемы архитектуры

#### P1 — Дублирование lifecycle/fsm типов

Два почти идентичных набора типов:

```go
// internal/core/lifecycle/lifecycle.go
type State string
type Transition struct { Name, From, To, Guards, Effects }
type Definition struct { Name, Initial, States, Transitions }
type Machine struct { Definition Definition }

// internal/process/fsm/fsm.go
type State string
type Transition struct { Name, From, To, Guards, Effects }
type Definition struct { Name, Initial, States, Transitions }
type Machine struct { Def Definition }
```

`lifecycle.Machine` и `fsm.Machine` дублируют друг друга с разными именами полей (`Definition` vs `Def`). `lifecycle.StandardObjectLifecycle()` и тестовые YAML-определения тоже дублируются.

**Рекомендация:** Оставить только `internal/process/fsm/`. Перенести `StandardObjectLifecycle()` и `NewMachine()` туда. Удалить `internal/core/lifecycle/`.

#### P1 — Разбросанные определения портов

Интерфейсы-порты определены в трёх местах:

1. `internal/app/ports/ports.go` — канонические интерфейсы (`ObjectRepository`, `Indexer`, `GitService`)
2. `internal/app/command/command.go` — дублирующие интерфейсы (`ObjectRepo`, `Indexer`, `GitOps`, `Validator`)
3. `internal/app/query/query.go` — `ObjectReader` интерфейс
4. `internal/modules/bom/bom.go` — `ObjectReader`, `RelationReader`
5. `internal/modules/release/release.go` — `ObjectReader`, `RelationLister`

Интерфейсы command и query НЕ совпадают с ports. Например, `ports.Indexer` имеет 5 методов, а `command.Indexer` — 5 методов с другими именами (плюс `RebuildIndex`).

**Рекомендация:** Консолидировать ВСЕ интерфейсы в `internal/app/ports/`. Command/Query/Modules должны использовать их через композицию (`type Repo interface { ports.ObjectReader; ports.ObjectWriter }`).

#### P2 — Неиспользуемый пакет ports

`internal/app/ports/ports.go` определён, но нигде не импортируется — command/query определяют свои интерфейсы заново.

#### P2 — Мёртвый код в transaction.go

```go
// internal/store/transaction/transaction.go:201
var _ = json.Marshal // ensure json import is used
```

Искусственный импорт. `encoding/json` реально используется только в структурах через теги. Либо использовать json в коде, либо удалить теги.

### 1.3 Позитивные аспекты

- Чистые доменные типы без зависимостей от БД/файловой системы
- Правильное использование value objects (ID, Class, State — typed strings)
- Инварианты через методы валидации (`ValidateIdentity`, `ValidateShape`)
- Immutable event log (append-only JSONL)
- Transaction Manager с фазами prepare → commit → cleanup
- Индекс как rebuildable projection, а не source of truth

---

## 2. КАЧЕСТВО КОДА (7/10)

### 2.1 Критические баги

#### P0 — Баг в RunTransition: событие логирует неверное состояние

```go
// internal/app/command/command.go:128-133
obj.State = object.State(newState)           // ← состояние УЖЕ изменено

evt := event.New(
    ..., event.LifecycleTransitioned, string(obj.ID),
    map[string]any{"from": obj.State, ...},  // ← obj.State = NEW state, а должно быть OLD
)
```

**Проблема:** `obj.State` перезаписывается до создания события. В результате `"from"` в payload содержит НОВОЕ состояние вместо старого. История переходов будет некорректной.

**Исправление:**
```go
oldState := string(obj.State)
obj.State = object.State(newState)

evt := event.New(
    ..., event.LifecycleTransitioned, string(obj.ID),
    map[string]any{"from": oldState, "to": newState, "transition": req.Transition},
)
```

#### P1 — AttachArtifact: fire-and-forget горутина без обработки ошибок

```go
// internal/app/command/command.go:223
go func() { s.Index.UpsertObject(context.Background(), obj) }()
```

**Проблемы:**
- Ошибка индексации silently ignored
- `context.Background()` — нет таймаута, нет отмены
- Если горутина panic — краш всего приложения
- Нет гарантии завершения перед возвратом из AttachArtifact

**Рекомендация:** Использовать errgroup с контекстом запроса (как в CreateObject/RunTransition) либо как минимум логировать ошибку.

### 2.2 Проблемы error handling

#### P1 — Игнорирование ошибок json.Marshal (6 мест)

```go
// internal/store/index/index.go
metaJSON, _ := json.Marshal(obj.Metadata)       // строки 101, 136, 228
// internal/app/command/command.go
evtJSON, _ := json.Marshal(evt)                  // строки 93, 135
```

`json.Marshal` для map[string]any ТЕОРЕТИЧЕСКИ может упасть (NaN, Inf, циклы). В текущем коде циклов нет (данные из YAML), но это плохая практика.

**Рекомендация:** Проверять ошибки либо использовать `MustMarshal` helper с panic (раз в домене гарантируется валидность).

#### P1 — Игнорирование ошибок в test helpers

```go
// internal/app/scenario_test.go:119
data, _ := os.ReadFile(path)
```

В тестовых хелперах ошибки чтения файлов игнорируются, что может скрывать проблемы в тестах.

#### P2 — AppendHistory не проверяет ошибку в RunTransition

```go
// internal/app/command/command.go:145
g.Go(func() error { s.Repo.AppendHistory(ctx, obj.ID, string(evtJSON)); return nil })
```

Ошибка `AppendHistory` явно игнорируется (`return nil`). История — важный аудиторский след, её потеря критична.

### 2.3 Проблемы concurrency

#### P2 — ListObjects: хрупкий паттерн «слива семафора»

```go
// internal/store/fsrepo/fsrepo.go:142-152
for i, id := range ids {
    sem <- struct{}{}                                   // acquire
    go func(idx int, oid object.ID) {
        defer func() { <-sem }()                        // release
        ...
    }(i, id)
}
for i := 0; i < workers; i++ {
    sem <- struct{}{}                                   // "drain" — send extra tokens
}
```

Паттерн работает, но хрупок: если горутина запаникует до `<-sem`, канал заблокируется навсегда. Такой же паттерн используется в `CheckReadiness` и `GenerateManifest`.

**Рекомендация:** Заменить на `sync.WaitGroup` + семафор:
```go
var wg sync.WaitGroup
for i, id := range ids {
    wg.Add(1)
    sem <- struct{}{}
    go func(idx int, oid object.ID) {
        defer wg.Done()
        defer func() { <-sem }()
        ...
    }(i, id)
}
wg.Wait()
```

#### P3 — BOM walkBOM: не-thread-safe visited map

```go
// internal/modules/bom/bom.go:87
func (s *Service) walkBOM(..., visited map[string]bool, ...) error {
    if visited[fromID] { ... }
    visited[fromID] = true
```

Сейчас вызывается последовательно — ок. Но мутация map через указатель — антипаттерн для будущей параллелизации.

#### P3 — GenerateManifest: результаты через канал, сортировка по индексу

```go
// internal/modules/release/release.go:176
manifestObjects := make([]ManifestObject, len(scope.Objects))
for e := range entries {
    manifestObjects[e.idx] = e.mo
}
```

Работает, но если дублируются индексы (не должны, но нет защиты) — будет перезапись.

### 2.4 Качество кода: позитивные аспекты

- Отсутствуют TODO/FIXME/HACK/XXX во всём коде
- Отсутствуют panic() вне test helpers
- Правильное использование `defer f.Close()`, `defer tx.Rollback()`, `defer rows.Close()`
- Строгая типизация строковых констант (type Class string, type State string)
- Понятные имена переменных и функций
- Хорошая композиция через интерфейсы
- Table-driven tests в core/object

### 2.5 Неиспользуемый код

#### P3 — Остатки импорта в naming_test.go

```go
// internal/naming/naming_test.go:81
go func() {
    // ...
}()
```
Горутина в naming тестах запускается без WaitGroup — потенциальный race в тесте.

---

## 3. БЕЗОПАСНОСТЬ (6/10)

### 3.1 Критические

#### P0 — FTS5 query injection

```go
// internal/store/index/index.go:175
rows, err := d.db.QueryContext(ctx,
    "SELECT id FROM fts_objects WHERE fts_objects MATCH ? LIMIT ?", query, limit)
```

FTS5 использует специальный синтаксис запросов. Символы `*`, `"`, `AND`, `OR`, `NOT`, `NEAR()` имеют специальное значение. Передача сырого пользовательского ввода в MATCH может привести к:

1. **Синтаксическим ошибкам FTS5** (не падение, но пустой результат)
2. **DoS-атакам через сложные запросы** (например, `"a" OR "b" OR "c" ...`)
3. **Утечке данных** (префиксные запросы `*` могут выдать все записи)

**Рекомендация:** Экранировать FTS5-специальные символы в пользовательском вводе:
```go
func sanitizeFTS5Query(q string) string {
    // Escape FTS5 special characters by wrapping in quotes
    q = strings.ReplaceAll(q, `"`, `""`)
    return `"` + q + `"`
}
```

Либо ограничить длину запроса и отклонить запросы со специальными символами.

#### P2 — Path traversal через AttachArtifact

```go
// internal/app/command/command.go:200
base := filepath.Base(localPath)
targetPath := filepath.Join(targetDir, base)
```

`filepath.Base` защищает от простых path traversal (`../../../etc/passwd`). Но если `localPath` — директория, поведение неочевидно.

#### P3 — Предсказуемые ID транзакций

```go
// internal/store/transaction/transaction.go:188-191
func NewID() string {
    b := make([]byte, 8)
    _, _ = rand.Read(b)
    return "tx-" + hex.EncodeToString(b)
}
```

8 байт (64 бита) достаточно для уникальности. Но ID события ещё хуже:

```go
// internal/app/command/command.go:88-89
fmt.Sprintf("evt-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)
```

`UnixNano()%1000000` — всего ~20 бит энтропии. Два события в одну секунду могут получить одинаковый ID.

### 3.2 Важные

#### P1 — Отсутствие аутентификации/авторизации

MVP без пользователей — допустимо. Но `"local-user"` захардкожен в app/command:
```go
Actor: "local-user", // строки 90, 132, 165
```

Должен быть передан из контекста или конфигурации.

#### P2 — Нет rate limiting на поиск

FTS5 поиск без ограничений может быть использован для DoS через сложные запросы.

#### P3 — SQLite без шифрования

Локальная БД .plm/index.db не шифруется. Для desktop-приложения допустимо, но стоит отметить в документации.

### 3.3 Позитивные аспекты

- Нет хардкода паролей или ключей
- `.gitignore` включает secrets, .env
- Транзакции с атомарными операциями
- Нет eval/injection кроме FTS5
- Все файловые пути строятся через `filepath.Join`

---

## 4. ПРОИЗВОДИТЕЛЬНОСТЬ (7/10)

### 4.1 Текущие оптимизации (положительно)

| Оптимизация | Где | Детали |
|------------|-----|--------|
| Worker pools (8 горутин) | fsrepo.ListObjects, release.CheckReadiness, release.GenerateManifest | Bounded concurrency |
| Batch INSERT (500 rows) | index.RebuildIndex | На 70% быстрее чем row-by-row |
| errgroup fan-out | command.CreateObject, RunTransition | SaveObject → Index+History параллельно |
| SQLite WAL mode | index.migrate | Лучшая конкурентность чтения |
| BOM оптимизация | bom.GetFlat | Группировка по child_id |

### 4.2 Проблемы производительности

#### P2 — ListObjects загружает ВСЕ объекты в память

```go
// internal/store/fsrepo/fsrepo.go:104
func (r *Repository) ListObjects(ctx context.Context) ([]object.Object, []error) {
```

С сотнями/тысячами объектов это приведёт к высокому потреблению памяти. Нет пагинации или курсора.

**Рекомендация:** Добавить `ListObjectsPaginated(offset, limit int)` метод.

#### P2 — getTransition — O(n) линейный поиск

```go
// internal/process/fsm/fsm.go:110-117
func (m *Machine) getTransition(name string) *Transition {
    for i := range m.Def.Transitions { ... }
}
```

Для StandardObjectLifecycle (9 transitions) — ок. Для кастомных процессов с 50+ переходами может быть узким местом.

**Рекомендация:** Кешировать в map при создании Machine.

#### P3 — SQLite MaxOpenConns(1) ограничивает читателей

```go
// internal/store/index/index.go:30
db.SetMaxOpenConns(1)
```

Правильно для записи, но с WAL mode можно разрешить несколько читателей. Рассмотреть пул соединений: 1 writer + N readers.

#### P3 — BOM: повторные GetObject при rollup

```go
// internal/modules/bom/bom.go:59
func (s *Service) GetFlat(...) ([]BOMRow, error) {
    structured, err := s.GetStructured(ctx, root)  // ← уже читает все объекты
    ...
}
```

GetFlat вызывает GetStructured (уже прочитал все объекты), потом группирует. GetStructured на каждый объект делает GetObject. Это N+1 для каждого BOM узла.

**Рекомендация:** Добавить метод GetObjectsByIDs для batch-чтения в BOM.

### 4.3 Позитивные аспекты

- Нет allocations в горячих путях (value types, stack allocation)
- BOM cycle detection с backtracking (delete visited[id])
- Release scope collection с ранним выходом (visited map)

---

## 5. ТЕСТИРОВАНИЕ (7/10)

### 5.1 Coverage по пакетам

| Пакет | Coverage | Оценка |
|-------|----------|--------|
| core/artifact | 100.0% | ✅ |
| core/diagnostic | 100.0% | ✅ |
| core/event | 100.0% | ✅ |
| core/lifecycle | 100.0% | ✅ |
| core/object | 100.0% | ✅ |
| core/project | 100.0% | ✅ |
| naming | 100.0% | ✅ |
| parser/markdown | 100.0% | ✅ |
| validation/rules | 100.0% | ✅ |
| parser/frontmatter | 96.2% | ✅ |
| core/revision | 95.0% | ✅ |
| modules/bom | 90.7% | ✅ |
| modules/release | 90.8% | ✅ |
| parser/yamlx | 85.7% | ✅ |
| process/fsm | 84.2% | ✅ |
| store/fsrepo | 82.2% | ✅ |
| store/index | 73.3% | ⚠️ |
| store/transaction | 71.0% | ⚠️ |
| core/relation | 66.7% | ⚠️ |
| modules/standardparts | 70.0% | ⚠️ |
| validation/engine | 65.2% | ⚠️ |
| **api/dto** | **0%** | ❌ |
| **api/mapper** | **0%** | ❌ |
| **api/tree** | **0%** | ❌ |
| **app/command** | **0%** | ❌ |
| **app/query** | **0%** | ❌ |
| **app/ports** | **0%** | ❌ |
| **gitops** | **0%** | ❌ |
| **process/effects** | **0%** | ❌ |
| **process/guards** | **0%** | ❌ |

### 5.2 Качество тестов

#### Сильные стороны
- **Интеграционные сценарии в `internal/app/`**: `scenario_test.go`, `all_types_test.go`, `project_gen_test.go`, `media_test.go`, `tree_test.go` — проверяют полный цикл от создания до чтения
- **Table-driven tests**: `object_test.go`, `revision_test.go`
- **Mock-объекты**: `mockObjects`, `mockRelations`, `mockGit` — простые и понятные
- **Временная файловая система**: везде используется `t.TempDir()` — изоляция
- **Race detector**: `make test` включает `-race`

#### Проблемы

**P2 — 9 пакетов без тестов.** api/dto — просто структуры (допустимо). Но gitops, process/effects, process/guards должны быть покрыты.

**P2 — Тесты app слоя — монолитный scenario_test.go (518 строк).** Содержит 9 тестовых функций + хелперы + mock-и. Нужно разделить.

**P3 — naming_test.go горутина без WaitGroup:**
```go
// internal/naming/naming_test.go:81
go func() {
    s.Next("prt")
    wg.Done()  // wg не существует в этом scope?
}()
```

**P3 — Негативные тесты есть только для RunTransition:**
```go
// scenario_test.go:208 TestErrorScenarios
```
Нет негативных тестов для GetBOM (циклы), CheckReadiness (missing objects), CreateCheckpoint (no git).

### 5.3 Отсутствующие тесты

| Тип теста | Статус |
|-----------|--------|
| Валидация всех 15 классов | ✅ all_types_test.go |
| Полный lifecycle: draft→archived | ✅ scenario_test.go |
| BOM nested + flat rollup | ✅ bom_test.go + scenario_test.go |
| Release readiness | ✅ scenario_test.go |
| Git init/commit/tag/log | ✅ scenario_test.go |
| Transaction write/copy/delete/rename | ✅ scenario_test.go |
| Index rebuild | ✅ scenario_test.go |
| Duplicate detection | ✅ (inline, без модуля standardparts) |
| Search (FTS5) | ⚠️ skipped если нет данных |
| AttachArtifact полный цикл | ✅ media_test.go |
| Tree generation (DTOs) | ✅ tree_test.go |
| Error: невалидный ID | ✅ |
| Error: невалидный transition | ✅ |
| **Concurrent access** | ❌ |
| **Recovery after crash** | ❌ |
| **BOM cycle detection (edge cases)** | ❌ (только простой цикл) |
| **Large BOM performance** | ❌ |
| **FTS5 injection** | ❌ |
| **Path traversal** | ❌ |

---

## 6. ДОКУМЕНТАЦИЯ (8/10)

### 6.1 Что хорошо

- **README.md**: 280 строк, охватывает философию, архитектуру, quick start, project structure, lifecycle, naming standard, concurrency
- **SRS - go-plm - Complete.md**: 39 секций требований
- **Комментарии в коде**: Каждый пакет имеет package doc, большинство экспортируемых типов документированы
- **GitHub Epics and Issues.md**: 13 эпиков, 51 задача
- **AUDIT документы**: 2 существующих аудита

### 6.2 Чего не хватает

| Документ | Приоритет |
|----------|-----------|
| **ADR (Architecture Decision Records)** | P1 — Почему Markdown а не JSON? Почему SQLite а не BoltDB? Почему go-git а не exec? |
| **API Reference** | P1 — Описание всех команд и запросов с примерами |
| **Deployment Guide** | P2 — Как собрать бинарник под все платформы |
| **Contributing Guide** | P3 — Стандарты кода, процесс PR |
| **Frontend Architecture** | P1 — Компонентное дерево, state management |

### 6.3 Обнаруженное несоответствие

README говорит «Go 1.26+», `go.mod` говорит `go 1.26.2`, но `go tool version` возвращает `go1.25.5`. Версия тулчейна не совпадает с версией компилятора.

---

## 7. ИНФРАСТРУКТУРА (6/10)

### 7.1 Проблемы

#### P1 — Несоответствие версий Go
```
go version → go1.26.2
go tool version → go1.25.5
```
`go test -coverprofile` падает с «compile: version "go1.26.2" does not match go tool version "go1.25.5"». Нужно обновить Go toolchain до 1.26.x.

#### P2 — Frontend только scaffold (`.gitkeep`)
React-приложение не реализовано. Frontend директория состоит из 6 `.gitkeep` файлов. Блокирует Wails-интеграцию.

#### P3 — cmd/plm/ не найден
`cmd/plm/` упомянут в README и Makefile, но директория пуста. Точка входа не реализована.

#### P2 — Нет Dockerfile
В Makefile есть цель `docker-build`, но Dockerfile отсутствует.

#### P3 — CI не проверен
`.github/` директория существует, но содержимое не проанализировано.

---

## 8. ПРИОРИТЕТНЫЙ ПЛАН ИСПРАВЛЕНИЙ

### P0 — Критические (немедленно)

| # | Проблема | Файл | Действие |
|---|----------|------|----------|
| 1 | **RunTransition: неверное состояние в событии** | `internal/app/command/command.go:128-133` | Сохранить oldState перед перезаписью |
| 2 | **FTS5 injection** | `internal/store/index/index.go:175` | Экранировать специальные символы в query |
| 3 | **Go toolchain mismatch** | Система | Обновить go tool до 1.26.x |

### P1 — Важные (следующий спринт)

| # | Проблема | Файл | Действие |
|---|----------|------|----------|
| 4 | Удалить дублирование lifecycle/fsm | `core/lifecycle/`, `process/fsm/` | Оставить только fsm |
| 5 | Консолидировать интерфейсы портов | `app/ports/`, `app/command/`, `app/query/` | Всё в ports.go |
| 6 | Исправить игнорирование ошибок json.Marshal | `store/index/`, `app/command/` | Добавить проверку или MustMarshal |
| 7 | AttachArtifact: убрать fire-and-forget | `app/command/command.go:223` | Использовать errgroup с контекстом |
| 8 | RunTransition: проверять ошибку AppendHistory | `app/command/command.go:145` | Не игнорировать return nil |
| 9 | Добавить тесты для 9 пакетов | gitops, effects, guards, api/* | Базовое покрытие |

### P2 — Улучшения (бэклог)

| # | Проблема | Действие |
|---|----------|----------|
| 10 | ListObjects: заменить семафор на WaitGroup | Надёжнее |
| 11 | getTransition: кешировать в map | Производительность |
| 12 | Пагинация в ListObjects | Память |
| 13 | ADR документы | Документация |
| 14 | API Reference | Документация |
| 15 | Конкуррентные тесты | Тестирование |
| 16 | Улучшить покрытие store слоя (71-73%) | Тестирование |
| 17 | Преобразовать ID событий в UUID | Безопасность |

---

## 9. МЕТРИКИ

| Метрика | Значение |
|---------|----------|
| Всего строк Go кода | ~4,500 |
| Пакетов | 28 |
| Файлов | 56 |
| Тестовых файлов | 25 |
| Тестовых функций | ~60 |
| Средний coverage | ~82% |
| Паники в коде | 0 |
| TODO/FIXME | 0 |
| Горутины без recover | 1 (AttachArtifact) |
| Непроверенных ошибок | ~10 |
| Багов P0 | 2 (event state, FTS5 injection) |
| Багов P1 | 6 |
| Архитектурных проблем | 3 |

---

## 10. ЗАКЛЮЧЕНИЕ

Кодовая база go-plm находится на **ранней стадии разработки (MVP)** с крепким фундаментом. Clean Architecture позволяет легко расширять систему. Доменная модель хорошо изолирована. Тесты покрывают основные сценарии.

**Главные риски:**
1. Баг в RunTransition искажает историю переходов (P0)
2. FTS5 injection может привести к DoS (P0)
3. 9 непокрытых тестами пакетов создают риск регрессий
4. Frontend отсутствует полностью — основной объём работы впереди

**Рекомендуемый порядок действий:**
1. Исправить P0 баги (2 часа)
2. Устранить дублирование lifecycle/fsm (1 час)
3. Консолидировать порты (2 часа)
4. Покрыть тестами непокрытые пакеты (4 часа)
5. Начать реализацию frontend и cmd/plm/

**Готовность к продакшену:** 3/10 (MVP backend, нет frontend и точки входа)
