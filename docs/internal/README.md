# Internal docs (not for installers)

Эти файлы **не** публичные инструкции. Пишем по-русски. Установщик читает только `docs/INSTALL.md`, `ONBOARDING.md`, `DEPENDENCIES.md` (English).

Продуктовые записки Координатора живут **здесь**, не в `Common/docs` Alina Assist. Как Alina-команда пользуется Координатором (голос, spend, protocol) — по-прежнему в `Common/docs` и `.cursor/rules`.

## Продукт

- [COMMERCIAL_CONCEPT.md](./COMMERCIAL_CONCEPT.md) — продукт и коммерция
- [PUBLIC_ONBOARDING_PLAN.md](./PUBLIC_ONBOARDING_PLAN.md) — план беты установки

## Модель и шина

- [Доменная модель](./20260915-1805-EK-COORDINATOR_DOMAIN_MODEL.md) — Resource / Claim / handoff
- [Data layer](./20260907-2225-EK-COORDINATOR_DATA_LAYER.md) — `Common/data/`, jsonl, SQLite-кэш
- [Параллельные слоты](./20260909-1423-EK-COORDINATOR_PARALLEL_TASKS.md)
- [Peer scope](./20260912-1121-EK-COORDINATOR_PEER_SCOPE.md)
- [Status bus](./20260910-0732-EK-COORDINATOR_STATUS_BUS_ALTERNATIVES.md) — `coordinator-state`
- [Автоматизации на пульсе](./20260926-1417-EK-COORDINATOR_AUTOMATIONS.md)

## Организации и внешний контур

- [Команды и CrewConnector](./20261013-1711-EK-COORDINATOR_CREWS_AND_CONNECTORS.md)
- [Внешние знания и события (Relay)](./20261013-1741-EK-COORDINATOR_SYSTEM_CONNECTOR.md)
- [Relay: маркетплейс коннекторов](./20261020-1808-EK-RELAY_CONNECTOR_MARKETPLACE.md) — плитки UI, настройки проекта, секреты в Relay DB, GitHub + OpenRouter
