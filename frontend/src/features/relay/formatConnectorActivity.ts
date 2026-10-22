import type { TFunction } from 'i18next'
import type { ConnectorActivityEntry } from './connectorActivityTypes'

function num(detail: Record<string, unknown> | undefined, key: string): number | undefined {
  const v = detail?.[key]
  if (typeof v === 'number' && !Number.isNaN(v)) return v
  return undefined
}

export function formatConnectorActivityText(entry: ConnectorActivityEntry, t: TFunction): string {
  const d = entry.detail
  if (entry.kind === 'poll_check' && d) {
    const trigger =
      d.trigger === 'manual' ? t('relay.activity.triggerManual') : t('relay.activity.triggerScheduled')
    const checked = num(d, 'models_checked') ?? 0
    const noExp = num(d, 'no_expiration')
    const published = num(d, 'events_published') ?? 0
    const removed = num(d, 'removed') ?? 0
    const deprecating = num(d, 'deprecating') ?? 0
    const expired = num(d, 'expired') ?? 0
    if (published > 0 || removed > 0 || deprecating > 0 || expired > 0) {
      return t('relay.activity.pollCheckWithFindings', {
        trigger,
        checked,
        published,
        removed,
        deprecating,
        expired,
      })
    }
    if (noExp !== undefined && noExp === checked) {
      return t('relay.activity.pollCheckAllNoExpiration', { trigger, checked })
    }
    return t('relay.activity.pollCheckOk', { trigger, checked })
  }
  if (entry.kind === 'poll_error') {
    return entry.summary
  }
  return entry.summary
}

/** Compact technical one-liner for table cells. */
export function formatConnectorActivityShort(entry: ConnectorActivityEntry): string {
  const d = entry.detail
  if (entry.kind === 'poll_check' && d) {
    const trigger = d.trigger === 'manual' ? 'manual' : d.trigger === 'scheduled' ? 'scheduled' : String(d.trigger ?? '?')
    const checked = num(d, 'models_checked')
    const published = num(d, 'events_published') ?? 0
    const removed = num(d, 'removed') ?? 0
    const expired = num(d, 'expired') ?? 0
    const deprecating = num(d, 'deprecating') ?? 0
    const noExp = num(d, 'no_expiration')
    let s = `trigger=${trigger}`
    if (checked !== undefined) s += ` models_checked=${checked}`
    if (noExp !== undefined) s += ` no_expiration=${noExp}`
    s += ` published=${published} removed=${removed} expired=${expired} deprecating=${deprecating}`
    return s
  }
  if (entry.summary.trim()) return entry.summary.trim()
  return entry.kind
}
