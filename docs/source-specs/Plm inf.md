# Information Model Specification

# Markdown-PLM / go-plm

## Подробная информационная модель PLM/PDM-системы, включая стандартные детали

Версия: `0.1`  
Статус: `draft`  
Назначение: описание доменной информационной модели, сущностей, связей, атрибутов, правил хранения, индексов, BOM, стандартных деталей, замен, версий и трассируемости.

---

# 1. Назначение информационной модели

Информационная модель Markdown-PLM описывает, как система представляет инженерные данные:

```text
изделия
детали
сборки
чертежи
CAD/CAM-файлы
BOM
технологические процессы
контроль
закупки
стандартные детали
изменения
релизы
производственные записи
```

Главная цель модели — обеспечить:

```text
1. стабильную идентификацию объектов;
2. читаемое файловое хранение;
3. строгие связи между объектами;
4. автоматическое построение BOM и where-used;
5. трассируемость изменений;
6. поддержку стандартных деталей и замен;
7. подготовку данных для ERP/MES;
8. возможность восстановления проекта из Markdown/YAML без скрытой БД.
```

---

# 2. Основные принципы модели

## 2.1. Markdown/YAML — источник истины

Канонические данные хранятся в:

```text
project.yaml
config/*.yaml
objects/*/*.md
objects/*/history.yaml.log
templates/*.md
exports/*/*.yaml
```

SQLite-индекс, кэш и проекции являются ускорителями, но не источником истины.

---

## 2.2. ID объекта = имя файла = номер документа

Каждый объект имеет стабильный ID:

```text
a320-prt-0001-v1.0
```

Этот ID используется как:

```text
ID объекта в PLM
имя Markdown-файла
базовое имя артефактов
номер документа в title block
part number для интеграций
ключ в связях
```

Пример:

```text
objects/a320-prt-0001-v1.0/a320-prt-0001-v1.0.md
files/cad/a320-prt-0001-v1.0.step
files/drawings/a320-drw-0001-v1.0.pdf
```

---

## 2.3. Имя минимальное, атрибуты в YAML

В имени запрещено хранить:

```text
материал
толщину
цвет
поставщика
покрытие
способ производства
автора
дату
технические параметры
```

Неправильно:

```text
a320-prt-0001-v1.0-al5052-2mm-laser-blue.step
```

Правильно:

```text
a320-prt-0001-v1.0.step
```

Атрибуты:

```yaml
metadata:
  material: al5052
  thickness_mm: 2.0
  manufacturing_method: laser_cut
  surface_treatment: anodize
  finish_color: ral7035
```

---

## 2.4. Объект важнее файла

Файл — это артефакт объекта.

Например, деталь:

```text
a320-prt-0001-v1.0
```

может иметь:

```text
Markdown description
STEP model
PDF drawing
DXF cutting file
inspection report
material certificate
```

Но все эти файлы принадлежат одному инженерному объекту или связаны с ним через отдельные производные объекты.

---

## 2.5. Связи — отдельный первый класс модели

BOM, drawing links, manufacturing files, substitutions, where-used и change impact должны строиться не по папкам, а по relations.

Связи — это основа PLM.

---

# 3. Верхнеуровневая структура модели

Информационная модель состоит из следующих базовых сущностей:

```text
Project
Object
ObjectClass
Metadata
Relation
Artifact
LifecycleState
Process
ValidationDiagnostic
Event
Revision
BOM
StandardPart
Substitution
ReleasePackage
ChangeRequest
BuildRecord
IndexProjection
```

Главные абстракции:

```text
Object
Relation
Artifact
Process
Event
Validation
Projection
```

---

# 4. Project

## 4.1. Назначение

`Project` — корневой контейнер PLM-данных.

Проект задает:

```text
код проекта
название
стандарт именования
классы объектов
счетчики sequence
жизненные циклы
правила валидации
шаблоны
маршрутизацию файлов
модули
экспортные правила
```

---

## 4.2. project.yaml

```yaml
project:
  code: a320
  title: A320 Test Project
  description: Engineering PLM workspace
  version: 1
  created_at: 2026-06-06T10:00:00Z
  created_by: local-user

naming:
  standard: v10
  case: kebab
  pattern: "[project]-[class]-[sequence]-v[major].[minor]"
  project_code: a320
  max_length_without_ext: 64

sequences:
  prt: 0007
  asm: 0100
  drw: 0003
  cut: 0001
  bnd: 0001
  nc: 0001
  ins: 0001
  tpc: 0001
  bom: 0001
  rel: 0001

modules:
  objects: true
  bom: true
  release: true
  manufacturing: true
  procurement: true
  standardparts: true
  shopfloor: false
  change_management: true

storage:
  source_of_truth: markdown_yaml
  index: .plm/index.db

git:
  enabled: true
  auto_checkpoint: false
  release_tags: true
```

---

## 4.3. Project invariants

Проект считается валидным, если:

```text
project.code соответствует naming rules;
project.yaml читается;
config/*.yaml валидны;
все объекты имеют уникальные ID;
все sequence не конфликтуют с существующими объектами;
индекс может быть пересобран из файлов;
.git существует, если git.enabled = true.
```

---

# 5. Object

## 5.1. Назначение

`Object` — базовая инженерная сущность.

Примеры:

```text
Part
Assembly
Drawing
Cutting File
Bending File
NC Program
Inspection Plan
Tech Process Card
BOM
Specification
Document
Report
Certificate
Material
Tool
Library
Standard Part
Release Package
Change Request
Build Record
```

---

## 5.2. Object identity

Каждый объект имеет:

```text
id
project
class
sequence или semantic id
version
revision
state
title
```

Пример:

```yaml
id: a320-prt-0001-v1.0
project: a320
class: prt
sequence: "0001"
version: "1.0"
revision: "1.0"
state: draft
title: Bracket Motor
```

---

## 5.3. Базовый YAML frontmatter объекта

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
```

---

## 5.4. Object invariants

Объект валиден, если:

```text
id соответствует naming standard;
id совпадает с именем Markdown-файла;
project совпадает с project.yaml;
class существует в classes.yaml;
version валидна;
state существует в lifecycle;
required metadata заполнены;
relations указывают на существующие объекты;
artifacts указывают на существующие файлы или допустимые placeholders;
YAML frontmatter парсится без ошибок;
Markdown body существует.
```

---

# 6. ObjectClass

## 6.1. Назначение

`ObjectClass` описывает тип объекта, его required fields, допустимые связи, шаблон, правила именования и жизненный цикл.

---

## 6.2. Базовые классы

```yaml
classes:
  prt:
    title: Part
    description: Деталь
    sequence_mode: numeric
    template: templates/object/prt.md
    lifecycle: object_lifecycle
    required_metadata:
      - unit
      - make_buy
    optional_metadata:
      - material
      - thickness_mm
      - manufacturing_method
      - surface_treatment

  asm:
    title: Assembly
    description: Сборка
    sequence_mode: numeric
    template: templates/object/asm.md
    lifecycle: object_lifecycle
    required_metadata:
      - unit

  drw:
    title: Drawing
    description: Чертеж
    sequence_mode: numeric
    template: templates/object/drw.md
    lifecycle: document_lifecycle
    required_metadata:
      - drawing_for
      - format

  cut:
    title: Cutting File
    description: Файл резки
    sequence_mode: inherit_or_numeric
    template: templates/object/cut.md
    lifecycle: manufacturing_file_lifecycle
    required_metadata:
      - source_part
      - process

  bnd:
    title: Bending File
    description: Файл гибки
    sequence_mode: inherit_or_numeric
    template: templates/object/bnd.md
    lifecycle: manufacturing_file_lifecycle
    required_metadata:
      - source_part

  nc:
    title: NC Program
    description: Программа ЧПУ
    sequence_mode: inherit_or_numeric
    template: templates/object/nc.md
    lifecycle: manufacturing_file_lifecycle
    required_metadata:
      - source_part
      - machine

  ins:
    title: Inspection
    description: Контроль, КИМ, план проверки
    sequence_mode: inherit_or_numeric
    template: templates/object/ins.md
    lifecycle: document_lifecycle
    required_metadata:
      - source_object

  tpc:
    title: Tech Process Card
    description: Технологическая карта
    sequence_mode: inherit_or_numeric
    template: templates/object/tpc.md
    lifecycle: document_lifecycle
    required_metadata:
      - source_object

  std:
    title: Standard Part
    description: Стандартное или покупное изделие
    sequence_mode: semantic
    template: templates/object/std.md
    lifecycle: standard_part_lifecycle
    required_metadata:
      - std_class
      - std_id
      - make_buy
```

---

# 7. Metadata

## 7.1. Назначение

`metadata` хранит все инженерные, производственные, закупочные и классификационные атрибуты объекта.

---

## 7.2. Общие metadata-поля

```yaml
metadata:
  lifecycle_stage: development
  owner: local-user
  unit: pcs
  make_buy: make
  tags: []
  keywords: []
  description_short: ""
```

---

## 7.3. Инженерные атрибуты

```yaml
metadata:
  material: al5052
  material_grade: h19
  thickness_mm: 2.0
  mass_g: 120.5
  dimensions: 100x50x20
  tolerance_class: iso2768-m
  surface_treatment: anodize
  heat_treatment: t6
  finish_color: ral7035
```

---

## 7.4. Производственные атрибуты

```yaml
metadata:
  manufacturing_method: laser_cut
  secondary_processes:
    - bending
    - deburring
    - anodizing
  work_center: laser-01
  machine: trumpf-3030
  setup_time_min: 15
  cycle_time_min: 2.5
  nesting_pattern: n01
```

---

## 7.5. Закупочные атрибуты

```yaml
metadata:
  make_buy: buy
  supplier: acme-fasteners
  supplier_part_number: acme-m6x12-a2
  manufacturer: acme
  manufacturer_part_number: m6x12-a2
  lead_time_days: 14
  min_order_qty: 100
  unit_cost:
    amount: 0.08
    currency: usd
  approved_vendor_status: approved
```

---

## 7.6. Качество и compliance

```yaml
metadata:
  inspection_required: true
  critical_part: false
  safety_critical: false
  certificates_required:
    - material_certificate
    - coating_certificate
  standards:
    - iso-9001
    - as9100
```

---

# 8. Artifact

## 8.1. Назначение

`Artifact` — физический или сгенерированный файл, принадлежащий объекту.

Примеры:

```text
STEP
DXF
PDF
PNG
NC file
G-code
certificate
photo evidence
generated manifest
generated BOM
```

---

## 8.2. Artifact YAML

```yaml
artifacts:
  - id: art-0001
    kind: cad
    role: primary_step
    path: files/cad/a320-prt-0001-v1.0.step
    original_name: Bracket final.step
    checksum: sha256:...
    size_bytes: 2048123
    generated: false
    required: true
    status: present
    created_at: 2026-06-06T10:00:00Z
    created_by: local-user
```

---

## 8.3. Artifact kinds

```yaml
artifact_kinds:
  cad:
    extensions: [.step, .stp, .iges, .igs, .sldprt, .sldasm]
  drawing:
    extensions: [.pdf, .dxf, .dwg]
  manufacturing:
    extensions: [.dxf, .nc, .tap, .gcode]
  image:
    extensions: [.png, .jpg, .jpeg, .webp]
  certificate:
    extensions: [.pdf, .md]
  evidence:
    extensions: [.png, .jpg, .pdf, .txt]
  generated:
    extensions: [.yaml, .csv, .md, .pdf]
```

---

## 8.4. Placeholder artifacts

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

Readiness engine должен показывать blocker:

```text
Missing required drawing PDF
```

---

# 9. Relation

## 9.1. Назначение

`Relation` описывает связь между объектами.

Связи используются для:

```text
BOM
where-used
чертежей
производных файлов
инструкций
замен
релизов
изменений
контроля
закупок
```

---

## 9.2. Базовая структура relation

```yaml
relations:
  - type: contains
    to: a320-prt-0001-v1.0
    quantity: 2
    unit: pcs
    effectivity:
      from_version: "1.0"
      to_version: null
    metadata:
      position: "10"
      note: "Main bracket"
```

---

## 9.3. Relation fields

```yaml
type: relation type
to: target object id
quantity: optional quantity
unit: optional unit
effectivity: optional applicability
metadata: additional relation metadata
```

---

## 9.4. Stored and projected relations

Хранить нужно только авторскую связь.

Например:

```yaml
# В сборке хранится:
relations:
  - type: contains
    to: a320-prt-0001-v1.0
    quantity: 2
    unit: pcs
```

Обратная связь `used_in` строится индексом.

```text
prt used_in asm
```

Правило:

```text
храним один источник связи;
обратные и агрегированные связи строим как projections.
```

---

# 10. Relation types

## 10.1. Структурные связи

```yaml
contains:
  description: Assembly contains child object
  from: [asm]
  to: [prt, asm, std]
  fields:
    - quantity
    - unit
  stored: true
  inverse: used_in

used_in:
  description: Object is used in parent assembly
  stored: false
  projected_from: contains
```

---

## 10.2. Документные связи

```yaml
has_drawing:
  from: [prt, asm, std]
  to: [drw]
  stored: true
  inverse: drawing_of

drawing_of:
  from: [drw]
  to: [prt, asm, std]
  stored: true_or_projected

has_document:
  from: [prt, asm, std, rel]
  to: [doc, spec, cert, rep]
  stored: true
```

---

## 10.3. Производственные связи

```yaml
has_manufacturing_file:
  from: [prt, asm]
  to: [cut, bnd, nc, wld, 3dp]
  stored: false
  projected_from:
    - manufacturing_file_for

manufacturing_file_for:
  from: [cut, bnd, nc, wld, 3dp]
  to: [prt, asm]
  stored: true

process_plan_for:
  from: [tpc]
  to: [prt, asm]
  stored: true

work_instruction_for:
  from: [doc, tpc]
  to: [prt, asm, cut, bnd, nc, ins]
  stored: true

inspection_for:
  from: [ins]
  to: [prt, asm]
  stored: true
```

---

## 10.4. Изменения и релизы

```yaml
affects:
  from: [doc]
  to: [prt, asm, drw, cut, bnd, nc, ins, tpc, bom]
  stored: true

includes:
  from: [rel, bom, spec]
  to: [prt, asm, drw, cut, bnd, nc, ins, tpc, doc, cert]
  stored: true

released_in:
  from: [prt, asm, drw, cut, bnd, nc, ins, tpc]
  to: [rel]
  stored: false
  projected_from: includes
```

---

## 10.5. Замены и альтернативы

```yaml
substitutes:
  from: [std, prt]
  to: [std, prt]
  stored: true

equivalent_to:
  from: [std, prt]
  to: [std, prt]
  stored: true

replaces:
  from: [prt, std]
  to: [prt, std]
  stored: true

alternate_for:
  from: [std, prt]
  to: [std, prt]
  stored: true
```

---

# 11. BOM model

## 11.1. BOM как relation graph

BOM не должен быть отдельной ручной таблицей. BOM строится из `contains` relations.

Сборка:

```yaml
id: a320-asm-0100-v1.0
class: asm
relations:
  - type: contains
    to: a320-prt-0001-v1.0
    quantity: 2
    unit: pcs
  - type: contains
    to: a320-std-fst-iso4017-m6x30-v1.0
    quantity: 8
    unit: pcs
```

---

## 11.2. eBOM

Engineering BOM — конструкторский состав.

Включает:

```text
детали
сборки
стандартные детали
конструкторские документы
```

Не обязан включать:

```text
операции
оснастку
расходники
упаковку
производственные инструкции
```

---

## 11.3. mBOM

Manufacturing BOM — производственный состав.

Строится из:

```text
eBOM
manufacturing files
process plans
work instructions
inspection objects
tools
consumables
packaging
```

Пример mBOM projection:

```yaml
mbom:
  root: a320-asm-0100-v1.0
  rows:
    - type: component
      object_id: a320-prt-0001-v1.0
      quantity: 2
      unit: pcs
    - type: manufacturing_file
      object_id: a320-cut-0001-v1.0
      source_object: a320-prt-0001-v1.0
    - type: inspection
      object_id: a320-ins-0001-v1.0
      source_object: a320-prt-0001-v1.0
```

---

## 11.4. Released BOM

Released BOM — замороженная спецификация релиза.

Создается как generated artifact и/или formal object:

```text
a320-bom-0100-v1.0
```

Содержит только approved/released объекты.

---

## 11.5. BOM row model

```yaml
bom_row:
  row_id: bomrow-0001
  parent_id: a320-asm-0100-v1.0
  child_id: a320-prt-0001-v1.0
  child_class: prt
  quantity: 2
  unit: pcs
  position: "10"
  effectivity:
    from: a320-asm-0100-v1.0
    until: null
  make_buy: make
  status: valid
```

---

# 12. StandardPart

## 12.1. Назначение

`StandardPart` — стандартное, покупное, каталоговое или повторно используемое изделие.

Примеры:

```text
болты
гайки
шайбы
подшипники
резисторы
разъемы
двигатели
покупные датчики
профили
типовые материалы
типовые кабели
```

Стандартная деталь должна вести себя как полноценный PLM-объект, но иметь особую модель идентификации, поиска, дублей, замен и закупки.

---

## 12.2. Почему стандартные детали — отдельная модель

Обычная деталь `prt` создается как уникальная проектная разработка:

```text
a320-prt-0001-v1.0
```

Стандартная деталь создается по смысловому идентификатору:

```text
a320-std-fst-iso4017-m6x30-a2-v1.0
```

Она часто:

```text
используется в разных сборках;
имеет поставщиков;
имеет manufacturer part number;
имеет аналоги;
может быть заменена;
может быть глобальной, а не проектной;
может не иметь CAD-файла;
может иметь статус released сразу после добавления в библиотеку.
```

---

## 12.3. Формат ID стандартной детали

Базовый формат:

```text
[project]-std-[std_class]-[std_id]-v[major].[minor]
```

Примеры:

```text
a320-std-fst-iso4017-m6x30-a2-v1.0
a320-std-brg-6205-2rs-c3-v1.0
a320-std-elc-res-10k-0603-1pct-v1.0
a320-std-elc-conn-jst-xh-4p-v1.0
```

Для глобальной библиотеки можно использовать проектный код:

```text
std-std-fst-iso4017-m6x30-a2-v1.0
lib-std-brg-6205-2rs-c3-v1.0
```

Рекомендуемый вариант:

```text
std-[std_class]-[std_id]-v1.0
```

Но если naming engine требует project segment, использовать:

```text
global-std-fst-iso4017-m6x30-a2-v1.0
```

---

## 12.4. std_class

`std_class` классифицирует тип стандартного изделия.

```yaml
std_classes:
  fst:
    title: Fastener
    examples:
      - bolt
      - nut
      - washer
      - screw

  brg:
    title: Bearing
    examples:
      - ball bearing
      - roller bearing
      - linear bearing

  elc:
    title: Electrical Component
    examples:
      - resistor
      - capacitor
      - connector
      - led
      - fuse

  pne:
    title: Pneumatic Component
    examples:
      - fitting
      - valve
      - cylinder

  hyd:
    title: Hydraulic Component

  mat:
    title: Material

  cab:
    title: Cable

  mot:
    title: Motor

  sen:
    title: Sensor

  prof:
    title: Profile

  seal:
    title: Seal
```

---

## 12.5. StandardPart YAML

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
  lifecycle_stage: production
  category: fastener
  family: bolt
  standard: iso4017
  size: m6x30
  thread: m6
  length_mm: 30
  material: stainless_steel
  material_grade: a2
  surface_treatment: none
  unit: pcs
  mass_g: null

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

quality:
  certificates_required:
    - certificate_of_conformity
  inspection_required: false
  safety_critical: false

relations:
  - type: equivalent_to
    to: a320-std-fst-din933-m6x30-a2-v1.0
  - type: substitutes
    to: a320-std-fst-iso4017-m6x30-a4-v1.0

artifacts:
  - id: art-datasheet
    kind: document
    role: datasheet
    path: files/documents/iso4017-m6x30-a2-datasheet.pdf
    checksum: sha256:...

created:
  at: 2026-06-06T10:00:00Z
  by: local-user

updated:
  at: 2026-06-06T10:00:00Z
  by: local-user
---

# Hex bolt ISO 4017 M6x30 A2

## Description

## Standard

## Procurement

## Substitutions

## Notes
```

---

# 13. Standard parts library

## 13.1. Назначение библиотеки

`Standard Parts Library` — специализированный каталог объектов `class: std`.

Он должен обеспечивать:

```text
создание стандартных деталей;
поиск по параметрам;
нормализацию имен;
выявление дублей;
управление поставщиками;
управление заменами;
использование в BOM;
экспорт в ERP;
контроль approved vendor list;
массовый импорт.
```

---

## 13.2. Размещение стандартных деталей

Вариант A: внутри проекта

```text
objects/a320-std-fst-iso4017-m6x30-a2-v1.0/
```

Вариант B: общая библиотека внутри workspace

```text
libraries/standard-parts/
├── fasteners/
├── bearings/
├── electrical/
└── materials/
```

Вариант C: отдельный Git submodule

```text
libs/standard-parts/
```

Рекомендуемая архитектура:

```text
Project can reference local standard parts and shared library standard parts.
Shared standard library is read-only by default.
Project can pin exact standard part version.
```

---

## 13.3. Локальная и глобальная стандартная деталь

Локальная:

```yaml
scope: project
project: a320
id: a320-std-fst-iso4017-m6x30-a2-v1.0
```

Глобальная:

```yaml
scope: global
project: global
id: global-std-fst-iso4017-m6x30-a2-v1.0
```

Если проект использует глобальную деталь, связь в BOM может указывать на global ID.

---

## 13.4. Pinning standard parts

BOM должен ссылаться на конкретную версию стандартной детали:

```yaml
relations:
  - type: contains
    to: global-std-fst-iso4017-m6x30-a2-v1.0
    quantity: 8
    unit: pcs
```

Нельзя неявно использовать “последнюю версию”, если объект released.

Для draft-сборок можно разрешить floating reference:

```yaml
relations:
  - type: contains
    to: global-std-fst-iso4017-m6x30-a2
    version_policy: latest_approved
    quantity: 8
    unit: pcs
```

При release такая ссылка должна быть зафиксирована:

```yaml
to: global-std-fst-iso4017-m6x30-a2-v1.0
version_policy: pinned
```

---

# 14. Standard part normalized key

## 14.1. Назначение

Для поиска дублей нужна нормализованная строка.

Пример:

```text
ISO 4017 M6 x 30 A2
iso4017-m6x30-a2
ISO-4017_M6X30_A2
```

должны давать один ключ:

```text
fst:iso4017:m6:30:a2
```

---

## 14.2. normalized_key

```yaml
normalized:
  key: fst:iso4017:m6:30:a2
  tokens:
    std_class: fst
    standard: iso4017
    thread: m6
    length_mm: 30
    material_grade: a2
```

---

## 14.3. Правила нормализации fasteners

Для крепежа:

```text
lowercase
удалить пробелы и лишние дефисы
нормализовать стандарт: iso4017, din933, gost7805
нормализовать резьбу: m6, m8, m10
нормализовать длину: 30
нормализовать материал: a2, a4, 8.8, 10.9
```

Пример:

```yaml
input:
  standard: ISO 4017
  thread: M6
  length: 30 mm
  material_grade: A2

normalized:
  key: fst:iso4017:m6:30:a2
```

---

## 14.4. Правила нормализации bearings

```yaml
normalized:
  key: brg:6205:2rs:c3
```

Поля:

```text
bearing_series
seal_type
clearance
precision_class
```

---

## 14.5. Правила нормализации electronics

```yaml
normalized:
  key: elc:res:10k:0603:1pct:0.1w
```

Поля:

```text
component_type
value
package
tolerance
power
voltage
temperature_range
```

---

# 15. Duplicate detection for standard parts

## 15.1. Уровни дублей

Система должна различать:

```text
exact duplicate
probable duplicate
possible equivalent
different but substitutable
not duplicate
```

---

## 15.2. Exact duplicate

Совпадает:

```text
std_class
normalized_key
manufacturer_part_number
```

Действие:

```text
запретить создание;
предложить открыть существующую деталь.
```

---

## 15.3. Probable duplicate

Совпадает:

```text
std_class
normalized_key
```

но отличается поставщик или MPN.

Действие:

```text
предложить добавить поставщика к существующей детали;
или создать supplier offer;
не создавать новый std object без подтверждения.
```

---

## 15.4. Possible equivalent

Разные стандарты, но технически совместимые.

Пример:

```text
ISO 4017 M6x30 A2
DIN 933 M6x30 A2
```

Действие:

```text
создать relation equivalent_to;
не объединять автоматически.
```

---

## 15.5. Different but substitutable

Пример:

```text
A2 stainless bolt
A4 stainless bolt
```

Может заменить при разрешении инженера.

Действие:

```text
создать relation substitutes с условиями.
```

---

## 15.6. Duplicate score

```yaml
duplicate_score:
  score: 0.92
  level: probable_duplicate
  reasons:
    - same std_class
    - same normalized dimensions
    - same standard
    - different supplier
```

---

# 16. Standard part supplier model

## 16.1. Procurement block

Стандартная деталь может иметь несколько поставщиков.

```yaml
procurement:
  preferred_supplier: supplier-acme
  approved_suppliers:
    - supplier_id: supplier-acme
      supplier_part_number: acme-123456
      manufacturer: acme
      manufacturer_part_number: iso4017-m6x30-a2
      lead_time_days: 7
      min_order_qty: 100
      package_qty: 100
      unit_cost:
        amount: 0.08
        currency: usd
      status: approved
      last_checked_at: 2026-06-06

    - supplier_id: supplier-boltmarket
      supplier_part_number: bm-778899
      manufacturer: generic
      manufacturer_part_number: m6x30-a2
      lead_time_days: 3
      min_order_qty: 50
      unit_cost:
        amount: 0.11
        currency: usd
      status: conditional
```

---

## 16.2. Supplier status

```yaml
supplier_statuses:
  approved:
    description: Approved for production
  conditional:
    description: Allowed with engineering/procurement approval
  blocked:
    description: Not allowed
  obsolete:
    description: Supplier no longer used
  candidate:
    description: Candidate, not yet approved
```

---

# 17. Standard part lifecycle

## 17.1. Lifecycle states

```text
candidate
qualified
approved
released
blocked
obsolete
archived
```

---

## 17.2. Lifecycle YAML

```yaml
processes:
  standard_part_lifecycle:
    initial: candidate
    states:
      - candidate
      - qualified
      - approved
      - released
      - blocked
      - obsolete
      - archived

    transitions:
      qualify:
        from: candidate
        to: qualified
        guards:
          - required_std_metadata
          - normalized_key_exists
          - no_exact_duplicate

      approve:
        from: qualified
        to: approved
        guards:
          - procurement_data_complete
          - supplier_approved_or_not_required

      release:
        from: approved
        to: released
        guards:
          - no_blockers

      block:
        from: "*"
        to: blocked
        requires_reason: true

      obsolete:
        from: released
        to: obsolete
        requires_reason: true

      archive:
        from: obsolete
        to: archived
```

---

## 17.3. Правила использования в BOM

В released BOM можно использовать только standard parts в состояниях:

```text
approved
released
```

Draft BOM может использовать:

```text
candidate
qualified
approved
released
```

но release readiness должен блокировать candidate/qualified, если они не утверждены.

---

# 18. Substitution model

## 18.1. Назначение

`Substitution` описывает, чем можно заменить деталь или стандартную деталь.

---

## 18.2. Substitution relation

```yaml
relations:
  - type: substitutes
    to: a320-std-fst-iso4017-m6x30-a4-v1.0
    metadata:
      substitution_type: upgrade
      allowed_for:
        - prototype
        - production
      requires_approval: true
      reason: A4 stainless can replace A2 in corrosive environment
```

---

## 18.3. Типы замен

```yaml
substitution_types:
  exact:
    description: Полностью взаимозаменяемая деталь
  equivalent:
    description: Технически эквивалентная по стандарту
  upgrade:
    description: Улучшенная замена
  temporary:
    description: Временная замена
  emergency:
    description: Аварийная замена
  conditional:
    description: Замена с условиями
```

---

## 18.4. Substitution approval

Для production/released объектов замена должна проходить workflow:

```text
proposed → engineering_review → procurement_review → approved → applied
```

YAML:

```yaml
substitution_request:
  id: a320-doc-substitution-0001-v1.0
  source: a320-std-fst-iso4017-m6x30-a2-v1.0
  target: a320-std-fst-iso4017-m6x30-a4-v1.0
  reason: supplier shortage
  state: proposed
  affected_boms:
    - a320-asm-0100-v1.0
```

---

# 19. Standard part usage in BOM

## 19.1. BOM row with standard part

```yaml
relations:
  - type: contains
    to: a320-std-fst-iso4017-m6x30-a2-v1.0
    quantity: 8
    unit: pcs
    metadata:
      position: "80"
      usage_note: cover mounting screws
      allow_substitutes: true
```

---

## 19.2. Quantity rollup

Если стандартная деталь используется в разных подсборках:

```text
asm root
 ├─ subasm A → bolt qty 4
 └─ subasm B → bolt qty 6
```

Flat BOM:

```yaml
rows:
  - object_id: a320-std-fst-iso4017-m6x30-a2-v1.0
    total_quantity: 10
    unit: pcs
    sources:
      - parent: subasm-a
        quantity: 4
      - parent: subasm-b
        quantity: 6
```

---

## 19.3. Procurement rollup

Buy list groups standard parts by:

```text
object_id
supplier
manufacturer_part_number
unit
package_qty
```

Example:

```yaml
buy_list:
  - object_id: a320-std-fst-iso4017-m6x30-a2-v1.0
    title: Hex bolt ISO 4017 M6x30 A2
    required_qty: 124
    unit: pcs
    package_qty: 100
    order_qty: 200
    supplier: supplier-acme
    supplier_part_number: acme-123456
```

---

# 20. Material as standard object

## 20.1. Material can be standard part

Материал можно моделировать как `std` или `mat`.

Вариант A:

```text
a320-std-mat-al5052-h19-2mm-v1.0
```

Вариант B:

```text
a320-mat-al5052-h19-2mm-v1.0
```

Рекомендуется:

```text
для каталожного материала — class mat;
для покупного стандартного компонента — class std.
```

---

## 20.2. Material YAML

```yaml
id: a320-mat-al5052-h19-2mm-v1.0
class: mat
title: Aluminum 5052-H19 sheet 2.0 mm

metadata:
  material: al5052
  material_grade: h19
  form: sheet
  thickness_mm: 2.0
  density_g_cm3: 2.68
  unit: m2
  make_buy: buy

procurement:
  preferred_supplier: supplier-metal
  lead_time_days: 10
```

Part can reference material:

```yaml
relations:
  - type: uses_material
    to: a320-mat-al5052-h19-2mm-v1.0
```

or by metadata only:

```yaml
metadata:
  material: al5052
  thickness_mm: 2.0
```

For strict traceability, relation is better.

---

# 21. Revision model

## 21.1. Version vs revision

The system uses:

```text
version in ID: v1.0, v1.1, v2.0
revision in title block/frontmatter: "1.0", "1.1", "2.0"
```

For this project, version and revision may initially be equal, but the model keeps them separate.

---

## 21.2. Revision link

```yaml
revision_info:
  previous_version: a320-prt-0001-v1.0
  next_versions:
    - a320-prt-0001-v1.1
  revision_reason: manufacturing improvement
  compatibility: backward_compatible
```

---

## 21.3. Revision compatibility

```yaml
compatibility:
  type: backward_compatible
  allowed_values:
    - backward_compatible
    - breaking_change
    - documentation_only
    - manufacturing_only
```

If `breaking_change`, parent assemblies are not auto-updated.

---

# 22. LifecycleState

## 22.1. Object lifecycle

```text
draft
in_review
approved
released
blocked
obsolete
archived
```

---

## 22.2. Document lifecycle

```text
draft
review
approved
released
obsolete
archived
```

---

## 22.3. Manufacturing lifecycle

```text
draft
simulation
trial
approved
released
blocked
obsolete
```

---

## 22.4. Procurement lifecycle

```text
candidate
qualified
approved
released
blocked
obsolete
```

---

# 23. Event model

## 23.1. Назначение

`Event` фиксирует историю изменений.

Каждый объект имеет:

```text
history.yaml.log
```

---

## 23.2. Event YAML

```yaml
---
event_id: evt-20260606-000001
time: 2026-06-06T10:00:00Z
actor: local-user
type: object.created
object_id: a320-prt-0001-v1.0
payload:
  class: prt
  title: Bracket Motor
---
event_id: evt-20260606-000002
time: 2026-06-06T10:05:00Z
actor: local-user
type: relation.added
object_id: a320-asm-0100-v1.0
payload:
  relation:
    type: contains
    to: a320-prt-0001-v1.0
    quantity: 2
    unit: pcs
```

---

## 23.3. Event types

```yaml
event_types:
  - object.created
  - object.updated
  - object.deleted
  - metadata.updated
  - document.updated
  - relation.added
  - relation.removed
  - relation.updated
  - artifact.attached
  - artifact.removed
  - artifact.checksum_updated
  - lifecycle.transitioned
  - revision.created
  - release.created
  - release.exported
  - substitution.proposed
  - substitution.approved
  - validation.run
```

---

# 24. ValidationDiagnostic

## 24.1. Назначение

Диагностика описывает ошибку, предупреждение или информационную проверку.

---

## 24.2. Diagnostic YAML

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

---

## 24.3. Severity

```text
info
warning
blocker
error
```

---

# 25. ReleasePackage

## 25.1. Назначение

`ReleasePackage` фиксирует состав релиза.

---

## 25.2. Release YAML

```yaml
id: a320-rel-0100-v1.0
class: rel
title: Release for Main Assembly
state: draft

metadata:
  root_object: a320-asm-0100-v1.0
  release_type: production
  release_date: null

relations:
  - type: includes
    to: a320-asm-0100-v1.0
  - type: includes
    to: a320-prt-0001-v1.0
  - type: includes
    to: a320-std-fst-iso4017-m6x30-a2-v1.0
```

---

## 25.3. Release manifest

```yaml
release:
  id: a320-rel-0100-v1.0
  root_object: a320-asm-0100-v1.0
  created_at: 2026-06-06T10:00:00Z
  git_commit: abc123
  git_tag: release/a320-rel-0100-v1.0

objects:
  - id: a320-asm-0100-v1.0
    class: asm
    state: released
  - id: a320-std-fst-iso4017-m6x30-a2-v1.0
    class: std
    state: released

files:
  - object_id: a320-prt-0001-v1.0
    path: files/cad/a320-prt-0001-v1.0.step
    checksum: sha256:...
```

---

# 26. ChangeRequest

## 26.1. Назначение

`ChangeRequest` описывает изменение released-объектов.

---

## 26.2. ChangeRequest YAML

```yaml
id: a320-doc-change-0001-v1.0
class: doc
title: Change Request: Replace Bracket Material
state: proposed

metadata:
  change_type: engineering_change
  priority: normal
  reason: supplier shortage
  requested_by: local-user

relations:
  - type: affects
    to: a320-prt-0001-v1.0
  - type: affects
    to: a320-asm-0100-v1.0
```

---

# 27. BuildRecord

## 27.1. Назначение

`BuildRecord` фиксирует факт производства или сборки.

---

## 27.2. BuildRecord YAML

```yaml
id: a320-build-0001-v1.0
class: build
title: Build Main Assembly SN-0001
state: planned

metadata:
  source_object: a320-asm-0100-v1.0
  serial_number: sn-0001
  operator: null
  started_at: null
  completed_at: null

relations:
  - type: build_of
    to: a320-asm-0100-v1.0
  - type: uses_instruction
    to: a320-doc-wi-0100-v1.0
```

---

# 28. IndexProjection

## 28.1. Назначение

Индексные проекции обеспечивают быстрый UI и запросы.

---

## 28.2. Основные проекции

```text
objects
relations
artifacts
metadata
search
tree
bom
where_used
readiness
release
standard_parts
duplicates
procurement
change_impact
```

---

## 28.3. Правило восстановления

Индекс должен быть полностью восстановим из:

```text
project.yaml
objects/*/*.md
objects/*/history.yaml.log
config/*.yaml
```

---

# 29. SQLite index conceptual schema

## 29.1. objects

```sql
objects(
  id TEXT PRIMARY KEY,
  project TEXT,
  class TEXT,
  sequence TEXT,
  version TEXT,
  state TEXT,
  title TEXT,
  path TEXT,
  updated_at TEXT
)
```

---

## 29.2. relations

```sql
relations(
  id TEXT PRIMARY KEY,
  from_id TEXT,
  type TEXT,
  to_id TEXT,
  quantity REAL,
  unit TEXT,
  metadata_yaml TEXT
)
```

---

## 29.3. artifacts

```sql
artifacts(
  id TEXT PRIMARY KEY,
  object_id TEXT,
  kind TEXT,
  role TEXT,
  path TEXT,
  checksum TEXT,
  status TEXT
)
```

---

## 29.4. standard_parts

```sql
standard_parts(
  id TEXT PRIMARY KEY,
  std_class TEXT,
  std_id TEXT,
  normalized_key TEXT,
  manufacturer TEXT,
  manufacturer_part_number TEXT,
  preferred_supplier TEXT,
  state TEXT
)
```

---

## 29.5. procurement_items

```sql
procurement_items(
  object_id TEXT PRIMARY KEY,
  make_buy TEXT,
  supplier TEXT,
  supplier_part_number TEXT,
  manufacturer TEXT,
  manufacturer_part_number TEXT,
  lead_time_days INTEGER,
  unit_cost REAL,
  currency TEXT
)
```

---

## 29.6. search

```sql
search_index(
  object_id TEXT,
  text TEXT
)
```

---

# 30. Information model commands

## 30.1. Object commands

```text
CreateObject
UpdateObjectMetadata
UpdateObjectDocument
DeleteObject
ArchiveObject
ReviseObject
ValidateObject
```

---

## 30.2. Relation commands

```text
AddRelation
UpdateRelation
RemoveRelation
ReplaceRelationTarget
GetWhereUsed
GetRelationGraph
```

---

## 30.3. Artifact commands

```text
AttachArtifact
RemoveArtifact
UpdateArtifactRole
RecalculateChecksum
CreatePlaceholderArtifact
```

---

## 30.4. Standard part commands

```text
CreateStandardPart
ImportStandardParts
NormalizeStandardPart
FindStandardPartDuplicates
MergeStandardParts
AddSupplierOffer
ApproveSupplier
ProposeSubstitution
ApproveSubstitution
UseStandardPartInBOM
```

---

# 31. Information model queries

```text
GetObject
ListObjects
SearchObjects
GetObjectRelations
GetBOM
GetFlatBOM
GetWhereUsed
GetChangeImpact
GetStandardPart
SearchStandardParts
FindDuplicateStandardParts
GetSubstitutions
GetProcurementRollup
GetReleaseReadiness
```

---

# 32. Standard parts workflow

## 32.1. Create standard part

```text
User opens Standard Parts Center
→ clicks New Standard Part
→ selects std_class
→ fills parameters
→ backend normalizes key
→ backend checks duplicates
→ if no duplicate, creates std object
→ state = candidate or released depending policy
```

---

## 32.2. Add standard part to BOM

```text
Open Assembly
→ BOM tab
→ Add Item
→ Search Standard Parts
→ select standard part
→ set quantity
→ backend adds contains relation
→ BOM projection updates
→ procurement rollup updates
```

---

## 32.3. Duplicate detected

```text
User tries to create ISO 4017 M6x30 A2
→ backend computes normalized key
→ existing part found
→ UI offers:
   Open existing
   Add supplier to existing
   Create variant
   Cancel
```

---

## 32.4. Substitute standard part

```text
User opens BOM row
→ clicks Propose Substitute
→ selects alternative
→ backend creates substitution request
→ engineering approves
→ procurement approves
→ relation substitutes is added
→ BOM row can use replacement
```

---

## 32.5. Obsolete standard part

```text
Standard part marked obsolete
→ backend finds where-used
→ affected assemblies listed
→ replacement suggestions shown
→ user creates change request
```

---

# 33. Model invariants for standard parts

Standard part is valid if:

```text
id follows std naming;
std_class exists;
std_id exists;
normalized_key is generated;
make_buy = buy by default;
unit is set;
no exact duplicate exists;
procurement block exists if required by project policy;
supplier status is valid;
substitution relations point to existing std/prt objects;
released standard part has approved or released state.
```

---

# 34. ERP integration mapping

## 34.1. PLM object to ERP item

```yaml
erp_item:
  item_number: a320-std-fst-iso4017-m6x30-a2-v1.0
  description: Hex bolt ISO 4017 M6x30 A2
  item_type: purchased
  unit: pcs
  make_buy: buy
  manufacturer_part_number: acme-iso4017-m6x30-a2
  supplier_part_number: acme-123456
```

---

## 34.2. BOM to ERP

```yaml
erp_bom:
  parent: a320-asm-0100-v1.0
  rows:
    - item_number: a320-std-fst-iso4017-m6x30-a2-v1.0
      quantity: 8
      unit: pcs
```

---

## 34.3. Standard part procurement export

```yaml
procurement_export:
  - item_number: a320-std-fst-iso4017-m6x30-a2-v1.0
    description: Hex bolt ISO 4017 M6x30 A2
    supplier: supplier-acme
    supplier_part_number: acme-123456
    required_qty: 124
    order_qty: 200
    unit_cost: 0.08
    currency: usd
```

---

# 35. Summary

Информационная модель Markdown-PLM строится вокруг нескольких устойчивых идей:

```text
Object — основная инженерная сущность.
Relation — основа BOM, where-used, change impact и traceability.
Artifact — файл, принадлежащий объекту.
Metadata — все параметры, которые нельзя кодировать в имени.
Event — история изменений.
Projection — быстрый индекс, пересобираемый из YAML.
StandardPart — особый объект для стандартных и покупных компонентов.
```

Стандартные детали должны быть полноценными PLM-объектами, а не строками в таблице. Они имеют собственный ID, normalized key, lifecycle, procurement block, suppliers, substitutions, duplicate detection и usage tracking.

Главное правило:

```text
Не файл управляет системой.
Не папка управляет системой.
Не имя содержит смысл.

Системой управляют:
ID + YAML metadata + Relations + Lifecycle + Events.
```