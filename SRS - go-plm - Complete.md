# Software Requirements Specification (SRS)

# go-plm / Markdown-PLM v2.0

## Локальная Git-native PLM/PDM-система для инженерной документации

|                   |                                                    |
|-------------------|----------------------------------------------------|
| **Версия**        | 2.0.0-draft                                        |
| **Статус**        | Technical Specification (нормализованный)           |
| **Дата**          | 2026-06-07                                         |
| **Репозиторий**   | github.com/Homiakus/go-plm                          |
| **Платформы**     | Windows 10+, macOS 13+, Linux (WebKitGTK)           |
| **Архитектура**   | Desktop-first, file-based, Git-native, FSM-driven   |
| **Авторы**        | Системный аналитик (нормализация на основе черновиков ТЗ) |

---

# 1. Executive Summary

`go-plm` (Markdown-PLM) — локальная desktop PLM/PDM-система, где каждый инженерный объект (деталь, сборка, чертёж, инструкция, релиз, change request) хранится как **Markdown-файл с YAML frontmatter** на диске. Git обеспечивает историю изменений. Конечные автоматы (FSM) управляют жизненными циклами. SQLite служит пересобираемым индексом для быстрых запросов и поиска.

**Ключевое отличие от традиционных PLM**: система не требует серверной инфраструктуры, СУБД или облака. Проект — это директория с файлами, которую можно скопировать, заархивировать, сохранить в Git. Индекс и кэш являются производными и пересобираются из source of truth.

**Главная формула системы**:

```
Object + Relation + Artifact + Process + Validation + Event + Projection
```

**Главные принципы**:

1. **Files are source of truth** — Markdown/YAML на диске являются каноническими данными
2. **Git-native** — история изменений, checkpoint, release tags через Git
3. **FSM-driven** — жизненные циклы управляются конечными автоматами с guard validation
4. **Markdown-everywhere** — вся техническая документация в человекочитаемом формате
5. **Index is rebuildable** — SQLite-индекс может быть удалён и пересобран из файлов

**Целевая аудитория**: инженеры-конструкторы, технологи, закупщики, производственный персонал, администраторы проектов — работающие на индивидуальных рабочих станциях.

---

# 2. Назначение системы

Система предназначена для управления полным инженерным циклом изделия на локальной рабочей станции:

- детали, сборки, чертежи, CAD/CAM-файлы
- BOM (eBOM, mBOM, Released BOM)
- стандартные и покупные детали
- документация, технологические карты, рабочие инструкции
- жизненные циклы объектов
- релизы и релизные пакеты
- изменения (change requests, change impact)
- производственная готовность (manufacturing matrix)
- закупочные списки (procurement buy list)
- история изменений и Git-интеграция
- экспорт данных для ERP/MES

Система НЕ является: ERP, MES, WMS, серверной PLM, CAD-редактором, облачным сервисом.

---

# 3. Цели и ценность продукта

## 3.1. Бизнес-цели

| ID | Цель | Приоритет |
|----|------|-----------|
| BG-01 | Создать лёгкую локальную альтернативу тяжёлым PLM/PDM-системам | P0 |
| BG-02 | Упростить управление инженерными объектами, BOM, чертежами и релизами | P0 |
| BG-03 | Обеспечить прозрачную историю изменений через Git без требования Git-знаний от пользователя | P0 |
| BG-04 | Снизить риск выпуска неполной/неактуальной документации через валидацию и readiness checks | P1 |
| BG-05 | Подготовить основу для интеграции с ERP, MES, WMS | P2 |
| BG-06 | Обеспечить работу со сложными инженерными процессами через универсальные абстракции | P0 |

## 3.2. Технические цели

| ID | Цель | Приоритет |
|----|------|-----------|
| TG-01 | Реализовать надёжное файловое хранилище с атомарными транзакциями | P0 |
| TG-02 | Реализовать доменное ядро (Object, Relation, Artifact, Event, Diagnostic) | P0 |
| TG-03 | Реализовать конфигурируемый FSM engine | P0 |
| TG-04 | Реализовать validation/readiness engine | P0 |
| TG-05 | Реализовать пересобираемый индекс проекта (SQLite) | P0 |
| TG-06 | Реализовать Git-интеграцию как внутренний механизм версионирования | P1 |
| TG-07 | Реализовать desktop UI на Wails + React + TypeScript | P1 |
| TG-08 | Реализовать расширяемую архитектуру модулей | P0 |
| TG-09 | Реализовать шаблоны объектов, документов, процессов, релизов | P1 |

---

# 4. Границы системы

## 4.1. Входит в MVP (P0/P1)

- Создание/открытие проекта (project.yaml + config/*.yaml)
- Создание инженерных объектов через Create Object Wizard
- Markdown/YAML-редактор с Live Preview
- Панель свойств (metadata panel)
- Дерево объектов (Project Tree) с поиском и фильтрацией
- Relations между объектами (базовые типы)
- Простая BOM (structured + flat)
- Where-used (обратные связи)
- Lifecycle FSM (draft → in_review → approved → released)
- Validation engine (проверка объекта, блокирующих проблем)
- Git status / checkpoint (без прямых git-команд в UI)
- SQLite index (пересобираемый)
- Поиск (SQLite FTS5)
- Стандартные детали базового уровня (с duplicate detection)
- Release readiness и release manifest
- Экспорт BOM/manifest в YAML/CSV

## 4.2. НЕ входит в MVP (Out of Scope для v2.0)

- Полноценная многопользовательская серверная PLM
- Облачная синхронизация
- Сложная ролевая модель enterprise-уровня
- Электронная подпись (юридически значимая)
- CAD-редактор / полноценный CAD viewer
- Автоматический парсинг всех CAD-форматов
- Полноценная ERP/MES/WMS
- Сложный merge/conflict resolution Git
- Shop-floor execution (build records, QR tokens) — перенесён в Post-MVP
- Bleve advanced search — Post-MVP

## 4.3. Явные исключения

Система НЕ должна:
- Требовать установки Docker, Kubernetes, серверной СУБД
- Требовать постоянного подключения к интернету
- Хранить данные в проприетарном бинарном формате
- Требовать от пользователя знания Git-команд

---

# 5. Пользователи и роли

| Роль | Описание | Основные действия |
|------|----------|-------------------|
| **Инженер/конструктор** | Создаёт детали, сборки, чертежи, CAD-файлы | Create Part/Assembly/Drawing, attach STEP/PDF/DXF, fill metadata, add to BOM, submit for review |
| **Технолог** | Работает с техпроцессами, операциями, инструкциями | Create process plan, work instruction, cutting/bending/NC files, check manufacturing readiness |
| **Закупщик** | Работает со стандартными деталями, поставщиками, заменами | Browse standard parts, add supplier, detect duplicates, generate buy list, propose substitutions |
| **Производственный пользователь** | (Post-MVP) Работает с инструкциями и build records | Open instruction, walk through steps, capture evidence, complete build |
| **Администратор проекта** | Настраивает правила проекта | Create project, configure classes/naming/lifecycle/validation/templates, rebuild index, recover project |

**Важно**: в MVP все пользователи работают на одной локальной машине. Многопользовательский режим — вне scope MVP.

---

# 6. Основные сценарии использования (Use Cases)

## UC-01: Создание проекта
1. Пользователь открывает приложение → Landing Page
2. Нажимает "Create New Project"
3. Заполняет: project_code, project_title, project_path
4. Система создаёт структуру директорий, project.yaml, config/*.yaml, инициализирует Git, создаёт initial commit
5. Система открывает workspace с пустым Project Tree

## UC-02: Создание детали внутри сборки
1. Пользователь выбирает сборку в Project Tree
2. Нажимает "+ New → Part"
3. Wizard: Identity (автогенерация ID) → Properties (material, thickness, etc.) → Documents & Attachments
4. Preview показывает: ID, создаваемые файлы и папки, связи
5. Пользователь подтверждает создание
6. Система атомарно создаёт: директорию объекта, .md файл, history.yaml.log, копирует прикреплённые файлы, обновляет индекс
7. Система автоматически добавляет связь `contains` в сборку

## UC-03: Проверка готовности к релизу
1. Пользователь открывает сборку → вкладка "Validation"
2. Система показывает все diagnostics (info, warning, blocker)
3. Для каждого blocker — описание проблемы и suggested action
4. Пользователь исправляет проблемы или нажимает "Explain Blockers"
5. Когда все blockers устранены, кнопка "Release" становится активной

## UC-04: Создание стандартной детали
1. Пользователь открывает Standard Parts Center
2. Нажимает "+ Add Standard Part"
3. Заполняет: std_class (fst/brg/elc/...), стандарт, размеры, материал
4. Система вычисляет normalized_key и проверяет дубликаты
5. Если exact duplicate найден — система блокирует создание и предлагает открыть существующий
6. Если probable duplicate — запрашивает подтверждение
7. После создания — пользователь может добавить поставщиков и procurement data

---

# 7. Термины и определения (Glossary)

| Термин | Определение |
|--------|-------------|
| **Object** | Базовая инженерная сущность (Part, Assembly, Drawing, etc.) |
| **Object ID** | Уникальный идентификатор объекта в формате `[project]-[class]-[sequence]-v[major].[minor]` |
| **ObjectClass** | Тип объекта (prt, asm, drw, std, doc, cut, bnd, nc, ins, tpc, wi, bom, rel, cr, build) |
| **Relation** | Именованная связь между двумя объектами (contains, has_drawing, substitutes, etc.) |
| **Artifact** | Файл, связанный с объектом (STEP, PDF, DXF, PNG, сертификат, сгенерированный документ) |
| **Source of Truth** | Markdown/YAML-файлы на диске — каноническое хранилище данных |
| **Index** | SQLite база данных с проекциями для быстрых запросов — пересобираема из Source of Truth |
| **Checkpoint** | Git commit, создаваемый через UI (инженерное действие, не git command) |
| **Release Tag** | Git tag в формате `release/[project]-rel-[sequence]-v[version]` |
| **Process / Lifecycle** | Конечный автомат (FSM), управляющий состояниями объекта |
| **Guard** | Условие, которое MUST быть выполнено для перехода FSM |
| **Effect** | Действие, выполняемое при успешном переходе FSM |
| **Validation Diagnostic** | Результат проверки с severity (info/warning/blocker/error), кодом, сообщением и suggested actions |
| **Readiness** | Состояние готовности объекта/сборки к release |
| **Blocker** | Diagnostic с severity=blocker, препятствующий release |
| **eBOM** | Engineering BOM — конструкторский состав, строится из `contains` relations |
| **mBOM** | Manufacturing BOM — производственный состав, включает eBOM + manufacturing files + инструкции |
| **Released BOM** | Замороженная спецификация релиза, только approved/released объекты |
| **Standard Part** | Стандартное/покупное изделие класса `std` с особым форматом ID и duplicate detection |
| **Normalized Key** | Строка для поиска дублей стандартных деталей (например, `fst:iso4017:m6:30:a2`) |
| **Substitution** | Связь замены между стандартными деталями с workflow утверждения |
| **Why Can't I Assistant** | UI-компонент, объясняющий причины невозможности действия (release, approve, etc.) |
| **Projection** | Индексное представление данных для быстрых запросов (BOM projection, where-used projection) |
| **Transaction** | Атомарная группа файловых операций с temp-файлами, fsync и atomic rename |
| **Event** | Запись в history.yaml.log о факте изменения |
| **Command** | Операция, изменяющая состояние системы (CreateObject, RunTransition, etc.) |
| **Query** | Операция, читающая состояние без изменений (GetObject, SearchObjects, etc.) |

---

# 8. Принципы архитектуры

## 8.1. Files are Source of Truth
MUST: все канонические данные хранятся в читаемых файлах (project.yaml, objects/*/item.json, objects/*/*.md, config/*.yaml). Система НЕ должна иметь скрытого состояния, невосстановимого из файлов.

## 8.2. Git-native
MUST: Git используется для фиксации изменений, diff, checkpoint, release tags. Пользователь НЕ работает с git-командами напрямую — только через инженерные действия (Create Checkpoint, Show Changes, Compare with Release).

## 8.3. FSM-driven
MUST: все важные процессы управляются конечными автоматами (lifecycle объекта, lifecycle документа, lifecycle релиза, lifecycle change request). FSM конфигурируется через YAML.

## 8.4. Index is Rebuildable
MUST: SQLite-индекс может быть удалён и полностью пересобран из Source of Truth командой RebuildIndex.

## 8.5. Clean Architecture
MUST: зависимости направлены внутрь. `core` не знает о файлах, SQLite, Git, Wails. `modules` не знают об API и UI. `api` только адаптирует вызовы.

## 8.6. Command/Query Separation (CQS)
MUST: команды (Commands) изменяют состояние, запросы (Queries) только читают. Query никогда не пишет Source of Truth.

## 8.7. No Direct File Writes from Frontend
MUST: frontend никогда не пишет .md, .yaml, .db, .git напрямую. Все изменения — через backend Commands.

## 8.8. Store Relation Once
MUST: связь хранится один раз (авторская). Обратные связи (used_in, where-used) строятся как проекции индекса.

## 8.9. Minimal File Names
MUST: имя файла содержит только [project]-[class]-[sequence]-v[major].[minor]. Материал, толщина, покрытие, цвет, поставщик — в YAML metadata.

## 8.10. Transactional File Operations
MUST: все записи Source of Truth проходят через transaction layer (temp files → fsync → atomic rename). Прямая перезапись критичных файлов запрещена.

---

# 9. Информационная модель (Domain Model)

## 9.1. Верхнеуровневая структура

Базовые сущности:
```
Project → Object → Relation → Artifact → Event → Diagnostic → IndexProjection
           ↑                        ↑
        Process (FSM)          Validation
```

## 9.2. Project

**Назначение**: корневой контейнер PLM-данных.

**Обязательные поля**:
- `code` — код проекта (используется в ID объектов)
- `title` — название
- `naming.standard` — версия naming standard
- `naming.pattern` — шаблон ID
- `sequences` — счётчики последовательностей по классам

**YAML представление**: `project.yaml` в корне проекта.

**Инварианты**:
- project.code соответствует naming rules
- config/*.yaml валидны
- все объекты имеют уникальные ID
- индекс может быть пересобран из файлов

## 9.3. Object

**Назначение**: базовая инженерная сущность.

**Обязательные поля**:
- `id` — уникальный идентификатор
- `project` — код проекта
- `class` — класс объекта (prt, asm, drw, std, ...)
- `state` — текущее состояние lifecycle
- `title` — название

**Опциональные поля**:
- `metadata` — map[string]any инженерных/производственных/закупочных атрибутов
- `relations` — исходящие связи
- `artifacts` — связанные файлы
- `created` / `updated` — временные метки

**Хранение**: `objects/[id]/[id].md` с YAML frontmatter.

**Инварианты**:
- ID соответствует naming standard
- ID совпадает с именем .md файла
- class существует в classes.yaml
- state существует в lifecycle конфигурации
- required metadata заполнены
- relations указывают на существующие объекты
- YAML frontmatter парсится без ошибок

**YAML frontmatter структура**:
```yaml
id: a320-prt-0001-v1.0
project: a320
class: prt
sequence: "0001"
version: "1.0"
revision: "1.0"
state: draft
title: Bracket Motor
metadata:
  material: al5052
  thickness_mm: 2.0
  unit: pcs
  make_buy: make
relations: []
artifacts: []
created:
  at: 2026-06-06T10:00:00Z
  by: local-user
updated:
  at: 2026-06-06T10:00:00Z
  by: local-user
generation:
  generated: false
  source_of_truth: true
```

## 9.4. ObjectClass

**Назначение**: описывает тип объекта, его required/optional metadata, допустимые связи, шаблон, правила именования, lifecycle.

**Базовые классы**:

| Class | Title | Sequence Mode | Lifecycle |
|-------|-------|---------------|-----------|
| prt | Part | numeric (0001-9999) | object_lifecycle |
| asm | Assembly | numeric (0100-9999) | object_lifecycle |
| drw | Drawing | numeric | document_lifecycle |
| doc | Document | numeric | document_lifecycle |
| std | Standard Part | semantic | standard_part_lifecycle |
| mat | Material | numeric | object_lifecycle |
| cut | Cutting File | inherit_or_numeric | manufacturing_file_lifecycle |
| bnd | Bending File | inherit_or_numeric | manufacturing_file_lifecycle |
| nc | NC Program | inherit_or_numeric | manufacturing_file_lifecycle |
| ins | Inspection | inherit_or_numeric | document_lifecycle |
| tpc | Tech Process Card | inherit_or_numeric | document_lifecycle |
| wi | Work Instruction | inherit_or_numeric | document_lifecycle |
| bom | Formal BOM | numeric | document_lifecycle |
| rel | Release Package | numeric | release_lifecycle |
| cr | Change Request | numeric | change_request_lifecycle |
| build | Build Record | numeric | build_lifecycle (Post-MVP) |

## 9.5. Relation

**Назначение**: именованная связь между двумя объектами.

**Поля**:
- `from_id` — источник (хранится в объекте-источнике)
- `to_id` — цель
- `type` — тип связи
- `quantity` — опциональное количество
- `unit` — единица измерения
- `effectivity` — применимость (from_version, to_version)
- `metadata` — дополнительные данные

**Типы связей**:

| Type | From | To | Stored | Inverse |
|------|------|----|--------|---------|
| contains | asm | prt, asm, std | yes | used_in |
| has_drawing | prt, asm, std | drw | yes | drawing_of |
| has_cad | prt, asm | prt (CAD) | yes | cad_of |
| manufacturing_file_for | cut, bnd, nc | prt, asm | yes | has_manufacturing_file |
| process_plan_for | tpc | prt, asm | yes | has_process_plan |
| work_instruction_for | wi | prt, asm | yes | has_work_instruction |
| inspection_for | ins | prt, asm | yes | has_inspection |
| substitutes | std, prt | std, prt | yes | substituted_by |
| equivalent_to | std, prt | std, prt | yes | — (симметричная) |
| replaces | prt, std | prt, std | yes | replaced_by |
| affects | cr | any | yes | affected_by |
| includes | rel | any | yes | released_in |
| derived_from | any | any | yes | — |

**Правило хранения**: связь хранится один раз в объекте-источнике. Обратные связи — проекции индекса.

## 9.6. Artifact

**Назначение**: файл, связанный с объектом.

**Поля**:
- `id` — идентификатор артефакта
- `kind` — cad, drawing, manufacturing, image, certificate, evidence, generated, document
- `role` — роль (primary_step, drawing_pdf, datasheet, etc.)
- `path` — относительный путь к файлу
- `original_name` — исходное имя файла
- `checksum` — SHA-256
- `size_bytes` — размер
- `generated` — является ли сгенерированным
- `required` — обязателен ли для readiness
- `status` — present, missing, outdated

**Artifact Kinds**:

| Kind | Расширения |
|------|-----------|
| cad | .step, .stp, .iges, .igs, .sldprt, .sldasm |
| drawing | .pdf, .dxf, .dwg |
| manufacturing | .dxf, .nc, .tap, .gcode |
| image | .png, .jpg, .jpeg, .webp |
| certificate | .pdf, .md |
| evidence | .png, .jpg, .pdf, .txt |
| generated | .yaml, .csv, .md, .pdf |
| document | .md, .pdf, .docx |

**Placeholder**: если файл обязателен, но не создан — status=missing. Validation engine показывает blocker.

## 9.7. Event

**Назначение**: атомарная запись в истории изменений.

**Поля**:
- `event_id` — уникальный ID события
- `time` — временная метка
- `actor` — пользователь/система
- `type` — полный перечень:
  - `object.created`, `object.deleted`, `object.duplicated`
  - `metadata.updated`
  - `relation.added`, `relation.removed`
  - `artifact.attached`, `artifact.detached`
  - `lifecycle.transitioned`
  - `checkpoint.created`
  - `release.published`
- `object_id` — целевой объект
- `payload` — данные события

**Хранение**: `objects/[id]/history.yaml.log` — append-only файл.

**Формат**: одна JSON-строка на событие (JSONL), разделённая `
`. Не YAML (несмотря на расширение `.yaml.log` — историческая причина, формат JSONL).

**Пример содержимого**:
```jsonl
{"event_id":"evt-20260606-000001","time":"2026-06-06T10:00:00Z","actor":"local-user","type":"object.created","object_id":"a320-prt-0001-v1.0","payload":{"class":"prt","title":"Bracket Motor"}}
{"event_id":"evt-20260606-000002","time":"2026-06-06T11:00:00Z","actor":"local-user","type":"lifecycle.transitioned","object_id":"a320-prt-0001-v1.0","payload":{"from":"draft","to":"in_review","transition":"submit_review"}}
```

**Правила**:
- Только append, никогда не перезаписывается
- Каждая строка — валидный JSON
- `event_id` — уникальный (формат: `evt-YYYYMMDD-NNNNNN`)
- При повреждении строки — она пропускается при чтении с diagnostic warning

## 9.8. Diagnostic

**Назначение**: результат проверки (валидации).

**Поля**:
- `id` — идентификатор диагностики
- `severity` — info, warning, blocker, error
- `code` — код (REQUIRED_METADATA_MISSING, BROKEN_RELATION, CHECKSUM_OUTDATED)
- `object_id` — целевой объект
- `path` — путь к проблемному полю (metadata.material)
- `message` — человекочитаемое сообщение
- `suggested_actions` — список действий для исправления

## 9.9. StandardPart

**Назначение**: стандартное/покупное/каталоговое изделие.

**Особые поля** (дополнительно к Object):
- `std_class` — классификатор (fst, brg, elc, pne, hyd, mat, cab, mot, sen, prof, seal)
- `std_id` — смысловой идентификатор (iso4017-m6x30-a2)
- `normalized.key` — нормализованный ключ для duplicate detection
- `procurement` — preferred_supplier, manufacturer, manufacturer_part_number, supplier_part_number, lead_time_days, min_order_qty, unit_cost, approved_vendor_status
- `quality` — certificates_required, inspection_required, safety_critical

**Формат ID**: `[project]-std-[std_class]-[std_id]-v[major].[minor]`

## 9.10. ReleasePackage

**Назначение**: замороженный релизный пакет.

**Поля**:
- `id` — идентификатор релиза
- `root_object` — корневая сборка
- `scope` — список включённых объектов
- `manifest` — описание состава
- `git_commit` / `git_tag` — привязка к Git
- `checksums` — контрольные суммы всех файлов

## 9.11. ChangeRequest

**Назначение**: запрос на изменение.

**Поля**:
- `id` — идентификатор CR
- `affected_objects` — затронутые объекты
- `impact_report` — отчёт об анализе влияния
- `state` — proposed, impact_analysis, approved, implemented, verified, closed, rejected

## 9.12. BuildRecord (Post-MVP)

**Назначение**: запись о производственном выполнении.

**Поля**:
- `build_id`, `object_id`, `operator`, `started_at`, `completed_at`, `state`, `steps`, `evidence`

---

# 10. Модель хранения данных и файловая структура

## 10.1. Структура workspace

```
MyProject/
├── project.yaml                     # Конфигурация проекта
├── objects/                         # Инженерные объекты
│   └── a320-prt-0001-v1.0/
│       ├── a320-prt-0001-v1.0.md    # Source of Truth: YAML frontmatter + Markdown body
│       ├── history.yaml.log         # Append-only event log (JSONL)
│       ├── files/                   # Физические артефакты
│       │   ├── cad/
│       │   ├── drawings/
│       │   ├── manufacturing/
│       │   ├── inspection/
│       │   ├── certificates/
│       │   └── images/
│       └── generated/               # Сгенерированные файлы
│           ├── bom/
│           ├── reports/
│           └── exports/
├── config/                          # Конфигурация проекта
│   ├── classes.yaml
│   ├── naming.yaml
│   ├── lifecycle.yaml
│   ├── validation.yaml
│   ├── relation-types.yaml
│   ├── file-routing.yaml
│   └── exports.yaml
├── templates/                       # Шаблоны
│   ├── object/
│   ├── project/
│   ├── release/
│   ├── manufacturing/
│   └── snippets/
├── exports/                         # Экспортированные пакеты
├── libraries/                       # Библиотеки
│   └── standard-parts/
├── .plm/                            # Внутренние данные (производные)
│   ├── index.db                     # SQLite индекс
│   ├── transactions/                # Незавершённые транзакции
│   ├── cache/
│   └── recovery/
└── .git/                            # Git репозиторий
```

## 10.2. Уровни данных

| Уровень | Описание | Пересобираем | Примеры |
|---------|----------|-------------|---------|
| **Canonical Source of Truth** | Данные, достаточные для полного восстановления проекта | Нет (это исходник) | project.yaml, objects/*/*.md, config/*.yaml, templates/*.md |
| **Physical Artifacts** | Бинарные/физические файлы | Нет (пользовательские) | .step, .pdf, .dxf, .png |
| **Generated Files** | Сгенерированные документы | Да | generated/bom/*.yaml, generated/reports/*.md |
| **Cache/Index** | Индексы и кэш | Да | .plm/index.db, .plm/cache/* |

## 10.3. Правила записи

1. **MUST**: все записи Source of Truth проходят через transaction layer
2. **MUST**: прямой overwrite .md/.yaml файлов без temp/backup запрещён
3. **MUST**: history.yaml.log — append-only, никогда не перезаписывается
4. **SHOULD**: транзакции идемпотентны (operation_id)
5. **MUST**: индексы обновляются в рамках той же транзакции или сразу после

# 11. Функциональные требования

## 11.1. Управление проектами

### FR-PROJ-001. Создание проекта
**Приоритет**: P0 / MVP
**Описание**: Система MUST позволять создать новый проект из шаблона.
**Входные данные**: project_code, project_title, project_path, naming_standard, enabled_modules
**Критерии приёмки**:
1. Создана директория проекта
2. Создан project.yaml с валидной конфигурацией
3. Созданы config/*.yaml (classes, naming, lifecycle, validation, file-routing)
4. Создана структура templates/
5. Инициализирован Git-репозиторий (.git)
6. Создан initial commit
7. Создан пустой SQLite индекс (.plm/index.db)

### FR-PROJ-002. Открытие проекта
**Приоритет**: P0 / MVP
**Описание**: Система MUST открывать существующий проект: читать project.yaml, сканировать objects/, строить/обновлять индекс.
**Критерии приёмки**:
1. project.yaml прочитан и провалидирован
2. Все объекты в objects/ просканированы
3. Индекс синхронизирован с Source of Truth
4. UI показывает Project Tree с объектами
5. Ошибки парсинга отдельных объектов не блокируют открытие (объект помечается damaged)

### FR-PROJ-003. Валидация проекта
**Приоритет**: P1 / MVP
**Описание**: Система MUST проверять целостность проекта.
**Критерии приёмки**:
1. Проверено наличие и валидность project.yaml
2. Проверена валидность config/*.yaml
3. Проверена валидность всех object records (ID, metadata, relations)
4. Обнаружены битые связи (dangling relations)
5. Обнаружены дубликаты ID
6. Обнаружены устаревшие checksums
7. Результат — список Diagnostics с severity

### FR-PROJ-004. Recovery Mode
**Приоритет**: P1 / MVP
**Описание**: При ошибках целостности система MUST предложить варианты восстановления.
**Действия**:
1. Repair automatically (завершить/откатить незавершённые транзакции)
2. Open diagnostics (показать все проблемы)
3. Export recovery report (YAML/JSON)
4. Ignore and open read-only

---

## 11.2. Управление объектами

### FR-OBJ-001. Создание объекта (Create Object Wizard)
**Приоритет**: P0 / MVP
**Описание**: Система MUST позволять создать объект через 3-шаговый wizard.
**Шаги**:
1. **Identity** — класс, parent object (опционально), preview сгенерированного ID
2. **Properties** — title, metadata (материал, размеры, make/buy, etc.)
3. **Documents & Attachments** — прикрепление файлов, preview создаваемой структуры
**Критерии приёмки**:
1. Создана директория objects/[id]/
2. Создан [id].md с YAML frontmatter и шаблонным Markdown body
3. ID соответствует naming standard (section 16)
4. Создан history.yaml.log с событием object.created
5. Объект появляется в индексе и Project Tree
6. Автоматически созданы связи с parent object
7. Прикреплённые файлы скопированы в objects/[id]/files/
8. Транзакция атомарна (откат при сбое)

### FR-OBJ-002. Редактирование объекта
**Приоритет**: P0 / MVP
**Описание**: Система MUST позволять редактировать title, metadata, Markdown body, relations, artifacts.
**Критерии приёмки**:
1. Изменения сохраняются в .md (YAML frontmatter)
2. Создаётся событие в history.yaml.log
3. Индекс обновляется
4. Autosave с debounce (5 секунд по умолчанию)

### FR-OBJ-003. Удаление объекта
**Приоритет**: P1 / MVP
**Описание**: Удаление MUST быть безопасным: проверить where-used, показать impact, запросить подтверждение.
**Критерии приёмки**:
1. Проверены все where-used связи
2. Показан список затронутых объектов
3. Запрошено подтверждение
4. Объект перемещён в архив/trash (не физически удалён)
5. Событие записано в event log

### FR-OBJ-004. Дублирование объекта
**Приоритет**: P2 / Post-MVP
**Описание**: Система SHOULD позволять создать копию объекта с новым ID.

### FR-OBJ-005. Revision bump
**Приоритет**: P1 / MVP
**Описание**: Система MUST поддерживать повышение версии (minor/major) с созданием новой ревизии.
**Критерии приёмки**:
1. Создаётся новый объект с обновлённой версией
2. Связь derived_from/replaces между ревизиями
3. История изменений сохранена

---

## 11.3. Редактор Markdown/YAML

### FR-EDIT-001. Live Preview
**Приоритет**: P1 / MVP
**Описание**: Редактор MUST показывать Live Preview Markdown body (WYSIWYG или split view).
**Критерии приёмки**:
1. Поддержка Markdown: заголовки, списки, таблицы, ссылки, изображения, code blocks
2. Поддержка Mermaid-диаграмм
3. Preview обновляется без полной перезагрузки

### FR-EDIT-002. YAML Frontmatter Validation
**Приоритет**: P0 / MVP
**Описание**: Ошибки YAML frontmatter MUST показываться inline с указанием строки и столбца.
**Критерии приёмки**:
1. Синтаксические ошибки YAML подсвечиваются
2. Семантические ошибки (неизвестный class, невалидный state) показываются как diagnostics
3. Autocomplete для class, state, relation type

### FR-EDIT-003. Autosave
**Приоритет**: P1 / MVP
**Описание**: Система SHOULD поддерживать autosave с debounce 5 секунд.

### FR-EDIT-004. Diff view
**Приоритет**: P2 / Post-MVP
**Описание**: Система SHOULD показывать diff относительно последнего сохранения, последнего checkpoint, последнего release.

---

## 11.4. Relations и BOM

### FR-BOM-001. BOM View
**Приоритет**: P1 / MVP
**Описание**: Система MUST показывать BOM для выбранной сборки.
**Колонки**: позиция, объект, название, класс, ревизия, статус, количество, ед.изм., make/buy, замены, проблемы.

### FR-BOM-002. Where-used
**Приоритет**: P1 / MVP
**Описание**: Система MUST показывать все объекты, использующие выбранный объект.
**Критерии приёмки**: рекурсивный обход связей, отображение в виде дерева или таблицы.

### FR-BOM-003. Cycle Detection
**Приоритет**: P0 / MVP
**Описание**: Система MUST запрещать циклические зависимости в BOM.
**Критерии приёмки**: при добавлении связи `contains` проверяется отсутствие цикла; при обнаружении — blocker diagnostic.

### FR-BOM-004. BOM Export
**Приоритет**: P1 / MVP
**Описание**: Система MUST экспортировать BOM в CSV, JSON, YAML.

---

## 11.5. Lifecycle / FSM

### FR-FSM-001. Выполнение перехода
**Приоритет**: P0 / MVP
**Описание**: Система MUST позволять выполнить transition, если все guards пройдены.

### FR-FSM-002. Guard Validation
**Приоритет**: P0 / MVP
**Описание**: Перед переходом система MUST показывать: passed guards, failed guards, warnings, recommended actions.

### FR-FSM-003. Transition History
**Приоритет**: P0 / MVP
**Описание**: Каждый переход MUST записываться в history.yaml.log.

### FR-FSM-004. Configurable FSM
**Приоритет**: P0 / MVP
**Описание**: FSM MUST задаваться через YAML (config/lifecycle.yaml).

---

## 11.6. Validation / Readiness

### FR-VAL-001. Валидация объекта
**Приоритет**: P0 / MVP
**Описание**: Система MUST проверять объект по: valid name, required metadata, valid relations, valid artifacts, checksums actual, lifecycle consistency.

### FR-VAL-002. Release Readiness
**Приоритет**: P1 / MVP
**Описание**: Система MUST проверять готовность сборки к релизу: все объекты approved/released, BOM valid, no blockers, все файлы checksummed, обязательные чертежи прикреплены.

### FR-VAL-003. Explain Blockers (Why Can't I Assistant)
**Приоритет**: P1 / MVP
**Описание**: Система MUST объяснять каждый blocker: что не так, на каком объекте, какие действия нужны для исправления.

---

## 11.7. Release Package

### FR-REL-001. Создание Release Package
**Приоритет**: P1 / MVP
**Описание**: Система MUST создавать release package, включающий: manifest, список объектов, список файлов, checksums, Git tag.

### FR-REL-002. Release Manifest
**Приоритет**: P1 / MVP
**Описание**: Manifest MUST содержать: release_id, project, created_at, список объектов с ID/state/checksum, список файлов, Git commit/tag.

---

## 11.8. Standard Parts

### FR-STD-001. Библиотека стандартных изделий
**Приоритет**: P1 / MVP
**Описание**: Система MUST поддерживать создание, поиск, фильтрацию стандартных деталей.

### FR-STD-002. Duplicate Detection
**Приоритет**: P1 / MVP
**Описание**: Система MUST определять возможные дубли по normalized key, стандарту, размерам, материалу, manufacturer part number.

### FR-STD-003. Duplicate Resolution
**Приоритет**: P1 / MVP
**Описание**:
- Exact duplicate MUST блокировать создание
- Probable duplicate MUST требовать явного подтверждения
- Possible equivalent MUST создавать relation, не объединять автоматически

### FR-STD-004. Procurement Data
**Приоритет**: P1 / MVP
**Описание**: Стандартная деталь MUST поддерживать: preferred supplier, manufacturer, manufacturer_part_number, supplier_part_number, lead_time, min_order_qty, unit_cost, approved_vendor_status.

---

## 11.9. Procurement

### FR-PROC-001. Buy List
**Приоритет**: P2 / Post-MVP
**Описание**: Система SHOULD строить buy list из BOM, фильтруя make/buy позиции, с группировкой по поставщикам.

### FR-PROC-002. Substitution Workflow
**Приоритет**: P2 / Post-MVP
**Описание**: Замена MUST проходить workflow: proposed → engineering_review → approved → applied.

---

## 11.10. Change Impact

### FR-CHG-001. Impact Analysis
**Приоритет**: P2 / Post-MVP
**Описание**: Система SHOULD показывать: где используется объект, какие сборки/релизы/инструкции затронуты, нужен ли change request.

---

## 11.11. Git Integration

### FR-GIT-001. Git Status
**Приоритет**: P1 / MVP
**Описание**: Система MUST показывать modified, added, deleted, untracked файлы/объекты в понятном UI.

### FR-GIT-002. Create Checkpoint
**Приоритет**: P1 / MVP
**Описание**: Пользователь MUST иметь возможность создать checkpoint (git commit) через UI без знания git-команд.

### FR-GIT-003. Release Tag
**Приоритет**: P1 / MVP
**Описание**: При выпуске release package MUST автоматически создаваться Git tag.

---

## 11.12. Search

### FR-SRCH-001. Full-text Search
**Приоритет**: P1 / MVP
**Описание**: Система MUST обеспечивать полнотекстовый поиск по: object ID, title, class, state, metadata, Markdown body, artifact names.
**Целевое время отклика**: ≤ 200 мс для проекта до 10 000 объектов.

---

# 12. Нефункциональные требования

## 12.1. Надёжность (Reliability)

| ID | Требование | Приоритет |
|----|-----------|-----------|
| NFR-REL-01 | Все записи Source of Truth MUST быть атомарными (transaction layer) | P0 |
| NFR-REL-02 | Запрещён прямой overwrite критичных файлов без temp/backup | P0 |
| NFR-REL-03 | Все изменения MUST фиксироваться в history.yaml.log | P0 |
| NFR-REL-04 | Система MUST уметь восстановиться после прерванной транзакции | P1 |
| NFR-REL-05 | Система MUST проверять целостность проекта при открытии | P1 |
| NFR-REL-06 | Индекс MUST быть пересобираем из Source of Truth в любой момент | P0 |
| NFR-REL-07 | Операции записи SHOULD быть идемпотентны (operation_id) | P1 |

## 12.2. Производительность (Performance)

| Показатель | Цель MVP | Приоритет |
|-----------|----------|-----------|
| Масштаб проекта | до 10 000 объектов | P0 |
| Открытие проекта (с индексом) | ≤ 5 секунд | P1 |
| Поиск (FTS5) | ≤ 200 мс | P1 |
| Открытие объекта | ≤ 100 мс | P1 |
| Построение BOM (structured) | ≤ 500 мс | P2 |
| Where-used | ≤ 500 мс | P2 |
| Валидация одного объекта | ≤ 200 мс | P2 |
| Валидация проекта (10 000 объектов) | ≤ 30 секунд | P2 |
| Создание объекта (wizard) | ≤ 1 секунда (без копирования больших CAD-файлов) | P1 |

## 12.3. Безопасность (Security)

| ID | Требование | Приоритет |
|----|-----------|-----------|
| NFR-SEC-01 | HTTP JSON-RPC MUST слушать только localhost (127.0.0.1) | P0 |
| NFR-SEC-02 | API MUST использовать session token | P1 |
| NFR-SEC-03 | Секреты (.env, tokens) MUST быть в .gitignore | P0 |
| NFR-SEC-04 | Запрещено выполнение произвольных скриптов из проекта без явного разрешения | P0 |
| NFR-SEC-05 | Вложения MUST открываться безопасно (без автозапуска) | P1 |

## 12.4. Кроссплатформенность

| Платформа | Статус |
|-----------|--------|
| Windows 10+ | MUST support (WebView2) |
| macOS 13+ | MUST support |
| Linux | MUST support (WebKitGTK) |

## 12.5. Расширяемость

- Новые модули MUST регистрироваться через стандартные интерфейсы (commands, queries, validators, views, templates, processes, exports)
- Новые ObjectClass MUST добавляться через config/classes.yaml без изменения кода
- Новые Relation Types MUST добавляться через config/relation-types.yaml
- Новые Validation Rules MUST добавляться через config/validation.yaml и реализацию Rule interface

---

# 13. UI/UX Требования

## 13.1. Общая концепция

UI строится вокруг инженерного рабочего процесса, а не вокруг файловой системы или Git. Пользователь работает с инженерными сущностями (Part, Assembly, Drawing, Work Instruction, Release Package), а не с файлами .md, .yaml, .git.

**Принципы UI**:
1. Пользователь всегда видит текущий проект и выбранный объект
2. Пользователь всегда видит статус объекта
3. Пользователь всегда понимает, что мешает следующему действию
4. Все destructive-действия имеют preview
5. Ошибки формулируются как действия (не "Ошибка 500", а "Отсутствует материал. Заполнить?")
6. Git скрыт за инженерными действиями (Create Checkpoint, Show Changes, Compare with Release)
7. Keyboard-first работа с Command Palette (Ctrl+K)
8. Dense engineering UI (компактные таблицы, status badges, split panels)

## 13.2. Основной Layout

```
┌──────────────────────────────────────────────────────────────────────┐
│ Top Bar: [Logo] [Project ▼] [BADGE] [Search...] [⌘K] [+ New ▼]     │
│          [Git: Clean] [Tasks: 0] [⚙]                                 │
├─────────────┬────────────────────────────────┬───────────────────────┤
│ Left Rail   │ Main Workspace                 │ Right Inspector       │
│             │                                │                       │
│ Project     │ Editor / BOM / Matrix / Graph  │ Properties            │
│ Tree        │                                │ Lifecycle             │
│ [Filter...] │ [Document] [Metadata] [Rel.]   │ Relations (summary)   │
│             │ [BOM] [Files] [Validation]     │ Artifacts             │
│             │ [History] [Diff]               │ Actions               │
│             │                                │ Blockers              │
├─────────────┴────────────────────────────────┴───────────────────────┤
│ Bottom Bar: project | object | save status | validation | Git | idx   │
└──────────────────────────────────────────────────────────────────────┘
```

## 13.3. Основные экраны

| Экран | Назначение | Приоритет |
|-------|-----------|-----------|
| Landing Page | Стартовый экран: Open/Create Project, Recent Projects | P0 |
| Create Project Wizard | Многошаговое создание проекта | P0 |
| Project Tree | Дерево объектов с фильтрацией и поиском | P0 |
| Object Detail Screen | Редактор Markdown + вкладки (Metadata, Relations, BOM, Files, Validation, History) | P0 |
| Markdown Editor | Vditor с Live Preview | P0 |
| Properties Panel | Инлайн-редактирование metadata | P0 |
| Relations Panel | Управление связями объекта | P1 |
| BOM Screen | Structured/flat BOM для сборки | P1 |
| Standard Parts Center | Каталог стандартных деталей | P1 |
| Release Dashboard | Состояние готовности к релизу, блокеры | P1 |
| Validation Panel | Diagnostics сгруппированы по severity | P1 |
| Git Changes Screen | Инженерное представление git diff/status | P1 |
| Settings/Admin | Конфигурация проекта | P1 |
| Manufacturing Matrix | Матрица производственной готовности | P2 |
| Procurement Buy List | Закупочный список | P2 |
| Change Impact Graph | Граф влияния изменений | P2 |

## 13.4. Command Palette

**Горячая клавиша**: `Ctrl+K`

**Команды MVP**:
- Create Part / Assembly / Drawing / Document / Work Instruction / Release Package / Standard Part
- Add Child to BOM / Attach File
- Validate Object / Validate Project
- Submit for Review / Approve / Release
- Explain Blockers
- Show Where Used / Show Change Impact
- Generate BOM / Generate Buy List
- Create Checkpoint / Show Git Diff
- Rebuild Index
- Open Project Settings

**Правила**: каждая команда имеет состояние enabled/disabled. Если disabled — показана причина.

## 13.5. Create Object Wizard

**Шаг 1 — Identity**:
- Выбор класса (prt, asm, drw, std, ...)
- Parent object (опционально)
- Preview: сгенерированный ID, директория

**Шаг 2 — Properties**:
- Title
- Required metadata (зависят от класса)
- Optional metadata

**Шаг 3 — Documents & Attachments**:
- Прикрепление файлов (STEP, PDF, DXF, ...)
- Preview создаваемой структуры: ID, файлы, папки, связи, артефакты

**Backend контракт**:
```
Command: CreateObject(CreateObjectRequest) → CreateObjectResponse
CreateObjectRequest:
  class: prt
  title: "Bracket Motor"
  parent_object_id: "a320-asm-0100-v1.0" (optional)
  metadata: {material: al5052, thickness_mm: 2.0}
  files: [{local_path: "/home/user/cad/bracket.step", role: primary_cad}]
CreateObjectResponse:
  object_id: "a320-prt-0008-v1.0"
  diagnostics: [...]
  refresh: {tree: true, object: "a320-prt-0008-v1.0"}
```

## 13.6. Why Can't I Assistant

Для каждого недоступного действия система MUST показывать:
1. Причину невозможности
2. Список blocker diagnostics
3. Suggested actions с кнопками для перехода

---

# 14. Backend Architecture

## 14.1. Архитектурный стиль

```
Clean Architecture (lite) + Hexagonal Architecture + CQS

cmd/plm          ← entrypoint, DI, Wails bindings
  ↓
api/             ← DTO, mappers, HTTP JSON-RPC adapters
  ↓
app/command      ← Commands (изменяют состояние)
app/query        ← Queries (читают состояние)
app/service      ← Orchestration services
  ↓
core/            ← Domain model (object, relation, artifact, event, diagnostic)
modules/         ← Domain modules (bom, release, standardparts, procurement, ...)
  ↓
store/           ← File repo, transactions, index, recovery
parser/          ← YAML/Markdown parsing
naming/          ← Naming standard engine
validation/      ← Validation engine
process/         ← FSM engine, guards, effects
gitops/          ← Git integration
search/          ← FTS indexing
```

**Правила зависимостей**:
- `core` НЕ импортирует store, api, gitops, модули
- `modules` НЕ импортируют api, Wails, React
- `api` НЕ содержит бизнес-логику
- Циклические импорты запрещены

## 14.2. Структура Go-проекта

```
go-plm/
├── cmd/plm/                    # Entrypoint
│   ├── main.go                 # bootstrap.NewContainer(), desktop.Run()
│   ├── app.go                  # Application container setup
│   └── bindings.go             # Wails API bindings
├── internal/
│   ├── core/                   # Domain model (zero dependencies)
│   │   ├── object/             # Object, ObjectID, ObjectClass
│   │   ├── relation/           # Relation, RelationType
│   │   ├── artifact/           # Artifact, ArtifactKind, ArtifactRole
│   │   ├── revision/           # Revision, Version
│   │   ├── event/              # Event, EventType
│   │   ├── diagnostic/         # Diagnostic, Severity
│   │   ├── lifecycle/          # State, Transition (pure model)
│   │   └── project/            # Project, ProjectConfig (pure model)
│   ├── app/                    # Application layer
│   │   ├── command/            # CreateObject, RunTransition, ...
│   │   ├── query/              # GetObject, SearchObjects, GetBOM, ...
│   │   ├── service/            # ObjectService, ReleaseService, ...
│   │   ├── bus/                # Event bus (опционально)
│   │   └── ports/              # Repository interfaces
│   ├── parser/                 # YAML/Markdown parsing
│   │   ├── frontmatter/        # Split/join frontmatter + body
│   │   ├── markdown/           # Outline, links extraction
│   │   └── yamlx/              # Typed YAML decode/encode
│   ├── naming/                 # Naming Standard v10
│   │   ├── parser.go           # Parse ID
│   │   ├── generator.go        # NextID
│   │   ├── validator.go        # Validate ID
│   │   └── sequence.go         # Sequence store
│   ├── store/                  # Persistence
│   │   ├── fsrepo/             # Filesystem repository
│   │   ├── transaction/        # Atomic file transactions
│   │   ├── index/              # SQLite index (sqlite.go, migrations.go)
│   │   └── recovery/           # Transaction recovery
│   ├── validation/             # Validation engine
│   │   ├── engine/             # Engine (validate object/project/release)
│   │   ├── rules/              # Individual rule implementations
│   │   └── registry/           # Rule registry
│   ├── process/                # FSM
│   │   ├── fsm/                # Pure FSM engine
│   │   ├── guards/             # Guard implementations
│   │   └── effects/            # Effect implementations
│   ├── gitops/                 # Git operations
│   │   └── service.go          # Status, Checkpoint, Tag, Diff, Log
│   ├── search/                 # Search
│   │   ├── sqlitefts/          # SQLite FTS5 index & search
│   │   └── bleve/              # Future advanced search (Post-MVP)
│   ├── modules/                # Domain modules
│   │   ├── bom/                # BOM construction & validation
│   │   ├── release/            # Release package & readiness
│   │   ├── standardparts/      # Standard parts & duplicate detection
│   │   ├── procurement/        # Buy list & ERP export
│   │   ├── manufacturing/      # Manufacturing readiness matrix
│   │   ├── change/             # Change impact analysis
│   │   └── shopfloor/          # Build records (Post-MVP)
│   └── api/                    # API layer
│       ├── dto/                # Data Transfer Objects
│       ├── mapper/             # Domain → DTO mapping
│       └── wails/              # Wails-specific bindings
├── pkg/
│   └── plmclient/              # Public client library (будущее)
├── frontend/                   # React + TypeScript
│   ├── src/
│   │   ├── app/                # App shell, hooks, state
│   │   ├── features/           # Feature modules
│   │   ├── editor/             # Vditor setup
│   │   ├── components/         # Reusable UI components
│   │   └── bindings/           # API client types
│   ├── package.json
│   └── vite.config.ts
├── examples/                   # Demo projects
├── templates/                  # Markdown templates
├── docs/                       # Documentation
├── go.mod
├── go.sum
├── README.md
└── run.ps1
```

## 14.3. Command/Query Contracts

### Response Envelope
```go
type Response[T any] struct {
    OK          bool           `json:"ok"`
    Result      T              `json:"result,omitempty"`
    Diagnostics []Diagnostic   `json:"diagnostics,omitempty"`
    Refresh     RefreshHints   `json:"refresh,omitempty"`
    Error       *APIError      `json:"error,omitempty"`
}

type RefreshHints struct {
    Tree       bool     `json:"tree"`
    Objects    []string `json:"objects,omitempty"`
    BOM        []string `json:"bom,omitempty"`
    Validation bool     `json:"validation"`
    GitStatus  bool     `json:"git_status"`
}
```

### Commands (изменяют состояние)
| Command | Назначение | Приоритет |
|---------|-----------|-----------|
| CreateProject | Создание нового проекта | P0 |
| OpenProject | Открытие существующего проекта | P0 |
| CreateObject | Создание объекта через wizard | P0 |
| UpdateObjectDocument | Сохранение Markdown body | P0 |
| UpdateObjectMetadata | Обновление metadata | P0 |
| AttachArtifact | Копирование файла в объект | P1 |
| AddRelation | Добавление связи | P1 |
| RemoveRelation | Удаление связи | P1 |
| RunTransition | Выполнение FSM-перехода | P0 |
| CreateReleasePackage | Создание релизного пакета | P1 |
| CreateCheckpoint | Git commit | P1 |
| RebuildIndex | Пересборка индекса из Source of Truth | P0 |

### Queries (читают состояние)
| Query | Назначение | Приоритет |
|-------|-----------|-----------|
| GetProject | Конфигурация проекта | P0 |
| ListObjects | Список объектов (дерево/таблица) | P0 |
| GetObject | Полный объект с relations/artifacts | P0 |
| SearchObjects | Полнотекстовый поиск | P1 |
| GetBOM | Structured/flat BOM | P1 |
| GetWhereUsed | Обратные связи | P1 |
| GetChangeImpact | Анализ влияния | P2 |
| ValidateObject | Валидация одного объекта | P0 |
| ValidateProject | Валидация всего проекта | P1 |
| GetReleaseReadiness | Проверка готовности к релизу | P1 |
| GetGitStatus | Состояние Git | P1 |
| GetIndexStatus | Состояние индекса | P1 |

## 14.4. Data Flow Diagram

```
┌──────────┐   HTTP JSON-RPC    ┌──────────────┐    interfaces    ┌─────────────┐
│ Frontend │ ──────────────────▶ │  api/dto     │ ──────────────▶ │ app/command │
│ (React)  │ ◀────────────────── │  api/mapper  │ ◀────────────── │ app/query   │
└──────────┘   {ok, result,     └──────────────┘                 └──────┬──────┘
               diagnostics,                                                  │
               refresh}                                              ┌──────▼──────┐
                                                                     │ app/service │
                                                                     └──────┬──────┘
                                                                            │
                          ┌─────────────────────────────────────────────────┤
                          │                    │                            │
                   ┌──────▼──────┐    ┌───────▼───────┐    ┌───────────────▼──┐
                   │ validation  │    │ modules/bom    │    │ store/fsrepo     │
                   │ engine      │    │ modules/release│    │ store/transaction│
                   │ process/fsm │    │ stdparts/proc  │    │ store/index      │
                   └─────────────┘    └───────────────┘    └───────┬──────────┘
                                                                    │
                         ┌──────────────────────────────────────────┤
                  ┌──────▼──────┐    ┌──────────┐    ┌─────────────▼──┐
                  │ Filesystem  │    │ SQLite   │    │ Git (.git)     │
                  │ objects/*/  │    │ index.db │    │ go-git         │
                  │ config/     │    │ FTS5     │    │                │
                  │ templates/  │    │          │    │                │
                  └─────────────┘    └──────────┘    └────────────────┘
```

**Поток создания объекта**:
1. Frontend → `CreateObject` Command (через api/wails)
2. app/command → naming.NextID() → parser.RenderTemplate()
3. app/command → validation.ValidateObject() (dry-run)
4. app/command → store/transaction.Execute(plan) — атомарная запись
5. app/command → store/index.UpsertObject() — обновление индекса
6. app/command → history.Append(event) — запись в history.yaml.log
7. Response → Frontend с RefreshHints

# 15. Модульная архитектура

## 15.1. Принципы модулей

Каждый модуль (пакет) MUST:
1. Выполнять одну понятную функцию
2. Быть тестируемым отдельно
3. Не иметь циклических импортов
4. Зависеть от интерфейсов, а не от конкретной инфраструктуры
5. Не писать файлы напрямую (только через store/transaction)
6. Не обращаться к UI (api, Wails, React)
7. Не хранить глобальное изменяемое состояние
8. Явно документировать, что делает и чего не делает

## 15.2. Перечень модулей и ответственность

| Модуль | Назначение | Делает | НЕ делает | Зависит от |
|--------|-----------|--------|-----------|------------|
| **core/object** | Доменная модель объекта | Object, ObjectID, Class, State | Чтение Markdown, запись YAML, SQLite, Git | — |
| **core/relation** | Модель связи | Relation, RelationType | Проверка target object, where-used | — |
| **core/artifact** | Модель артефакта | Artifact, Kind, Role, Status | Копирование, checksums, открытие файлов | — |
| **core/event** | Модель события | Event, EventType | Запись history.yaml.log | — |
| **core/diagnostic** | Модель диагностики | Diagnostic, Severity | Исправление проблем | — |
| **parser/frontmatter** | Разделение Markdown | Split/Join frontmatter + body | Валидация бизнес-правил | goldmark, yaml.v3 |
| **parser/yamlx** | YAML encode/decode | Typed YAML | Graph validation | yaml.v3 |
| **parser/markdown** | Markdown backend | Outline, links extraction | UI rendering | goldmark |
| **naming** | Naming Standard v10 | Parse/Generate/Validate ID, Sequence store | Запись project.yaml | core/object |
| **store/fsrepo** | Файловый репозиторий | Чтение/сохранение объектов | BOM, FSM, Git, validation | core/*, store/transaction |
| **store/transaction** | Атомарные операции | Plan, temp files, rename, fsync, recovery | — | — |
| **store/index** | SQLite индекс | objects/relations/artifacts tables, FTS, projections | Считать себя source of truth | core/* |
| **store/recovery** | Восстановление | Поиск незавершённых транзакций, repair | — | store/transaction |
| **validation** | Валидация | Engine, Rule registry, rule implementations | Исправление данных, запись файлов | core/*, naming |
| **process/fsm** | Чистый FSM engine | Definition, Can, Next, Available | Чтение файлов, запись history, Git | — |
| **process/guards** | Guard implementations | required_metadata, valid_name, bom_valid, ... | — | core/*, validation |
| **process/effects** | Effect implementations | append_history, update_index, create_git_tag | — | core/*, store/* |
| **gitops** | Git интеграция | Status, Checkpoint, Tag, Log, Diff | BOM, release readiness, manifest | go-git |
| **search/sqlitefts** | FTS поиск | FTS5 index, search | Bleve facets/highlighting | store/index |
| **modules/bom** | BOM engine | Structured/flat BOM, cycle detection, quantity rollup | Создание объектов, запись YAML, release | core/* |
| **modules/release** | Release engine | Scope, readiness check, manifest | Git tag (вызывается app layer) | core/*, modules/bom |
| **modules/standardparts** | Standard Parts | Normalized key, duplicate detection, std validation | UI rendering | core/* |
| **modules/procurement** | Закупки (Post-MVP) | Buy list, supplier selection, ERP export | Создание std part | core/*, modules/bom |
| **modules/manufacturing** | Производство (Post-MVP) | Manufacturing matrix, readiness | Shop-floor execution | core/* |
| **modules/change** | Изменения (Post-MVP) | Where-used expansion, impact report | Изменение released объектов | core/* |
| **api** | API адаптер | DTO, mappers, Wails bindings | Бизнес-логика | app/* |

---

# 16. Naming Standard v10

## 16.1. Базовый формат ID

```
[project]-[class]-[sequence]-v[major].[minor]
```

**Примеры**:
- `a320-prt-0001-v1.0` — Part 0001, версия 1.0
- `a320-asm-0100-v1.0` — Assembly 0100, версия 1.0
- `a320-drw-0001-v1.0` — Drawing 0001, версия 1.0

## 16.2. Формат стандартной детали

```
[project]-std-[std_class]-[std_id]-v[major].[minor]
```

**Примеры**:
- `a320-std-fst-iso4017-m6x30-a2-v1.0` — болт ISO 4017 M6x30 A2
- `a320-std-brg-6205-2rs-c3-v1.0` — подшипник 6205-2RS C3
- `std-fst-iso4017-m6x30-a2-v1.0` — глобальная стандартная деталь

## 16.3. Правила имён

1. **MUST**: только `[project]-[class]-[sequence]-v[major].[minor]`
2. **MUST**: kebab-case, lowercase
3. **MUST NOT**: кодировать в имени материал, толщину, покрытие, цвет, дату, автора, поставщика, способ производства, статус
4. **SHOULD**: project_code до 16 символов, максимальная длина без расширения — 64 символа

## 16.4. Sequence allocation

| Class | Start | Width | Mode |
|-------|-------|-------|------|
| prt | 0001 | 4 | numeric |
| asm | 0100 | 4 | numeric (при создании проекта инициализируется 0100) |
| drw | 0001 | 4 | numeric |
| doc | 0001 | 4 | numeric |
| cut | 0001 | 4 | inherit_or_numeric |
| bnd | 0001 | 4 | inherit_or_numeric |
| nc | 0001 | 4 | inherit_or_numeric |
| ins | 0001 | 4 | inherit_or_numeric |
| tpc | 0001 | 4 | inherit_or_numeric |
| wi | 0001 | 4 | inherit_or_numeric |
| bom | 0001 | 4 | numeric |
| rel | 0001 | 4 | numeric |
| cr | 0001 | 4 | numeric |
| std | — | — | semantic |

## 16.5. Версионирование

SemVer-подобная схема: v[major].[minor]
- **minor** — совместимое изменение (исправление, уточнение)
- **major** — несовместимое изменение (другая геометрия, другой материал)
- Released объект НЕЛЬЗЯ менять напрямую — создаётся новая revision/version

## 16.6. Sequence Conflict Resolution

Для single-user режима (MVP) конфликты sequence невозможны.

**Для future multi-user**: при конкурентном создании объектов одного класса используется оптимистическая блокировка:
- `NextID()` атомарно инкрементирует sequence в project.yaml
- При конфликте (sequence уже занят) — повторная попытка с новым sequence
- Все операции создания идут через transaction layer с operation_id для идемпотентности

---

# 17. Lifecycle / FSM

## 17.1. Object Lifecycle

```
draft ──submit_review──▶ in_review ──approve──▶ approved ──release──▶ released ──obsolete──▶ obsolete ──archive──▶ archived
  ▲                         │                                                                │
  │────── reject ───────────┘                                                                │
  │                                                                                          │
  └────────────────────────── revise ────────────────────────────────────────────────────────┘

blocked ←── block ── (any state)
blocked ── unblock ──▶ draft
```

**Состояния**: draft, in_review, approved, released, blocked, obsolete, archived

**Переходы**:

| Transition | From | To | Guards | Effects |
|-----------|------|----|--------|---------|
| submit_review | draft | in_review | valid_name, required_metadata, no_broken_relations | — |
| reject | in_review | draft | — | — |
| approve | in_review | approved | no_blocking_issues, checksums_actual | — |
| release | approved | released | no_release_blockers, children_released, bom_valid, checksums_actual | create_git_tag |
| revise | released | draft | — | create_new_revision, create_change_event |
| block | * | blocked | — | — |
| unblock | blocked | draft | — | — |
| obsolete | released | obsolete | — | — |
| archive | obsolete | archived | — | — |

## 17.2. Release Lifecycle

**Состояния**: draft, validating, ready, released, failed, cancelled

**Переходы**:
| Transition | From | To | Guards |
|-----------|------|----|--------|
| validate | draft | validating | — |
| mark_ready | validating | ready | no_release_blockers, manifest_generated |
| publish | ready | released | — |
| fail | validating | failed | — |
| cancel | * | cancelled | — |

## 17.3. Change Request Lifecycle

**Состояния**: proposed, impact_analysis, approved, implemented, verified, closed, rejected

## 17.4. Standard Part Lifecycle

**Состояния**: draft, approved, released, obsolete (упрощённый lifecycle)

## 17.5. Guard Definitions

| Guard ID | Описание | Приоритет |
|----------|----------|-----------|
| valid_name | ID соответствует naming standard | P0 |
| required_metadata | Все required metadata заполнены | P0 |
| no_broken_relations | Все relation target существуют | P0 |
| checksums_actual | Все checksums актуальны | P1 |
| bom_valid | BOM: quantity > 0, unit present, no cycles, target exists | P1 |
| children_released | Все дочерние объекты (contains) в state=released | P1 |
| standard_parts_approved | Все std parts в BOM approved/released | P1 |
| no_release_blockers | Нет diagnostics с severity=blocker | P1 |
| drawing_exists | Обязательный чертёж прикреплён | P2 |
| manufacturing_ready | Все manufacturing файлы готовы | P2 |
| procurement_ready | Все закупочные позиции approved | P2 |

---

# 18. Validation Engine

## 18.1. Уровни валидации

1. **Syntax** — валидность YAML, Markdown
2. **Schema** — соответствие ObjectClass schema
3. **Naming** — соответствие naming standard
4. **Metadata** — required fields, допустимые значения
5. **Relations** — target exists, no dangling refs
6. **Artifacts** — required present, checksums
7. **BOM** — no cycles, quantity > 0, unit present
8. **Lifecycle** — state в допустимом lifecycle, valid transition
9. **Release** — все условия для release выполнены
10. **Standard Parts** — duplicate detection, supplier data
11. **Project Integrity** — уникальность ID, восстанавливаемость индекса

## 18.2. Diagnostic Model

```yaml
diagnostic:
  id: diag-0001
  severity: blocker       # info | warning | blocker | error
  code: REQUIRED_METADATA_MISSING
  object_id: a320-prt-0001-v1.0
  path: metadata.material
  message: "Required metadata field 'material' is missing"
  suggested_actions:
    - fill_material
```

## 18.3. Интерфейсы

```go
type Rule interface {
    ID() string
    Check(ctx context.Context, target Target) ([]diagnostic.Diagnostic, error)
}

type Engine struct {
    Registry *Registry
}

func (e *Engine) ValidateObject(ctx context.Context, id object.ID) ([]diagnostic.Diagnostic, error)
func (e *Engine) ValidateProject(ctx context.Context) ([]diagnostic.Diagnostic, error)
func (e *Engine) ValidateRelease(ctx context.Context, id object.ID) ([]diagnostic.Diagnostic, error)
```

**Правило**: validation НЕ исправляет данные самостоятельно. Quick fix — отдельная application Command.

---

# 19. BOM / eBOM / mBOM

## 19.1. eBOM (Engineering BOM)

Строится из `contains` relations. Рекурсивный обход от корневой сборки.

**Структура**:
```yaml
bom:
  root: a320-asm-0100-v1.0
  rows:
    - row_id: bomrow-0001
      parent_id: a320-asm-0100-v1.0
      child_id: a320-prt-0001-v1.0
      child_class: prt
      quantity: 2
      unit: pcs
      position: "10"
      make_buy: make
      status: valid
```

## 19.2. Flat BOM

Рекурсивный `contains` → multiply quantities → group by child_id.

## 19.3. mBOM (Manufacturing BOM)

Расширение eBOM: добавляются manufacturing files, process plans, work instructions, inspection objects, tools, consumables, packaging.

## 19.4. Released BOM

Создаётся ТОЛЬКО из approved/released объектов. Проверки:
- нет draft children
- все обязательные чертежи есть
- все checksum актуальны
- стандартные детали approved/released
- нет release blockers

## 19.5. BOM Module API

```go
type Service struct {
    Objects   ObjectReader
    Relations RelationReader
}

func (s *Service) GetStructured(ctx context.Context, root object.ID) (StructuredBOM, error)
func (s *Service) GetFlat(ctx context.Context, root object.ID) (FlatBOM, error)
func (s *Service) Validate(ctx context.Context, root object.ID) ([]diagnostic.Diagnostic, error)
```

---

# 20. Standard Parts Center

## 20.1. Почему отдельная модель

Стандартные детали отличаются от проектных (prt):
- Используются во многих сборках
- Имеют смысловой идентификатор (не numeric sequence)
- Часто покупаются (make_buy: buy)
- Имеют поставщиков, manufacturer part numbers
- Имеют аналоги и замены
- Могут быть глобальными (не привязаны к проекту)
- Нуждаются в duplicate detection
- Часто сразу released

## 20.2. Классификация (std_class)

| std_class | Название | Примеры |
|-----------|----------|---------|
| fst | Fastener | bolt, nut, washer, screw, rivet |
| brg | Bearing | ball bearing, roller bearing, linear bearing |
| elc | Electrical | resistor, capacitor, connector, LED, fuse, relay |
| pne | Pneumatic | fitting, valve, cylinder, filter, regulator |
| hyd | Hydraulic | fitting, valve, cylinder, pump |
| mat | Material | sheet, bar, tube, profile (стандартный) |
| cab | Cable | cable, wire, harness component |
| mot | Motor | stepper, servo, DC motor, gearmotor |
| sen | Sensor | temperature, pressure, proximity, encoder |
| prof | Profile | aluminum extrusion, rail |
| seal | Seal | o-ring, gasket, lip seal |

## 20.3. Normalized Key

**Назначение**: строка для быстрого поиска дублей.

**Правила**: lowercase, убрать пробелы, нормализовать стандарт, параметры.

**Примеры**:
- Крепёж: `fst:iso4017:m6:30:a2`
- Подшипник: `brg:6205:2rs:c3`
- Резистор: `elc:res:10k:0603:1pct:0.1w`

## 20.4. Duplicate Detection

| Уровень | Значение | Действие |
|---------|----------|----------|
| exact_duplicate | Абсолютный дубль (совпадает std_class + standard + все параметры) | **BLOCK** создание |
| probable_duplicate | Вероятный дубль (совпадает std_class + большинство параметров) | **CONFIRM** — запросить подтверждение |
| possible_equivalent | Возможный эквивалент (другой стандарт, те же параметры) | **RELATE** — создать relation equivalent_to |
| substitutable | Может быть заменой | **SUGGEST** — предложить relation substitutes |
| not_duplicate | Точно разное | Пропустить |

## 20.5. Standard Part YAML

```yaml
id: a320-std-fst-iso4017-m6x30-a2-v1.0
project: a320
class: std
std_class: fst
std_id: iso4017-m6x30-a2
version: "1.0"
state: released
title: "Hex bolt ISO 4017 M6x30 A2"
metadata:
  make_buy: buy
  category: fastener
  family: bolt
  standard: iso4017
  size: m6x30
  thread: m6
  length_mm: 30
  material: stainless_steel
  material_grade: a2
  unit: pcs
normalized:
  key: "fst:iso4017:m6:30:a2"
procurement:
  preferred_supplier: supplier-acme
  manufacturer: acme
  manufacturer_part_number: acme-iso4017-m6x30-a2
  supplier_part_number: acme-123456
  lead_time_days: 7
  min_order_qty: 100
  package_qty: 100
  unit_cost: {amount: 0.08, currency: usd}
  approved_vendor_status: approved
quality:
  certificates_required: [certificate_of_conformity]
  inspection_required: false
  safety_critical: false
relations:
  - type: equivalent_to
    to: a320-std-fst-din933-m6x30-a2-v1.0
  - type: substitutes
    to: a320-std-fst-iso4017-m6x30-a4-v1.0
```

## 20.6. Standard Parts Module API

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

---

# 21. Procurement & Buy List (Post-MVP)

## 21.1. Buy List Generation

Строится из BOM: фильтр make_buy=buy, группировка по поставщикам, quantity rollup.

## 21.2. Substitution Workflow

```
proposed → engineering_review → approved → applied
                                    ↓
                                rejected
```

## 21.3. ERP Export

Форматы: CSV, JSON, ERP import package (настраиваемый mapping).

---

# 22. Release Package

## 22.1. Release Scope

Определяется рекурсивным обходом `contains` + `includes` от корневой сборки.

## 22.2. Release Readiness Check

- Все объекты в scope: state = approved или released
- BOM валиден (no cycles, quantities, units)
- Все checksums актуальны
- Обязательные чертежи прикреплены
- Стандартные детали approved/released
- Нет blocker diagnostics

## 22.3. Manifest Structure

```yaml
release:
  id: a320-rel-0100-v1.0
  root_object: a320-asm-0100-v1.0
  created_at: 2026-06-06T10:00:00Z
  created_by: local-user
  git_commit: abc123
  git_tag: release/a320-rel-0100-v1.0
objects:
  - id: a320-asm-0100-v1.0
    class: asm
    state: released
    checksum: sha256:...
  - id: a320-prt-0001-v1.0
    class: prt
    state: released
    checksum: sha256:...
files:
  - object_id: a320-prt-0001-v1.0
    path: files/cad/a320-prt-0001-v1.0.step
    checksum: sha256:...
```

---

# 23. Change Management (Post-MVP)

## 23.1. Change Impact Radar

Для выбранного объекта показывает:
- Где используется (where-used, все уровни)
- Какие сборки затронуты
- Какие релизы затронуты
- Какие инструкции затронуты
- Какие закупочные позиции затронуты
- Нужен ли формальный Change Request

## 23.2. Change Request Lifecycle

```
proposed → impact_analysis → approved → implemented → verified → closed
                 ↓                           ↓
             rejected                    rejected
```

---

# 24. Git Integration

## 24.1. Принцип

Пользователь НЕ видит git-команды. Все Git-операции скрыты за инженерными действиями:
- "Save" → stage + commit (при автосохранении или ручном)
- "Create Checkpoint" → git commit с сообщением
- "Show Changes" → domain-level diff (объекты, metadata, BOM, а не строки)
- "Compare with Release" → diff между текущим состоянием и release tag
- "Create Release Tag" → git tag release/[id]

## 24.2. Git Operations API

```go
type Service interface {
    Status(ctx context.Context) (Status, error)
    Checkpoint(ctx context.Context, req CheckpointRequest) (CheckpointResult, error)
    CreateTag(ctx context.Context, tag string, message string) error
    Log(ctx context.Context, limit int) ([]Commit, error)
    Diff(ctx context.Context, base string, head string) (Diff, error)
}
```

---

# 25. Search & Indexing

## 25.1. MVP: SQLite FTS5

**Индексируемые поля**:
- object id, title, class, state
- metadata (ключи и значения)
- Markdown headings и body
- artifact names (original_name)
- standard part normalized key
- supplier/manufacturer part numbers

**API**: `Search(ctx, query) → []SearchResult`

## 25.2. Index Schema Migration

SQLite индекс имеет версионированную схему. Миграции хранятся в `store/index/migrations.go`.

**Формат миграции**:
```go
type Migration struct {
    Version int
    Up      string  // SQL для применения
}
```

**Принцип**: миграции forward-only. При повышении версии старый индекс удаляется и пересобирается через `RebuildIndex` из Source of Truth. Индекс всегда пересобираем — down-миграции не нужны.

**Версионирование**: `PRAGMA user_version` хранит текущую версию схемы. При несовпадении версии кода и индекса — автоматический `RebuildIndex` при открытии проекта.

## 25.3. Post-MVP: Bleve

- Fuzzy search
- Facets (class, state, std_class, supplier)
- Highlighting
- Duplicate candidates scoring
- Advanced standard parts search

---

# 26. Security (см. секцию 12.3)

**Основные требования безопасности** — в секции 12.3 (NFR-SEC-01..05).

**Дополнительное требование**:
| ID | Требование | Приоритет |
|----|-----------|-----------|
| SEC-06 | .gitignore включает: .env, *.token, .plm/cache/, .plm/transactions/ | P0 |

---

# 27. Reliability & Recovery

## 27.1. Transaction Layer

**Принцип**: все записи Source of Truth проходят через transaction layer.

**Алгоритм записи**:
1. Создать transaction plan (WriteOp, CopyOp, DeleteOp, RenameOp)
2. Создать temp-файлы в .plm/transactions/
3. Записать содержимое в temp-файлы
4. fsync() на temp-файлах
5. Atomic rename (os.Rename) в целевой путь
6. fsync() на директории
7. Записать commit marker
8. Очистить temp

**Recovery**:
- При старте проверить .plm/transactions/ на незавершённые транзакции
- Предложить: Complete (доделать rename) или Rollback (удалить temp)

## 27.2. Идемпотентность

Каждая операция записи имеет `operation_id`. При повторном вызове с тем же id — возврат предыдущего результата без изменений.

---

# 28. Performance Requirements (см. секцию 12.2)

Целевые показатели производительности — в секции 12.2 (Нефункциональные требования / Производительность).

---

# 29. Testing Strategy

## 29.1. Уровни тестирования

| Уровень | Инструмент | Что тестирует |
|---------|-----------|---------------|
| Unit tests (Go) | Go test + testify | Каждый модуль изолированно |
| Integration tests (Go) | Go test + test fixtures | Цепочки: parser→naming→store→index |
| Golden tests | Go test + golden files | YAML/Markdown парсинг и генерация |
| Fixture tests | Go test + testdata/ | Готовые проекты для проверки целостности |
| Frontend unit tests | Vitest | React компоненты, хуки |
| E2E tests | Playwright | Сценарии UI (create project → create object → validate → release) |
| Performance tests | Go benchmark | Индексация N объектов, поиск, BOM |
| Recovery tests | Go test + fault injection | Прерванные транзакции, повреждённые файлы |

## 29.2. Минимальные тесты по модулям

**naming**:
- parse valid ID → ParsedID
- reject invalid ID (wrong format, bad characters)
- generate next ID (sequence increment)
- handle std ID (semantic, not numeric)

**bom**:
- build structured BOM (2-level assembly)
- build flat BOM with quantity rollup
- detect cycle (A→B→C→A)
- validate BOM row (missing quantity, missing unit)

**transaction**:
- atomic write success
- interrupted transaction recovery
- rollback on error
- idempotent retry with same operation_id

**validation**:
- required_metadata missing → blocker
- dangling relation → blocker
- checksum outdated → warning

**standardparts**:
- exact duplicate blocks creation
- probable duplicate requires confirmation
- normalized key generation for different std_classes

---

# 30. Observability / Logging / Diagnostics

## 30.1. Логирование

- **Backend**: structured logging (slog) — level, module, operation_id, duration, error
- **Frontend**: console.error для ошибок, debug mode для детального лога API-вызовов
- **History**: все изменения в history.yaml.log (append-only JSONL)

## 30.2. Diagnostics

- Все проверки возвращают []Diagnostic с severity
- Diagnostics доступны через API: ValidateObject, ValidateProject, GetReleaseReadiness
- UI показывает diagnostics сгруппированными по severity

## 30.3. Admin Commands

| Command | Назначение |
|---------|-----------|
| RebuildIndex | Полная пересборка SQLite из Source of Truth |
| ValidateProject | Полная проверка целостности |
| ExportDiagnostics | Экспорт всех diagnostics в JSON/YAML |
| ShowRecoveryItems | Показать незавершённые транзакции |

---

# 31. Конфигурация системы

## 31.1. Project Config (project.yaml)

```yaml
project:
  code: a320
  title: "A320 Test Project"
  version: 1
  created_at: "2026-06-06T10:00:00Z"
naming:
  standard: v10
  pattern: "[project]-[class]-[sequence]-v[major].[minor]"
  case: kebab
sequences:
  prt: 0007
  asm: 0100
  # ...
modules:
  objects: true
  bom: true
  release: true
  manufacturing: true
  procurement: true
  standardparts: true
  shopfloor: false    # Post-MVP
  change_management: true
storage:
  source_of_truth: markdown_yaml
  index: .plm/index.db
git:
  enabled: true
  auto_checkpoint: false
  release_tags: true
```

## 31.2. Object Classes (config/classes.yaml)

```yaml
classes:
  prt:
    title: Part
    sequence_mode: numeric
    template: templates/object/prt.md
    lifecycle: object_lifecycle
    required_metadata: [unit, make_buy]
    optional_metadata: [material, thickness_mm, manufacturing_method, surface_treatment]
  asm:
    title: Assembly
    sequence_mode: numeric
    template: templates/object/asm.md
    lifecycle: object_lifecycle
    required_metadata: [unit]
  # ...
  std:
    title: Standard Part
    sequence_mode: semantic
    template: templates/object/std.md
    lifecycle: standard_part_lifecycle
    required_metadata: [std_class, std_id, make_buy]
```

## 31.3. Lifecycle Config (config/lifecycle.yaml)

```yaml
processes:
  object_lifecycle:
    initial: draft
    states: [draft, in_review, approved, released, blocked, obsolete, archived]
    transitions:
      submit_review:
        from: draft
        to: in_review
        guards: [valid_name, required_metadata, no_broken_relations]
      approve:
        from: in_review
        to: approved
        guards: [no_blocking_issues, checksums_actual]
      release:
        from: approved
        to: released
        guards: [no_release_blockers, children_released, bom_valid]
        effects: [create_git_tag]
      # ...
```

## 31.4. Validation Config (config/validation.yaml)

```yaml
validators:
  valid_name:
    type: naming
  required_metadata:
    type: required_fields
  no_broken_relations:
    type: relation_integrity
  checksums_actual:
    type: checksum
  children_released:
    type: relation_state
    relation: contains
    allowed_states: [released]
  bom_valid:
    type: bom
    checks: [quantity_positive, unit_present, no_cycles, target_exists]
  no_release_blockers:
    type: blockers
```

---

# 32. API / Command / Query Contracts

## 32.1. Коммуникация

- **Транспорт**: HTTP JSON-RPC 2.0 на localhost
- **Формат**: JSON
- **Response envelope**: `{ok, result, diagnostics, refresh, error}`

## 32.2. Полный перечень Commands

| Command | Request | Response | Приоритет |
|---------|---------|----------|-----------|
| CreateProject | {code, title, path, naming, modules} | {project} | P0 |
| OpenProject | {path} | {project, tree} | P0 |
| CreateObject | {class, title, metadata, parent_id, files} | {object_id, diagnostics, refresh} | P0 |
| UpdateObjectDocument | {object_id, markdown} | {diagnostics, refresh} | P0 |
| UpdateObjectMetadata | {object_id, metadata} | {diagnostics, refresh} | P0 |
| AttachArtifact | {object_id, local_path, role} | {artifact, refresh} | P1 |
| AddRelation | {from_id, to_id, type, quantity, unit} | {relation, diagnostics} | P1 |
| RemoveRelation | {from_id, to_id, type} | {refresh} | P1 |
| RunTransition | {object_id, transition} | {new_state, diagnostics, effects} | P0 |
| CreateReleasePackage | {root_object_id} | {release_id, manifest} | P1 |
| CreateCheckpoint | {message, scope} | {commit_hash} | P1 |
| RebuildIndex | {} | {status, stats} | P0 |
| DeleteObject | {object_id} | {archived_path} | P1 |

## 32.3. Полный перечень Queries

| Query | Params | Response | Приоритет |
|-------|--------|----------|-----------|
| GetProject | {} | {project_config, stats} | P0 |
| ListObjects | {filter, sort} | {objects[]} | P0 |
| GetObject | {object_id} | {object, relations, artifacts} | P0 |
| SearchObjects | {query, limit} | {results[]} | P1 |
| GetBOM | {root_id, type} | {bom_rows[]} | P1 |
| GetWhereUsed | {object_id} | {parents[]} | P1 |
| GetChangeImpact | {object_id} | {impact_report} | P2 |
| ValidateObject | {object_id} | {diagnostics[]} | P0 |
| ValidateProject | {} | {diagnostics[]} | P1 |
| GetReleaseReadiness | {root_id} | {readiness, blockers[]} | P1 |
| GetGitStatus | {} | {status} | P1 |
| GetIndexStatus | {} | {stats} | P1 |
| GetObjectHistory | {object_id, limit} | {events[]} | P1 |
| ListAvailableCommands | {context} | {commands[]} | P0 |

---

# 33. Roadmap и этапы реализации

## Phase 0 — Repository Skeleton (Week 1-2)
**Цель**: инициализация проекта, CI, минимальная структура
- go.mod, структура директорий, .gitignore, README
- GitHub Actions CI (lint, test, build)
- Makefile / Taskfile для сборки
**Acceptance**: `go build ./...` успешно

## Phase 1 — Core Domain (Week 2-4)
**Цель**: чистые доменные типы
- core/object, core/relation, core/artifact, core/event, core/diagnostic
- unit tests (100% покрытие core)
**Acceptance**: все core-типы определены, иммутабельны, покрыты тестами

## Phase 2 — Filesystem Store (Week 3-5)
**Цель**: надёжное файловое хранение
- store/fsrepo: чтение/запись объектов
- store/transaction: атомарные операции, recovery
- unit + integration tests
**Acceptance**: создание/чтение/удаление объекта через fsrepo + атомарность транзакций

## Phase 3 — Parser / Naming / Validation (Week 4-7)
**Цель**: парсинг YAML/Markdown, naming engine, validation engine
- parser/frontmatter, parser/yamlx, parser/markdown
- naming: parse/generate/validate ID, sequence store
- validation/engine + базовые rules
- tests: golden tests для парсинга, unit tests для naming, fixture tests для validation
**Acceptance**: парсинг полного .md файла → ObjectRecord; генерация ID по naming standard

## Phase 4 — SQLite Index (Week 6-8)
**Цель**: пересобираемый индекс
- store/index: objects, relations, artifacts, FTS5, projections
- RebuildIndex из Source of Truth
- unit + integration tests
**Acceptance**: RebuildIndex после удаления .plm/index.db восстанавливает все данные

## Phase 5 — FSM Lifecycle (Week 7-9)
**Цель**: конфигурируемый FSM engine
- process/fsm: pure FSM
- process/guards: guard implementations
- process/effects: effect implementations
- RunTransition command
- tests: все переходы с guards
**Acceptance**: объект проходит draft → in_review → approved → released с guard validation

## Phase 6 — Desktop MVP UI (Week 8-14)
**Цель**: работающее desktop-приложение
- Wails 3 + React + TypeScript + Vite + Tailwind
- Landing Page, Create/Open Project
- Project Tree, Object Detail Screen
- Markdown Editor (Vditor) с Live Preview
- Properties Panel, Relations Panel
- Command Palette (Ctrl+K)
- Create Object Wizard (3 шага)
- Git Status / Checkpoint UI
**Acceptance**: открыть проект, создать объект, отредактировать, изменить lifecycle

## Phase 7 — BOM (Week 12-15)
**Цель**: BOM engine + UI
- modules/bom: structured/flat BOM, cycle detection
- BOM View в UI, where-used
**Acceptance**: BOM строится из сборки, циклы детектятся

## Phase 8 — Standard Parts (Week 14-17)
**Цель**: стандартные детали
- modules/standardparts: normalize, find duplicates, validate
- Standard Parts Center UI
- Duplicate detection и resolution
**Acceptance**: создание std part с duplicate check, поиск по normalized key

## Phase 9 — Git UX (Week 15-17)
**Цель**: инженерный Git-интерфейс
- gitops service
- UI: Git status, Create Checkpoint, Show Changes (domain-level diff)
**Acceptance**: checkpoint создаётся через UI, пользователь не видит git-команд

## Phase 10 — Release Package (Week 16-19)
**Цель**: релизный пакет
- modules/release: scope, readiness, manifest
- Release Dashboard UI
- Release tag в Git
**Acceptance**: создание release package с manifest и Git tag

## Phase 11 — Manufacturing / Procurement (Week 18-22) — Post-MVP
- Manufacturing matrix, Process plans, Work instructions
- Procurement buy list, substitution workflow
- ERP export

## Phase 12 — Change Impact (Week 20-23) — Post-MVP
- Change impact radar
- Change request lifecycle

## Phase 13 — Hardening (Week 22-26)
**Цель**: стабильность, производительность, документация
- Performance optimization (10K objects)
- E2E tests (Playwright)
- Documentation (user + developer)
- Bug fixes
**Acceptance**: все E2E тесты проходят, perf targets достигнуты

---

# 34. MVP Scope (Checklist)

- [ ] FR-PROJ-001: Создание проекта
- [ ] FR-PROJ-002: Открытие проекта
- [ ] FR-PROJ-003: Валидация проекта
- [ ] FR-PROJ-004: Recovery Mode
- [ ] FR-OBJ-001: Create Object Wizard
- [ ] FR-OBJ-002: Редактирование объекта
- [ ] FR-OBJ-003: Удаление объекта
- [ ] FR-OBJ-005: Revision bump
- [ ] FR-EDIT-001: Live Preview
- [ ] FR-EDIT-002: YAML validation
- [ ] FR-EDIT-003: Autosave
- [ ] FR-BOM-001: BOM View
- [ ] FR-BOM-002: Where-used
- [ ] FR-BOM-003: Cycle detection
- [ ] FR-BOM-004: BOM export (CSV/YAML)
- [ ] FR-FSM-001, 002, 003, 004: Lifecycle FSM
- [ ] FR-VAL-001, 002, 003: Validation + Explain Blockers
- [ ] FR-REL-001, 002: Release package
- [ ] FR-STD-001, 002, 003, 004: Standard parts
- [ ] FR-GIT-001, 002, 003: Git integration
- [ ] FR-SRCH-001: Full-text search
- [ ] NFR-REL-01..07: Reliability
- [ ] NFR-SEC-01..05: Security
- [ ] All perf targets ≤ 10K objects

---

# 35. Out of Scope (явно исключено из MVP)

- Многопользовательская серверная PLM
- Облачная синхронизация
- Ролевая модель enterprise-уровня
- Электронная подпись
- CAD-редактор/CAD-viewer
- ERP/MES/WMS (кроме базового экспорта)
- Shop-floor execution (Build Records, QR tokens)
- Bleve advanced search
- Сложный Git merge/conflict resolution
- Мобильная версия

---

# 36. Acceptance Criteria (Definition of Done)

Функциональность считается готовой, если:
1. Реализована в соответствии со спецификацией
2. Покрыта unit-тестами (≥ 80% coverage для core, naming, validation, modules)
3. Покрыта интеграционными/фикстурными тестами (для store, parser, index, lifecycle)
4. Все diagnostics (errors, warnings, blockers) обработаны и отображаются в UI
5. Транзакции атомарны (rollback при сбое)
6. Индекс пересобираем (RebuildIndex из Source of Truth даёт идентичный результат)
7. UI соответствует layout и behaviour спецификации
8. Документация (README, godoc) обновлена

---

# 37. Risks and Mitigations

| Риск | Вероятность | Влияние | Митигация |
|------|------------|---------|-----------|
| Wails 3 нестабилен (альфа/бета) | Medium | High | Зафиксировать версию, иметь fallback на Wails 2 или Tauri |
| Производительность SQLite FTS5 на 10K+ объектов | Low | Medium | Оптимизация индекса, lazy loading, pagination |
| Сложность синхронизации frontmatter ↔ index | Medium | Medium | Строгий transaction layer, RebuildIndex как аварийное восстановление |
| Git merge conflicts при параллельной работе (будущее) | High (для multi-user) | High | Не в MVP. При переходе к multi-user — продумать стратегию разрешения конфликтов |
| Сложность duplicate detection для всех std_classes | Medium | Medium | Начинать с fst (fasteners), добавлять классы итеративно |
| Порог входа для инженеров без Git-опыта | Medium | Medium | Полное скрытие Git за инженерными действиями, обучение через UI |
| Кроссплатформенные проблемы Wails (Linux/WebKitGTK) | Medium | Medium | Раннее тестирование на всех платформах, CI матрица |

---

# 38. Open Questions (Decision Required)

### OQ-01: Глобальная библиотека стандартных деталей
**Assumption**: В MVP стандартные детали хранятся внутри проекта (objects/...).
**Question**: Нужна ли общая глобальная библиотека std parts, общая для нескольких проектов?
**Recommendation**: MVP — только проектные std parts. Глобальная библиотека — Phase 8+.

### OQ-02: Формат item.json vs только .md
**Conflict**: Plm main tz.md использует item.json, Plm inf.md говорит "YAML frontmatter = source of truth".
**Recommendation**: Принять единый подход — только .md с YAML frontmatter. item.json — generated/derived (если нужен).
**Decision required**: Подтвердить отказ от item.json в пользу единого .md.

### OQ-03: CodeMirror 6 vs Vditor (RESOLVED)
**Decision**: Принят Vditor как единственный редактор (WYSIWYG + instant render + split preview).
CodeMirror 6 исключён. Vditor покрывает Markdown body. YAML frontmatter редактируется через structured Properties Panel, а не raw text.

### OQ-04: Автосохранение vs ручное сохранение
**Question**: Autosave с debounce 5s или явная кнопка Save?
**Recommendation**: Autosave 5s + ручная кнопка Save. Autosave не создаёт checkpoint (git commit), только сохраняет файл.
**Decision required**: Подтвердить.

### OQ-05: Shop-floor в MVP или нет
**Conflict**: Plm main tz.md включает shop-floor, MODULES_SPEC помечает как Post-MVP.
**Recommendation**: Shop-floor (build records, QR tokens) однозначно Post-MVP. MVP = engineering readiness, не execution.
**Decision required**: Подтвердить исключение shop-floor из MVP.

### OQ-06: SQLite или JSON-индекс
**Conflict**: Plm main tz.md говорит "SQLite", Plm tz.md упоминает ".plm/index.json".
**Recommendation**: Принять SQLite (modernc.org/sqlite, без CGO) с FTS5. JSON-индекс deprecated.
**Decision required**: Подтвердить SQLite как единственный формат индекса.

### OQ-07: Bleve в MVP или Post-MVP
**Recommendation**: Bleve — Post-MVP. FTS5 достаточно для MVP.
**Decision required**: Подтвердить.

### OQ-08: Wails 3 (альфа) стабильность
**Assumption**: Wails 3 используется как desktop shell.
**Risk**: На момент разработки Wails 3 может быть нестабилен.
**Recommendation**: Стартовать с Wails 2, запланировать миграцию на Wails 3 когда стабилизируется.
**Decision required**: Выбрать версию Wails.

---

# 39. Appendices

## Appendix A — Requirement Index (ID Mapping)

| ID | Название | Тип | Приоритет | Модуль | MVP/Post |
|----|----------|-----|-----------|--------|----------|
| FR-PROJ-001 | Создание проекта | Functional | P0 | app/command | MVP |
| FR-PROJ-002 | Открытие проекта | Functional | P0 | app/command | MVP |
| FR-PROJ-003 | Валидация проекта | Functional | P1 | validation | MVP |
| FR-PROJ-004 | Recovery Mode | Functional | P1 | store/recovery | MVP |
| FR-OBJ-001 | Create Object Wizard | Functional | P0 | app/command | MVP |
| FR-OBJ-002 | Редактирование объекта | Functional | P0 | app/command | MVP |
| FR-OBJ-003 | Удаление объекта | Functional | P1 | app/command | MVP |
| FR-OBJ-004 | Дублирование объекта | Functional | P2 | app/command | Post-MVP |
| FR-OBJ-005 | Revision bump | Functional | P1 | app/command | MVP |
| FR-EDIT-001 | Live Preview | Functional | P1 | frontend/editor | MVP |
| FR-EDIT-002 | YAML validation | Functional | P0 | parser + frontend | MVP |
| FR-EDIT-003 | Autosave | Functional | P1 | frontend | MVP |
| FR-EDIT-004 | Diff view | Functional | P2 | frontend | Post-MVP |
| FR-BOM-001 | BOM View | Functional | P1 | modules/bom | MVP |
| FR-BOM-002 | Where-used | Functional | P1 | modules/bom | MVP |
| FR-BOM-003 | Cycle detection | Functional | P0 | modules/bom | MVP |
| FR-BOM-004 | BOM Export | Functional | P1 | modules/bom | MVP |
| FR-FSM-001 | Execute Transition | Functional | P0 | process/fsm | MVP |
| FR-FSM-002 | Guard Validation | Functional | P0 | process/guards | MVP |
| FR-FSM-003 | Transition History | Functional | P0 | process/effects | MVP |
| FR-FSM-004 | Configurable FSM | Functional | P0 | process/fsm | MVP |
| FR-VAL-001 | Validate Object | Functional | P0 | validation | MVP |
| FR-VAL-002 | Release Readiness | Functional | P1 | modules/release | MVP |
| FR-VAL-003 | Explain Blockers | Functional | P1 | validation + UI | MVP |
| FR-REL-001 | Create Release Package | Functional | P1 | modules/release | MVP |
| FR-REL-002 | Release Manifest | Functional | P1 | modules/release | MVP |
| FR-STD-001 | Standard Parts Library | Functional | P1 | modules/standardparts | MVP |
| FR-STD-002 | Duplicate Detection | Functional | P1 | modules/standardparts | MVP |
| FR-STD-003 | Duplicate Resolution | Functional | P1 | modules/standardparts | MVP |
| FR-STD-004 | Procurement Data | Functional | P1 | modules/standardparts | MVP |
| FR-PROC-001 | Buy List | Functional | P2 | modules/procurement | Post-MVP |
| FR-PROC-002 | Substitution Workflow | Functional | P2 | modules/procurement | Post-MVP |
| FR-CHG-001 | Impact Analysis | Functional | P2 | modules/change | Post-MVP |
| FR-GIT-001 | Git Status | Functional | P1 | gitops | MVP |
| FR-GIT-002 | Create Checkpoint | Functional | P1 | gitops | MVP |
| FR-GIT-003 | Release Tag | Functional | P1 | gitops | MVP |
| FR-SRCH-001 | Full-text Search | Functional | P1 | search/sqlitefts | MVP |
| NFR-REL-01 | Atomic writes | Non-functional | P0 | store/transaction | MVP |
| NFR-REL-02 | No direct overwrite | Non-functional | P0 | store/transaction | MVP |
| NFR-REL-03 | Event log | Non-functional | P0 | store/fsrepo | MVP |
| NFR-REL-04 | Recovery | Non-functional | P1 | store/recovery | MVP |
| NFR-REL-05 | Integrity check | Non-functional | P1 | validation | MVP |
| NFR-REL-06 | Rebuildable index | Non-functional | P0 | store/index | MVP |
| NFR-REL-07 | Idempotency | Non-functional | P1 | store/transaction | MVP |
| NFR-SEC-01 | Localhost only | Non-functional | P0 | api | MVP |
| NFR-SEC-02 | Session token | Non-functional | P1 | api | MVP |
| NFR-SEC-03 | Secrets in .gitignore | Non-functional | P0 | project setup | MVP |
| NFR-SEC-04 | No script execution | Non-functional | P0 | store/fsrepo | MVP |
| NFR-SEC-05 | Safe attachments | Non-functional | P1 | frontend | MVP |
