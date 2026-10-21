# Relay: маркетплейс коннекторов и зашифрованные настройки

Author: **Evgeny KOSIVTSOV** (`EK`)
Created: 2026-10-20 18:08
Branch: `main`
Services: `.coordinator`, `Common`, `Core`
Status: **phase 1 in progress** (built-in catalog, SQLite secrets, UI grid)

Связано: [Relay — внешние события](./20261013-1741-EK-COORDINATOR_SYSTEM_CONNECTOR.md), [data layer](./20260907-2225-EK-COORDINATOR_DATA_LAYER.md).

## Зачем

Relay уже принимает **стандартизированный поток** `ExternalEvent` через внутренний API (Gateway, registry). Дальше нужно:

1. **Включать** только нужные коннекторы на установке / у крупного клиента — без пересборки Координатора.
2. **Хранить** разные настройки и секреты коннекторов **на стороне Relay**, зашифрованно, не в git.
3. Дать в UI понятный **маркетплейс**: каталог в коде, подключение и конфигурация через интерфейс.

Цель дистрибуции: один код маркетплейса, разный **набор включённых** коннекторов и образов Relay под клиента (см. родительский пропозал Relay).

## Принципы

| Принцип | Смысл |
|---------|--------|
| Каталог в коде | Все известные коннекторы зарегистрированы в Relay (маркетплейс). Неизвестный id из UI не подключить. |
| Включение в настройках проекта | Какие `source` активны — в настройках **проекта** (git-safe), без секретов. |
| Секреты только на Relay | Пароли, API keys, webhook secrets — **локальная БД Relay**, шифрование at rest. Координатор их не хранит. |
| Один выход | Любой коннектор (webhook, poll, sidecar) → `[]ExternalEvent` → Gateway → `progress/external/events/…`. |
| UI в Координаторе | Раздел реле: маркетплейс + модалки. Запись секретов — через **admin** API Relay. |

## Маркетплейс в UI

Раздел в настройках Координатора (рядом с «Сигналы» / подключением к Relay): **«Маркетплейс коннекторов»** (рабочее имя; в коде `relay_marketplace`).

### Внешний вид

- **Плитки** (grid), у каждой коннектора: **иконка**, короткое имя, одна строка описания.
- Плитка **кликабельна целиком**.
- **Подключённый** коннектор визуально отмечен (бейдж «Подключён», обводка/фон, иконка «в сети» — детали UX при реализации).
- **Не подключённый** — нейтральная плитка, призыв «Подключить».

### Поведение по клику

| Состояние | Действие |
|-----------|----------|
| Уже подключён | Модалка **«Настройки»**: поля конфигурации, «Сохранить», «Отключить» (с подтверждением). Секреты — маскированные поля; смена секрета перезаписывает в Relay. |
| Ещё не подключён | Модалка **«Подключение»**: тот же или упрощённый набор полей (у коннекторов **разные** формы — см. `config_schema`). Кнопка «Подключить» → валидация → запись в Relay + `enabled` в проекте. |

Формы не хардкодить одной простынёй: у каждого коннектора в каталоге — **схема полей** (id, label, type: string / secret / url / duration, required, help).

### Права

- Просмотр маркетплейса и статуса «подключён» — участник с доступом к настройкам Relay (как сейчас пара `.relay_access`).
- Сохранение секретов и enable/disable — **admin** Relay (как выдача токенов).

## Данные: три места

```text
┌─────────────────────────────────────────────────────────────┐
│  Каталог (код Relay)                                         │
│  id, name, icon, config_schema, capabilities (webhook|poll) │
└───────────────────────────┬─────────────────────────────────┘
                            │ только чтение в UI
┌───────────────────────────▼─────────────────────────────────┐
│  Настройки проекта (Common/data/settings/, git)                │
│  relay_connectors.enabled: ["github", "llm_model_lifecycle"] │
│  relay_connectors.<id>.non_secret: { poll_interval, … }       │
└───────────────────────────┬─────────────────────────────────┘
                            │ enable/disable, не секреты
┌───────────────────────────▼─────────────────────────────────┐
│  Relay PostgreSQL (отдельная БД, MigratorV2)                 │
│  connector_secrets (AES-GCM), external_events, stats         │
│  master key: RELAY_CONFIG_KEY / .relay_config_key на диске   │
└─────────────────────────────────────────────────────────────┘
```

**Почему так:** `coordinator-state` и `Common main` не должны тащить OpenRouter keys и webhook secrets. Участник команды видит в git только «GitHub включён», не пароль.

Синхронизация: UI сохраняет non-secret в Common (или в bus settings через API координатора); секреты — отдельным вызовом `PUT /api/v1/connectors/{id}/config` на Relay. После сохранения Relay **перечитывает** registry: активны только `enabled` ∩ зарегистрированные в коде.

## Внутренний API Relay (расширение)

Уже есть: `GET /sources`, webhooks, `POST /connectors/{id}/events`.

Добавить (proposal):

| Метод | Назначение |
|-------|------------|
| `GET /api/v1/connectors/marketplace` | Каталог плиток: id, name, detail, icon, connected, enabled, config_schema |
| `GET /api/v1/connectors/{id}/config` | Non-secret + маски секретов (••••) |
| `PUT /api/v1/connectors/{id}/config` | Admin: merge non-secret + новые секреты → encrypt → SQLite |
| `POST /api/v1/connectors/{id}/enable` | Admin: включить (запись в project settings через координатор или relay отдаёт payload для сохранения) |
| `POST /api/v1/connectors/{id}/disable` | Остановить poll/webhook, убрать из enabled |

Координатор backend проксирует marketplace к Relay при наличии `.relay_access`; запись в `relay_connectors` в settings — через существующий путь настроек Common.

## База данных Relay

- Отдельный Postgres: `RELAY_DATABASE_URL` или `RELAY_POSTGRES_*` (см. `relay/infra/migrator/README.md`).
- Миграции: `relay/infra/migrator/` (как Core/WMD); при старте relay — `up` автоматически.
- Таблицы: `connector_secrets`, `external_events` (дублирует поток для поиска и статистики), `connector_daily_stats` (заготовка).
- JSONL в `progress/external/events/` остаётся шиной для Coordinator; в PG пишется параллельно при `Gateway.Publish`.

## Шифрование

- Секреты в `connector_secrets`: `ciphertext`, `nonce`, `updated_at`.
- Ключ: `RELAY_CONFIG_KEY` или автоген при первом старте в gitignore-файле (как bearer token).
- Алгоритм: AES-GCM или age — выбрать при реализации; в пропозале важно **не plaintext на диске**.

Non-secret можно дублировать в SQLite для Relay offline или читать только из project settings — при реализации выбрать один источник truth для poll interval (предпочтение: non-secret в project settings, секреты только SQLite).

## Первые два коннектора

### 1. GitHub (`github`)

- **Для кого:** соседний контур InCountry (ревью, CI).
- **Тип:** webhook (`POST /connectors/github/webhook` или legacy `/github/webhook`).
- **Конфиг:** webhook secret (secret), опционально список org/repo (non-secret, фаза 2).
- **События:** уже реализованы (`review.*`, `ci.*`) → конверт `ExternalEvent`.

### 2. LLM model lifecycle (`llm_model_lifecycle`) — только инстанс Alina

- **Где код:** `.coordinator-relay/connectors/llm_model_lifecycle/` (не stock pack).
- **Для кого:** Alina Assist — мониторинг **окончания срока жизни LLM-моделей** у провайдеров (сейчас OpenRouter; позже другие провайдеры в том же коннекторе).
- **Тип:** **poll** (scheduler в Relay).
- **Логика (реализовано):**
  1. Список моделей: `Common/data/settings/relay_llm_model_watch.json` на `RELAY_DATA` + опционально union с Core `GET /internal/api/v1/relay/llm-model-slugs` (`X-Coordinator-Relay-Key` = Core `COORDINATOR_RELAY_GATEWAY_SECRET`).
  2. OpenRouter `GET /api/v1/models` → поле `expiration_date`.
  3. События: `llm_model.deprecating` / `expired` / `removed`, `data.provider` = `openrouter`.
  4. Доставка: Postgres + `bus_outbox` → push `coordinator-state`; Pulse читает `external.events[]`.
- **Конфиг:** `openrouter_api_key`, `poll_interval_hours`, `warn_days_before`, `core_base_url`, `coordinator_relay_gateway_secret`, `merge_core_models`.

## Жизненный цикл коннектора

```text
Плитка в маркетплейсе (каталог в коде)
    → Подключение (модалка, admin)
    → Секреты → Relay SQLite (encrypt)
    → enabled в project settings
    → Relay стартует webhook route / poll goroutine
    → События → Gateway → PG + bus_outbox → coordinator-state
    → Координатор pull state → Pulse `external.events[]`
```

**Отключение:** stop poll, disable в settings, секреты можно оставить или стереть (выбор в модалке «Отключить»).

## Дистрибуция и крупные клиенты

- **Stock Relay:** в бинарь входят все коннекторы маркетплейса; клиент включает нужные в UI.
- **Custom build:** тот же UI, в `main` регистрируется урезанный `NewRegistry(...)` — в маркетплейсе видны только разрешённые плитки.
- **Sidecar-коннектор:** в каталоге нет in-process реализации; плитка помечена «внешний», инструкция + `POST /events` (как в родительском пропозале).

## UI: копирайт и иконки

- Иконки: Lucide или статические SVG в `frontend` по `connector_id` (github → brand-safe generic git; openrouter → custom или нейтральная «cloud»).
- Не дублировать длинные пути и «локальная машина» в подзаголовках (`coordinator-ui-copy.mdc`).
- RU/EN в `locales` для названий плиток и модалок.

## Фазы

### Фаза 0 — этот документ

Согласовать маркетплейс, три слоя данных, два первых коннектора.

### Фаза 1 — каркас

- Postgres + MigratorV2 + encrypt на Relay;
- `GET /api/v1/connectors/marketplace`, enable/disable;
- UI: grid плиток, модалка подключения для GitHub (secret + test webhook);
- `relay_connectors` в project settings (schema в Common).

### Фаза 2 — Alina OpenRouter

- Poll-коннектор + события deprecation;
- Форма с интервалом и key;
- Core endpoint (отдельный task в Core) — опциональная интеграция.

### Фаза 3

- Pulse читает `external.events`;
- напоминания Отто по `class` / kind;
- фильтры по repo / model в UI.

## Не входит в v1 маркетплейса

- Загрузка коннекторов как `.so` / плагины без пересборки.
- Публикация сторонних коннекторов вне git-репозитория Relay.
- Хранение секретов в `team.json` или `coordinator-state`.
- Полноценный «магазин» с версиями и оплатой — только метафора UI, не billing.

## Открытые решения

1. Non-secret коннектора: только `Common/settings` или дубль в Relay DB?
2. Кто пишет `enabled[]` — только Coordinator API при «Подключить» или Relay пушит в Common (team mode)?
3. Иконки GitHub/OpenRouter: generic vs brand guidelines.
4. Core endpoint: один общий secret на Relay или per-environment.

## Репозитории (pack + инстанс)

| Путь | Git | Роль |
|------|-----|------|
| `.coordinator/relay/` | открытый Coordinator | Upstream pack, `PACK_VERSION`, обновления ядра |
| `.coordinator-relay/` | `Alina-Assist/coordinator-relay` | Инстанс Alina: копия pack + `connectors/` |
| Синхронизация | `./.coordinator/utils/relay_sync_pack.sh` | Накат pack в инстанс (не трогает `connectors/`, README инстанса) |
| Инструкция | `Common/docs/CoordinatorRelay/20261020-1925-EK-RELAY_PACK_SYNC.md` | Pack → sync → `go test` → push coordinator-relay → пересборка / Coolify |

Целевой шаг: `relay-core` как Go-модуль; до перехода — pack + sync script + явный цикл пересборки инстанса.

## Связь с кодом (на момент пропозала)

Реализовано: `connector.Registry`, `gateway.Gateway`, конверт `ExternalEvent`, GitHub in-process, маршруты webhook/events.

Реализовано (фаза 1): `GET /api/v1/connectors/marketplace`, PUT/GET config, disable, Postgres (`connector_secrets`, `external_events`), AES-GCM, `relay_connectors.json`, UI плиток и модалка GitHub, прокси в Coordinator.

Не реализовано: test webhook, отдельные процессы-плагины, провайдеры кроме OpenRouter в `llm_model_lifecycle`.
