# AI Team Coordinator

Локальный дашборд команды. Живёт рядом с `.cursor`: служебный корень воркспейса, не сервис продукта.

Данные и доки **не** здесь. Их пути задаются в `.env` (скопировать `.env.example`).

Смотреть дашборд: **http://localhost:4321** (`./utils/run.sh` пересобирает Go + UI и поднимает всё на этом порту). Vite `:5175` — не адрес для проверки.

```bash
cp .env.example .env   # поправить пути, если дерево проекта другое
./utils/run.sh         # пересобрать и открыть http://localhost:4321
```

Для Alina Assist по умолчанию:

| Переменная | Смысл | Пример |
|---|---|---|
| `WORKSPACE_ROOT` | корень клона со всеми репами | `..` |
| `DATA_DIR` | настройки, progress, jsonl, кэш SQLite | `../Common/data` |
| `DOCS_DIR` | документы задач | `../Common/docs` |
| `BUS_DIR` | git-шина (pull/push настроек) | `../Common` |
| `CURSOR_DIR` | правила IDE, fallback автора | `../.cursor` |
| `PORT` | дашборд (Go + собранный UI) | `4321` |

Скрипты записи событий и запуска — в `utils/` (`sync_event.sh`, `log_event.sh`, `run.sh`). Хуки Cursor и `coolify_deploy.sh` вызывают их оттуда.

Контракт шины: [data layer](../Common/docs/20260907-2225-EK-COORDINATOR_DATA_LAYER.md), [lifecycle](../Common/docs/20260907-2207-EK-COORDINATOR_TASK_LIFECYCLE.md).
