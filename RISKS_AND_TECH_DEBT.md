# RISKS AND TECHNICAL DEBT — go-plm v2.0

## Critical

### Event ID collision risk
- **Finding:** Event IDs use `time.Now().UnixNano()%1000000` — ~20 бит энтропии
- **Evidence:** `internal/app/command/command.go:88-89`
- **Confidence:** High
- **Impact:** Два события в одну секунду получат одинаковый ID → дубликаты в history.jsonl
- **Recommendation:** Использовать `github.com/google/uuid` (уже есть в go.mod как indirect dep)

### SQLite FTS5 schema: no content sync trigger
- **Finding:** FTS5 virtual table `fts_objects` создаётся без триггеров синхронизации с `objects` таблицей
- **Evidence:** `internal/store/index/index.go:89-92`
- **Confidence:** High
- **Impact:** FTS5 индекс не обновляется автоматически при UpsertObject/DeleteObject — поиск может возвращать устаревшие данные
- **Recommendation:** Добавить триггеры: `CREATE TRIGGER ... AFTER INSERT ON objects BEGIN INSERT INTO fts_objects ... END;`

### ListObjects: unbounded memory
- **Finding:** `ListObjects` загружает ВСЕ объекты в память без пагинации
- **Evidence:** `internal/store/fsrepo/fsrepo.go:104` — `return objects, errs` для всех объектов
- **Confidence:** High
- **Impact:** С >10 000 объектов — OOM risk. Каждый .md файл читается в память целиком
- **Recommendation:** Добавить `ListObjectsPaginated(offset, limit)` с потоковым чтением

## High

### BOM N+1 queries
- **Finding:** `walkBOM` делает `GetObject` для каждого узла BOM
- **Evidence:** `internal/modules/bom/bom.go:111` — `childObj, err := s.Objects.GetObject(ctx, object.ID(rel.ToID))`
- **Confidence:** High
- **Impact:** BOM из 100 компонентов = 100+ чтений с диска. С глубокой вложенностью — ещё больше
- **Recommendation:** Batch preload: собрать все child ID → один запрос → map

### Recovery mode not surfaced to UI
- **Finding:** `transaction.Manager.Recover()` вызывается при старте, но результаты не показываются
- **Evidence:** `internal/app/service/service.go` — авто-recover в Open(), без уведомления
- **Confidence:** High
- **Impact:** Пользователь не знает о восстановленных/потерянных транзакциях
- **Recommendation:** Показывать notification в UI после recover

### No input validation for JSON-RPC params
- **Finding:** `callMethod` использует reflection — нет валидации типов до вызова методов
- **Evidence:** `cmd/plm/server.go:158-258`
- **Confidence:** Medium
- **Impact:** Неправильные типы параметров вызывают панику в reflect
- **Recommendation:** Добавить type assertion с recovery вокруг reflect.Call

### Missing FTS5 content sync
- **Finding:** FTS5 virtual table не синхронизируется с таблицей objects
- **Evidence:** `internal/store/index/index.go:89-92` — FTS5 создаётся без content sync triggers
- **Confidence:** High
- **Impact:** Поиск возвращает устаревшие данные после обновления объектов
- **Recommendation:** Добавить `INSERT INTO fts_objects(fts_objects) VALUES('rebuild')` после UpsertObject

## Medium

### Hardcoded port 8470
- **Finding:** Порт жёстко задан в `cmd/plm/main.go:runServer` — `srv.Start(8470)`
- **Evidence:** `cmd/plm/main.go`
- **Impact:** Конфликт портов с другими приложениями
- **Recommendation:** Флаг `--port` или авто-выбор свободного порта

### getTransition O(n)
- **Finding:** Линейный поиск по transitions при каждом вызове
- **Evidence:** `internal/process/fsm/fsm.go:110-117`
- **Confidence:** High
- **Impact:** При 50+ transitions в definition — микрозамедление
- **Recommendation:** Кешировать в map при `New()`/`NewFromYAML()`

### SQLite MaxOpenConns(1) limits readers
- **Finding:** `SetMaxOpenConns(1)` блокирует параллельные чтения
- **Evidence:** `internal/store/index/index.go:30`
- **Confidence:** Medium
- **Impact:** С WAL mode можно иметь concurrent readers
- **Recommendation:** Увеличить до 4-8 в WAL mode

### No graceful shutdown of HTTP server
- **Finding:** `main()` использует `select {}` — нет обработки SIGINT/SIGTERM
- **Evidence:** `cmd/plm/main.go`
- **Confidence:** High
- **Impact:** При Ctrl+C горутины могут не завершиться, транзакции — не откатиться
- **Recommendation:** `signal.NotifyContext` + `srv.Shutdown()`

### Duplicate standard parts check: local only
- **Finding:** `CreateObject` проверяет дубликаты std только через `standardparts.Validate(obj)` — без сравнения с существующими
- **Evidence:** `internal/app/command/command.go`
- **Confidence:** Medium
- **Impact:** Дубликаты std деталей могут быть созданы
- **Recommendation:** Загрузить все std-объекты и вызвать `standardparts.FindDuplicates`

### Autosave: debounce without cleanup on unmount
- **Finding:** `MarkdownEditor` использует `setTimeout` без очистки при размонтировании
- **Evidence:** `frontend/src/components/MarkdownEditor.tsx:46`
- **Confidence:** Medium
- **Impact:** Таймер может сработать после unmount → setState на несуществующем компоненте
- **Recommendation:** `useEffect(() => () => clearTimeout(timerRef.current), [])`

## Low

### Unused `json` import in transaction.go
- **Finding:** `var _ = json.Marshal` для сохранения импорта — код-запах
- **Evidence:** `internal/store/transaction/transaction.go:201`
- **Confidence:** High
- **Impact:** Косметический
- **Recommendation:** Использовать json для сериализации Plan или удалить теги

### mockGit lacks full interface
- **Finding:** `mockGit` в тестах не реализует `CreateTag`
- **Evidence:** `internal/app/scenario_test.go:426-431`
- **Impact:** Тесты не покрывают создание тегов
- **Recommendation:** Расширить mock

### Frontend: no code splitting
- **Finding:** Vite собирает весь JS в один бандл (282KB)
- **Evidence:** `frontend/dist/assets/index-*.js`
- **Impact:** Медленная первая загрузка на слабых соединениях
- **Recommendation:** `React.lazy()` + `Suspense` для крупных компонентов

### No frontend tests
- **Finding:** 0 тестов для React-компонентов
- **Evidence:** Нет `*.test.tsx` файлов
- **Impact:** Регрессии UI не отлавливаются
- **Recommendation:** Vitest + React Testing Library для ключевых компонентов

### Hardcoded "local-user"
- **Finding:** Actor всегда "local-user"
- **Evidence:** `internal/app/command/command.go` (6 мест)
- **Impact:** Нет аудита действий в многопользовательском режиме (будущее)
- **Recommendation:** Вынести в конфигурацию/контекст

---

## Refactoring Roadmap

### First 1–2 days (быстрые безопасные улучшения)
1. Fix Event ID: использовать `uuid.New()` вместо `UnixNano()%1000000`
2. Fix graceful shutdown: `signal.NotifyContext` + `srv.Shutdown()`
3. Fix autosave cleanup: `clearTimeout` в useEffect cleanup
4. Remove `var _ = json.Marshal` hack
5. Add FTS5 sync triggers

### First week (средние улучшения)
1. Add `ListObjectsPaginated(offset, limit)`
2. Cache `getTransition` results in map
3. Increase `MaxOpenConns` for SQLite readers
4. Add input validation in JSON-RPC handler (recover around reflect)
5. Surface recovery results to UI
6. Batch preload objects in BOM walkBOM

### First month (архитектурные изменения)
1. Frontend code splitting (React.lazy)
2. Frontend tests (Vitest + RTL)
3. Duplicate std parts: cross-reference with all existing
4. Multiuser support preparation (actor from context)
5. Database migrations framework for SQLite schema changes
