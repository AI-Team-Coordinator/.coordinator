# Координатор: команды (crew) и CrewConnector

Author: **Evgeny KOSIVTSOV** (`EK`)
Created: 2026-10-13 17:11
Branch: `main`
Services: `.coordinator`, `.cursor`, `Common`

Связано: [доменная модель](./20260915-1805-EK-COORDINATOR_DOMAIN_MODEL.md), [data layer](./20260907-2225-EK-COORDINATOR_DATA_LAYER.md), [parallel tasks](./20260909-1423-EK-COORDINATOR_PARALLEL_TASKS.md), [peer scope](./20260912-1121-EK-COORDINATOR_PEER_SCOPE.md).

## Зачем

Сейчас ростер плоский: `team.json` → список **Person** с `role` и `focus[]` (id **сервисов** из `project_profile.json`). Pulse строится по людям; overlap смотрит на **Claim** (`services[]` слота), без организационного контекста.

Нужно:

1. Группировать людей в **команды** (отделы / crew) с M:N («один человек — в нескольких командах»).
2. Задать у каждой команды **направление интереса** — за что она отвечает и что отслеживает в целом (текст).
3. Описать **как команды связаны смыслом** — не git и не lock, а «команда X интересуется работой команды Y: …» — через сущность **CrewConnector**.
4. Хранить это в шине настроек (`Common/data/settings/`), править и **просматривать в настройках Координатора** (структурированно + полные текстовые формулировки).

Дальше — отдельный scoped Pulse по команде и влияние внешних команд на пульс; в этом документе — модель данных и границы с существующим кодом.

## Семейство «коннекторов»

Слово **коннектор** в Координаторе — общее имя для **типизированной связи** между сущностями, которую можно описать текстом и показать в настройках. Типы **не смешиваются** в одном массиве без явного поля `kind`.

| Тип (код) | Контекст | В этом документе |
|-----------|----------|------------------|
| **CrewConnector** | Команда → команда | Да, полная спецификация |
| Коннекторы внешних систем | Не сущность Координатора | Живут в плоскости **Relay**: [внешние знания и события](./20261013-1741-EK-COORDINATOR_SYSTEM_CONNECTOR.md) |

В коде, API и JSON для связей **между командами** не использовать голое `connector` / `connectors` — только **CrewConnector** и ключ **`crew_connectors`**. Адаптеры Jira/Confluence/GitHub — домен Relay, не эта модель.

В UI (ru) для CrewConnector: **«Связь команд»**; в английском интерфейсе — «Crew link» или «Crew connector».

## Термины

| Термин | В коде / JSON | В UI (ru) |
|--------|---------------|-----------|
| Команда / отдел | **crew** | «Команда» |
| Связь команд | **CrewConnector** | «Связь команд» |
| Направление интереса команды | **interest_brief** на crew | «Направление интереса» |
| Пункт интереса по связи | элемент **interests[]** на CrewConnector | «Интерес» |

Не использовать id `department` в JSON — путаница с HR. Не путать с `project_profile.groups` (группы **сервисов**) и с `members[].focus[]` (какие репозитории человек обычно трогает).

## Место в доменной модели

Из [доменной модели](./20260915-1805-EK-COORDINATOR_DOMAIN_MODEL.md):

- **Person**, **Agent**, **Task**, **Resource**, **Claim**, **Lock**, **Dependency** (на слоте: `related_tasks[]` / `related_docs[]`) — без изменений смысла.
- **Crew** — группа Person; не исполнитель и не зона в git.
- **CrewConnector** — направленное смысловое ребро **Crew → Crew**; не событие jsonl и не overlap.

```text
Person ∈ Crew (M:N)
CrewConnector: observer_crew ──interests[]──► subject_crew
Claim: Person+Task ──services[]──► Resource   (как сейчас)
Dependency: Task ──► Task/Doc                 (как сейчас)
(внешние системы — плоскость Relay, [пропозал](./20261013-1741-EK-COORDINATOR_SYSTEM_CONNECTOR.md))
```

CrewConnector дополняет Dependency: dependency — «эта **задача** ждёт то», CrewConnector — «эта **команда** следит за тем, что делает другая команда».

## Crew (команда)

### Поля

| Поле | Тип | Обязательно | Смысл |
|------|-----|-------------|--------|
| `id` | slug | да | Стабильный id (`core-dev`, `qa`, `devops`) |
| `label` | `{ en, ru }` | да | Как в `project_profile.groups` |
| `kind` | enum | да | `internal` \| `affiliated` \| `external` — для будущего tier Pulse |
| `interest_brief` | string (markdown) | рекомендуется | Направление интереса команды: миссия, зона ответственности, что отслеживаем **в целом** |
| `notes` | string (markdown) | нет | Свободные заметки для людей |

Состав: `members[].crews[]` в ростере (M:N), плюс опционально `primary_crew` на человеке — дефолтный фильтр Pulse / контекст Отто.

### Пример

```json
{
  "id": "core-dev",
  "label": { "ru": "Core-разработка", "en": "Core development" },
  "kind": "internal",
  "interest_brief": "Платформенный API, биллинг, контракты Inbox, релизы Core и смежных сервисов."
}
```

## CrewConnector (связь команд)

**CrewConnector** отвечает на вопрос: **как команда-наблюдатель (`from_crew_id`) связана с работой команды-объекта (`to_crew_id`)**.

Направление важно: «QA → Core» и «Core → QA» — разные CrewConnector с разными `interests[]`.

### Поля

| Поле | Тип | Обязательно | Смысл |
|------|-----|-------------|--------|
| `id` | slug | да | Уникальный id связи (`qa-watches-core`) |
| `from_crew_id` | crew id | да | Кто **интересуется** (наблюдатель) |
| `to_crew_id` | crew id | да | Чья **работа** в фокусе (объект) |
| `interests` | массив | да | Смысловые пункты интереса (см. ниже) |
| `summary` | string | нет | Одна строка для списка в UI |
| `notes` | string (markdown) | нет | Общий комментарий к связи |

Правила:

- `from_crew_id` ≠ `to_crew_id`.
- Пара `(from_crew_id, to_crew_id)` — не больше одного CrewConnector (при необходимости расширять `interests[]`, а не плодить дубликаты).
- `from` / `to` могут быть `external` — связь всё равно хранится и видна в настройках; влияние на Pulse — фаза 2.

### Элемент `interests[]`

Каждый пункт — **текстовое описание**, но с минимальной структурой для таблицы и фильтров в UI:

| Поле | Тип | Обязательно | Смысл |
|------|-----|-------------|--------|
| `title` | string | да | Короткая формулировка («Релизы API Inbox», «Регресс биллинга») |
| `detail` | string (markdown) | нет | Развёрнутое описание: когда важно, кого пинговать, что считать событием |
| `tags` | string[] | нет | Опциональные метки для группировки в UI (`release`, `sla`, `security`) — не валидируются против сервисов |

Смысл фразы в продукте: **«{from.label} интересуется в работе {to.label}: {title}»** — `detail` раскрывается по клику.

Пример (элемент массива `crew_connectors`):

```json
{
  "id": "qa-watches-core",
  "from_crew_id": "qa",
  "to_crew_id": "core-dev",
  "summary": "QA следит за релизами и контрактами Core",
  "interests": [
    {
      "title": "Релизы API и миграции",
      "detail": "Нужен heads-up за 24ч до prod deploy Core; smoke по чек-листу Inbox JWT.",
      "tags": ["release"]
    },
    {
      "title": "Изменения биллинга",
      "detail": "Любой PR с `domain/billing` — уведомление в канал QA.",
      "tags": ["billing"]
    }
  ]
}
```

## Хранилище

### Файл

Отдельный файл рядом с ростером (проще миграция, чем ломать `team.json` v1):

`Common/data/settings/org_topology.json`

```json
{
  "version": 1,
  "crews": [ /* Crew */ ],
  "crew_connectors": [ /* CrewConnector */ ]
}
```

Поле **`system_connectors`** в этом файле **не добавлять**. Подписки на внешние системы — конфиг Relay, не `org_topology.json`.

Ростер остаётся `team.json`; при переходе на v2 ростера — только добавление `crews[]` / `primary_crew` у `members`, без дублирования текстов CrewConnector в двух местах.

Синхронизация: как `team.json` и `project_profile.json` — коммит в `Common` `main`, API Координатора читает/пишет через существующий путь настроек (`FileRepository`, sync settings).

### Валидация (backend)

- Все `crew_id` в CrewConnector и в `members[].crews[]` существуют в `crews[]`.
- Уникальность `id` crew и CrewConnector.
- `interests[].title` не пустой, длина лимитов (как у `name` в team: порядка 80–200 символов на title, detail — разумный max для markdown).
- Slug-паттерн для id — как у service id в profile (`[a-z][a-z0-9_-]{0,63}`).

### Локальный контекст просмотра

По аналогии с `.current_author` (gitignore):

`Common/data/.current_crew` — опционально, **какая секция развёрнута первой** или последний выбранный crew в переключателе Pulse (не жёсткий фильтр «только эта команда»). `primary_crew` в `team.json` — подсказка по умолчанию для этого выбора.

## Решения: Pulse и отображение людей

### Два режима Pulse (фаза 2)

Режим по умолчанию зависит от **роли просмотра**, а не от `primary_crew` как единственного фильтра.

| Роль просмотра | Кто | Дефолт на вкладке Pulse | Переключатель |
|----------------|-----|-------------------------|---------------|
| **Участник** (`pulse_scope: member`, по умолчанию для всех в ростере) | Разработчик, QA, … в одной или нескольких crew | **Сгруппировано по «моим» командам**: секция на каждый `crew_id` из `members[].crews[]` текущего alias; внутри секции — карточки всех участников этой crew (как сейчас MemberCard + слоты) | Можно открыть Pulse **другой** crew (read-only или awareness tier для `external`) и при необходимости **org-wide** («вся организация») |
| **Менеджер / C-level** (`pulse_scope: executive`) | Основатель, руководители | **Org-wide**: один плоский или сводный Pulse по всем людям из ростера, **без** фильтра по crew | Можно сузить до одной crew (drill-down) |

Поле **`pulse_scope`** на человеке в `team.json`: `member` (default) | `executive`. Не путать с `role` (`founder`, …) — это подпись в UI; `pulse_scope` — только логика дашборда. На старте можно выставлять `executive` вручную founder’ам; позже — отдельная роль в настройках.

Конфликты overlap и stray repos при org-wide и при crew-scoped **считаются по полному снимку** (как сейчас); смена вида Pulse — только presentation layer, не ослабление stop/warn.

API (черновик):

```text
GET /api/pulse?view=grouped|global&crew=<id>   # crew опционален; view по умолчанию из pulse_scope текущего alias
```

### Где дублировать карточку человека

| Экран | Дублирование | Правило |
|-------|----------------|---------|
| **Организация → команда** (просмотр / редактирование состава) | **Да** | Список участников crew — полный ростер этой команды; человек в QA и Core **перечислен в обеих** карточках команд. Это ожидаемо. |
| **Pulse, режим участника (grouped)** | **Да, в рамках секций** | В секции «QA» человек с `crews: [qa, core-dev]` показан **вместе с остальными QA**; в секции «Core-dev» — снова со своими слотами. Одна и та же живая карточка (те же слоты), два контекста группировки. |
| **Pulse, org-wide (executive)** | **Нет** | Одна карточка на alias, бейджи всех `crews[]` на карточке. |

Итого: дублирование согласовано для **страницы департаментов** и для **crew-секций Pulse у участника**; глобальный Pulse для executive — одна карточка на человека.

## UI настроек Координатора

Новая секция (или вкладка) **«Организация»**:

1. **Команды** — таблица: label, kind, interest_brief (свёрнуто / разворот markdown), состав (aliases из `team.json`).
2. **Связи команд** — список CrewConnector: `from → to`, summary, число пунктов `interests`.
3. **Карточка CrewConnector** — заголовок пары команд, `notes`, таблица интересов (`title`, tags, preview `detail`).
4. **Граф (опционально, фаза 1b)** — только чтение: узлы = crew, рёбра = CrewConnector (направленные).

Интеграции Jira/Confluence/GitHub — не вкладка «Связи команд». Координатор показывает только уже попавшие в шину знания/события; адаптеры — в Relay. См. [пропозал](./20261013-1741-EK-COORDINATOR_SYSTEM_CONNECTOR.md).

Редактирование: те же формы, что для команды — save через API → запись `org_topology.json` + sync git (team mode).

Просмотр без правки: `access: member` видит; `access: admin` — save (как для team).

## Поведение Координатора (дорожная карта)

### Фаза 1 — данные и настройки

- `org_topology.json`, CRUD API, валидация, UI «Организация» (crews + crew_connectors).
- `team.json`: `crews[]`, `primary_crew`, `pulse_scope` у members.
- Pulse без смены overlap: опционально бейджи crew на `MemberCard`.

### Фаза 2 — scoped Pulse и «осведомлённость»

- Режимы Pulse: `grouped` (участник, секции «мои» crew) и `global` (executive по умолчанию); переключатель crew / «вся организация» — см. § «Решения: Pulse».
- Полоска **awareness**: для выбранной команды показать входящие и исходящие **CrewConnector** («кто за нами следит» / «за кем следим») — только заголовки `interests[].title`, без автоматических stop/warn.
- `kind: external` — compact / awareness tier (см. обсуждение в чате 2026-10-07).

### Фаза 3 — задачи и Отто

- Опционально `Crew:` в шапке task-doc и `crew_id` в снимке слота.
- Отто в правилах: при peer overlap или релизе в зоне `to_crew` подсказать затронутые CrewConnector (читать topology, не хардкод).

### Не входит в v1

- Адаптеры внешних систем и UI Relay (это другая плоскость).
- Автоматическое создание Dependency из CrewConnector.
- Lock / overlap по crew вместо Resource.
- Версионирование истории текстов интересов (достаточно git blame на `org_topology.json`).

## Отличия от существующих полей

| Механизм | Уровень | Тип связи |
|----------|---------|-----------|
| `members[].focus[]` | Person | Сервисы из profile (техн.) |
| `interest_brief` | Crew | Текст: общая миссия команды |
| `CrewConnector.interests[]` | Crew → Crew | Текст: что именно X хочет знать о работе Y |
| `related_tasks[]` / `related_docs[]` | Task | Конкретная задача / док |
| `handoff` / `decision` | Event | Журнал действий |

## Открытые решения

1. Нужен ли обратный индекс «все CrewConnector, где `to_crew_id` = мы» на главной Pulse без захода в настройки?
2. Один файл `org_topology.json` vs два (`crews.json`, `crew_connectors.json`) — пока один файл проще атомарного коммита.
3. Внешние системы — не CrewConnector и не SystemConnector в Pulse; см. [Relay](./20261013-1741-EK-COORDINATOR_SYSTEM_CONNECTOR.md).

## Черновик расширения API

```text
GET  /api/org-topology
PUT  /api/org-topology          (admin; body = full file or patch — при реализации выбрать)
GET  /api/pulse?view=grouped|global&crew=<id>   (фаза 2; default view из pulse_scope)
```

Модели Go: `model.Crew`, `model.CrewConnector`, `model.CrewConnectorInterest` в пакете `coordinator` или `model/org.go`; DTO для фронта в `dto/org_topology.dto.go`. Не называть тип `Connector` без префикса Crew/System.
