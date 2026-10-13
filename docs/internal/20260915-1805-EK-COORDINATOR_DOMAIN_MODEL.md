# Доменная модель Координатора

Author: **Evgeny KOSIVTSOV** (`EK`)
Created: 2026-09-15 18:05
Branch: `main`
Services: `.coordinator`, `.cursor`, `Common`

Первый срез: вынести Resource и Claim из `services[]`, не переезжать всеми пакетами сразу.

Связано: [параллельные слоты](./20260909-1423-EK-COORDINATOR_PARALLEL_TASKS.md), [чужой репо](./20260912-1121-EK-COORDINATOR_PEER_SCOPE.md), [data layer](./20260907-2225-EK-COORDINATOR_DATA_LAYER.md).

## Зачем

Pulse, радар и git-инспекция уже говорят «claim», но в коде это CSV сервисов у слота. Из одной строки одновременно следует намерение, стоп своему второму чату и предупреждение коллеге. Нет способа сказать «мне нужен чужой контракт, я им не владею».

## Восемь понятий

| Сущность | Смысл | Сейчас |
|---|---|---|
| **Person** | Человек (`alias`) | `domain/identity` |
| **Agent** | Исполнитель внутри Person (чат Cursor) | `Claim.Agents` из `session_ids` / вкладок; Pulse `agents[]` |
| **Session** | Вкладка Composer | `session_id` агента |
| **Task** | Работа; слот и статистика — проекции | пока `model.Task` |
| **Resource** | Зона: сервис, опционально путь | `domain/coordination` |
| **Claim** | «В этой задаче это моя зона» | проекция из `services[]` + исполнители |
| **Lock** | Политика overlap: свой product/workspace → stop, чужой product → warn | не отдельный объект |
| **Dependency** | «Мне нужно то, что здесь происходит» | тип, ещё не пишется в снимок |
| **Handoff** | Передача зоны или задачи другому человеку/агенту | событие `handoff` |
| **Decision** | Короткий «почему»: выбор и отвергнутое | событие `decision` |
| **Event** | Журнал jsonl | `domain/timeline` + `model.Event` |

Research остаётся рядом с Task, не в этой восьмёрке.

## Handoff и Decision

Не отдельные таблицы. Строка jsonl, payload типизирован в `domain/timeline`.

```text
./.coordinator/utils/sync_event.sh EK handoff <TASK_ID> <branch> Core to_alias=AS summary="passing billing" session_id=<from>
./.coordinator/utils/sync_event.sh EK decision <TASK_ID> main summary="advisory claims" findings="rejected file locks"
```

Поля: `agent_id` (кто отдал / кто записал), `to_alias` / `to_agent` (кому). Handoff сам не вычёркивает сервис из `services[]` — снимок по-прежнему живой.

Overlap по-прежнему между Person+Task, не между двумя агентами одной задачи.

## Agent

Pulse рисует человека. Агент — Composer-вкладка, которая исполняет claims этой задачи. Несколько `session_ids` на слоте — несколько исполнителей, одна зона.

## Resource

Иерархия, не файл:

```text
LLM
LLM/stt
billing/invoices
```

`Common` — шина, в overlap не входит. `.cursor` / `.coordinator` — workspace: свой второй слот стопается, чужой нет.

Файлы не клеим: это lock-хук ветки, не зона ответственности.

## Claim vs Lock vs Dependency

- **Claim** — намерение. Карта и предупреждение.
- **Lock** — реакция на пересечение claims, не вторая запись в git.
- **Dependency** — другое ребро: ждать/читать, не владеть. Первая запись в снимок: `related_tasks[]` / `related_docs[]` на фиксах (ребёнок → предки). Reopen той же фичи — тот же `task_id`, не новое ребро.

Снимок `.current_task_*`: `services[]` — источник claims. Можно писать зону (`LLM/stt`), не только репо. Агенты пишут `handoff` и `decision` в jsonl по правилу `coordinator/claims.mdc`; лента на доске их показывает.

## Пакеты

```text
backend/domain/identity/        Person, Agent, Session
backend/domain/coordination/    Resource, Claim, Dependency, overlap
backend/domain/timeline/        Kind, Handoff, Decision
backend/domain/coordinator/     HTTP, Pulse, аналитика — вызывает overlap
```

Не восемь тонких пакетов. FinOps, setup, GitHub — не тащить в эту модель.

## Не входит

- клейм файлов
- Agent как субъект команды вместо Person
- авто-handoff при `task_completed` (закрытие задачи — снятие claims, не передача)
- проекция claims только из событий, без `services[]`
