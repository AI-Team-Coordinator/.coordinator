import type { ExternalEvent } from '../../shared/types/api'

export function externalEventNeedsAction(ev: ExternalEvent): boolean {
  if (!ev.requires_reaction) return false
  const st = ev.triage?.status
  return st === 'open' || st === 'in_progress' || !st
}

export function formatExternalEventShort(ev: ExternalEvent): string {
  const raw = (ev.summary || ev.title || '').trim()
  if (!raw) return ev.kind
  if (raw.length <= 140) return raw
  return `${raw.slice(0, 137)}…`
}
