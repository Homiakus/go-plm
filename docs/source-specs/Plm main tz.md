# Главное техническое задание

# go-plm / Markdown-PLM

## Локальная Git-native PLM/PDM-система на Go, Markdown, YAML и Git

Версия: `1.0-draft`  
Статус: `основное техническое задание`  
Назначение: разработка desktop PLM/PDM-системы для управления инженерными объектами, документацией, BOM, стандартными деталями, релизами, изменениями и производственной готовностью.

---

# 1. Назначение системы

`go-plm` — это локальная desktop PLM/PDM-система для инженерных и производственных проектов.

Система должна позволять управлять:

```text
деталями;
сборками;
чертежами;
CAD/CAM-файлами;
BOM/eBOM/mBOM;
стандартными деталями;
документацией;
технологическими картами;
рабочими инструкциями;
релизами;
изменениями;
производственной готовностью;
закупочными списками;
историей изменений.
```

Система строится вокруг принципа:

```text
Markdown/YAML = источник истины
Go backend = логика, надежность, транзакции, Git, индекс
SQLite = пересобираемый индекс
Git = история изменений и релизы
React UI = инженерная рабочая среда
```

---

# 2. Главная цель проекта

Создать легкую, локальную, прозрачную и расширяемую PLM/PDM-систему, которую можно использовать без тяжелой серверной инфраструктуры.

Система должна закрывать главную инженерную цепочку:

```text
создать объект
→ описать его
→ прикрепить файлы
→ связать с другими объектами
→ включить в BOM
→ проверить готовность
→ провести через lifecycle
→ выпустить релиз
→ передать данные в производство/закупки/ERP
```

---

# 3. Основные пользователи

## 3.1. Инженер / конструктор

Работает с деталями, сборками, CAD-файлами, чертежами, спецификациями и изменениями.

Основные действия:

```text
создать деталь;
создать сборку;
прикрепить STEP/PDF/DXF;
добавить объект в BOM;
создать чертеж;
заполнить материал, массу, размеры;
отправить объект на проверку;
посмотреть, что мешает выпуску.
```

## 3.2. Технолог

Работает с производственными файлами, техпроцессами, операциями, инструкциями и матрицей готовности.

Основные действия:

```text
создать техкарту;
создать work instruction;
создать файл резки/гибки/ЧПУ;
задать операции;
задать контрольные точки;
проверить производственную готовность.
```

## 3.3. Закупщик

Работает со стандартными деталями, поставщиками, заменами и закупочным списком.

Основные действия:

```text
найти стандартную деталь;
добавить поставщика;
проверить дубли;
сформировать buy list;
экспортировать закупочный список;
предложить замену.
```

## 3.4. Производственный пользователь

Работает с инструкциями и build records.

Основные действия:

```text
открыть инструкцию;
пройти шаги сборки;
зафиксировать evidence;
отметить контроль;
завершить build.
```

## 3.5. Администратор проекта

Настраивает правила проекта.

Основные действия:

```text
создать проект;
настроить классы объектов;
настроить naming standard;
настроить lifecycle;
настроить validation rules;
настроить шаблоны;
пересобрать индекс;
восстановить проект после ошибки.
```

---

# 4. Основные принципы системы

## 4.1. Markdown/YAML как источник истины

Все канонические данные должны храниться в человекочитаемых файлах:

```text
project.yaml
config/*.yaml
objects/*/*.md
objects/*/history.yaml.log
templates/*.md
exports/*/*.yaml
```

SQLite, кэш и поисковые индексы являются производными данными и должны пересобираться из файлов.

---

## 4.2. Объект важнее файла

Пользователь создает не файл, а инженерный объект.

Объект состоит из:

```text
ID;
класса;
версии;
статуса;
YAML frontmatter;
Markdown body;
relations;
artifacts;
history;
index projection.
```

Физические файлы являются артефактами объекта.

---

## 4.3. ID = имя файла = номер документа

Каждый инженерный объект имеет стабильный ID:

```text
a320-prt-0001-v1.0
```

Этот ID используется как:

```text
идентификатор объекта;
имя Markdown-файла;
базовое имя CAD/PDF/DXF-файлов;
номер документа;
ключ связи;
part number для экспорта.
```

---

## 4.4. Минимальное имя файла

В имени запрещено кодировать:

```text
материал;
толщину;
покрытие;
цвет;
дату;
автора;
поставщика;
способ производства;
статус.
```

Неправильно:

```text
a320-prt-0001-v1.0-al5052-2mm-laser-blue.step
```

Правильно:

```text
a320-prt-0001-v1.0.step
```

Атрибуты должны храниться в YAML:

```yaml
metadata:
  material: al5052
  thickness_mm: 2.0
  manufacturing_method: laser_cut
  surface_treatment: anodize
  finish_color: ral7035
```

---

## 4.5. Связи как основа PLM

BOM, where-used, change impact, чертежи, релизы, инструкции и замены строятся через `relations`.

Главная формула:

```text
Object + Relation + Artifact + Process + Validation + Event + Projection
```

---

## 4.6. Git-native подход

Git используется для:

```text
истории изменений;
checkpoint;
diff;
release tag;
сравнения релизов;
восстановления.
```

Пользователь не должен работать с Git-командами напрямую. В UI должны быть инженерные действия:

```text
Create Checkpoint
Show Changes
Compare with Release
Create Release Tag
Restore Checkpoint
```

---

# 5. Границы MVP

## 5.1. Входит в MVP

```text
создание/открытие проекта;
создание объектов;
Markdown/YAML-документы;
редактор Markdown;
панель свойств;
дерево объектов;
relations;
простая BOM;
where-used;
lifecycle FSM;
validation engine;
Git status;
checkpoint;
SQLite index;
поиск;
стандартные детали базового уровня;
release readiness;
release manifest;
экспорт BOM/manifest в YAML/CSV.
```

## 5.2. Не входит в MVP

```text
полноценная ERP;
полноценная MES;
полноценный WMS;
серверная многопользовательская PLM;
облачная синхронизация;
сложная ролевая модель;
электронная подпись;
полноценный CAD viewer;
автоматический парсинг всех CAD-форматов;
сложный merge/conflict resolution Git.
```

---

# 6. Технологический стек

## 6.1. Backend

```text
Go
gopkg.in/yaml.v3
goldmark
modernc.org/sqlite
SQLite FTS5
go-git
custom FSM engine
custom validation engine
custom file transaction engine
```

## 6.2. Frontend

```text
React
TypeScript
Vite
Tailwind CSS
shadcn/ui
Radix UI
Vditor
CodeMirror 6
TanStack Table
Zustand
React Flow / xyflow
Mermaid
```

## 6.3. Тестирование и CI

```text
Go test
Vitest
Playwright
GitHub Actions
Storybook optional
```

---

# 7. Основная файловая структура проекта PLM

```text
MyProject/
├── project.yaml
├── objects/
│   └── a320-prt-0001-v1.0/
│       ├── a320-prt-0001-v1.0.md
│       ├── history.yaml.log
│       ├── files/
│       │   ├── cad/
│       │   ├── drawings/
│       │   ├── manufacturing/
│       │   ├── inspection/
│       │   ├── certificates/
│       │   └── images/
│       └── generated/
│           ├── bom/
│           ├── reports/
│           └── exports/
├── config/
│   ├── classes.yaml
│   ├── naming.yaml
│   ├── lifecycle.yaml
│   ├── validation.yaml
│   ├── relation-types.yaml
│   ├── file-routing.yaml
│   └── exports.yaml
├── templates/
│   ├── object/
│   ├── project/
│   ├── release/
│   ├── manufacturing/
│   └── snippets/
├── exports/
├── libraries/
│   └── standard-parts/
├── .plm/
│   ├── index.db
│   ├── transactions/
│   ├── cache/
│   └── recovery/
└── .git/
```

---

# 8. Naming standard

## 8.1. Базовый формат ID

```text
[project]-[class]-[sequence]-v[major].[minor]
```

Примеры:

```text
a320-prt-0001-v1.0
a320-asm-0100-v1.0
a320-drw-0001-v1.0
a320-cut-0001-v1.0
```

## 8.2. Формат стандартной детали

```text
[project]-std-[std_class]-[std_id]-v[major].[minor]
```

Примеры:

```text
a320-std-fst-iso4017-m6x30-a2-v1.0
a320-std-brg-6205-2rs-c3-v1.0
a320-std-elc-res-10k-0603-1pct-v1.0
```

## 8.3. Версионирование

Используется SemVer-подобная схема:

```text
v1.0
v1.1
v2.0
```

Правила:

```text
minor version — совместимое изменение;
major version — несовместимое изменение;
released-объект нельзя менять напрямую;
для released-объекта создается новая revision/version.
```

---

# 9. Основные классы объектов

```text
prt   — Part / деталь
asm   — Assembly / сборка
drw   — Drawing / чертеж
doc   — Document / документ
std   — Standard Part / стандартная деталь
mat   — Material / материал
cut   — Cutting File / файл резки
bnd   — Bending File / файл гибки
nc    — NC Program / программа ЧПУ
ins   — Inspection / контроль
tpc   — Tech Process Card / технологическая карта
wi    — Work Instruction / рабочая инструкция
bom   — Formal BOM / формальная спецификация
rel   — Release Package / релизный пакет
cr    — Change Request / запрос на изменение
build — Build Record / производственная запись
```

---

# 10. Структура Markdown-объекта

Каждый объект хранится в `.md` файле с YAML frontmatter.

```markdown
---
id: a320-prt-0001-v1.0
project: a320
class: prt
sequence: "0001"
version: "1.0"
revision: "1.0"
state: draft
title: Bracket Motor

metadata:
  lifecycle_stage: development
  owner: local-user
  unit: pcs
  make_buy: make
  material: al5052
  thickness_mm: 2.0
  manufacturing_method: laser_cut

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
---

# Bracket Motor

## Назначение

## Материал

## Производственные примечания

## Контроль качества
```

---

# 11. Relations model

## 11.1. Назначение

`relations` описывают связи между объектами.

Основные типы:

```text
contains;
used_in;
has_drawing;
drawing_of;
has_document;
manufacturing_file_for;
process_plan_for;
work_instruction_for;
inspection_for;
substitutes;
equivalent_to;
replaces;
affects;
includes;
released_in.
```

## 11.2. Пример BOM-связи

```yaml
relations:
  - type: contains
    to: a320-prt-0001-v1.0
    quantity: 2
    unit: pcs
    metadata:
      position: "10"
```

## 11.3. Stored vs projected relations

Хранится только авторская связь.

Пример:

```yaml
# В сборке хранится:
relations:
  - type: contains
    to: a320-prt-0001-v1.0
    quantity: 2
    unit: pcs
```

Обратная связь `used_in` строится индексом.

Правило:

```text
Store relation once.
Build inverse relations through projections.
```

---

# 12. Artifact model

## 12.1. Назначение

`Artifact` — файл, связанный с объектом.

Примеры:

```text
STEP;
PDF;
DXF;
DWG;
NC;
G-code;
PNG;
сертификат;
отчет;
generated BOM;
release manifest.
```

## 12.2. Пример artifact

```yaml
artifacts:
  - id: art-0001
    kind: cad
    role: primary_step
    path: files/cad/a320-prt-0001-v1.0.step
    original_name: bracket-final.step
    checksum: sha256:...
    size_bytes: 2048123
    generated: false
    required: true
    status: present
```

## 12.3. Placeholder artifact

Если файл обязателен, но еще не создан:

```yaml
artifacts:
  - id: art-drawing-pdf
    kind: drawing
    role: drawing_pdf
    path: null
    required: true
    status: missing
```

Validation engine должен показывать blocker.

---

# 13. BOM model

## 13.1. eBOM

Engineering BOM строится из `contains` relations.

```yaml
relations:
  - type: contains
    to: a320-prt-0001-v1.0
    quantity: 2
    unit: pcs
```

## 13.2. Flat BOM

Backend должен уметь строить плоский BOM:

```text
root assembly
→ recursive contains
→ multiply quantities
→ group by object ID
→ output rows
```

## 13.3. mBOM

Manufacturing BOM строится из:

```text
eBOM;
manufacturing files;
process plans;
work instructions;
inspection objects;
tools;
consumables;
packaging.
```

## 13.4. Released BOM

Released BOM создается только из approved/released объектов.

Проверки:

```text
нет draft children;
все обязательные чертежи есть;
все checksum актуальны;
стандартные детали approved/released;
нет release blockers.
```

---

# 14. Стандартные детали

## 14.1. Назначение

Стандартная деталь — это полноценный PLM-объект класса `std`.

Примеры:

```text
болты;
гайки;
шайбы;
подшипники;
резисторы;
разъемы;
датчики;
двигатели;
профили;
кабели;
покупные компоненты.
```

## 14.2. Почему стандартные детали отдельные

Они отличаются от проектных деталей:

```text
используются во многих сборках;
часто покупаются;
имеют поставщиков;
имеют manufacturer part number;
имеют аналоги и замены;
могут быть глобальными;
часто сразу released/approved;
нуждаются в duplicate detection.
```

## 14.3. YAML стандартной детали

```markdown
---
id: a320-std-fst-iso4017-m6x30-a2-v1.0
project: a320
class: std
std_class: fst
std_id: iso4017-m6x30-a2
version: "1.0"
revision: "1.0"
state: released
title: Hex bolt ISO 4017 M6x30 A2

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
  key: fst:iso4017:m6:30:a2

procurement:
  preferred_supplier: supplier-acme
  manufacturer: acme
  manufacturer_part_number: acme-iso4017-m6x30-a2
  supplier_part_number: acme-123456
  lead_time_days: 7
  min_order_qty: 100
  package_qty: 100
  unit_cost:
    amount: 0.08
    currency: usd
  approved_vendor_status: approved

relations:
  - type: equivalent_to
    to: a320-std-fst-din933-m6x30-a2-v1.0
  - type: substitutes
    to: a320-std-fst-iso4017-m6x30-a4-v1.0
---

# Hex bolt ISO 4017 M6x30 A2
```

## 14.4. Duplicate detection

Система должна определять:

```text
exact duplicate;
probable duplicate;
possible equivalent;
different but substitutable;
not duplicate.
```

Ключ нормализации:

```text
fst:iso4017:m6:30:a2
```

Если найден дубль, UI должен предложить:

```text
Open existing;
Add supplier to existing;
Create variant;
Cancel.
```

## 14.5. Standard Parts Center

Должен поддерживать:

```text
поиск стандартных деталей;
фильтры по классу, стандарту, материалу, поставщику;
создание новой стандартной детали;
импорт библиотеки;
поиск дублей;
управление заменами;
approved vendor list;
where-used;
экспорт в ERP/buy list.
```

---

# 15. Lifecycle / FSM

## 15.1. Object lifecycle

```text
draft
→ in_review
→ approved
→ released
→ obsolete
→ archived
```

Дополнительное состояние:

```text
blocked
```

## 15.2. Пример lifecycle config

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
          - bom_valid
          - children_released
          - checksums_actual

      revise:
        from: released
        to: draft
        effects:
          - create_new_revision
          - create_change_event
```

## 15.3. Guard engine

Guards:

```text
valid_name;
required_metadata;
no_broken_relations;
checksums_actual;
bom_valid;
children_released;
standard_parts_approved;
no_release_blockers;
drawing_exists;
manufacturing_ready;
procurement_ready.
```

## 15.4. Effects

Effects:

```text
create_new_revision;
append_history_event;
create_git_tag;
generate_manifest;
update_index;
recalculate_readiness;
create_change_request.
```

---

# 16. Validation engine

## 16.1. Уровни валидации

```text
syntax validation;
schema validation;
naming validation;
metadata validation;
relation validation;
artifact validation;
BOM validation;
lifecycle validation;
release validation;
standard parts validation;
procurement validation;
project integrity validation.
```

## 16.2. Diagnostic model

```yaml
diagnostic:
  id: diag-0001
  severity: blocker
  code: REQUIRED_METADATA_MISSING
  object_id: a320-prt-0001-v1.0
  path: metadata.material
  message: Required metadata field material is missing
  suggested_actions:
    - fill_material
```

## 16.3. Severity

```text
info;
warning;
blocker;
error.
```

## 16.4. Why Can't I Assistant

Система должна объяснять, почему действие невозможно.

Пример:

```text
Нельзя выпустить объект a320-asm-0100-v1.0.

Причины:
1. Дочерний объект a320-prt-0004-v1.0 находится в draft.
2. Отсутствует чертеж.
3. Checksum STEP-файла устарел.
4. Стандартная деталь не approved.

Действия:
[Open child]
[Create drawing]
[Recalculate checksum]
[Open standard part]
```

---

# 17. Автогенерация объектов и файлов

## 17.1. Общий pipeline создания объекта

```text
CreateObjectDraft
→ GenerateIdentity
→ ResolveTemplate
→ AutofillMetadata
→ BuildRelations
→ PlanFileStructure
→ ValidateDraft
→ ExecuteFileTransaction
→ AppendHistory
→ UpdateIndex
→ ReturnObjectSnapshot
```

## 17.2. Preview before create

Перед созданием UI должен показать preview:

```text
Generated ID:
a320-prt-0008-v1.0

Files to create:
objects/a320-prt-0008-v1.0/a320-prt-0008-v1.0.md
objects/a320-prt-0008-v1.0/history.yaml.log

Folders:
files/cad
files/drawings
generated

Relations:
a320-asm-0100-v1.0 contains a320-prt-0008-v1.0 qty=1 pcs

Artifacts:
C:\cad\bracket.step → files/cad/a320-prt-0008-v1.0.step
```

## 17.3. File transaction

Создание должно быть транзакционным:

```text
create transaction plan;
write temp files;
fsync;
atomic rename;
append history;
update index;
commit transaction marker;
cleanup.
```

При сбое система должна предложить recovery:

```text
Complete transaction;
Rollback transaction;
Open details.
```

## 17.4. Автосвязи

Примеры:

```text
создание Part из Assembly
→ добавить contains в Assembly;

создание Drawing из Part
→ добавить has_drawing/drawing_of;

создание Cutting File из Part
→ добавить manufacturing_file_for;

создание Inspection из Part
→ добавить inspection_for;

создание Release Package из Assembly
→ добавить includes для объектов релиза.
```

---

# 18. UI требования

## 18.1. Основной layout

```text
┌────────────────────────────────────────────────────────────┐
│ Top Bar: Project | Search | Command Palette | Git | Tasks   │
├───────────────┬────────────────────────┬───────────────────┤
│ Left Sidebar  │ Main Workspace         │ Right Inspector   │
│ Project Tree  │ Editor/BOM/Graph/etc.  │ Properties/State  │
├───────────────┴────────────────────────┴───────────────────┤
│ Bottom Bar: Save | Validation | Git | Index | Tasks         │
└────────────────────────────────────────────────────────────┘
```

## 18.2. Главные экраны

```text
Landing Page;
Create Project Wizard;
Project Tree;
Object Detail Screen;
Markdown Editor;
Metadata Panel;
Relations Panel;
BOM Screen;
Standard Parts Center;
Release Screen;
Manufacturing Screen;
Procurement Screen;
Change Impact Screen;
Git Changes Screen;
Settings/Admin Screen.
```

## 18.3. Command Palette

Горячая клавиша:

```text
Ctrl+K
```

Команды:

```text
Create Part;
Create Assembly;
Create Drawing;
Create Standard Part;
Attach File;
Add BOM Item;
Validate Object;
Validate Project;
Submit Review;
Approve;
Release;
Explain Blockers;
Show Where Used;
Show Change Impact;
Generate Buy List;
Create Checkpoint;
Rebuild Index.
```

## 18.4. Object Detail Screen

Header:

```text
[CLASS] Title
ID | Revision | State | Validation
[Primary Action] [Validate] [Checkpoint] [More]
```

Tabs:

```text
Document;
Metadata;
Relations;
BOM;
Files;
Validation;
History;
Diff.
```

## 18.5. Right Inspector

Показывает:

```text
Properties;
Lifecycle;
Readiness;
Relations summary;
Artifacts;
Actions.
```

## 18.6. Bottom Bar

Показывает:

```text
current project;
selected object;
save status;
validation status;
Git status;
index status;
running tasks.
```

---

# 19. Backend архитектура

## 19.1. Целевая структура backend

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
    ├── standardparts/
    └── change/
```

## 19.2. Command / Query separation

Commands изменяют систему:

```text
CreateProject;
OpenProject;
CreateObject;
UpdateObjectDocument;
UpdateObjectMetadata;
AttachArtifact;
AddRelation;
RemoveRelation;
RunTransition;
CreateReleasePackage;
CreateCheckpoint;
RebuildIndex.
```

Queries читают данные:

```text
GetProject;
ListObjects;
GetObject;
SearchObjects;
GetBOM;
GetWhereUsed;
GetChangeImpact;
ValidateObject;
GetGitStatus;
GetIndexStatus;
GetReleaseReadiness.
```

## 19.3. Правило записи

Только backend имеет право писать:

```text
Markdown;
YAML;
history.yaml.log;
SQLite index;
Git.
```

Frontend не пишет файлы напрямую.

---

# 20. SQLite index

## 20.1. Назначение

SQLite хранит быстрые проекции.

Файл:

```text
.plm/index.db
```

## 20.2. Что индексируется

```text
objects;
relations;
artifacts;
metadata;
BOM projections;
where-used;
standard parts normalized keys;
procurement projections;
validation summaries;
release readiness;
search index.
```

## 20.3. Правило

SQLite не является источником истины.

Если индекс удален или поврежден:

```text
RebuildIndex
```

должен восстановить его из Markdown/YAML.

---

# 21. Search

## 21.1. MVP search

Используется SQLite FTS5.

Индексировать:

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
supplier part number;
manufacturer part number.
```

## 21.2. Advanced search

На следующем этапе можно добавить Bleve для:

```text
fuzzy search;
facets;
highlighting;
advanced scoring;
duplicate candidates;
поиска стандартных деталей.
```

---

# 22. Git integration

## 22.1. Возможности MVP

```text
Git status;
changed files;
changed objects;
create checkpoint;
Git log;
create release tag;
compare with last release.
```

## 22.2. Checkpoint

Checkpoint — это Git commit с пользовательским сообщением.

UI:

```text
[Create Checkpoint]
Message:
Included objects:
Included files:
```

Backend:

```text
stage changed files;
commit;
append project event;
return Git status.
```

## 22.3. Release tag

Формат:

```text
release/a320-rel-0100-v1.0
```

---

# 23. Release package

## 23.1. Назначение

Release package фиксирует состав утвержденного релиза.

## 23.2. Состав

```text
release object;
manifest.yaml;
ebom.yaml;
mbom.yaml;
checksums.yaml;
files-list.yaml;
release-notes.md;
included Markdown files;
included artifacts;
Git tag.
```

## 23.3. Manifest example

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

files:
  - object_id: a320-prt-0001-v1.0
    path: files/cad/a320-prt-0001-v1.0.step
    checksum: sha256:...
```

---

# 24. Procurement / ERP export

## 24.1. Buy List

Buy list строится из flat BOM и standard parts.

Поля:

```text
object_id;
title;
quantity;
unit;
make_buy;
supplier;
manufacturer;
manufacturer_part_number;
supplier_part_number;
lead_time_days;
unit_cost;
order_qty.
```

## 24.2. ERP export

Форматы MVP:

```text
YAML;
CSV.
```

Пример:

```yaml
erp_bom:
  parent: a320-asm-0100-v1.0
  rows:
    - item_number: a320-std-fst-iso4017-m6x30-a2-v1.0
      quantity: 8
      unit: pcs
```

---

# 25. Manufacturing

## 25.1. Manufacturing Matrix

Показывает готовность объекта к производству:

```text
Object;
State;
BOM;
Drawing;
CAD;
Cutting file;
Bending file;
NC program;
Process plan;
Work instruction;
Inspection;
Safety;
Ready.
```

## 25.2. Work Instructions

Рабочая инструкция хранится как Markdown/YAML объект.

Содержит:

```text
инструменты;
материалы;
требования безопасности;
операции;
шаги;
проверки;
evidence requirements;
acceptance criteria.
```

## 25.3. Shop Floor

Не входит в MVP, но модель должна предусмотреть:

```text
BuildRecord;
QR token;
step execution;
evidence capture;
deviation report;
completion report.
```

---

# 26. Change management

## 26.1. Change Request

Released-объект нельзя менять напрямую.

При попытке изменить released-объект система должна предложить:

```text
Create Change Request;
Analyze Impact;
Create New Revision.
```

## 26.2. Impact analysis

Показывает:

```text
где используется объект;
какие сборки затронуты;
какие релизы затронуты;
какие инструкции затронуты;
какие закупочные позиции затронуты;
какие build records затронуты.
```

---

# 27. Нефункциональные требования

## 27.1. Надежность

```text
атомарная запись файлов;
транзакции файловых операций;
recovery mode;
history event для каждого изменения;
возможность пересобрать индекс;
валидация проекта при открытии;
никаких silent failures.
```

## 27.2. Производительность MVP

Целевые показатели:

```text
до 10 000 объектов в проекте;
открытие проекта с готовым индексом до 5 секунд;
поиск до 300 мс;
открытие объекта до 150 мс;
построение BOM до 1 секунды;
where-used до 1 секунды;
валидация одного объекта до 300 мс.
```

## 27.3. Безопасность

```text
не выполнять произвольные скрипты из проекта;
не открывать вложения без подтверждения;
запрещать path traversal;
не хранить секреты в Git;
писать только внутри project root;
проверять file routing.
```

## 27.4. Кроссплатформенность

```text
Windows 10/11;
macOS;
Linux.
```

---

# 28. Тестирование

## 28.1. Backend unit tests

```text
naming engine;
YAML parser;
frontmatter parser;
Markdown parser;
file transaction engine;
FSM transitions;
validation guards;
relations;
BOM cycle detection;
standard part normalization;
duplicate detection;
release readiness;
Git wrapper.
```

## 28.2. Integration tests

```text
create project;
create object;
attach file;
add relation;
build BOM;
run lifecycle;
create checkpoint;
create release package;
rebuild index.
```

## 28.3. Frontend tests

```text
object tree;
create object wizard;
metadata forms;
BOM table;
standard parts table;
validation panel;
command palette;
editor dirty state.
```

## 28.4. E2E tests

```text
создать проект;
создать деталь;
создать сборку;
добавить деталь в BOM;
прикрепить STEP;
запустить validation;
увидеть blocker;
исправить blocker;
создать checkpoint.
```

---

# 29. Этапы разработки

## Этап 0. Подготовка репозитория

Результат:

```text
go.mod;
cmd/plm;
internal skeleton;
frontend skeleton;
CI;
README с реальными командами;
examples/demo-project.
```

Критерии приемки:

```text
go test ./... проходит;
frontend build проходит;
CI запускается;
приложение открывает пустой shell.
```

---

## Этап 1. Core domain

Реализовать:

```text
Object;
Relation;
Artifact;
Event;
Revision;
ProjectConfig;
ObjectClass;
Diagnostics.
```

Критерии:

```text
объекты сериализуются;
валидация базовых полей работает;
unit tests есть.
```

---

## Этап 2. Filesystem store

Реализовать:

```text
CreateProject;
OpenProject;
CreateObject;
ReadObject;
UpdateObject;
ListObjects;
history.yaml.log;
atomic write;
file transactions.
```

Критерии:

```text
объект создается на диске;
Markdown/YAML валиден;
history пишется;
transaction recovery работает.
```

---

## Этап 3. Naming, parser, validation

Реализовать:

```text
naming v10;
frontmatter parser;
YAML validation;
Markdown parser;
required metadata;
relation integrity;
checksum validation.
```

---

## Этап 4. SQLite index

Реализовать:

```text
index.db;
objects table;
relations table;
artifacts table;
search FTS;
rebuild index;
index diagnostics.
```

---

## Этап 5. FSM lifecycle

Реализовать:

```text
configurable FSM;
guards;
effects;
RunTransition;
transition history;
Explain blockers.
```

---

## Этап 6. Desktop MVP UI

Реализовать:

```text
Wails app;
React shell;
project open/create;
project tree;
object detail;
Markdown editor;
metadata panel;
relations panel;
validation panel;
right inspector;
bottom status bar.
```

---

## Этап 7. BOM

Реализовать:

```text
contains relations;
BOM table;
flat BOM;
where-used;
cycle detection;
BOM export.
```

---

## Этап 8. Standard Parts Center

Реализовать:

```text
std object;
std naming;
normalized key;
standard parts table;
duplicate detection;
supplier block;
substitutions;
use in BOM.
```

---

## Этап 9. Git UX

Реализовать:

```text
Git status;
changed objects;
checkpoint;
diff;
log;
release tag.
```

---

## Этап 10. Release package

Реализовать:

```text
release object;
release readiness;
manifest.yaml;
ebom.yaml;
checksums.yaml;
export package;
release tag.
```

---

## Этап 11. Manufacturing / Procurement

Реализовать:

```text
manufacturing matrix;
process plan;
work instruction;
buy list;
ERP export YAML/CSV.
```

---

## Этап 12. Change impact

Реализовать:

```text
where-used graph;
impact graph;
change request;
affected objects report.
```

---

# 30. Definition of Done

Фича считается готовой, если:

```text
есть backend command/query;
есть domain logic;
есть валидация;
есть UI;
есть diagnostics;
изменения пишутся в Markdown/YAML;
history event создается;
index обновляется или инвалидируется;
есть unit tests;
есть integration tests для критичной логики;
ошибки отображаются понятным языком;
данные можно восстановить из source of truth.
```

---

# 31. Главные риски

## 31.1. Слишком большой scope

Митигировать:

```text
сначала core;
потом объектный UI;
потом BOM;
потом standard parts;
потом release;
потом manufacturing/procurement.
```

## 31.2. Повреждение файлов

Митигировать:

```text
file transactions;
atomic write;
backup before overwrite;
history log;
recovery mode;
tests.
```

## 31.3. Сложный UX

Митигировать:

```text
command palette;
right inspector;
Why Can't I assistant;
actionable diagnostics;
preview before write;
Git abstraction.
```

## 31.4. Хаос модулей

Митигировать общей моделью:

```text
Object;
Relation;
Artifact;
Process;
Validation;
Event;
Projection.
```

---

# 32. Финальный результат

После реализации система должна позволять инженеру работать так:

```text
создаю проект;
создаю деталь;
прикрепляю CAD;
создаю чертеж;
добавляю деталь в сборку;
получаю BOM;
добавляю стандартные детали;
проверяю readiness;
исправляю blockers;
создаю checkpoint;
выпускаю release package;
экспортирую BOM/buy list.
```

Система должна постоянно отвечать на вопросы:

```text
что это за объект?
где он используется?
из чего он состоит?
какие файлы к нему относятся?
какой у него статус?
что мешает выпуску?
что изменилось?
что сломается, если его изменить?
какие стандартные детали используются?
что нужно купить?
можно ли это выпустить?
```

---

# 33. Главная архитектурная формула

```text
Markdown/YAML = source of truth
Go = domain logic and reliability
SQLite = fast projections
Git = history and releases
React = engineering UX
```

И главная доменная формула:

```text
Object + Relation + Artifact + Process + Validation + Event + Projection
```

Все модули системы должны строиться поверх этой модели.