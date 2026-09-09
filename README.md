# AI Team Coordinator

Локальный дашборд команды. Живёт рядом с `.cursor`: служебный корень воркспейса, не сервис продукта.

Данные и доки **не** здесь. Их пути задаются в `.env` (скопировать `.env.example`).

Смотреть дашборд: **http://localhost:5175** (`npm run dev` / `./utils/frontend.sh start`). API: **http://127.0.0.1:4321** (`./utils/backend.sh`). Оба процесса отвязаны от чата Cursor и не гаснут при закрытии вкладки.

```bash
cp .env.example .env   # поправить пути, если дерево проекта другое
./utils/run.sh         # поднять API (если нет) + Vite HMR
# после правок Go:
./utils/backend.sh restart
```

| Переменная | Смысл | Пример |
|---|---|---|
| `WORKSPACE_ROOT` | корень клона со всеми репами | `..` |
| `DATA_DIR` | настройки, progress, jsonl, кэш SQLite | `../Common/data` |
| `DOCS_DIR` | документы задач | `../Common/docs` |
| `BUS_DIR` | git-шина (pull/push настроек) | `../Common` |
| `CURSOR_DIR` | правила IDE, fallback автора | `../.cursor` |
| `PORT` | Go API | `4321` |

Скрипты: `utils/backend.sh`, `utils/frontend.sh`, `utils/run.sh`, `sync_event.sh`. Хуки Cursor вызывают `sync_event.sh` оттуда.

Контракт шины: [data layer](../Common/docs/20260907-2225-EK-COORDINATOR_DATA_LAYER.md), [lifecycle](../Common/docs/20260907-2207-EK-COORDINATOR_TASK_LIFECYCLE.md).
