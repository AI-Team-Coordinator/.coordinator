# Coordinator: единый data-слой и локальное хранилище событий

Author: **Evgeny KOSIVTSOV** (`EK`)
Created: 2026-09-07 22:25
Branch: `main`
Services: Common, `.cursor`

Связанные документы: [AI Team Coordinator](../../../Common/docs/20260907-1758-EK-AI_TEAM_COORDINATOR.md), [UI architecture](../../../Common/docs/20260907-1848-EK-COORDINATOR_UI_ARCHITECTURE.md), [жизненный цикл задачи](../../../Common/docs/20260907-2207-EK-COORDINATOR_TASK_LIFECYCLE.md).

## Зачем

Сейчас данные координатора размазаны по трём местам:

| Где | Что |
|-----|-----|
| `Common/progress/` | снимок задачи, jsonl событий |
| `Common/logs/` | сырые логи Coolify, SQL, отчёты LogAnalyzer |
| `.cursor/` | `team.json`, `project_profile.json` (+ IDE: rules, hooks) |

Отсюда два git-пуша на смену состава команды, два корня у скриптов и ощущение, что «журнал» — это файл, который неудобно фильтровать.

Нужно: **один каталог правды в Common**, **лёгкий запросный слой у бэкенда**, без второго источника правды в бинарной БД в git.

## Решение в двух частях

### A. Каталог `Common/data/` — всё, что команда синхронизирует через git

```
Common/data/
  .current_author               # gitignore: кто за этой машиной
  .current_author.example       # шаблон, в git
  settings/
    team.json
    project_profile.json
  progress/
    .current_task_<ALIAS>
    events/<ALIAS>/2026.jsonl   # один файл на год, не бесконечный EK.jsonl
    events/<ALIAS>/2027.jsonl
  logs/                         # gitignore: сырые дампы
  .cache/                       # gitignore: sqlite
```

Скрипты (`.coordinator/utils/sync_event.sh`, `.coordinator/utils/log_event.sh`, `coolify_fetch_logs.sh`, `prod_db_query.sh`, Coordinator API) читают и пишут только сюда. Корень данных **один**: `Common/data`.

**Переезд настроек из `.cursor`:** `team.json` и `project_profile.json` — данные продукта, не конфиг IDE. После сохранения состава команды или создания единицы Coordinator сам коммитит whitelist и пушит **Common** (`main`).

**`.current_author` тоже в `Common/data/`.** Cursor его не читает (это не rules/hooks). Его читают агент, хук ветки, `docs.mdc`, `sync_event.sh`, Coordinator. Логично держать рядом с `team.json`: «кто есть в команде» vs «кто сейчас за этим клоном». Файл **не коммитить** — у каждого разработчика свой алиас; в git только `.current_author.example`. В `Common/.gitignore`: `data/.current_author`, `data/logs/`, `data/.cache/`.

**Что остаётся в `.cursor`:** только то, без чего IDE слепа — `rules/`, `hooks/`, `skills/`.

**`author.md`:** не держать вторым источником. Алиасы — в `team.json`. Документы либо читают JSON, либо `author.md` генерируется из него (один write-path).

**Что не складывать как «логи коммитов»:** история git уже в каждом сервисном репо. В `progress/events` пишем только **намеренные события** (`task_started`, `task_completed`, `repo_merged`, `deploy_*`). Не зеркалировать каждый `git commit`. Merge сервиса в журнал пишет Cursor-хук, не git hook в каждом репо.

Сырые логи Coolify / SQL — в `data/logs/`, **не коммитить** (объём). В git только settings + progress.

### B. SQLite рядом с бэкендом — не файл в git

Идея «поднять лёгкую БД вместе с координатором и пушить файл БД» даёт индексы и SQL, но **плохой источник правды в git**:

- sqlite — бинарник: конфликт двух `git pull` почти не смерживается;
- diff в PR нечитаемый;
- два агента на `Common/main` одновременно ломают файл (`-wal`/`-shm`);
- это ломает модель «Common = текстовая шина».

**Как сделать и индексы, и командную синхронизацию:**

1. **Источник правды в git** — append-only jsonl по человеку и **году** (`events/EK/2026.jsonl`). Два append в конец файла обычно мержатся; у каждого алиаса свой файл.
2. **При старте Coordinator** — открыть **локальный** SQLite (`Common/data/.cache/coordinator.sqlite`, gitignore, или `:memory:`). Прочитать jsonl текущего года (и N прошлых при необходимости), заполнить таблицы и индексы по `alias`, `task_id`, `event`, `timestamp`, `service`.
3. **API журнала и Stats** — SELECT из SQLite. Новые события: append в jsonl **и** INSERT в SQLite. Пуш в git — текстовые файлы, не `.sqlite`.
4. Если кэш отстал или файл побит — удалить `.cache` и пересобрать из jsonl.

### Насколько раздуется jsonl

Журнал координатора — это метаданные (~200–400 байт на строку), не логи Coolify.

Грубая оценка на одного человека: 20–40 событий в рабочий день (старт/финиш задачи, деплой). За год ≈ 5–10 тысяч строк ≈ **1–4 МБ**. Команда из пяти человек — десятки мегабайт в git за год, если никогда не резать.

Для Go прочитать 10 МБ jsonl при старте — миллисекунды. Упирается не парсер, а **вечный рост одного файла в git** (каждый push трогает всё более длинный `EK.jsonl`).

Поэтому не один `EK.jsonl` навсегда, а **ротация по календарному году**:

- писать только в `events/<ALIAS>/<YYYY>.jsonl`;
- прошлые годы — замороженные файлы, их почти не коммитят;
- SQLite при старте читает текущий год всегда и N прошлых по необходимости (Stats «за месяц» не нужен 2019-й).

Если когда-нибудь понадобится архив старше N лет — вынести `events/*/2024.jsonl` из активного клона (отдельный архив / git history уже хранит). До этого порога года резать не обязательно.

Не класть в jsonl тела логов деплоя — только `event`, `task_id`, `service`, `status`. Тогда файл не раздувается «случайным» stdout Coolify.

Пока событий мало, jsonl и так быстрый. SQLite нужен, когда появятся фильтры «деплои Core за неделю», «cycle time по человеку», пагинация журнала. Архитектуру (год + кэш) закладываем сразу, драйвер можно подключить во второй итерации.

Драйвер: чистый Go (`modernc.org/sqlite` или аналог), без CGO — Coordinator остаётся одним бинарником.

Не предлагаем Postgres/Redis: координатор local-first, без облачного сервиса.

## Слои (как сейчас, только пути)

1. **Снимок** — `data/progress/.current_task_*` (пульс, `services[]`, опционально `doc` / `summary` — intent задачи, опционально `git_report` — метаданные локального git без диффов).
2. **Факты git** — живой осмотр реп (local / pushed / merged). Не дублировать файлом.
3. **Действия** — jsonl → (опционально) SQLite для запросов.
4. **Деплой** — те же jsonl через `log_event.sh` из `coolify_deploy.sh`, не отдельная «база деплоев».

Гейт чистого `main` и незакрытая задача — [lifecycle](../../../Common/docs/20260907-2207-EK-COORDINATOR_TASK_LIFECYCLE.md) / `coordinator/protocol.mdc`; от data-слоя не зависят.

## Этапы

1. **Каталог и переезд** — сделано: `Common/data/`, settings из `.cursor`, progress/jsonl по году, хук и скрипты смотрят сюда. Сохранение команды / новой единицы само пушит Common.
2. **Кэш SQLite в бэкенде** — сделано: `infra/db`, файл `data/.cache/coordinator.sqlite` (gitignore), при отсутствии пересобирается из jsonl. API `/events` и `/stats` читают SQL.
3. **Журнал в UI** — сделано: фильтры alias / тип события / сервис, пагинация.

## Не делаем

- Коммит `*.sqlite` в Common как источник правды.
- Отдельный сервер БД.
- Перенос `rules/` / `hooks/` из `.cursor`.
- Запись каждого git-коммита в jsonl.

## Критерий готовности этапа 1

- Состав команды и карта сервисов редактируются в Coordinator и уезжают одним `git push` Common.
- Хук ветки и `sync_event.sh` смотрят только в `Common/data/progress/`.
- `.cursor` без `team.json` / `project_profile.json` / `.current_author`.
- Алиас машины читается из `Common/data/.current_author` (файл не в git).
