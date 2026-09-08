# AI Team Coordinator

Локальный дашборд команды. Живёт рядом с `.cursor`: служебный корень воркспейса, не сервис продукта.

Данные и доки **не** здесь. Их пути задаются в `.env` (скопировать `.env.example`).

```bash
cp .env.example .env   # поправить пути, если дерево проекта другое
./dev.sh               # API :4321 + Vite :5175
./run.sh               # API + собранный UI на :4321
```

Для Alina Assist по умолчанию:

| Переменная | Смысл | Пример |
|---|---|---|
| `WORKSPACE_ROOT` | корень клона со всеми репами | `..` |
| `DATA_DIR` | настройки, progress, jsonl, кэш SQLite | `../Common/data` |
| `DOCS_DIR` | документы задач | `../Common/docs` |
| `BUS_DIR` | git-шина (pull/push настроек) | `../Common` |
| `CURSOR_DIR` | правила IDE, fallback автора | `../.cursor` |
| `PORT` | HTTP API | `4321` |

Скрипты записи событий (`sync_event.sh`, `log_event.sh`) в этом каталоге читают тот же `.env`. Хуки Cursor и `coolify_deploy.sh` вызывают их отсюда.

Контракт шины: [data layer](../Common/docs/20260907-2225-EK-COORDINATOR_DATA_LAYER.md), [lifecycle](../Common/docs/20260907-2207-EK-COORDINATOR_TASK_LIFECYCLE.md).
