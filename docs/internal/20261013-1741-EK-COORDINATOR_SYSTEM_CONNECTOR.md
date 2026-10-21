# Координатор: внешние знания и события; плоскость Relay

Author: **Evgeny KOSIVTSOV** (`EK`)
Created: 2026-10-13 17:41
Branch: `main`
Services: `.coordinator`, `.cursor`, `Common`

Связано: [команды и CrewConnector](./20261013-1711-EK-COORDINATOR_CREWS_AND_CONNECTORS.md), [маркетплейс коннекторов](./20261020-1808-EK-RELAY_CONNECTOR_MARKETPLACE.md), [доменная модель](./20260915-1805-EK-COORDINATOR_DOMAIN_MODEL.md), [data layer](./20260907-2225-EK-COORDINATOR_DATA_LAYER.md), [status bus](./20260910-0732-EK-COORDINATOR_STATUS_BUS_ALTERNATIVES.md), [автоматизации на пульсе](./20260926-1417-EK-COORDINATOR_AUTOMATIONS.md).

Статус: **proposal**, решения по имени и срезу **согласованы 2026-10-14** (см. конец). Код ещё не пишем.

Имя файла историческое (`SYSTEM_CONNECTOR`). Сущность **SystemConnector на стороне Координатора нет** — см. ниже.

## Две плоскости

Это разные продукты с общей шиной, не «коннектор внутри Pulse».

Две связи, их нельзя смешивать:

| Плоскость | Направление | Зачем |
|-----------|-------------|--------|
| **Данные** | Relay → git-шина → Координатор | Durable: ExternalEvent / ExternalKnowledge. Jira и GitHub Координатор не трогает |
| **Управление** | Координатор (настройки) ↔ Relay API | Подключить сотрудника, токен, права; админ — адаптеры. Не ingest |

```text
  Jira / Confluence / CI / …
           ▲
           │  данные: ходит только Relay
           │
  ┌────────┴────────┐         git-шина (coordinator-state)
  │  .coordinator/  │  ──пишет──►  progress/external/…
  │     relay       │
  │  (адаптеры,     │
  │   выдача токенов)
  └────────┬────────┘
           │
           │  управление: HTTP из настроек Координатора
           │  (адрес + локальный access token)
           ▼
  ┌─────────────────┐         читает шину; Pulse / Отто / UsePolicy
  │  .coordinator/  │
  │ backend+frontend│
  └─────────────────┘
```

| Плоскость | Где живёт | Делает | Не делает |
|-----------|-----------|--------|-----------|
| **Координатор** | `.coordinator/backend`, `frontend` | Потребляет знания и события; смысловая реакция (UsePolicy); **пара** с Relay из настроек | Не ходит в Jira / GitHub / Confluence, не хранит их токены, не знает адаптеров |
| **Relay** | `.coordinator/relay` | Ходит во внешние системы; адаптеры; пишет в шину; **выдаёт** токены сотрудникам | Не открывает слоты, не overlap, не голос Отто |

## Имя: Relay

Каталог: **`.coordinator/relay`** — рядом с `backend/` и `frontend/`, отдельный процесс (как API на `:4321`, только свой бинарь и порт).

В коде, `go.mod` и английском UI — **Relay**. В русской речи и подписях UI — **реле**: односторонняя ретрансляция сигнала (внешняя система → нормализованная запись в шине), не двустороннее зеркало.

| Имя | Почему нет |
|-----|------------|
| cloud sync | Врёт: runtime — инфра **клиента**, не наш cloud |
| remote sync | «Sync» звучит как зеркало Jira |
| Remote | Слишком пусто: «удалённое» не говорит, что сервис **ретранслирует** события в шину |
| ingest / harvester | Точно по роли, плохо как имя продукта |

Адаптеры (Jira, Confluence, GitHub) — **домен Relay**, имена вида `JiraAdapter` внутри `.coordinator/relay`. На стороне Координатора слова Connector для внешних систем **нет**: там уже есть CrewConnector, и смешивать нельзя. Внутри Relay тоже не называем адаптер Connector — иначе путаница с CrewConnector.

Старое рабочее **SystemConnector** снимаем. Это был конфиг подписки, замаскированный под сущность Pulse.

## Что видит Координатор

Два типа записей (плюс позже UsePolicy). Источник правды по-прежнему шина; с **парой** Координатор ещё может **подтянуть** свежее с Relay API (read). Без пары оверлей скрыт.

### ExternalEvent

Дискретное «случилось»: QA failed, тикет сдвинулся, сборка красная.

Не строка в `events/EK/2026.jsonl` (`model.Event` = действия человека). Соседний журнал, оверлей Pulse как conflicts/stray:

```text
Pulse: members[] + conflicts[] + stray[] + external.events[] + external.knowledge[]
```

Конверт (`relay/model.ExternalEvent`, `schema_version: 1`):

| Поле | Смысл |
|------|--------|
| `received_at` | Когда Relay принял и записал строку (UTC, ставит шина) |
| `occurred_at` | Когда случилось во внешней системе, если адаптер знает |
| `source` | Адаптер / коннектор (`github`, позже `jira`) |
| `class` | Как читать: `info`, `warning`, `reminder` (набор расширяется) |
| `kind` | Что случилось (`review.commented`, `ci.failed`, …) |
| `id` | Стабильный id в шине |
| `external_id` | Id во внешней системе |
| `title` / `summary` / `url` | Для Pulse и Отто |
| `tags` | Опциональные метки фильтра |
| `data` | JSON-объект адаптера (у GitHub: `repo`, `branch`, `pr_author_login`, …) |

Позже на конверте: `crew_id`, `task_id`, `alias` — когда Relay замапит на команду/слот.

### ExternalKnowledge

Снимок, не лог: страница Confluence, выжимка epic. Markdown/текст + `updated_at` + `external_id`. Не зеркало wiki.

### UsePolicy

Смысл **на стороне Координатора**: когда Отто смотрит на эти записи. Relay про политику не знает.

Примеры якорей: `before_task_started`, `notify`, `brief_on_view`. Первая версия — текст, как `interest_brief`. Жёсткий stop из Jira в overlap — не v1.

## Что такое коннектор (только Relay)

Коннектор — **плагин к внутреннему API Relay**. Сколько угодно сложная механика снаружи (webhook, poll, очередь, отдельный бинарь на проекте) на выходе даёт **стандартизированный поток** `[]ExternalEvent`. Дальше только Relay: шлюз → git-шина. Ядро не знает Jira/GitHub.

```text
внешняя система
    → Connector (in-process Go или свой сервис)
    → []ExternalEvent  — единый конверт (class, kind, data, …)
    → Gateway.Publish / Gateway.Webhook
    → progress/external/events/<source>/<год>.jsonl
```

### Внутренний API Relay

| Вход | Кто вызывает | Смысл |
|------|----------------|--------|
| `POST /connectors/{source}/webhook` | Внешняя система | Сырой webhook → `Connector.Normalize` → шина |
| `POST /api/v1/connectors/{source}/events` | Bearer (координатор или sidecar-коннектор) | Уже нормализованный JSON-массив событий → шина |
| `GET /api/v1/sources` | Координатор | Каталог `Descriptor` зарегистрированных коннекторов |

In-process: `connector.Registry` + `gateway.Gateway` в `app.New` (сейчас stock `github`). Другой проект добавляет свой `Connector` в registry или шлёт на `/events`.

```text
read(external_id)  ←  Connector.Read  ←  запрос Координатора (refresh, позже)
```

У адаптера два входа, оба на Relay:

| Вход | Когда | Куда кладётся |
|------|--------|----------------|
| **Async** (webhook или poll) | Во внешней системе что-то случилось, ноутбук может быть закрыт | Запись в шину. Это основной путь |
| **Read** | Человек открыл задачу и шина отстала, или нужен текст, которого в событии нет | Ответ в API Координатору. По желанию Relay дописывает снимок в knowledge, чтобы следующий раз взять из шины |

Секреты Jira / GitHub / Confluence — только env Relay. Access token сотрудника — только на его машине.

Конфиг адаптеров живёт **у Relay**, не в `org_topology.json`. Поле **`system_connectors`** в crews-файл **не добавлять**.

Витрина адаптеров для чтения может дублироваться в шину. Живое администрирование — только Relay API по admin-токену, UI в настройках Координатора.

## Зачем async, а зачем read

**Async нужен**, когда факт должен дойти до команды без действия человека: смена статуса, красный CI, комментарий ревью. Иначе Pulse и Отто видят это только у того, кто в этот момент открыл браузер.

**Read нужен**, когда человеку уже понятен ключ (номер тикета в имени ветки, ссылка в событии) и надо **дочитать** описание. Это не второй ingest и не повод пускать Координатор во внешнюю систему с ноутбука: запрос идёт Координатор → Relay → система. Учётные данные остаются на Relay.

Разбор ключа из ветки (`PROJ-123` в `feature/PROJ-123-…`) делает **Координатор локально**. В Jira за текстом он не ходит.

## Пример паттерна: Jira (не обязательство продукта)

Мотив — команда вне этого репозитория (работа в InCountry): ветки завязаны на номера тикетов, описания задач в Jira, рядом Confluence, GitHub Actions и свой QA. Для каркаса это **иллюстрация адаптера**, не бэклог Alina Assist и не обещание подключить ту команду.

| Вопрос | Куда |
|--------|------|
| Ветка содержит ключ тикета | Координатор, локальный разбор имени ветки. Relay не нужен |
| Статус, QA failed, комментарий, который должна увидеть другая crew | **Async** Jira-адаптер → ExternalEvent в шине |
| «Что в описании тикета?» при открытой задаче | **Read** того же адаптера, если снимка в knowledge нет или он старше порога |
| Страница Confluence | **Async** снимок: `.md` в шине + `url` в шапке. Не зеркало wiki |

Координатор с ноутбука в Jira **не** ходит: иначе секрет на каждой машине и второй клиент рядом с адаптером.

## Пример паттерна: GitHub (ревью и Actions)

Комментарий на PR и красный workflow — типичный async: автор в этот момент не смотрит вкладку Review.

| Сигнал | Запись | Что не класть в шину |
|--------|--------|----------------------|
| Review comment, review requested, changes requested | ExternalEvent `review.commented` / `review.requested`: url, автор, путь файла, короткий excerpt | Полный diff и тред целиком |
| `workflow_run` / check conclusion | ExternalEvent `ci.failed` / `ci.passed`: имя workflow, url, sha | Лог job |

Дочитать тред — **Read** того же GitHub-адаптера.

Не дублировать то, что Координатор уже пишет сам в человеческий jsonl (`deploy_*` у Alina Assist из Coolify). Адаптер GitHub для другой команды описывает **её** CI; он не второй источник тех же deploy-событий.

## Хранилище (шина)

Пишет **Relay**. Читает **Координатор** (SQLite-кэш как у jsonl людей).

```text
Common/data/progress/external/
  events/<origin_id>/<YYYY>.jsonl
  knowledge/<origin_id>/<external_id>.md
```

Ветка `origin/coordinator-state`, не `Common` `main`. Не класть сырой JSON Jira целиком.

Идемпотентность: `source` + `external_id`.

## Управление Relay из настроек Координатора

Relay сначала поднимают в инфре клиента. Потом участник в Координаторе подключает доступ — без живого Relay парить нечего.

Это не `access: admin|member` из `team.json`. Права на Relay живут на **токене, который выдал Relay**.

### Когда в установке

После health. В настройках — блок **«Реле»**.

### Пара

1. Адрес: подсказка `relay.default_url` в `coordinator.json` (в git, без секрета); на машине можно переопределить.
2. Координатор стучится на Relay: health + handshake.
3. Первый admin: bootstrap-секрет в env Relay, первый успешный handshake получает scope `admin`. Дальше admin перевыпускает токены (`read` / `admin`) уже на Relay. Привязка к GitHub org — не v1.
4. Relay отдаёт access token. Координатор кладёт его в `Common/data/.relay_access` (gitignore): `url`, `token`, scopes. В git, в `team.json` и в шину токен **не** пишется.

Повторный заход: локальный токен → проверка, что Relay его принимает. Отзыв — на Relay; локально стирается файл.

### Два уровня токена

| Scope | Что может сотрудник из Координатора |
|-------|--------------------------------------|
| **read** | Оверлей Pulse, refresh с Relay, если шина отстала, Read по ключу (тикет, тред). Адаптеры не трогает |
| **admin** | Всё из read плюс включение адаптеров, фильтры, маппинг на crew. Запросы идут в Relay API |

Сменить read→admin = перевыпустить токен на Relay, не флажок в `team.json`.

Форма «новый адаптер» сохраняется **на Relay**. Секреты внешней системы на диск сотрудника не кладутся.

### Что без пары

Без валидного `.relay_access` оверлей `external.*` **скрыт целиком**, даже если записи уже приехали в `coordinator-state`. Иначе клон без допуска видит чужой Jira/GitHub контекст.

### Не путать с шиной

Пара не заменяет запись в `coordinator-state`. Relay по-прежнему **пишет** события в git. Токен — кто с этой машины имеет право забирать обновления и менять Relay. Ноутбук не становится ingest-воркером.

## Пример потока

1. Jira: тикет → Failed.
2. Relay (`JiraAdapter`) нормализует → ExternalEvent `qa.failed`, `crew_id=core-dev`.
3. Запись в state-шину.
4. Координатор с **read**-токеном показывает её на Pulse. Дочитать описание — Read того же адаптера, не прямой вызов Jira.

## Фазы

### Фаза 0 — этот документ

Плоскости и имя **Relay** согласованы.

### Фаза 1 — каркас приложения и шина

Отдельный процесс `.coordinator/relay` (см. ниже). Каталог `progress/external/`. Первый адаптер в коде — **GitHub** (ревью и Actions). Координатор пока только читает jsonl, когда до него дойдёт очередь. Пара `.relay_access` и UI — не в этом каркасе.

### Фаза 2 — пара из настроек

Health уже есть у процесса. Дальше: bootstrap → admin, `.relay_access`, refresh с реле. Страница «Сигналы» в UI Координатора запрашивает у Relay список включённых источников (например `GET /api/v1/sources`, Bearer из `.relay_access`); название и описание отдаёт Relay, не захардкоженный перечень адаптеров в frontend.

### Фаза 3 — UsePolicy у Отто

Advisory, без stop overlap.

### Не входит

- Координатор ходит в Jira / Confluence / GitHub сам.
- Двусторонний sync тикетов и PR.
- Автослот из Jira.
- SystemConnector в модели Pulse.
- CrewConnector и адаптеры Relay в одном массиве.
- Токен Relay в git или в `team.json`.
- Дубль `deploy_*` Alina Assist из GitHub, если то же событие уже пишет Координатор.

## Согласованные решения (2026-10-14)

| # | Решение |
|---|---------|
| 1 | Продукт **Relay**, каталог `.coordinator/relay`. По-русски **реле**. «Remote» не используем |
| 2 | Первый адаптер в коде — **GitHub** (ревью и Actions). Jira остаётся примером паттерна, не этим каркасом |
| 3 | `kind` — строка + `tags[]`. Известные kind документируются; новый адаптер не требует релиза Координатора |
| 4 | Knowledge — файл `.md` в шине; в шапке `url`, `updated_at`, `title`, `crew_id` |
| 5 | Токен: bootstrap-секрет → первый admin; дальше выдача на Relay. GitHub-org login — не v1 |
| 6 | `relay.default_url` в `coordinator.json` + локальный `.relay_access` |
| 7 | Без пары `external.*` в UI не показывать |
| 8 | Ключ тикета из ветки парсит Координатор. Текст и статусы — только Relay |
| 9 | Ревью и GitHub Actions — async-события Relay; полный diff в шину не класть |
| 10 | В меню Координатора по-русски **Сигналы** (страница «Внешние сигналы»). В доках и в коде по-прежнему Relay |

## Приложение

Свой модуль **`.coordinator/relay`**, `go.mod` = `module relay`, свой `main`, порт **`:4322`**. Не пакет внутри `backend/`: другие секреты и другой процесс. Репозиторий тот же (`.coordinator` всегда на `main`).

Повтор слоёв Координатора, без SQLite и без UI:

```text
main
  app          config, http.Server, ServeMux
  domain/relay controller → service
  domain/relay/github   первый адаптер
  domain/relay/bus      дописывает jsonl
  model        ExternalEvent
  infra        лог
```

```text
POST /github/webhook
  подпись X-Hub-Signature-256
  → github.Adapter.Normalize
  → bus.Append
  → progress/external/events/github/<YYYY>.jsonl
```

`Adapter`: `Normalize` (webhook) и `Read` (дочитать тред; в каркасе возвращает «ещё нет»). Координатор этот модуль не импортирует: потом читает те же файлы, как человеческий jsonl.

Секрет webhook и ключ GitHub App — только env реле. В каркасе используется webhook secret; App ID и приватный ключ появятся, когда понадобится дочитать API.
