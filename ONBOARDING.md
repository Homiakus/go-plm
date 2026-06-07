# ONBOARDING GUIDE — go-plm

## Для нового разработчика: порядок изучения

### 1. Первое знакомство (15 минут)

```bash
# Клонировать и собрать
git clone https://github.com/Homiakus/go-plm.git
cd go-plm
cd frontend && npm install && npm run build && cd ..
go build -o build/plm.exe ./cmd/plm/

# Создать тестовый проект и запустить
./build/plm init test 'My Test'
./build/plm test
# Открыть http://127.0.0.1:8470 в браузере
```

### 2. Файлы, которые нужно прочитать ПЕРВЫМИ

| # | Файл | Что узнаешь | Время |
|---|------|-------------|-------|
| 1 | `README.md` | Философия, архитектура, Quick Start | 5 мин |
| 2 | `internal/core/object/object.go` | Главная сущность системы — Object | 5 мин |
| 3 | `cmd/plm/main.go` | Как стартует программа (CLI + server) | 5 мин |
| 4 | `cmd/plm/server.go` | HTTP сервер, JSON-RPC, embedded frontend | 5 мин |
| 5 | `internal/app/service/service.go` | Как собираются все зависимости (DI) | 5 мин |
| 6 | `internal/app/command/command.go` | Все write-операции | 10 мин |
| 7 | `frontend/src/App.tsx` | Корневой React-компонент | 5 мин |

### 3. Ключевые сценарии для изучения

| Сценарий | Как проследить |
|----------|---------------|
| **Создание объекта** | Frontend: CreateWizard → App.handleCreateObject → api.createObject → JSON-RPC → Go: API.CreateObject → command.CreateObject → naming.NextID → fsrepo.SaveObject |
| **Жизненный цикл** | Frontend: ObjectEditor → RunTransition → Go: command.RunTransition → fsm.Can/Next → SaveObject → side effects (git tag) |
| **BOM** | Frontend: BOMViewer → api.getBOM → Go: query.GetBOM → bom.GetStructured → walkBOM (recursive) |
| **Поиск** | Frontend: SearchPanel → api.searchObjects → Go: index.SearchObjects → FTS5 MATCH |
| **Релиз** | Frontend: ReleaseDashboard → api.createReleasePackage → Go: command.CreateReleasePackage → scope collection |

### 4. Команды

```bash
make build          # Сборка всего (frontend + Go)
make test           # Запуск всех тестов
make lint           # go vet + golangci-lint
make frontend-dev   # Dev-сервер React (HMR)
make frontend-build # Только сборка frontend
make run            # Сборка + запуск
make demo           # Новый проект + запуск
```

### 5. Безопасные для изменений зоны

- **Frontend components** (`frontend/src/components/*.tsx`) — можно менять без опасений
- **DTO types** (`internal/api/dto/dto.go`) — добавить поле → обновить mapper
- **Validation rules** (`internal/validation/rules/rules.go`) — добавить правило → зарегистрировать
- **FSM guards** (`internal/process/guards/guards.go`) — добавить guard → зарегистрировать
- **Query methods** (`internal/app/query/query.go`) — read-only, безопасно

### 6. Зоны, требующие осторожности

- **`internal/core/object/object.go`** — изменение Object затрагивает ВСЁ
- **`internal/store/fsrepo/fsrepo.go`** — изменение формата .md ломает данные
- **`internal/naming/naming.go`** — изменение формата ID ломает все объекты
- **`internal/store/transaction/transaction.go`** — атомарность критична для целостности данных
- **`internal/store/index/index.go`** — изменение схемы SQLite требует миграции

### 7. Что где лежит (быстрый поиск)

| Нужно | Иди в |
|-------|------|
| Добавить API-метод | `internal/api/wails/api.go` |
| Добавить команду (write) | `internal/app/command/command.go` |
| Добавить запрос (read) | `internal/app/query/query.go` |
| Добавить поле в объект | `internal/core/object/object.go` + `internal/api/dto/dto.go` + `internal/api/mapper/mapper.go` |
| Добавить тип связи | `internal/core/relation/relation.go` |
| Добавить класс объекта | `internal/core/object/object.go` (Class const) |
| Добавить правило валидации | `internal/validation/rules/rules.go` |
| Добавить guard FSM | `internal/process/guards/guards.go` |
| Изменить жизненный цикл | `config/lifecycle.md` + `internal/process/fsm/fsm.go` |
| Добавить UI-компонент | `frontend/src/components/` |
| Добавить UI-страницу | `frontend/src/App.tsx` (viewMode switch) |

### 8. Типичные ошибки новичка

1. **Забыть обновить mapper** — добавил поле в Object → нужно в dto.ObjectDTO и mapper.ObjectToDTO/DTOToObject
2. **nil вместо []** — Go возвращает nil для пустых слайсов; JSON-RPC сериализует в null; фронтенд падает. Всегда `make([]T, 0)` или `|| []`
3. **Не проверить `ctx.Err()`** — долгие операции должны уважать отмену контекста
4. **Прямая запись в .md** — только через transaction.Manager
5. **TDZ в TypeScript** — const должен быть объявлен ДО использования (особенно в module scope)
