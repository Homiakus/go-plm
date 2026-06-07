# Техническое задание

# Markdown-PLM / go-plm

## Локальная Git-native PLM/PDM-система для инженерной документации, BOM, релизов и производственных процессов

Версия документа: `0.1`  
Статус: `draft`  
Дата: `2026-06-06`  
Репозиторий: `Homiakus/go-plm`  
Целевая платформа: Windows / macOS / Linux  
Архитектурный стиль: Desktop-first, file-based, Git-native, FSM-driven

---

# 1. Назначение системы

Markdown-PLM — это локальная desktop PLM/PDM-система для управления инженерными объектами, документацией, спецификациями, жизненными циклами, релизами, производственной готовностью, закупочными списками и изменениями.

Система должна позволять инженерам, конструкторам, технологам и производственным пользователям управлять изделием как набором связанных инженерных объектов, где каждый объект хранится в виде файлов на диске:

```text
Markdown + YAML frontmatter + JSON metadata + attachments + history log
```

Основной принцип: **файлы являются источником истины**, Git фиксирует историю изменений, а бизнес-процессы управляются конечными автоматами.

---

# 2. Цели проекта

## 2.1. Бизнес-цели

1. Создать легкую локальную альтернативу тяжелым PLM/PDM-системам.
2. Упростить управление инженерными объектами, BOM, чертежами и релизами.
3. Обеспечить прозрачную историю изменений через Git.
4. Снизить риск выпуска неполной или неактуальной документации.
5. Дать инженеру понятный UX без необходимости работать напрямую с Git, YAML и файловой структурой.
6. Подготовить основу для интеграции с ERP, MES, WMS и системами документооборота.
7. Обеспечить работу со сложными инженерными процессами через абстракции: Object, Relation, Artifact, Process, Validation, Event.

## 2.2. Технические цели

1. Реализовать надежное файловое хранилище с атомарными операциями.
2. Реализовать универсальное доменное ядро.
3. Реализовать конфигурируемый FSM engine.
4. Реализовать validation/readiness engine.
5. Реализовать индекс проекта для быстрых запросов.
6. Реализовать Git-интеграцию как внутренний механизм версионирования.
7. Реализовать desktop UI на Wails + React + TypeScript.
8. Реализовать расширяемую архитектуру модулей.
9. Реализовать шаблоны объектов, документов, процессов и релизов.
10. Реализовать основу для будущей интеграции с ERP.

---

# 3. Основные пользователи

## 3.1. Инженер / конструктор

Создает детали, сборки, чертежи, документацию, CAD-файлы, редактирует свойства и связи.

Основные действия:

- создать деталь;
- создать сборку;
- добавить чертеж;
- прикрепить STEP/DXF/PDF;
- заполнить материал, массу, размеры;
- добавить объект в BOM;
- отправить объект на проверку;
- посмотреть, что мешает релизу.

## 3.2. Технолог

Создает производственные маршруты, операции, инструкции и матрицу готовности.

Основные действия:

- создать process plan;
- добавить операции;
- создать work instruction;
- определить проверки;
- указать оснастку;
- указать safety warnings;
- проверить готовность изделия к производству.

## 3.3. Закупщик

Работает с buy list, поставщиками, стандартными изделиями и заменами.

Основные действия:

- получить список закупки из BOM;
- увидеть make/buy позиции;
- проверить поставщиков;
- предложить замену;
- проверить дубли стандартных изделий;
- экспортировать закупочный список.

## 3.4. Производственный пользователь

Работает с shop-floor player.

Основные действия:

- открыть build по QR-коду;
- пройти шаги инструкции;
- отметить проверки;
- приложить evidence;
- зафиксировать отклонение;
- завершить производственную операцию.

## 3.5. Администратор проекта

Настраивает шаблоны, naming standard, workflow, роли и проектные правила.

Основные действия:

- создать проект;
- настроить object classes;
- настроить naming convention;
- настроить lifecycle;
- настроить validation rules;
- настроить export rules;
- восстановить проект после ошибок.

---

# 4. Границы системы

## 4.1. Входит в систему

1. Локальное хранение инженерных объектов.
2. Markdown/YAML-редактор.
3. Управление объектами: детали, сборки, документы, чертежи, инструкции, релизы.
4. Управление BOM.
5. Управление связями между объектами.
6. Управление файлами и вложениями.
7. Жизненные циклы через FSM.
8. Валидация объектов и релизов.
9. История изменений.
10. Git status/diff/commit/tag.
11. Release package.
12. Manufacturing readiness matrix.
13. Work instructions.
14. Shop-floor execution.
15. Procurement buy list.
16. Standard parts library.
17. Change impact analysis.
18. Export для ERP/MES/документооборота.

## 4.2. Не входит в MVP

1. Полноценная многопользовательская серверная PLM.
2. Полноценная ERP.
3. Полноценная MES.
4. Полноценный WMS.
5. CAD-редактор.
6. Облачная синхронизация.
7. Сложная ролевая модель уровня enterprise.
8. Электронная подпись по юридическим стандартам.
9. Полный документооборот с внешними согласующими.
10. Полноценная работа с бинарными CAD-форматами внутри приложения.

---

# 5. Ключевые принципы

## 5.1. Files are source of truth

Все данные должны храниться в читаемых файлах:

```text
project.yaml
objects/*/item.json
objects/*/*.md
objects/*/history.log
objects/*/files/*
objects/*/generated/*
config/*
templates/*
exports/*
```

Приложение не должно иметь скрытого состояния, которое невозможно восстановить из файлов проекта.

## 5.2. Git-native

Git используется для:

- фиксации изменений;
- просмотра diff;
- создания checkpoint;
- создания release tag;
- анализа истории;
- отката изменений;
- сравнения релизов.

Пользовательский интерфейс не должен требовать от инженера знания Git-команд.

## 5.3. FSM-driven

Все важные процессы должны управляться state machine:

- жизненный цикл объекта;
- жизненный цикл документа;
- жизненный цикл релиза;
- жизненный цикл change request;
- жизненный цикл производственного build;
- жизненный цикл закупочной замены.

## 5.4. Markdown-everywhere

Техническая документация должна быть редактируемой, читаемой и версионируемой.

Структура объекта:

```markdown
---
id: demo-prt-0001-v1.0
class: prt
title: Bracket Motor
revision: 1.0
state: draft
material: Aluminum 6061
---

# Bracket Motor

## Description

## Manufacturing Notes

## Inspection Notes
```

## 5.5. Абстракции вместо хаоса модулей

В основе системы должны быть универсальные абстракции:

```text
Object
Relation
Artifact
Process
Validation
Event
Projection
Command
Query
```

BOM, релизы, закупки, инструкции и shop-floor должны быть модулями поверх этого ядра.

---

# 6. Доменная модель

## 6.1. Object

Object — базовая инженерная сущность.

Примеры:

- Part;
- Assembly;
- Drawing;
- Document;
- WorkInstruction;
- ProcessPlan;
- ReleasePackage;
- BuildRecord;
- StandardPart;
- ChangeRequest;
- ProcurementItem.

### Поля ObjectRecord

```go
type ObjectRecord struct {
    ID          string                 `json:"id"`
    Class       string                 `json:"class"`
    Title       string                 `json:"title"`
    Revision    Revision               `json:"revision"`
    State       string                 `json:"state"`
    Metadata    map[string]any         `json:"metadata"`
    Relations   []Relation             `json:"relations"`
    Artifacts   []ArtifactRef          `json:"artifacts"`
    Checksums   map[string]string      `json:"checksums"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}
```

## 6.2. Relation

Relation описывает связь между объектами.

Примеры:

```text
assembly contains part
part has drawing
part has cad model
part uses standard part
process plan uses work instruction
release includes object
change request affects object
standard part substitutes part
```

### Поля Relation

```go
type Relation struct {
    FromID      string      `json:"from_id"`
    ToID        string      `json:"to_id"`
    Type        string      `json:"type"`
    Quantity    *float64    `json:"quantity,omitempty"`
    Unit        string      `json:"unit,omitempty"`
    Effectivity Effectivity `json:"effectivity,omitempty"`
    Metadata    map[string]any `json:"metadata,omitempty"`
}
```

## 6.3. Artifact

Artifact — файл, связанный с объектом.

Примеры:

- Markdown document;
- STEP;
- DXF;
- PDF;
- PNG/JPG;
- certificate;
- inspection report;
- generated BOM;
- release manifest;
- build evidence.

### Поля Artifact

```go
type Artifact struct {
    ID        string    `json:"id"`
    Kind      string    `json:"kind"`
    Role      string    `json:"role"`
    Path      string    `json:"path"`
    Checksum  string    `json:"checksum"`
    Generated bool      `json:"generated"`
    CreatedAt time.Time `json:"created_at"`
}
```

## 6.4. Process

Process — конфигурируемый жизненный цикл.

```go
type ProcessDefinition struct {
    Name        string       `json:"name"`
    States      []State      `json:"states"`
    Transitions []Transition `json:"transitions"`
}
```

## 6.5. Validation

Validation — набор правил, проверяющих объект, связь, релиз или проект.

Примеры проверок:

- valid_name;
- required_metadata;
- no_broken_relations;
- no_bom_cycles;
- all_children_released;
- checksums_actual;
- drawing_attached;
- no_release_blockers;
- procurement_complete.

## 6.6. Event

Event — атомарное событие истории.

```json
{
  "event_id": "evt_000001",
  "time": "2026-06-06T10:00:00Z",
  "actor": "local-user",
  "object_id": "demo-prt-0001-v1.0",
  "type": "object.created",
  "payload": {
    "class": "prt",
    "title": "Bracket Motor"
  }
}
```

---

# 7. Файловая структура проекта

## 7.1. Структура workspace

```text
MyProject/
├── project.yaml
├── objects/
│   └── demo-prt-0001-v1.0/
│       ├── item.json
│       ├── demo-prt-0001-v1.0.md
│       ├── history.log
│       ├── files/
│       └── generated/
├── config/
│   ├── naming.yaml
│   ├── classes.yaml
│   ├── lifecycle.yaml
│   ├── validation.yaml
│   ├── file-routing.yaml
│   └── exports.yaml
├── templates/
│   ├── object/
│   ├── document/
│   ├── release/
│   ├── manufacturing/
│   └── procurement/
├── exports/
├── .plm/
│   ├── index.json
│   ├── transactions/
│   ├── cache/
│   └── recovery/
└── .git/
```

## 7.2. project.yaml

```yaml
project:
  code: demo
  title: Demo Engineering Project
  version: 1
  created_at: 2026-06-06T10:00:00Z

naming:
  standard: v10
  pattern: "[project]-[class]-[identifier]-v[major].[minor]"
  project_code: demo

storage:
  source_of_truth: files
  index: .plm/index.json

git:
  enabled: true
  auto_commit: false
  release_tags: true

features:
  bom: true
  manufacturing: true
  procurement: true
  shopfloor: true
  standardparts: true
  release: true
```

---

# 8. Naming Standard

## 8.1. Формат имени

```text
[project]-[class]-[identifier]-v[major].[minor][-status].[ext]
```

Примеры:

```text
demo-prt-0001-v1.0.md
demo-asm-0100-v2.0.md
a320-asm-0100-v2.0.step
std-fst-iso4017-m8x30-v1.0.md
```

## 8.2. Классы объектов

```yaml
classes:
  prt:
    title: Part
    prefix: prt
    required_fields:
      - title
      - material
      - unit
  asm:
    title: Assembly
    prefix: asm
    required_fields:
      - title
  drw:
    title: Drawing
    prefix: drw
    required_fields:
      - title
      - related_object
  doc:
    title: Document
    prefix: doc
  wi:
    title: Work Instruction
    prefix: wi
  pp:
    title: Process Plan
    prefix: pp
  rel:
    title: Release Package
    prefix: rel
  cr:
    title: Change Request
    prefix: cr
  std:
    title: Standard Part
    prefix: std
```

---

# 9. Жизненные циклы

## 9.1. Object Lifecycle

```text
draft → in_review → approved → released → obsolete → archived
```

Дополнительное состояние:

```text
blocked
```

## 9.2. Переходы

```yaml
processes:
  object_lifecycle:
    initial: draft
    states:
      - draft
      - in_review
      - approved
      - released
      - blocked
      - obsolete
      - archived

    transitions:
      submit_review:
        from: draft
        to: in_review
        guards:
          - valid_name
          - required_metadata
          - no_broken_relations

      reject:
        from: in_review
        to: draft
        requires_reason: true

      approve:
        from: in_review
        to: approved
        guards:
          - no_blocking_issues
          - checksums_actual

      release:
        from: approved
        to: released
        guards:
          - no_release_blockers
          - checksums_actual
          - children_released
          - bom_valid

      revise:
        from: released
        to: draft
        effects:
          - bump_minor_revision
          - create_change_event

      block:
        from: "*"
        to: blocked
        requires_reason: true

      unblock:
        from: blocked
        to: draft
        requires_reason: true

      obsolete:
        from: released
        to: obsolete
        requires_reason: true

      archive:
        from: obsolete
        to: archived
        requires_reason: true
```

## 9.3. Release Lifecycle

```yaml
processes:
  release_lifecycle:
    initial: draft
    states:
      - draft
      - validating
      - ready
      - released
      - failed
      - cancelled

    transitions:
      validate:
        from: draft
        to: validating

      mark_ready:
        from: validating
        to: ready
        guards:
          - no_release_blockers
          - manifest_generated

      publish:
        from: ready
        to: released
        effects:
          - create_git_tag
          - export_release_package
          - append_release_event

      fail:
        from: validating
        to: failed

      cancel:
        from: "*"
        to: cancelled
```

## 9.4. Change Request Lifecycle

```yaml
processes:
  change_request:
    initial: proposed
    states:
      - proposed
      - impact_analysis
      - approved
      - implemented
      - verified
      - closed
      - rejected

    transitions:
      analyze:
        from: proposed
        to: impact_analysis

      approve:
        from: impact_analysis
        to: approved
        guards:
          - impact_report_exists

      implement:
        from: approved
        to: implemented

      verify:
        from: implemented
        to: verified
        guards:
          - affected_objects_valid

      close:
        from: verified
        to: closed

      reject:
        from:
          - proposed
          - impact_analysis
        to: rejected
        requires_reason: true
```

---

# 10. UX-требования

## 10.1. Общие требования

1. Пользователь всегда должен видеть текущий проект.
2. Пользователь всегда должен видеть текущий выбранный объект.
3. Пользователь всегда должен видеть статус объекта.
4. Пользователь всегда должен понимать, что мешает следующему действию.
5. Все destructive-действия должны иметь preview.
6. Ошибки должны формулироваться как действия.
7. Git должен быть скрыт за понятными инженерными действиями.
8. UI должен поддерживать keyboard-first работу.
9. Должна быть command palette.
10. Должна быть быстрая навигация по объектам, связям и статусам.

## 10.2. Главный экран

Основная компоновка:

```text
┌────────────────────────────────────────────────────────────┐
│ Top Bar: Project | Search | Command Palette | Git Status   │
├───────────────┬────────────────────────┬───────────────────┤
│ Project Tree  │ Editor / View          │ Properties        │
│               │                        │ Lifecycle         │
│ Filters       │ Markdown + Preview     │ Relations         │
│ Views         │ BOM / Matrix / Diff    │ Attachments       │
└───────────────┴────────────────────────┴───────────────────┘
```

## 10.3. Command Palette

Горячая клавиша:

```text
Ctrl+K
```

Команды:

```text
Create Part
Create Assembly
Create Drawing
Create Work Instruction
Create Release Package
Add Child to BOM
Show Where Used
Show Change Impact
Submit for Review
Approve
Release
Explain Blockers
Generate Buy List
Export Release Package
Rebuild Index
Validate Project
Open Git Diff
```

## 10.4. Why Can't I Assistant

Ассистент должен объяснять невозможность действия.

Пример:

```text
Нельзя выпустить объект demo-asm-0100-v1.0.

Причины:
1. Дочерний объект demo-prt-0004-v1.0 находится в состоянии draft.
2. Отсутствует утвержденный чертеж.
3. Checksum файла bracket.step устарел.
4. BOM содержит позицию без количества.

Действия:
[Открыть draft-объекты]
[Добавить чертеж]
[Пересчитать checksum]
[Исправить BOM]
```

## 10.5. Project Views

Система должна поддерживать несколько представлений:

```text
Tree View
Table View
BOM View
Lifecycle View
Graph View
Release View
Manufacturing View
Procurement View
History View
Git View
```

## 10.6. Bulk Operations

Пользователь должен иметь возможность:

```text
выбрать несколько объектов;
запустить validate;
отправить на review;
создать release package;
пересчитать checksums;
экспортировать BOM;
изменить статус;
добавить общий тег;
```

---

# 11. Функциональные требования

# 11.1. Управление проектами

## FR-PROJ-001. Создание проекта

Система должна позволять создать новый проект по шаблону.

Входные данные:

```text
project_code
project_title
project_path
template
naming_standard
enabled_modules
```

Результат:

```text
создана директория проекта;
создан project.yaml;
созданы config-файлы;
созданы templates;
инициализирован Git;
создан initial commit;
создан индекс проекта.
```

## FR-PROJ-002. Открытие проекта

Система должна открывать существующий проект, читать `project.yaml`, сканировать объекты и строить индекс.

## FR-PROJ-003. Валидация проекта

Система должна проверять:

```text
наличие project.yaml;
валидность config-файлов;
валидность object records;
битые связи;
дубли ID;
устаревшие checksums;
битый history.log;
несоответствие Markdown/YAML/item.json;
```

## FR-PROJ-004. Recovery Mode

При ошибках система должна предложить:

```text
Repair automatically
Open diagnostics
Export recovery report
Ignore and open read-only
```

---

# 11.2. Управление объектами

## FR-OBJ-001. Создание объекта

Система должна создавать объект через wizard.

Шаги wizard:

```text
1. Identity
2. Properties
3. Documents & Attachments
```

## FR-OBJ-002. Редактирование объекта

Система должна позволять редактировать:

```text
title
metadata
Markdown body
relations
artifacts
lifecycle state
```

## FR-OBJ-003. Удаление объекта

Удаление должно быть безопасным:

```text
проверить where-used;
показать impact;
запросить подтверждение;
переместить в archived/trash;
сохранить event log.
```

## FR-OBJ-004. Дублирование объекта

Система должна уметь создать новый объект на основе существующего.

## FR-OBJ-005. Revision bump

Система должна поддерживать:

```text
bump patch/minor/major;
создание новой ревизии;
связь между ревизиями;
историю изменений.
```

---

# 11.3. Редактор Markdown/YAML

## FR-EDIT-001. Live Preview

Редактор должен показывать Markdown preview.

## FR-EDIT-002. YAML frontmatter validation

Ошибки YAML должны показываться inline.

## FR-EDIT-003. Metadata sync

Изменения YAML должны синхронизироваться с `item.json`.

## FR-EDIT-004. Autosave

Система должна поддерживать autosave с debounce.

## FR-EDIT-005. Diff view

Пользователь должен видеть изменения относительно:

```text
последнего сохранения;
последнего commit;
последнего release;
выбранной ревизии.
```

---

# 11.4. Relations и BOM

## FR-REL-001. Добавление связи

Система должна позволять добавить связь между объектами.

Типы связей:

```text
contains
has_drawing
has_cad
uses_instruction
uses_process_plan
substitutes
affects
includes
derived_from
```

## FR-BOM-001. BOM View

BOM должен показывать:

```text
позиция
объект
название
класс
ревизия
статус
количество
единица измерения
make/buy
замены
проблемы
```

## FR-BOM-002. Where-used

Система должна показывать, где используется объект.

## FR-BOM-003. Cycle detection

Система должна запрещать циклические зависимости в BOM.

## FR-BOM-004. BOM export

Система должна экспортировать BOM:

```text
CSV
JSON
YAML
ERP import package
```

---

# 11.5. Lifecycle / FSM

## FR-FSM-001. Выполнение перехода

Система должна позволять выполнить transition, если guards пройдены.

## FR-FSM-002. Guard validation

Перед переходом система должна показать:

```text
passed guards
failed guards
warnings
recommended actions
```

## FR-FSM-003. Transition history

Каждый переход должен записываться в history.log.

## FR-FSM-004. Configurable FSM

FSM должен задаваться через YAML.

---

# 11.6. Validation / Readiness

## FR-VAL-001. Валидация объекта

Проверки:

```text
valid name
required metadata
valid relations
valid artifacts
checksums actual
lifecycle consistency
```

## FR-VAL-002. Release readiness

Проверки:

```text
all objects approved/released
BOM valid
no blockers
all files checksummed
required drawings attached
manufacturing data complete
procurement data complete
```

## FR-VAL-003. Explain blockers

Система должна объяснять каждый blocker.

---

# 11.7. Release Package

## FR-REL-001. Создание release package

Release package должен включать:

```text
manifest.json
BOM
object records
Markdown documents
PDF exports
attachments
checksums
history summary
Git tag
```

## FR-REL-002. Manifest

Пример:

```json
{
  "release_id": "demo-rel-0001-v1.0",
  "project": "demo",
  "created_at": "2026-06-06T10:00:00Z",
  "objects": [
    {
      "id": "demo-asm-0100-v1.0",
      "state": "released",
      "checksum": "sha256:..."
    }
  ],
  "files": [],
  "bom": [],
  "git": {
    "commit": "...",
    "tag": "release/demo-rel-0001-v1.0"
  }
}
```

## FR-REL-003. Почему нельзя выпустить

Система должна иметь отдельный экран:

```text
Why Can't I Release
```

---

# 11.8. Manufacturing Matrix

## FR-MFG-001. Матрица готовности

Матрица должна показывать:

```text
объект
операции
чертежи
CAD
BOM
инструкции
оснастка
проверки
статус
blockers
```

## FR-MFG-002. Process Plan

Process Plan должен содержать:

```text
operations
sequence
work center
tools
materials
inspection points
estimated time
work instructions
```

## FR-MFG-003. Work Instructions

Work Instruction должен содержать:

```text
steps
checks
evidence requirements
safety warnings
tools
images
acceptance criteria
```

---

# 11.9. Shop Floor

## FR-SF-001. Build Record

Build Record должен фиксировать конкретное выполнение.

```json
{
  "build_id": "build-0001",
  "object_id": "demo-asm-0100-v1.0",
  "operator": "local-user",
  "started_at": "...",
  "completed_at": null,
  "state": "in_progress",
  "steps": []
}
```

## FR-SF-002. QR Token

Система должна генерировать QR token для build или инструкции.

## FR-SF-003. Evidence Capture

Система должна позволять приложить:

```text
фото
файл
комментарий
результат измерения
подтверждение проверки
```

---

# 11.10. Procurement

## FR-PROC-001. Buy List

Buy List строится из BOM.

Поля:

```text
part_id
title
quantity
unit
make_buy
supplier
manufacturer_part_number
lead_time
substitutes
status
```

## FR-PROC-002. Substitution workflow

Процесс замены:

```text
proposed → engineering_review → approved → applied → rejected
```

## FR-PROC-003. Export to ERP

Экспорт:

```text
CSV
JSON
ERP import package
```

---

# 11.11. Standard Parts

## FR-STD-001. Библиотека стандартных изделий

Система должна поддерживать стандартные детали.

## FR-STD-002. Duplicate detection

Система должна определять возможные дубли по:

```text
name
standard
dimensions
material
manufacturer part number
normalized key
```

---

# 11.12. Change Impact Radar

## FR-CHG-001. Impact analysis

Для объекта система должна показывать:

```text
где используется;
какие сборки затронуты;
какие релизы затронуты;
какие инструкции затронуты;
какие закупочные позиции затронуты;
какие build records затронуты;
нужен ли change request.
```

## FR-CHG-002. Impact report

Отчет должен сохраняться как artifact.

---

# 11.13. Git Integration

## FR-GIT-001. Git status

Система должна показывать:

```text
modified
added
deleted
untracked
staged
committed
```

## FR-GIT-002. Checkpoint

Пользователь должен создавать checkpoint без знания Git.

## FR-GIT-003. Release tag

При выпуске release package должен создаваться tag.

## FR-GIT-004. Diff

Пользователь должен видеть diff на уровне:

```text
файлов;
объектов;
metadata;
BOM;
relations.
```

---

# 12. Нефункциональные требования

## 12.1. Надежность

1. Все записи на диск должны быть атомарными.
2. Запрещено прямое перезаписывание критичных файлов без backup/temp.
3. Все изменения должны фиксироваться в event log.
4. Система должна уметь восстановиться после прерванной операции.
5. Система должна проверять целостность проекта при запуске.
6. Ошибки должны быть диагностируемыми.

## 12.2. Производительность

Целевые показатели MVP:

```text
проект до 10 000 объектов;
открытие проекта до 5 секунд при наличии индекса;
поиск до 200 мс;
открытие объекта до 100 мс;
построение BOM до 500 мс;
where-used до 500 мс;
валидация одного объекта до 200 мс;
валидация проекта до 30 секунд для 10 000 объектов.
```

## 12.3. Безопасность

1. Нельзя выполнять произвольные скрипты из проекта без разрешения.
2. Вложения должны открываться безопасно.
3. HTTP JSON-RPC должен слушать только localhost.
4. API должен иметь session token.
5. Секреты не должны попадать в Git.
6. В `.gitignore` должны быть env, hash, temp, cache.

## 12.4. Кроссплатформенность

Поддержка:

```text
Windows 10+
macOS 13+
Linux with WebKitGTK
```

## 12.5. Расширяемость

Новые модули должны добавляться поверх core API:

```text
module registers:
commands
queries
validators
views
templates
processes
exports
```

---

# 13. Архитектура

## 13.1. Целевая структура репозитория

```text
go-plm/
├── go.mod
├── go.sum
├── README.md
├── run.ps1
├── cmd/
│   └── plm/
│       ├── main.go
│       ├── desktop.go
│       ├── assets.go
│       └── services/
├── internal/
│   ├── core/
│   │   ├── object/
│   │   ├── relation/
│   │   ├── artifact/
│   │   ├── revision/
│   │   └── identity/
│   ├── store/
│   │   ├── filesystem/
│   │   ├── transaction/
│   │   ├── index/
│   │   └── recovery/
│   ├── process/
│   │   ├── fsm/
│   │   ├── guards/
│   │   └── effects/
│   ├── validation/
│   ├── parser/
│   ├── naming/
│   ├── history/
│   ├── gitops/
│   ├── modules/
│   │   ├── bom/
│   │   ├── release/
│   │   ├── manufacturing/
│   │   ├── procurement/
│   │   ├── shopfloor/
│   │   ├── standardparts/
│   │   └── change/
│   └── api/
├── frontend/
│   ├── package.json
│   ├── vite.config.ts
│   └── src/
│       ├── app/
│       ├── features/
│       ├── editor/
│       ├── components/
│       └── bindings/
├── examples/
│   └── demo-project/
├── templates/
├── docs/
└── .github/
    └── workflows/
        └── ci.yml
```

## 13.2. Backend layers

```text
API layer
Application commands/queries
Domain services
Core domain
Store
Filesystem
Git
```

## 13.3. Frontend layers

```text
App shell
Feature modules
Reusable components
API client
State management
Editor extensions
```

---

# 14. API

## 14.1. Стиль API

IPC между frontend и backend:

```text
HTTP JSON-RPC over localhost
```

## 14.2. Commands

```text
CreateProject
OpenProject
ValidateProject
CreateObject
UpdateObject
DeleteObject
AttachFile
AddRelation
RemoveRelation
RunTransition
CreateReleasePackage
GenerateBOM
GenerateBuyList
CreateBuildRecord
CommitCheckpoint
CreateReleaseTag
RebuildIndex
```

## 14.3. Queries

```text
GetProject
ListObjects
GetObject
GetObjectHistory
GetBOM
GetWhereUsed
GetChangeImpact
GetReadiness
GetReleaseBlockers
GetGitStatus
GetDiff
SearchObjects
ListTemplates
ListProcesses
```

## 14.4. Пример JSON-RPC

Request:

```json
{
  "jsonrpc": "2.0",
  "method": "object.create",
  "params": {
    "class": "prt",
    "title": "Bracket Motor",
    "metadata": {
      "material": "Aluminum 6061"
    }
  },
  "id": "req-001"
}
```

Response:

```json
{
  "jsonrpc": "2.0",
  "result": {
    "id": "demo-prt-0001-v1.0",
    "path": "objects/demo-prt-0001-v1.0"
  },
  "id": "req-001"
}
```

---

# 15. Шаблоны

# 15.1. Шаблон детали

```markdown
---
id: "{{id}}"
class: "prt"
title: "{{title}}"
revision: "1.0"
state: "draft"
material: ""
unit: "pcs"
mass: null
make_buy: "make"
---

# {{title}}

## Назначение

## Материал

## Основные размеры

## Производственные примечания

## Контроль качества

## Связанные документы

## История изменений
```

# 15.2. Шаблон сборки

```markdown
---
id: "{{id}}"
class: "asm"
title: "{{title}}"
revision: "1.0"
state: "draft"
unit: "pcs"
---

# {{title}}

## Назначение сборки

## Состав

BOM управляется через relations.

## Последовательность сборки

## Критичные требования

## Контроль

## Связанные документы
```

# 15.3. Шаблон чертежа

```markdown
---
id: "{{id}}"
class: "drw"
title: "{{title}}"
revision: "1.0"
state: "draft"
related_object: "{{related_object}}"
format: "A3"
---

# {{title}}

## Связанный объект

## Версия чертежа

## Примечания

## Файлы

- PDF:
- DXF:
```

# 15.4. Шаблон work instruction

```markdown
---
id: "{{id}}"
class: "wi"
title: "{{title}}"
revision: "1.0"
state: "draft"
related_object: "{{related_object}}"
---

# {{title}}

## Назначение инструкции

## Требуемые инструменты

## Материалы

## Требования безопасности

## Шаги

### Шаг 1

**Действие:**  
**Проверка:**  
**Evidence required:** yes/no  
**Критерий приемки:**  

### Шаг 2

**Действие:**  
**Проверка:**  
**Evidence required:** yes/no  
**Критерий приемки:**  

## Завершение операции
```

# 15.5. Шаблон process plan

```markdown
---
id: "{{id}}"
class: "pp"
title: "{{title}}"
revision: "1.0"
state: "draft"
related_object: "{{related_object}}"
---

# {{title}}

## Маршрут производства

| № | Операция | Участок | Время | Инструкция | Контроль |
|---|----------|---------|-------|------------|----------|
| 10 | | | | | |
| 20 | | | | | |

## Оснастка

## Материалы

## Контрольные точки

## Риски
```

# 15.6. Шаблон change request

```markdown
---
id: "{{id}}"
class: "cr"
title: "{{title}}"
revision: "1.0"
state: "proposed"
affected_objects: []
---

# {{title}}

## Причина изменения

## Описание изменения

## Затронутые объекты

## Анализ влияния

## Риски

## Решение

## План внедрения

## Проверка после внедрения
```

# 15.7. Шаблон release package

```markdown
---
id: "{{id}}"
class: "rel"
title: "{{title}}"
revision: "1.0"
state: "draft"
release_scope: []
---

# {{title}}

## Цель релиза

## Состав релиза

## Проверки готовности

## Blockers

## Manifest

## Git Tag

## Export Package

## Approval
```

# 15.8. Шаблон standard part

```markdown
---
id: "{{id}}"
class: "std"
title: "{{title}}"
revision: "1.0"
state: "released"
standard: ""
manufacturer: ""
manufacturer_part_number: ""
material: ""
dimensions: ""
---

# {{title}}

## Стандарт

## Основные параметры

## Производитель

## Замены

## Примечания
```

# 15.9. Шаблон build record

```markdown
---
id: "{{id}}"
class: "build"
title: "{{title}}"
state: "planned"
object_id: "{{object_id}}"
operator: ""
started_at: null
completed_at: null
---

# {{title}}

## Объект сборки

## Оператор

## Шаги выполнения

## Evidence

## Отклонения

## Результат
```

---

# 16. Шаблоны конфигураций

## 16.1. config/lifecycle.yaml

```yaml
version: 1

processes:
  object_lifecycle:
    initial: draft
    states:
      - draft
      - in_review
      - approved
      - released
      - blocked
      - obsolete
      - archived

    transitions:
      submit_review:
        from: draft
        to: in_review
        guards:
          - valid_name
          - required_metadata
          - no_broken_relations

      approve:
        from: in_review
        to: approved
        guards:
          - no_blocking_issues
          - checksums_actual

      release:
        from: approved
        to: released
        guards:
          - no_release_blockers
          - children_released
          - bom_valid

      revise:
        from: released
        to: draft
        effects:
          - bump_minor_revision

      block:
        from: "*"
        to: blocked
        requires_reason: true

      unblock:
        from: blocked
        to: draft
```

## 16.2. config/validation.yaml

```yaml
version: 1

validators:
  valid_name:
    type: naming

  required_metadata:
    type: required_fields

  no_broken_relations:
    type: relation_integrity

  no_bom_cycles:
    type: graph_cycle_detection

  checksums_actual:
    type: checksum

  children_released:
    type: relation_state
    relation: contains
    allowed_states:
      - released

  bom_valid:
    type: bom
    checks:
      - quantity_positive
      - unit_present
      - no_cycles
      - target_exists

  no_release_blockers:
    type: blockers
```

## 16.3. config/file-routing.yaml

```yaml
version: 1

routes:
  markdown:
    extensions:
      - .md
    target: root

  cad:
    extensions:
      - .step
      - .stp
      - .iges
      - .igs
      - .sldprt
      - .sldasm
    target: files/cad

  drawings:
    extensions:
      - .pdf
      - .dxf
      - .dwg
    target: files/drawings

  images:
    extensions:
      - .png
      - .jpg
      - .jpeg
      - .webp
    target: files/images

  generated:
    extensions:
      - .html
      - .typ
      - .generated.pdf
    target: generated
```

## 16.4. config/exports.yaml

```yaml
version: 1

exports:
  release_package:
    include:
      - manifest
      - object_records
      - markdown
      - attachments
      - generated
      - bom
      - checksums
      - history_summary

  erp_bom:
    format: csv
    fields:
      - item
      - part_id
      - title
      - revision
      - quantity
      - unit
      - make_buy
      - supplier
      - manufacturer_part_number
```

---

# 17. Этапы разработки

# Этап 0. Подготовка репозитория

## Цель

Привести репозиторий в соответствие с заявленной архитектурой.

## Работы

1. Добавить `go.mod`.
2. Добавить `cmd/plm/main.go`.
3. Добавить базовую структуру `internal`.
4. Добавить `frontend/package.json`.
5. Добавить минимальный React/Vite frontend.
6. Добавить `examples/demo-project`.
7. Добавить GitHub Actions CI.
8. Обновить README с реальными командами запуска.

## Результат

Репозиторий собирается, тесты запускаются, приложение имеет пустой shell.

## Критерии приемки

```text
go test ./... проходит;
npm install проходит;
npm run build проходит;
README соответствует структуре проекта;
CI запускается на push/pull_request.
```

---

# Этап 1. Core Domain

## Цель

Создать доменное ядро.

## Работы

1. ObjectRecord.
2. Relation.
3. Artifact.
4. Revision.
5. ProjectConfig.
6. Diagnostics.
7. Event.
8. Базовые unit tests.

## Результат

Можно создать доменный объект в памяти и сериализовать его в JSON/YAML.

## Критерии приемки

```text
есть типы core domain;
есть unit tests;
есть JSON serialization;
есть validation базовых полей.
```

---

# Этап 2. Filesystem Store

## Цель

Реализовать надежное файловое хранение.

## Работы

1. OpenProject.
2. CreateProject.
3. CreateObject.
4. ReadObject.
5. UpdateObject.
6. ListObjects.
7. Atomic write.
8. History append.
9. Recovery transaction folder.

## Результат

Система может создать проект и объект на диске.

## Критерии приемки

```text
объект создается в objects/<id>;
создается item.json;
создается markdown file;
создается history.log;
запись атомарная;
после crash можно восстановиться.
```

---

# Этап 3. Parser / Naming / Validation

## Цель

Реализовать чтение Markdown/YAML и базовую защиту данных.

## Работы

1. YAML frontmatter parser.
2. Markdown body parser.
3. Naming standard parser.
4. Required metadata validation.
5. Relation integrity validation.
6. Checksum validation.

## Результат

Система умеет читать и валидировать инженерные объекты.

## Критерии приемки

```text
валидный markdown читается;
невалидный YAML дает diagnostics;
невалидное имя отклоняется;
битые связи обнаруживаются;
checksums рассчитываются.
```

---

# Этап 4. FSM Engine

## Цель

Реализовать конфигурируемые жизненные циклы.

## Работы

1. ProcessDefinition.
2. State.
3. Transition.
4. Guard registry.
5. Effect registry.
6. RunTransition.
7. Transition history.

## Результат

Работает жизненный цикл объекта.

## Критерии приемки

```text
draft → in_review работает;
invalid transition запрещается;
failed guard возвращает объяснение;
transition пишется в history.log;
FSM загружается из YAML.
```

---

# Этап 5. Desktop MVP

## Цель

Создать первый полезный пользовательский интерфейс.

## Работы

1. Wails shell.
2. React app shell.
3. Open project screen.
4. Project tree.
5. Object editor.
6. Properties panel.
7. Lifecycle buttons.
8. Validation panel.
9. API bridge.

## Результат

Пользователь может открыть проект, создать объект, отредактировать его и провести через lifecycle.

## Критерии приемки

```text
приложение запускается;
проект открывается;
объект создается;
markdown редактируется;
status виден;
transition запускается;
ошибки показываются в UI.
```

---

# Этап 6. Relations / BOM

## Цель

Реализовать структуру изделия.

## Работы

1. AddRelation.
2. RemoveRelation.
3. BOM View.
4. Quantity/unit.
5. Where-used.
6. Cycle detection.
7. BOM export.

## Результат

Можно управлять составом сборки.

## Критерии приемки

```text
деталь добавляется в сборку;
BOM отображается таблицей;
where-used работает;
циклы запрещаются;
BOM экспортируется в CSV/JSON.
```

---

# Этап 7. Release Readiness

## Цель

Сделать надежный выпуск релиза.

## Работы

1. Release package object.
2. Release readiness validation.
3. Why Can't I Release.
4. Manifest generation.
5. Export folder.
6. Git release tag.

## Результат

Пользователь может выпустить корректный release package.

## Критерии приемки

```text
release blockers объясняются;
manifest создается;
release package экспортируется;
Git tag создается;
некорректный релиз запрещается.
```

---

# Этап 8. Git UX

## Цель

Скрыть сложность Git за инженерным UX.

## Работы

1. Git status.
2. Checkpoint.
3. Diff viewer.
4. Release tag.
5. Compare with release.
6. Restore from checkpoint.

## Результат

Пользователь работает с версиями без знания Git-команд.

## Критерии приемки

```text
измененные объекты видны;
checkpoint создается;
diff отображается;
можно сравнить с релизом;
можно восстановить checkpoint.
```

---

# Этап 9. Manufacturing

## Цель

Добавить производственную готовность.

## Работы

1. Manufacturing matrix.
2. Process plan.
3. Work instruction.
4. Operation readiness.
5. Safety warnings.
6. Inspection points.

## Результат

Система показывает готовность изделия к производству.

## Критерии приемки

```text
матрица готовности строится;
process plan связан с объектом;
work instruction связан с операцией;
blockers отображаются;
готовность считается автоматически.
```

---

# Этап 10. Procurement

## Цель

Добавить закупочный контур.

## Работы

1. Buy list from BOM.
2. Make/buy field.
3. Supplier fields.
4. Substitution workflow.
5. Standard parts.
6. Duplicate detection.
7. ERP export.

## Результат

Можно получить закупочный список из инженерного BOM.

## Критерии приемки

```text
buy list генерируется;
make/buy учитывается;
замены проходят workflow;
дубли standard parts находятся;
ERP export формируется.
```

---

# Этап 11. Shop Floor

## Цель

Добавить выполнение инструкций на производстве.

## Работы

1. Build record.
2. QR token.
3. Shop-floor player.
4. Step execution.
5. Evidence capture.
6. Deviations.
7. Completion report.

## Результат

Пользователь может пройти производственную инструкцию пошагово.

## Критерии приемки

```text
build record создается;
QR token открывает build;
шаги выполняются последовательно;
evidence сохраняется;
отклонения фиксируются;
completion report создается.
```

---

# Этап 12. Change Impact

## Цель

Добавить анализ последствий изменений.

## Работы

1. Where-used graph.
2. Impact report.
3. Affected releases.
4. Affected instructions.
5. Affected procurement.
6. Change request workflow.

## Результат

Перед изменением объекта пользователь видит последствия.

## Критерии приемки

```text
impact graph строится;
затронутые объекты видны;
отчет сохраняется;
change request создается;
изменение released-объекта требует CR.
```

---

# Этап 13. Hardening / Reliability

## Цель

Повысить надежность перед реальным использованием.

## Работы

1. Recovery mode.
2. Project diagnostics.
3. Stress tests.
4. Golden tests.
5. Large project fixtures.
6. File corruption tests.
7. UI smoke tests.
8. Backup/export/import.

## Результат

Система устойчива к ошибкам данных и прерываниям операций.

## Критерии приемки

```text
проект с 10 000 объектов открывается;
битые объекты диагностируются;
индекс пересобирается;
прерванная транзакция восстанавливается;
тесты проходят в CI.
```

---

# 18. План тестирования

## 18.1. Unit tests

```text
naming parser
frontmatter parser
FSM transitions
guards
validators
relations
BOM cycle detection
checksums
filesystem transactions
```

## 18.2. Integration tests

```text
create project
create object
edit object
attach file
run lifecycle
generate BOM
create release
export package
```

## 18.3. Golden tests

Фикстуры:

```text
simple part
simple assembly
assembly with invalid BOM
release-ready project
project with broken relation
project with stale checksum
```

## 18.4. UI tests

```text
open project
create object wizard
edit markdown
run transition
show validation error
open BOM view
create release package
```

---

# 19. Definition of Done

Фича считается готовой, если:

```text
реализована backend-логика;
реализован UI;
есть unit tests;
есть integration tests;
есть diagnostics для ошибок;
есть документация;
есть пример в demo-project;
фича работает на Windows;
ошибки отображаются понятным языком;
данные сохраняются в файлах;
изменение фиксируется в history.log.
```

---

# 20. MVP Scope

Минимальная полезная версия должна включать:

```text
Create/Open Project
Create/Edit Object
Markdown + YAML editor
Project Tree
Properties Panel
Filesystem Store
History Log
Naming Validation
Object Validation
FSM Lifecycle
Git Status
Checkpoint
Relations
Simple BOM
Where-used
Release Blockers
Release Package Manifest
```

Не включать в MVP:

```text
Shop Floor
Procurement
Advanced Manufacturing
ERP integration
PDF generation
Complex permissions
Cloud sync
```

---

# 21. Риски

## 21.1. Слишком большой scope

Риск: попытка реализовать PLM, PDM, MES, ERP и Git UI одновременно.

Митигировать:

```text
сначала core;
потом UX;
потом BOM/release;
потом manufacturing/procurement/shopfloor.
```

## 21.2. Ненадежное файловое хранение

Риск: повреждение данных при прерывании записи.

Митигировать:

```text
atomic write;
transaction log;
backup before write;
recovery mode;
golden tests.
```

## 21.3. Сложный UX

Риск: инженер не захочет работать с YAML/Git/сложными статусами.

Митигировать:

```text
properties panel;
command palette;
why-can't-I assistant;
actionable diagnostics;
Git abstraction.
```

## 21.4. Хаос модулей

Риск: BOM, release, procurement, shopfloor будут реализованы как несвязанные подсистемы.

Митигировать:

```text
единые Object/Relation/Artifact/Process/Event;
CQRS;
module registry;
validation engine.
```

---

# 22. Приоритеты реализации

## P0

```text
repo skeleton
go.mod
core domain
filesystem store
object CRUD
markdown parser
naming validation
basic FSM
desktop shell
```

## P1

```text
project tree
editor
properties panel
validation panel
history log
Git status
checkpoint
relations
BOM
```

## P2

```text
release package
release blockers
manifest
where-used
change impact
diff viewer
```

## P3

```text
manufacturing matrix
work instructions
procurement
standard parts
substitutions
shop-floor player
ERP export
```

---

# 23. Ожидаемый результат

После реализации система должна стать легкой инженерной PLM/PDM-платформой, которая:

```text
хранит инженерные данные прозрачно в файлах;
позволяет управлять деталями, сборками, документами и BOM;
проводит объекты через жизненный цикл;
объясняет, почему нельзя выпустить объект;
фиксирует историю через event log и Git;
поддерживает release package;
готовит основу для производства, закупок и ERP-интеграции;
остается простой за счет правильных абстракций.
```

Ключевая формула архитектуры:

```text
Object + Relation + Artifact + Process + Validation + Event + Projection
```

Все остальные функции должны быть модулями поверх этого ядра.


---
Да, в твоей концепции **JSON почти везде лучше убрать**. Ты прав: если философия проекта — **Markdown + YAML + Git-readable files**, то `item.json`, `manifest.json`, `index.json` выглядят чужеродно.

Я бы разделил так:

```text
YAML — человекочитаемые данные, source of truth
Markdown + YAML frontmatter — инженерные документы
JSONL — только машинный append-only event log, если нужен быстрый поток событий
SQLite/bleve/badger/cache — только пересобираемый индекс, не source of truth
```

## Что заменить

Было в ТЗ:

```text
objects/<id>/item.json
objects/<id>/history.log
exports/<release>/manifest.json
.plm/index.json
```

Лучше:

```text
objects/<id>/object.yaml
objects/<id>/<id>.md
objects/<id>/history.yaml.log или history.yml
exports/<release>/manifest.yaml
.plm/index.db или .plm/index.yaml.cache
```

## Правильная структура объекта

```text
objects/
  demo-prt-0001-v1.0/
    object.yaml
    demo-prt-0001-v1.0.md
    history.yaml.log
    files/
      cad/
      drawings/
      images/
    generated/
```

`object.yaml`:

```yaml
id: demo-prt-0001-v1.0
class: prt
title: Bracket Motor
revision: "1.0"
state: draft

metadata:
  material: Aluminum 6061
  unit: pcs
  mass_g: 120
  make_buy: make

relations:
  - type: has_drawing
    to: demo-drw-0001-v1.0

artifacts:
  - id: art-step-001
    kind: cad
    role: primary_step
    path: files/cad/bracket.step
    checksum: sha256:...
```

Markdown-файл тогда остается для описания:

```markdown
---
id: demo-prt-0001-v1.0
class: prt
title: Bracket Motor
revision: "1.0"
state: draft
---

# Bracket Motor

## Назначение

## Производственные примечания

## Контроль качества
```

## Еще лучше: один Markdown как source of truth

Можно вообще отказаться от отдельного `object.yaml` для простых объектов:

```text
objects/demo-prt-0001-v1.0/
  demo-prt-0001-v1.0.md
  history.yaml.log
  files/
```

А все метаданные хранить во frontmatter:

```markdown
---
id: demo-prt-0001-v1.0
class: prt
title: Bracket Motor
revision: "1.0"
state: draft

metadata:
  material: Aluminum 6061
  unit: pcs
  mass_g: 120

relations:
  - type: has_drawing
    to: demo-drw-0001-v1.0

artifacts:
  - kind: cad
    role: primary_step
    path: files/cad/bracket.step
    checksum: sha256:...
---

# Bracket Motor
```

Это самый чистый вариант для Markdown-PLM.

## Где JSON все-таки может быть оправдан

Я бы оставил JSON только в трех случаях:

1. **IPC/API между frontend и backend**  
    Там JSON нормален, потому что браузер/React/TypeScript естественно работают с JSON.
    
2. **Внутренний кэш/индекс**  
    Но лучше не `index.json`, а SQLite или другой быстрый индекс. И обязательно: индекс пересобираемый, не источник истины.
    
3. **JSONL event log**, если нужна высокая надежность append-only записи  
    Но если важнее человекочитаемость, можно сделать `history.yaml.log`.
    

## Я бы переписал принцип в ТЗ так

```text
YAML is the canonical structured data format.
Markdown frontmatter is the preferred object metadata format.
JSON is allowed only for transport protocols, generated machine output, or rebuildable caches.
No canonical domain data should be stored only in JSON.
```

## Исправленная файловая модель

```text
MyProject/
├── project.yaml
├── objects/
│   └── demo-prt-0001-v1.0/
│       ├── demo-prt-0001-v1.0.md
│       ├── history.yaml.log
│       ├── files/
│       └── generated/
├── config/
│   ├── naming.yaml
│   ├── classes.yaml
│   ├── lifecycle.yaml
│   ├── validation.yaml
│   ├── file-routing.yaml
│   └── exports.yaml
├── templates/
├── exports/
│   └── demo-rel-0001-v1.0/
│       ├── manifest.yaml
│       ├── bom.yaml
│       └── files/
└── .plm/
    ├── index.db
    ├── transactions/
    ├── cache/
    └── recovery/
```

## Вывод

Да, для этого проекта **YAML должен быть основным форматом**, потому что он лучше соответствует идее инженерной системы, которую можно читать, diff-ить и править руками.

Я бы убрал из ТЗ `item.json`, `manifest.json`, `index.json` и заменил на:

```text
object.yaml или YAML frontmatter
manifest.yaml
bom.yaml
history.yaml.log
index.db как rebuildable cache
```

Самый красивый принцип:

```text
Markdown для инженерного смысла.
YAML для структуры.
Git для истории.
SQLite/индекс только для скорости.
JSON только для API.
```

---
Да. Для `go-plm` я бы выбрал стек так, чтобы он был **максимально Go-native**, без лишней инфраструктуры, с понятной сборкой в один desktop-бинарник, YAML как основной формат, Git как история, SQLite только как пересобираемый индекс.

## Мой выбор: лучший стек под Go

```text
Desktop:        Wails v2 сейчас, Wails v3 позже
Backend:        Go 1.23/1.24+
Frontend:       React + TypeScript + Vite
UI:             Tailwind + shadcn/ui или Radix UI
Editor:         CodeMirror 6
Markdown:       goldmark
YAML:           gopkg.in/yaml.v3
Index/cache:    SQLite через modernc.org/sqlite
Git:            go-git/v5
Search:         SQLite FTS5 сначала, Bleve позже
FSM:            свой маленький FSM engine
API:            Wails bindings + internal command/query layer
Logs/events:    YAMLL или JSONL, но domain source of truth — YAML
Tests:          Go test + Vitest + Playwright
CI:             GitHub Actions
```

---

# 1. Desktop shell

## Рекомендация: Wails v2 сейчас

В README у тебя заявлен Wails 3, но на официальной странице Wails “Next Version” прямо указано, что это unreleased documentation, а актуальная стабильная версия — `v2.12.0` .

Поэтому я бы сделал так:

```text
Сейчас: Wails v2
Позже: миграция на Wails v3, когда он станет стабильным
```

Почему Wails подходит:

- Go backend в одном приложении;
- React/TypeScript frontend;
- desktop без Electron;
- нормальная кроссплатформенность;
- удобно упаковывать как один продукт;
- подходит под локальную PLM/PDM.

Wails поддерживает Windows, macOS и Linux, но требует WebView2 на Windows и WebKitGTK/libgtk на Linux .

**Альтернативы:**

|Вариант|Оценка|
|---|---|
|Wails|Лучший выбор под Go desktop|
|Tauri|Отличный, но backend Rust-first|
|Fyne|Go-native, но UI слабее для сложного PLM|
|Gio|Мощно, но дорого по разработке UI|
|Electron + Go API|Слишком тяжело|

**Итог:**  
Для `go-plm` — **Wails v2 + React + TypeScript**.

---

# 2. Frontend

## Рекомендация

```text
React 19
TypeScript
Vite
Tailwind CSS
Radix UI или shadcn/ui
TanStack Query
Zustand
CodeMirror 6
React Flow — для графов зависимостей
```

## Почему так

PLM-интерфейс будет сложный:

```text
project tree
BOM table
properties panel
markdown editor
lifecycle buttons
validation diagnostics
change impact graph
manufacturing matrix
release readiness
```

Для такого UI лучше React + TypeScript, а не чистый HTML или Go-native GUI.

## UI-библиотеки

Я бы выбрал:

```text
Tailwind CSS
+
Radix UI primitives
+
свои engineering components
```

или:

```text
shadcn/ui
```

Но важное замечание: не надо строить весь продукт вокруг готового UI-kit. Нужны свои доменные компоненты:

```text
ObjectBadge
LifecycleStateBadge
RelationPicker
BOMGrid
ReadinessMatrix
ArtifactList
ValidationPanel
ReleaseBlockerCard
WhereUsedGraph
```

---

# 3. Markdown editor

## Рекомендация: CodeMirror 6

Это лучший выбор для такого проекта.

Нужны расширения:

```text
Markdown syntax
YAML frontmatter highlighting
inline diagnostics
autocomplete for object IDs
hover preview for relations
link resolver
command palette
split preview
```

## Markdown parser на backend

Рекомендую:

```text
github.com/yuin/goldmark
```

`goldmark` — Markdown parser на Go, он CommonMark-compliant, расширяемый, имеет AST и написан на pure Go .

Почему это важно:

- можно парсить Markdown на backend;
- можно валидировать структуру документа;
- можно извлекать заголовки, блоки, таблицы;
- можно делать generated docs;
- можно делать preview/export pipeline.

---

# 4. YAML

## Рекомендация

```text
gopkg.in/yaml.v3
```

Это нормальный основной YAML-пакет для Go. Он умеет encode/decode YAML, поддерживает большую часть YAML 1.2, имеет `yaml.Node`, `Encoder`, `Decoder`, `KnownFields` и стабильный v3 API .

Для твоей системы я бы сделал так:

```text
YAML = canonical domain data
Markdown frontmatter = object metadata
SQLite = rebuildable index
JSON = только IPC/API, если нужно
```

## Формат объекта

Лучший вариант:

```text
objects/demo-prt-0001-v1.0/
  demo-prt-0001-v1.0.md
  history.yaml.log
  files/
  generated/
```

А метаданные во frontmatter:

```yaml
---
id: demo-prt-0001-v1.0
class: prt
title: Bracket Motor
revision: "1.0"
state: draft

metadata:
  material: Aluminum 6061
  unit: pcs
  mass_g: 120

relations:
  - type: has_drawing
    to: demo-drw-0001-v1.0

artifacts:
  - kind: cad
    role: primary_step
    path: files/cad/bracket.step
    checksum: sha256:...
---
```

`object.yaml` я бы оставил **опциональным** только для очень больших объектов, где frontmatter становится слишком тяжелым.

---

# 5. Локальный индекс

## Рекомендация: SQLite через `modernc.org/sqlite`

Для индекса лучше не YAML и не JSON. Индекс — это не source of truth, а быстрый пересобираемый кэш.

```text
.plm/index.db
```

Рекомендую:

```text
modernc.org/sqlite
```

Это CGo-free SQLite driver для Go; SQLite сам по себе in-process, serverless, zero-configuration и transactional database engine .

Почему это хорошо:

- не нужен внешний сервер;
- работает в одном бинарнике;
- быстрые запросы;
- можно FTS5 для поиска;
- можно хранить projections;
- можно пересобрать из Markdown/YAML;
- меньше боли, чем с самописным index.yaml.

## Что хранить в индексе

```text
objects
relations
artifacts
states
revisions
where_used
bom_flat
search_index
validation_status
release_readiness
```

Важно:

```text
YAML/Markdown = truth
SQLite = cache/projection
```

Если `index.db` удалили — приложение должно пересобрать его из файлов.

---

# 6. Git

## Рекомендация: `go-git/v5`

```text
github.com/go-git/go-git/v5
```

`go-git` — pure Go Git implementation, умеет работать с Git-репозиториями через idiomatic Go API, поддерживает high-level и low-level операции .

Для `go-plm` этого достаточно для:

```text
status
diff
commit checkpoint
tag release
log
compare releases
show changed files
```

Но важный нюанс: `go-git` не надо использовать как полноценную замену всему Git CLI для сложных операций. Например, merge/conflict resolution лучше оставить на будущее или делать через внешний git, если понадобится.

## UX-абстракции над Git

Пользователь должен видеть не Git-команды, а инженерные действия:

```text
Create checkpoint
Show changes
Compare with last release
Create release tag
Restore checkpoint
```

---

# 7. FSM / workflow

## Рекомендация: свой FSM engine

Не надо брать тяжелый workflow engine. Для desktop PLM лучше свой небольшой механизм:

```go
type ProcessDefinition struct {
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

Конфигурация в YAML:

```yaml
processes:
  object_lifecycle:
    initial: draft
    states:
      - draft
      - in_review
      - approved
      - released
      - obsolete
      - archived

    transitions:
      submit_review:
        from: draft
        to: in_review
        guards:
          - valid_name
          - required_metadata

      release:
        from: approved
        to: released
        guards:
          - no_release_blockers
          - bom_valid
          - checksums_actual
```

Почему свой FSM лучше:

- меньше зависимостей;
- проще тестировать;
- проще объяснять пользователю;
- можно делать “Why can’t I release?”;
- guards/effects легко подключаются как Go-интерфейсы.

---

# 8. API внутри приложения

Я бы не делал сразу HTTP JSON-RPC, если Wails bindings хватает.

Лучше так:

```text
frontend
  ↓
Wails bindings
  ↓
Application service
  ↓
Command / Query layer
  ↓
Domain services
  ↓
Store / Index / Git
```

Внутри Go:

```go
type AppService struct {
    Commands *CommandBus
    Queries  *QueryBus
}
```

Команды:

```text
CreateObject
UpdateObject
AttachFile
AddRelation
RunTransition
CreateRelease
RebuildIndex
CommitCheckpoint
```

Запросы:

```text
GetObject
ListObjects
GetBOM
GetWhereUsed
GetReadiness
GetGitStatus
SearchObjects
```

Это даст чистую архитектуру без привязки к Wails. Потом можно будет добавить HTTP API, CLI или серверный режим.

---

# 9. Search

## MVP

```text
SQLite FTS5
```

Индексировать:

```text
id
title
class
state
metadata
markdown headings
markdown body
artifact names
```

## Позже

```text
Bleve
```

Bleve можно добавить, если нужен более мощный full-text search, фасеты, scoring, fuzzy search. Но для MVP SQLite FTS достаточно.

---

# 10. Файловые операции

## Рекомендация

Использовать стандартную библиотеку Go + свои atomic helpers:

```text
os
io/fs
path/filepath
crypto/sha256
```

Критическое правило:

```text
write temp file → fsync → atomic rename → append history → update index
```

Нужен пакет:

```text
internal/store/transaction
```

Пример:

```go
type FileTransaction struct {
    ID      string
    Root    string
    Writes  []PlannedWrite
    Renames []PlannedRename
    Events  []Event
}
```

Для PLM это важнее, чем выбор красивой библиотеки.

---

# 11. Event log

Я бы выбрал:

```text
history.yaml.log
```

Формат: один YAML-документ на событие, разделитель `---`.

```yaml
---
event_id: evt_000001
time: 2026-06-06T10:00:00Z
actor: local-user
type: object.created
object_id: demo-prt-0001-v1.0
payload:
  class: prt
  title: Bracket Motor
---
event_id: evt_000002
time: 2026-06-06T10:05:00Z
actor: local-user
type: lifecycle.transitioned
object_id: demo-prt-0001-v1.0
payload:
  from: draft
  to: in_review
  transition: submit_review
```

JSONL быстрее и стандартнее для машинной обработки, но для твоей философии лучше YAML log. Если понадобится производительность — можно добавить бинарный/SQLite event projection, но не как source of truth.

---

# 12. Таблица полного стека

|Слой|Выбор|Почему|
|---|---|---|
|Desktop|Wails v2|Go-native desktop, стабильнее v3 сейчас|
|Future desktop|Wails v3|После стабилизации|
|Backend|Go 1.23/1.24+|Быстро, надежно, один бинарник|
|Frontend|React + TypeScript|Лучший баланс для сложного UI|
|Build frontend|Vite|Быстро и просто|
|UI|Tailwind + Radix/shadcn|Гибкие компоненты|
|Editor|CodeMirror 6|Лучший web editor для Markdown/YAML|
|Markdown parser|goldmark|Pure Go, CommonMark, extensible|
|YAML|gopkg.in/yaml.v3|Стандартный зрелый YAML для Go|
|Index|SQLite / modernc.org/sqlite|CGo-free, local, transactional|
|Search|SQLite FTS5|Достаточно для MVP|
|Git|go-git/v5|Pure Go Git API|
|FSM|Custom|Нужно объяснимое поведение|
|Logs|YAMLL|Человекочитаемая история|
|Tests backend|Go test + testify|Просто и стабильно|
|Tests frontend|Vitest|Для TS/React|
|E2E|Playwright|Проверка desktop/web UI|
|CI|GitHub Actions|Достаточно для repo|
|Packaging|Wails build + NSIS позже|Desktop delivery|

---

# 13. Рекомендуемая структура репозитория

```text
go-plm/
├── go.mod
├── go.sum
├── README.md
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
│   │   └── event/
│   ├── app/
│   │   ├── commands/
│   │   ├── queries/
│   │   └── services/
│   ├── store/
│   │   ├── filesystem/
│   │   ├── transaction/
│   │   ├── index/
│   │   └── recovery/
│   ├── parser/
│   │   ├── markdown/
│   │   └── frontmatter/
│   ├── naming/
│   ├── validation/
│   ├── process/
│   │   ├── fsm/
│   │   ├── guards/
│   │   └── effects/
│   ├── gitops/
│   └── modules/
│       ├── bom/
│       ├── release/
│       ├── manufacturing/
│       ├── procurement/
│       ├── shopfloor/
│       ├── standardparts/
│       └── change/
├── frontend/
│   ├── package.json
│   ├── vite.config.ts
│   └── src/
│       ├── app/
│       ├── features/
│       ├── components/
│       ├── editor/
│       ├── api/
│       └── state/
├── examples/
│   └── demo-project/
├── templates/
├── docs/
└── .github/
    └── workflows/
        └── ci.yml
```

---

# 14. Что не брать

Я бы **не брал** в MVP:

```text
PostgreSQL
Docker requirement
Kubernetes
GraphQL
Temporal
Kafka
NATS
Redis
MongoDB
Electron
full-blown plugin VM
complex permissions engine
```

Почему: это локальная инженерная desktop PLM. Чем меньше внешней инфраструктуры, тем надежнее и проще.

---

# 15. Итоговый выбор

Лучший стек под `go-plm`:

```text
Go + Wails v2 + React/TypeScript + CodeMirror 6
YAML frontmatter + goldmark + gopkg.in/yaml.v3
SQLite index через modernc.org/sqlite
Git через go-git/v5
Custom FSM + validation engine
YAML source of truth, SQLite cache, Git history
```

Главная архитектурная формула:

```text
Markdown/YAML = данные
Go = доменная логика и надежность
SQLite = скорость
Git = история
React/CodeMirror = удобный инженерный UX
```

Это даст систему, которую можно реально собрать, поддерживать и развивать без превращения в тяжелый enterprise-монолит.