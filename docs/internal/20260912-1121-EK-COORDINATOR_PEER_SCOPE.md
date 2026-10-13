# Чужой репо — предупреждение, свой — стоп

Author: **Evgeny KOSIVTSOV** (`EK`)
Created: 2026-09-12 11:21
Branch: `main`
Services: `.coordinator`, `.cursor`, `Common`

Конфликт — один working tree, не «один репозиторий на команду». Предупреждения пишем в журнал событий — след решения Отто.

Связано: [параллельные слоты](./20260909-1423-EK-COORDINATOR_PARALLEL_TASKS.md), [голос](../../../Common/docs/20260909-1629-EK-COORDINATOR_VOICE.md).

## Модель

| Ситуация | Свой слот | Чужой слот |
|---|---|---|
| Тот же product-репо, разные ветки | стоп | предупредить, стартовать |
| Одна ветка (не `main`) | стоп | стоп |
| Тот же `task_id` / тот же FIX-топик | стоп | стоп и спросить |
| Похожий смысл, тот же модуль | стоп и спросить | предупредить |
| `Common` | не считается | не считается |
| `.cursor` / `.coordinator` | стоп | не считается |

Свой стоп: на диске один `HEAD`. Чужой warning: разные клоны, разные ветки; merge позже. Отто называет чужую задачу по смыслу, не «Core overlap».

## Журнал

События jsonl:

- `coordinator_warning` — warning (чужой тот же product-репо, похожий смысл)
- `coordinator_stop` — critical (свой overlap, одна ветка, один `task_id`)

Поля: `alias` (кто увидел), `task_id`, `service`, `summary` (текст как в радаре), `findings` (отпечаток для дедупа). Одна и та же коллизия не пишется повторно, пока жива.

Агент при чужом product-репо: говорит warning и вызывает `sync_event.sh … coordinator_warning`. Дашборд при Pulse дописывает строку, если отпечатка ещё нет.

## Пульс

Радар: critical — красный стоп; только warning — янтарный. Timeline показывает `coordinator_warning` / `coordinator_stop`.
