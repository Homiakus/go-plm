# Autogeneration Specification

# Markdown-PLM / go-plm

## Логика создания, автозаполнения файлов и связей между объектами

Версия: `0.1`  
Статус: `draft`  
Назначение: описание механизма автоматического создания объектов, файлов, YAML frontmatter, папок, связей, BOM, производных документов и индексов.

---

# 1. Главный принцип

Система должна автоматически создавать не “файл”, а полноценный инженерный объект.

Объект в Markdown-PLM — это:

```text
Object = ID + YAML frontmatter + Markdown body + files + relations + history + index projection
```

Пользователь нажимает:

```text
+ New → Part
```

А backend должен создать:

```text
objects/a320-prt-0001-v1.0/
├── a320-prt-0001-v1.0.md
├── history.yaml.log
├── files/
│   ├── cad/
│   ├── drawings/
│   ├── images/
│   ├── certificates/
│   └── evidence/
└── generated/
    ├── bom/
    ├── reports/
    └── exports/
```

Все канонические данные объекта должны храниться в Markdown/YAML. Индекс SQLite используется только для скорости и может быть пересобран.

---

# 2. Уровни данных

Система должна различать 4 уровня данных.

## 2.1. Canonical source of truth

Источник истины:

```text
project.yaml
config/*.yaml
objects/*/*.md
objects/*/history.yaml.log
templates/*.md
```

Именно эти файлы должны быть достаточны для восстановления проекта.

## 2.2. Physical artifacts

Физические файлы:

```text
STEP
DXF
PDF
PNG/JPG
NC programs
cutting files
bending files
inspection files
certificates
```

Они не должны хранить главные метаданные в имени. Имя должно быть минимальным и стабильным.

## 2.3. Generated files

Сгенерированные файлы:

```text
generated/bom/*.yaml
generated/reports/*.md
generated/exports/*.csv
generated/release/*.yaml
```

Они могут быть пересозданы из source of truth.

## 2.4. Cache / index

Пересобираемые данные:

```text
.plm/index.db
.plm/cache/*
.plm/tasks/*
```

Они не являются источником истины.

---

# 3. Naming engine

## 3.1. Базовая формула имени

Все основные объекты должны генерироваться по формату:

```text
[project]-[class]-[sequence]-v[major].[minor]
```

Пример:

```text
a320-prt-0001-v1.0
a320-asm-0100-v1.0
a320-drw-0001-v1.0
a320-cut-0001-v1.0
```

Расширение добавляется только для физических файлов:

```text
a320-prt-0001-v1.0.md
a320-prt-0001-v1.0.step
a320-drw-0001-v1.0.pdf
```

## 3.2. Правило минимального имени

Запрещено кодировать в имени:

```text
материал
толщину
покрытие
метод производства
поставщика
цвет
автора
дату
статус жизненного цикла
```

Неправильно:

```text
a320-prt-0001-v1.0-al5052-2mm-laser-anodized.step
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
```

## 3.3. Регистр

По умолчанию система должна генерировать имена в `kebab-case`:

```text
a320-prt-0001-v1.0
```

Допускается отображение в UI в верхнем регистре, но физическое имя и ID должны быть lowercase.

## 3.4. Генерация sequence

Для каждого проекта и класса backend должен хранить счетчик последовательностей.

Пример:

```yaml
sequences:
  prt: 0007
  asm: 0100
  drw: 0004
  cut: 0002
  bnd: 0001
  nc: 0003
```

При создании новой детали:

```text
project = a320
class = prt
last sequence = 0007
next sequence = 0008
id = a320-prt-0008-v1.0
```

Счетчик должен восстанавливаться из существующих объектов, если `project.yaml` поврежден или отсутствует.

## 3.5. Правила sequence по классам

Рекомендуемые диапазоны:

```yaml
sequence_policy:
  prt:
    start: 0001
    width: 4
  asm:
    start: 0100
    width: 4
  drw:
    start: 0001
    width: 4
  cut:
    start: 0001
    width: 4
  bnd:
    start: 0001
    width: 4
  nc:
    start: 0001
    width: 4
  ins:
    start: 0001
    width: 4
  std:
    mode: semantic
```

Для `std` используется отдельная логика:

```text
[project]-std-[std_class]-[std_id]-v[major].[minor]
```

Пример:

```text
a320-std-fst-gost7805-m6x12-v1.0
a320-std-brg-6205-2rs-c3-v1.0
```

---

# 4. Object creation pipeline

Создание любого объекта должно проходить через единый pipeline.

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

## 4.1. CreateObjectDraft

Frontend передает backend черновик:

```yaml
class: prt
title: Bracket Motor
parent_object_id: a320-asm-0100-v1.0
metadata:
  material: al5052
  thickness_mm: 2.0
files:
  - local_path: C:\cad\bracket.step
    role: primary_cad
```

Черновик еще не создает файлы. Он нужен для preview.

## 4.2. GenerateIdentity

Backend генерирует:

```yaml
id: a320-prt-0008-v1.0
project: a320
class: prt
sequence: "0008"
version: "1.0"
revision: "1.0"
state: draft
```

Также формируются пути:

```yaml
object_dir: objects/a320-prt-0008-v1.0
markdown_path: objects/a320-prt-0008-v1.0/a320-prt-0008-v1.0.md
history_path: objects/a320-prt-0008-v1.0/history.yaml.log
```

## 4.3. ResolveTemplate

Backend выбирает шаблон по классу объекта.

Пример:

```text
class = prt
template = templates/object/part.md
```

Если шаблон не найден, используется built-in fallback template.

## 4.4. AutofillMetadata

Backend заполняет metadata из нескольких источников:

```text
1. project.yaml defaults
2. class schema defaults
3. selected template defaults
4. parent object context
5. user input
6. file-derived metadata
7. relation-derived metadata
```

Приоритет:

```text
user input > file-derived > parent context > template > class defaults > project defaults
```

## 4.5. BuildRelations

Backend автоматически создает связи на основе контекста.

Например, если пользователь создает деталь внутри сборки:

```yaml
relations:
  - type: contained_in
    to: a320-asm-0100-v1.0
  - type: contains_inverse
    from: a320-asm-0100-v1.0
    to: a320-prt-0008-v1.0
    quantity: 1
    unit: pcs
```

Физически в YAML объекта лучше хранить исходящие связи, а обратные связи строить через индекс.

## 4.6. PlanFileStructure

Backend строит план файловой транзакции:

```yaml
transaction:
  create_dirs:
    - objects/a320-prt-0008-v1.0
    - objects/a320-prt-0008-v1.0/files/cad
    - objects/a320-prt-0008-v1.0/files/drawings
    - objects/a320-prt-0008-v1.0/generated
  write_files:
    - path: objects/a320-prt-0008-v1.0/a320-prt-0008-v1.0.md
      content: rendered_template
    - path: objects/a320-prt-0008-v1.0/history.yaml.log
      content: initial_event
  copy_files:
    - from: C:\cad\bracket.step
      to: objects/a320-prt-0008-v1.0/files/cad/a320-prt-0008-v1.0.step
```

## 4.7. ValidateDraft

Перед записью backend должен проверить:

```text
ID валиден
ID уникален
папка не существует
имя соответствует standard v10
required metadata заполнены
relations валидны
parent object существует
file routing валиден
нет конфликта artifact role
```

## 4.8. ExecuteFileTransaction

Файлы создаются атомарно:

```text
1. Создать transaction directory в .plm/transactions/
2. Подготовить temp-файлы
3. Проверить checksums
4. Сделать atomic rename/copy
5. fsync
6. Записать commit marker транзакции
7. Очистить temp
```

Если операция прервана, recovery mode должен уметь завершить или откатить транзакцию.

## 4.9. AppendHistory

Каждое создание объекта записывает событие:

```yaml
---
event_id: evt-20260606-000001
time: 2026-06-06T10:00:00Z
actor: local-user
type: object.created
object_id: a320-prt-0008-v1.0
payload:
  class: prt
  title: Bracket Motor
  source: create_wizard
  template: templates/object/part.md
```

## 4.10. UpdateIndex

После записи backend обновляет индекс:

```text
objects table
relations table
artifacts table
metadata table
search index
tree projection
bom projection
where-used projection
validation projection
```

## 4.11. ReturnObjectSnapshot

Backend возвращает frontend готовый snapshot:

```yaml
object:
  id: a320-prt-0008-v1.0
  class: prt
  title: Bracket Motor
  state: draft
  path: objects/a320-prt-0008-v1.0/a320-prt-0008-v1.0.md
relations: []
artifacts:
  - role: primary_cad
    path: files/cad/a320-prt-0008-v1.0.step
diagnostics: []
refresh:
  tree: true
  object: a320-prt-0008-v1.0
  bom: a320-asm-0100-v1.0
```

---

# 5. YAML frontmatter structure

Каждый объект должен иметь frontmatter.

## 5.1. Базовая структура

```markdown
---
id: a320-prt-0008-v1.0
project: a320
class: prt
sequence: "0008"
version: "1.0"
revision: "1.0"
state: draft
title: Bracket Motor

metadata:
  lifecycle_stage: development
  owner: local-user
  unit: pcs

relations: []

artifacts: []

created:
  at: 2026-06-06T10:00:00Z
  by: local-user

updated:
  at: 2026-06-06T10:00:00Z
  by: local-user
---

# Bracket Motor
```

## 5.2. Почему metadata отдельно

Поля верхнего уровня используются ядром системы:

```yaml
id:
project:
class:
sequence:
version:
state:
title:
relations:
artifacts:
```

А инженерные параметры идут в `metadata`:

```yaml
metadata:
  material: al5052
  thickness_mm: 2.0
  surface_treatment: anodize
  manufacturing_method: laser_cut
```

Так проще отделять системные поля от предметных.

---

# 6. Template engine

## 6.1. Типы шаблонов

Система должна поддерживать шаблоны:

```text
project templates
object templates
document templates
artifact templates
process templates
release templates
manufacturing templates
procurement templates
```

Структура:

```text
templates/
├── project/
│   ├── simple-mechanical/
│   └── manufacturing/
├── object/
│   ├── prt.md
│   ├── asm.md
│   ├── drw.md
│   ├── cut.md
│   ├── bnd.md
│   ├── nc.md
│   ├── ins.md
│   ├── wi.md
│   ├── pp.md
│   └── std.md
├── generated/
│   ├── bom.yaml
│   ├── manifest.yaml
│   └── buy-list.yaml
└── snippets/
    ├── safety-warning.md
    ├── inspection-check.md
    └── title-block.md
```

## 6.2. Переменные шаблонов

Доступные переменные:

```yaml
object:
  id:
  project:
  class:
  sequence:
  version:
  title:
  state:

user:
  name:
  email:

time:
  now:
  date:

parent:
  id:
  class:
  title:

metadata:
  material:
  thickness_mm:
  unit:
```

## 6.3. Пример part template

```markdown
---
id: "{{ object.id }}"
project: "{{ object.project }}"
class: "{{ object.class }}"
sequence: "{{ object.sequence }}"
version: "{{ object.version }}"
revision: "{{ object.version }}"
state: draft
title: "{{ object.title }}"

metadata:
  lifecycle_stage: development
  owner: "{{ user.name }}"
  unit: pcs
  material: "{{ metadata.material }}"
  thickness_mm: {{ metadata.thickness_mm }}
  make_buy: make

relations:
{{ relations_yaml }}

artifacts:
{{ artifacts_yaml }}

created:
  at: "{{ time.now }}"
  by: "{{ user.name }}"

updated:
  at: "{{ time.now }}"
  by: "{{ user.name }}"
---

# {{ object.title }}

## Назначение

## Материал

## Основные размеры

## Производственные примечания

## Контроль качества

## Связанные документы
```

---

# 7. Автозаполнение по типам объектов

# 7.1. Part / prt

## Создаваемые файлы

```text
objects/a320-prt-0008-v1.0/
├── a320-prt-0008-v1.0.md
├── history.yaml.log
├── files/
│   ├── cad/
│   ├── drawings/
│   ├── manufacturing/
│   ├── inspection/
│   └── certificates/
└── generated/
```

## Автозаполняемые поля

```yaml
id: generated
project: from project.yaml
class: prt
sequence: next prt sequence
version: "1.0"
revision: "1.0"
state: draft
title: user input

metadata:
  lifecycle_stage: development
  owner: current user
  unit: pcs
  make_buy: make
  material: user input or empty
  thickness_mm: user input or null
  manufacturing_method: null
```

## Автоматические связи

Если создается из контекста assembly:

```yaml
relations:
  - type: used_in
    to: a320-asm-0100-v1.0
```

Если прикреплен CAD:

```yaml
artifacts:
  - role: primary_cad
    kind: cad
    path: files/cad/a320-prt-0008-v1.0.step
    checksum: sha256:...
```

Если прикреплен чертеж:

```yaml
relations:
  - type: has_drawing
    to: a320-drw-0008-v1.0
```

---

# 7.2. Assembly / asm

## Создаваемые файлы

```text
objects/a320-asm-0101-v1.0/
├── a320-asm-0101-v1.0.md
├── history.yaml.log
├── files/
│   ├── cad/
│   ├── drawings/
│   └── documents/
└── generated/
    ├── bom/
    └── release/
```

## Автозаполняемые поля

```yaml
metadata:
  lifecycle_stage: development
  owner: current user
  unit: pcs
  assembly_type: mechanical
```

## Автоматические связи

Если создается как подсборка внутри другой сборки:

```yaml
relations:
  - type: used_in
    to: parent_assembly_id
```

BOM хранится как связи `contains`:

```yaml
relations:
  - type: contains
    to: a320-prt-0001-v1.0
    quantity: 2
    unit: pcs
  - type: contains
    to: a320-prt-0002-v1.0
    quantity: 1
    unit: pcs
```

---

# 7.3. Drawing / drw

## Логика создания

Чертеж почти всегда создается из контекста детали или сборки.

Пользователь:

```text
Open Part → Files or Relations → Create Drawing
```

Backend:

```text
1. Берет sequence связанного объекта, если политика проекта требует same-sequence drawing
2. Генерирует ID чертежа
3. Создает Drawing object
4. Создает связь part → drawing
5. Создает связь drawing → part
```

## Варианты sequence

### Вариант A: свой sequence для drw

```text
a320-prt-0008-v1.0
a320-drw-0003-v1.0
```

### Вариант B: sequence наследуется от объекта

```text
a320-prt-0008-v1.0
a320-drw-0008-v1.0
```

Рекомендуется вариант B для простых проектов.

## YAML

```yaml
id: a320-drw-0008-v1.0
class: drw
title: Drawing for Bracket Motor
state: draft

metadata:
  drawing_for: a320-prt-0008-v1.0
  format: A3
  projection: first_angle
  title_block_number: a320-drw-0008-v1.0

relations:
  - type: drawing_of
    to: a320-prt-0008-v1.0
```

## Связь в исходном объекте

В детали автоматически добавляется:

```yaml
relations:
  - type: has_drawing
    to: a320-drw-0008-v1.0
```

---

# 7.4. Cutting file / cut

Создается из детали, если manufacturing method = laser_cut, waterjet или plasma.

## Trigger

```text
Part metadata.manufacturing_method = laser_cut
или
User → Generate Manufacturing Files → Cutting File
```

## Создаваемые файлы

```text
objects/a320-cut-0008-v1.0/
├── a320-cut-0008-v1.0.md
├── history.yaml.log
├── files/
│   └── cutting/
│       └── a320-cut-0008-v1.0.dxf
└── generated/
```

## Автозаполнение

```yaml
id: a320-cut-0008-v1.0
class: cut
title: Cutting file for Bracket Motor
state: draft

metadata:
  source_part: a320-prt-0008-v1.0
  process: laser_cut
  material: from source part
  thickness_mm: from source part
  nesting_pattern: null
  machine: null
  estimated_cycle_time_sec: null

relations:
  - type: manufacturing_file_for
    to: a320-prt-0008-v1.0
  - type: derived_from
    to: a320-prt-0008-v1.0
```

## Обратная связь в детали

```yaml
relations:
  - type: has_manufacturing_file
    to: a320-cut-0008-v1.0
```

---

# 7.5. Bending file / bnd

Создается для листовой детали, если есть гибка.

## Trigger

```text
metadata.manufacturing_method contains bending
или
User → Generate Manufacturing Files → Bending File
```

## YAML

```yaml
id: a320-bnd-0008-v1.0
class: bnd
title: Bending file for Bracket Motor
state: draft

metadata:
  source_part: a320-prt-0008-v1.0
  material: al5052
  thickness_mm: 2.0
  bending_machine: null
  tool: null
  bend_count: null
  bending_sequence: []

relations:
  - type: manufacturing_file_for
    to: a320-prt-0008-v1.0
  - type: derived_from
    to: a320-prt-0008-v1.0
```

---

# 7.6. NC Program / nc

Создается для деталей с мехобработкой.

## Trigger

```text
metadata.manufacturing_method = milling
metadata.manufacturing_method = turning
metadata.manufacturing_method = cnc
```

## YAML

```yaml
id: a320-nc-0008-v1.0
class: nc
title: NC program for Bracket Motor
state: draft

metadata:
  source_part: a320-prt-0008-v1.0
  machine: null
  controller: null
  postprocessor: null
  tool_list: []
  setup_time_min: null
  cycle_time_min: null

relations:
  - type: manufacturing_file_for
    to: a320-prt-0008-v1.0
  - type: derived_from
    to: a320-prt-0008-v1.0
```

---

# 7.7. Inspection / ins

Создается для контрольной программы или плана проверки.

## Trigger

```text
User → Create Inspection Plan
или
Release readiness требует inspection
```

## YAML

```yaml
id: a320-ins-0008-v1.0
class: ins
title: Inspection plan for Bracket Motor
state: draft

metadata:
  source_object: a320-prt-0008-v1.0
  inspection_type: dimensional
  equipment: null
  sampling_plan: null
  critical_dimensions: []

relations:
  - type: inspection_for
    to: a320-prt-0008-v1.0
```

---

# 7.8. Work Instruction / wi

Создается для операции или объекта.

## Trigger

```text
Manufacturing Matrix → Missing Work Instruction → Create
```

## YAML

```yaml
id: a320-wi-0008-v1.0
class: wi
title: Work instruction for Bracket Motor
state: draft

metadata:
  source_object: a320-prt-0008-v1.0
  operation: null
  safety_level: normal
  estimated_time_min: null

relations:
  - type: work_instruction_for
    to: a320-prt-0008-v1.0
```

---

# 7.9. Process Plan / tpc or pp

Если используется class `tpc`, он соответствует Tech Process Card.

## YAML

```yaml
id: a320-tpc-0008-v1.0
class: tpc
title: Process card for Bracket Motor
state: draft

metadata:
  source_object: a320-prt-0008-v1.0
  route:
    - op: 10
      name: cutting
      work_center: laser
      instruction: null
    - op: 20
      name: bending
      work_center: press_brake
      instruction: null
    - op: 30
      name: inspection
      work_center: qc
      instruction: null

relations:
  - type: process_plan_for
    to: a320-prt-0008-v1.0
```

---

# 7.10. BOM / bom and Specification / spec

BOM can be generated as a projection or as a formal object.

## Projection BOM

Для повседневной работы BOM строится из relations:

```yaml
relations:
  - type: contains
    to: a320-prt-0001-v1.0
    quantity: 2
    unit: pcs
```

## Formal BOM object

Для релиза создается объект:

```text
a320-bom-0100-v1.0
```

YAML:

```yaml
id: a320-bom-0100-v1.0
class: bom
title: BOM for Main Assembly
state: draft

metadata:
  source_assembly: a320-asm-0100-v1.0
  bom_type: ebom
  generated_from_index_at: 2026-06-06T10:00:00Z

relations:
  - type: bom_for
    to: a320-asm-0100-v1.0
  - type: includes
    to: a320-prt-0001-v1.0
    quantity: 2
    unit: pcs
```

---

# 8. File routing engine

## 8.1. Назначение

Когда пользователь прикрепляет файл, backend должен автоматически определить:

```text
kind
role
target folder
target filename
checksum
relations
metadata extraction
```

## 8.2. Пример file-routing.yaml

```yaml
routes:
  cad:
    extensions: [.step, .stp, .iges, .igs, .sldprt, .sldasm]
    target: files/cad
    kind: cad
    default_role: primary_cad

  drawings:
    extensions: [.pdf, .dxf, .dwg]
    target: files/drawings
    kind: drawing
    default_role: drawing_file

  cutting:
    extensions: [.dxf]
    target: files/manufacturing/cutting
    kind: manufacturing
    default_role: cutting_file
    when:
      class: cut

  nc:
    extensions: [.nc, .tap, .gcode]
    target: files/manufacturing/nc
    kind: manufacturing
    default_role: nc_program

  images:
    extensions: [.png, .jpg, .jpeg, .webp]
    target: files/images
    kind: image
    default_role: illustration

  certificates:
    extensions: [.pdf]
    target: files/certificates
    kind: certificate
    default_role: certificate
    when:
      class: cert
```

## 8.3. Правило переименования вложений

По умолчанию вложенный файл получает ID объекта:

```text
a320-prt-0008-v1.0.step
a320-prt-0008-v1.0.pdf
a320-cut-0008-v1.0.dxf
```

Если файлов одного типа несколько:

```text
a320-prt-0008-v1.0-primary.step
a320-prt-0008-v1.0-simplified.step
a320-prt-0008-v1.0-supplier.step
```

Но role хранится в YAML, а не только в имени.

## 8.4. Artifact record

```yaml
artifacts:
  - id: art-0001
    kind: cad
    role: primary_cad
    path: files/cad/a320-prt-0008-v1.0.step
    original_name: Bracket Motor final.step
    checksum: sha256:...
    size_bytes: 2048123
    created_at: 2026-06-06T10:00:00Z
    created_by: local-user
```

---

# 9. Автоматическое создание связей

# 9.1. Типы связей

Система должна поддерживать базовые relation types:

```yaml
relation_types:
  contains:
    description: Assembly contains child object
    allowed_from: [asm]
    allowed_to: [prt, asm, std]
    fields: [quantity, unit]

  used_in:
    description: Reverse projection for contains
    stored: false

  has_drawing:
    allowed_from: [prt, asm, std]
    allowed_to: [drw]

  drawing_of:
    allowed_from: [drw]
    allowed_to: [prt, asm, std]

  has_manufacturing_file:
    allowed_from: [prt, asm]
    allowed_to: [cut, bnd, nc, wld, 3dp]

  manufacturing_file_for:
    allowed_from: [cut, bnd, nc, wld, 3dp]
    allowed_to: [prt, asm]

  has_inspection:
    allowed_from: [prt, asm]
    allowed_to: [ins]

  inspection_for:
    allowed_from: [ins]
    allowed_to: [prt, asm]

  process_plan_for:
    allowed_from: [tpc]
    allowed_to: [prt, asm]

  work_instruction_for:
    allowed_from: [doc, tpc]
    allowed_to: [prt, asm, cut, bnd, nc, ins]

  derived_from:
    allowed_from: [cut, bnd, nc, ins, drw, rep]
    allowed_to: [prt, asm]

  substitutes:
    allowed_from: [std, prt]
    allowed_to: [std, prt]

  affects:
    allowed_from: [doc]
    allowed_to: [prt, asm, drw, tpc, cut, bnd, nc, ins]
```

## 9.2. Stored vs projected relations

Не все связи нужно дублировать в YAML.

Хранимые связи:

```text
asm contains prt
prt has_drawing drw
cut manufacturing_file_for prt
ins inspection_for prt
```

Проекционные связи:

```text
prt used_in asm
drw drawing_of inverse
prt has_manufacturing_file cut
```

Они строятся индексом.

Правило:

```text
Store the authoring relation once.
Build inverse relations through index.
```

---

# 10. Автосвязи при создании из контекста

## 10.1. Создание Part из Assembly

Сценарий:

```text
Open Assembly → BOM Tab → Add Item → Create New Part
```

Backend должен:

```text
1. Создать prt
2. Добавить contains relation в asm
3. Добавить used_in projection в index
4. Обновить BOM projection
5. Обновить tree projection
```

В assembly YAML:

```yaml
relations:
  - type: contains
    to: a320-prt-0008-v1.0
    quantity: 1
    unit: pcs
```

В part YAML можно не писать обратную связь. Она строится индексом.

## 10.2. Создание Drawing из Part

Backend:

```text
1. Создать drw
2. В drw добавить drawing_of → part
3. В part добавить has_drawing → drw
4. Если прикреплен PDF, добавить artifact
5. Обновить readiness part
```

## 10.3. Создание Manufacturing files из Part

Если Part имеет:

```yaml
metadata:
  manufacturing_method: laser_cut
```

UI может предложить:

```text
Generate Cutting File
```

Backend:

```text
1. Создает cut object
2. Наследует material, thickness_mm, source_part
3. Создает relation cut → part
4. Создает projection part → cut
5. Создает placeholder DXF artifact slot
```

## 10.4. Создание Process Plan из Assembly

Backend:

```text
1. Создает tpc object
2. Анализирует BOM assembly
3. Предлагает маршрут операций
4. Создает process_plan_for → assembly
5. Создает пустые operation slots
```

Пример auto route:

```yaml
route:
  - op: 10
    name: prepare_components
    source: bom
  - op: 20
    name: assembly
    source: default
  - op: 30
    name: inspection
    source: default
```

---

# 11. Автозаполнение из parent context

Когда объект создается из контекста другого объекта, часть полей наследуется.

## 11.1. Part from Assembly

Наследуется:

```yaml
project: from assembly
lifecycle_stage: from assembly
owner: current user
unit: pcs
```

Не наследуется:

```text
material
thickness
manufacturing_method
supplier
```

## 11.2. Drawing from Part

Наследуется:

```yaml
title: Drawing for {{ part.title }}
metadata:
  drawing_for: part.id
  material: part.metadata.material
  source_revision: part.version
```

## 11.3. Manufacturing file from Part

Наследуется:

```yaml
metadata:
  source_part: part.id
  material: part.metadata.material
  thickness_mm: part.metadata.thickness_mm
  manufacturing_method: part.metadata.manufacturing_method
```

## 11.4. Inspection from Part

Наследуется:

```yaml
metadata:
  source_object: part.id
  material: part.metadata.material
  critical_dimensions: part.metadata.critical_dimensions
```

---

# 12. Автозаполнение из файлов

# 12.1. Прикрепление CAD файла

При attach STEP/STP:

Backend должен:

```text
1. Определить extension
2. Назначить kind = cad
3. Назначить role = primary_cad
4. Скопировать в files/cad
5. Рассчитать checksum
6. Добавить artifact record
7. Попробовать извлечь базовые file metadata
8. Обновить readiness
```

Извлекаемые поля без CAD-парсера:

```yaml
file:
  original_name:
  extension:
  size_bytes:
  modified_at:
  checksum:
```

Если позже будет CAD parser:

```yaml
metadata:
  bounding_box:
  mass:
  volume:
  material_from_cad:
```

## 12.2. Прикрепление PDF чертежа

Если файл прикреплен как drawing PDF:

```text
Если текущий объект prt/asm:
  предложить создать Drawing object
Если текущий объект drw:
  добавить как artifact drawing_pdf
```

## 12.3. Прикрепление DXF

DXF может быть:

```text
drawing
cutting file
flat pattern
```

UI должен спросить роль, если контекст неоднозначен.

Если текущий объект `cut`:

```yaml
role: cutting_dxf
```

Если текущий объект `drw`:

```yaml
role: drawing_dxf
```

Если текущий объект `prt`:

```text
UI предлагает:
[Attach as Drawing DXF]
[Create Cutting File]
[Attach as Reference]
```

---

# 13. Создание placeholder artifacts

Не все файлы появляются сразу. Система должна уметь создавать “слоты”.

Пример: у детали нужен чертеж, но файла еще нет.

```yaml
artifacts:
  - id: art-placeholder-drawing
    kind: drawing
    role: drawing_pdf
    path: null
    required: true
    status: missing
```

Readiness показывает:

```text
Drawing PDF missing
[Attach]
[Create Drawing Object]
```

При attach файл заменяет placeholder.

---

# 14. BOM generation logic

## 14.1. eBOM

eBOM строится из `contains` relations.

```yaml
relations:
  - type: contains
    to: a320-prt-0001-v1.0
    quantity: 2
    unit: pcs
```

## 14.2. Flatten BOM

Backend должен уметь строить плоский BOM:

```text
assembly root
→ recursive contains
→ multiply quantities
→ group by object ID
→ output rows
```

Пример:

```yaml
bom:
  root: a320-asm-0100-v1.0
  type: flat
  rows:
    - object_id: a320-prt-0001-v1.0
      quantity: 4
      unit: pcs
      level: 2
```

## 14.3. mBOM

mBOM строится из eBOM + manufacturing relations:

```text
parts
+ cut/bnd/nc/wld/3dp files
+ inspection
+ process plan
+ work instructions
+ tools
```

Backend:

```text
1. Load eBOM
2. For each part find manufacturing files
3. Find process plans
4. Find inspections
5. Find tools
6. Build manufacturing matrix
```

## 14.4. Released BOM

Released BOM создается только из released/approved объектов.

Validation:

```text
root assembly released or approved
all children released or approved
no draft children
all required drawings attached
all required manufacturing files ready
checksums actual
```

---

# 15. Release package autogeneration

Когда пользователь создает release package, backend должен создать:

```text
objects/a320-rel-0100-v1.0/
├── a320-rel-0100-v1.0.md
├── history.yaml.log
└── generated/
    ├── manifest.yaml
    ├── ebom.yaml
    ├── mbom.yaml
    ├── checksums.yaml
    ├── files-list.yaml
    └── release-notes.md
```

## 15.1. manifest.yaml

```yaml
release:
  id: a320-rel-0100-v1.0
  root_object: a320-asm-0100-v1.0
  created_at: 2026-06-06T10:00:00Z
  created_by: local-user
  git_commit: null
  git_tag: null

objects:
  - id: a320-asm-0100-v1.0
    class: asm
    state: released
    checksum: sha256:...

files:
  - object_id: a320-prt-0001-v1.0
    role: primary_cad
    path: files/cad/a320-prt-0001-v1.0.step
    checksum: sha256:...
```

## 15.2. Автозаполнение release notes

Backend может автоматически создать основу:

```markdown
# Release a320-rel-0100-v1.0

## Scope

Root object: a320-asm-0100-v1.0

## Included Objects

## Changes Since Previous Release

## Known Issues

## Validation Summary

## Approval
```

## 15.3. Git tag

После успешного релиза:

```text
tag = release/a320-rel-0100-v1.0
```

В manifest обновляется:

```yaml
git:
  commit: abc123
  tag: release/a320-rel-0100-v1.0
```

---

# 16. Change request autogeneration

Когда released-объект изменяется, система не должна просто разрешить edit. Она должна предложить Change Request.

## 16.1. Trigger

```text
User edits released object
User clicks Revise
User changes released BOM
User replaces released part
```

## 16.2. Backend flow

```text
1. Detect protected object state
2. Block direct edit
3. Offer Create Change Request
4. Run impact analysis
5. Create CR object
6. Link affected objects
7. Create working revision if approved
```

## 16.3. YAML

```yaml
id: a320-doc-change-0001-v1.0
class: doc
title: Change Request for Bracket Motor
state: proposed

metadata:
  change_type: engineering_change
  reason: null
  priority: normal

relations:
  - type: affects
    to: a320-prt-0008-v1.0
  - type: affects
    to: a320-asm-0100-v1.0
```

---

# 17. Revision autogeneration

## 17.1. Minor revision

Если изменение совместимое:

```text
a320-prt-0008-v1.0 → a320-prt-0008-v1.1
```

Backend:

```text
1. Copy object folder
2. Update ID in frontmatter
3. Update version/revision
4. Link previous_version
5. Set state = draft
6. Copy artifacts or create placeholders
7. Append history
```

YAML:

```yaml
previous_version: a320-prt-0008-v1.0
revision_reason: minor compatible improvement
```

## 17.2. Major revision

Если изменение несовместимое:

```text
a320-prt-0008-v1.0 → a320-prt-0008-v2.0
```

Backend должен дополнительно:

```text
1. Mark previous relation compatibility
2. Require impact analysis
3. Require update BOM decision
4. Notify affected parent assemblies
```

---

# 18. Standard part creation

## 18.1. STD identity

STD не использует обычный sequence.

Формат:

```text
[project]-std-[std_class]-[std_id]-v1.0
```

Пример:

```text
a320-std-fst-gost7805-m6x12-v1.0
```

## 18.2. Автозаполнение

```yaml
id: a320-std-fst-gost7805-m6x12-v1.0
class: std
title: Bolt GOST 7805 M6x12
state: released

metadata:
  std_class: fst
  standard: gost7805
  size: m6x12
  material: null
  manufacturer: null
  manufacturer_part_number: null
  make_buy: buy
```

## 18.3. Duplicate detection

Перед созданием STD backend должен искать дубли:

```text
same std_class
same normalized std_id
same manufacturer_part_number
same dimensions
```

Если найден дубль:

```text
STD already exists:
a320-std-fst-gost7805-m6x12-v1.0

Actions:
[Open Existing]
[Create Variant]
[Cancel]
```

---

# 19. Автоматическое обновление связей при rename/revision

## 19.1. Rename forbidden by default

Так как ID = имя файла = номер документа, переименование released-объекта запрещено.

Разрешено только:

```text
title change
metadata change
revision creation
```

## 19.2. Revision creates new object ID

Новая версия — это новый объект:

```text
a320-prt-0008-v1.0
a320-prt-0008-v1.1
```

Связи не переписываются автоматически без решения пользователя.

UI должен спросить:

```text
Update parent assemblies to use new revision?

[No, keep old revision]
[Update selected parents]
[Create change request]
```

Backend:

```text
Command: ReplaceRelationTarget
```

---

# 20. Validation rules for autogeneration

Перед созданием объекта:

```text
ID must match naming standard
ID must be unique
class must exist
template must exist or fallback available
required metadata must be present
relations must be allowed by relation schema
target objects must exist
artifact roles must be valid
file target path must be safe
no path traversal
no overwrite without confirmation
```

После создания объекта:

```text
created files exist
frontmatter parseable
history event written
index updated
search finds object
tree projection includes object
relations projection valid
Git status sees new files
```

---

# 21. Preview mode

Каждое создание должно иметь preview до записи.

UI должен показывать:

```text
Generated ID:
a320-prt-0008-v1.0

Files to create:
objects/a320-prt-0008-v1.0/a320-prt-0008-v1.0.md
objects/a320-prt-0008-v1.0/history.yaml.log

Folders to create:
files/cad
files/drawings
generated

Relations:
a320-asm-0100-v1.0 contains a320-prt-0008-v1.0 qty=1 pcs

Artifacts:
C:\cad\bracket.step → files/cad/a320-prt-0008-v1.0.step
```

Backend query:

```text
PreviewCreateObject
```

Backend command after confirmation:

```text
CreateObject
```

---

# 22. Backend commands

## 22.1. PreviewCreateObject

Input:

```yaml
class: prt
title: Bracket Motor
context:
  parent_object_id: a320-asm-0100-v1.0
metadata:
  material: al5052
files:
  - local_path: C:\cad\bracket.step
    role: primary_cad
```

Output:

```yaml
generated:
  id: a320-prt-0008-v1.0
  object_dir: objects/a320-prt-0008-v1.0
  markdown_path: objects/a320-prt-0008-v1.0/a320-prt-0008-v1.0.md

planned_files:
  create:
    - objects/a320-prt-0008-v1.0/a320-prt-0008-v1.0.md
    - objects/a320-prt-0008-v1.0/history.yaml.log
  copy:
    - from: C:\cad\bracket.step
      to: objects/a320-prt-0008-v1.0/files/cad/a320-prt-0008-v1.0.step

planned_relations:
  - from: a320-asm-0100-v1.0
    type: contains
    to: a320-prt-0008-v1.0
    quantity: 1
    unit: pcs

diagnostics: []
```

## 22.2. CreateObject

Executes preview plan.

Important:

```text
CreateObject must never recompute a different ID than preview unless preview expired.
```

Therefore preview should return:

```yaml
preview_token: create-preview-abc123
expires_at: 2026-06-06T10:05:00Z
```

CreateObject input:

```yaml
preview_token: create-preview-abc123
confirm: true
```

---

# 23. Idempotency and safety

Creation commands should be idempotent with operation IDs.

```yaml
operation_id: op-20260606-0001
```

If frontend retries after timeout, backend should return existing result if operation already completed.

Rules:

```text
same operation_id + same payload → return previous result
same operation_id + different payload → error
```

---

# 24. Recovery after failed creation

If crash happens during creation, `.plm/transactions` contains recovery plan.

On next project open:

```text
Recovery mode detected unfinished transaction:
Create object a320-prt-0008-v1.0

State:
✓ folder created
✓ markdown written
✗ history not written
✗ index not updated

Actions:
[Complete]
[Rollback]
[Open Details]
```

Backend must support:

```text
CompleteTransaction
RollbackTransaction
```

---

# 25. Autogeneration scenarios

## Scenario A. Create part inside assembly

User:

```text
Assembly → BOM → Add Item → Create New Part
```

Backend creates:

```text
objects/a320-prt-0008-v1.0/a320-prt-0008-v1.0.md
```

Updates assembly:

```yaml
relations:
  - type: contains
    to: a320-prt-0008-v1.0
    quantity: 1
    unit: pcs
```

Updates index:

```text
where-used: a320-prt-0008-v1.0 → a320-asm-0100-v1.0
bom: a320-asm-0100-v1.0 includes a320-prt-0008-v1.0
```

## Scenario B. Attach STEP to part

User:

```text
Part → Files → Attach STEP
```

Backend:

```text
copies file to files/cad/a320-prt-0008-v1.0.step
calculates checksum
adds artifact record
updates readiness
```

## Scenario C. Create drawing from part

User:

```text
Part → Create Drawing
```

Backend:

```text
creates a320-drw-0008-v1.0
adds drawing relation
creates placeholder drawing PDF
updates readiness
```

## Scenario D. Generate manufacturing files

User:

```text
Part → Generate Manufacturing Files
```

If metadata:

```yaml
manufacturing_method: laser_cut_bending
```

Backend creates:

```text
a320-cut-0008-v1.0
a320-bnd-0008-v1.0
a320-ins-0008-v1.0
```

Relations:

```text
cut → manufacturing_file_for → part
bnd → manufacturing_file_for → part
ins → inspection_for → part
```

## Scenario E. Create release package

User:

```text
Assembly → Release → Create Package
```

Backend:

```text
validates assembly
generates release object
generates manifest.yaml
generates ebom.yaml
generates mbom.yaml
generates checksums.yaml
optionally creates Git tag
```

---

# 26. Final design rule

Every generated file must answer:

```text
Why does this file exist?
Which object owns it?
Which source object created it?
Which process created it?
Can it be regenerated?
Is it source of truth or derived?
```

Therefore every generated object/file should have metadata:

```yaml
generation:
  generated: true
  generator: release_package_generator
  source_objects:
    - a320-asm-0100-v1.0
  generated_at: 2026-06-06T10:00:00Z
  generated_by: local-user
  can_regenerate: true
```

For source objects:

```yaml
generation:
  generated: false
  source_of_truth: true
```

---

# 27. Summary

The autogeneration system must make object creation safe, predictable and traceable.

Core rules:

```text
1. User creates engineering intent, not files manually.
2. Backend generates ID from naming standard.
3. YAML/Markdown is canonical.
4. Files are routed automatically.
5. Relations are created from context.
6. Inverse relations are projections, not duplicated manually.
7. Every change writes history.
8. Every generated file is traceable to source object.
9. Every creation has preview.
10. Every operation is recoverable.
```