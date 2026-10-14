import { unixFromISO, startOfLocalDay } from '../metrics/cursorSpend'
import type { HourParticipant, HourSpend } from '../../shared/types/api'

export function hourBelongsToUser(hour: HourSpend, alias?: string): boolean {
  if (!alias) return true
  if (hour.alias === alias) return true
  if (hour.aliases?.includes(alias)) return true
  if (hour.participants?.some((p) => p.alias === alias)) return true
  const tagged = Boolean(hour.alias) || (hour.aliases && hour.aliases.length > 0) || hour.participants?.some((p) => p.alias)
  return !tagged
}

export function filterPersonalHours(hours: HourSpend[] | undefined, alias?: string, fromUnix = 0): HourSpend[] {
  const out: HourSpend[] = []
  for (const hour of hours ?? []) {
    if (fromUnix > 0 && hour.hour + 3600 <= fromUnix) continue
    if (!hourBelongsToUser(hour, alias)) continue
    out.push(hour)
  }
  return out
}

export function padTodayPersonalHours(hours: HourSpend[] | undefined, alias?: string, now = new Date()): HourSpend[] {
  const cycleStart = startOfLocalDay(now)
  const filtered = filterPersonalHours(hours, alias, 0)
  const byHour = new Map<number, HourSpend>()
  for (const hour of filtered) {
    byHour.set(hour.hour, hour)
  }
  const out: HourSpend[] = []
  for (let i = 0; i < 24; i += 1) {
    const ts = cycleStart + i * 3600
    out.push(
      byHour.get(ts) ?? { hour: ts, cursor_models_pct: 0, other_models_pct: 0, ondemand_usd: 0, participants: [] }
    )
  }
  return out
}

function startOfDayUnix(ts: number): number {
  const d = new Date(ts * 1000)
  return Math.floor(new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime() / 1000)
}

export function buildPersonalCycleDays(
  hours: HourSpend[] | undefined,
  alias?: string,
  billingCycleStartISO?: string,
  now = new Date()
): { rows: HourSpend[]; periodStart: number; periodEnd: number } {
  const loc = now
  const todayStart = startOfLocalDay(loc)
  let periodStart = unixFromISO(billingCycleStartISO)
  if (periodStart <= 0 || periodStart > todayStart) {
    periodStart = todayStart
  } else {
    periodStart = startOfDayUnix(periodStart)
  }
  const periodEnd = todayStart

  const filtered = filterPersonalHours(hours, alias, periodStart)
  const byDay = new Map<number, HourSpend>()
  const parts = new Map<number, Map<string, HourParticipant>>()

  for (const hour of filtered) {
    const day = startOfDayUnix(hour.hour)
    let agg = byDay.get(day)
    if (!agg) {
      agg = { hour: day, cursor_models_pct: 0, other_models_pct: 0, ondemand_usd: 0, participants: [] }
      byDay.set(day, agg)
      parts.set(day, new Map())
    }
    agg.cursor_models_pct += hour.cursor_models_pct ?? 0
    agg.other_models_pct += hour.other_models_pct ?? 0
    agg.ondemand_usd += hour.ondemand_usd ?? 0
    if (hour.budget_usd != null) {
      agg.budget_usd = (agg.budget_usd ?? 0) + hour.budget_usd
    }
    for (const p of hour.participants ?? []) {
      const key = `${p.kind}\0${p.id || p.alias}`
      parts.get(day)!.set(key, p)
    }
  }

  const rows: HourSpend[] = []
  for (let ts = periodStart; ts <= periodEnd; ts += 86400) {
    const agg = byDay.get(ts)
    if (agg) {
      agg.participants = [...(parts.get(ts)?.values() ?? [])]
      rows.push(agg)
    } else {
      rows.push({ hour: ts, cursor_models_pct: 0, other_models_pct: 0, ondemand_usd: 0, participants: [] })
    }
  }
  return { rows, periodStart, periodEnd }
}

export function formatPeriodRange(startUnix: number, endUnix: number, language: string): string {
  const fmt = (ts: number) =>
    new Date(ts * 1000).toLocaleDateString(language.startsWith('ru') ? 'ru-RU' : 'en-US', {
      day: 'numeric',
      month: 'short',
    })
  if (startUnix === endUnix) return fmt(startUnix)
  return `${fmt(startUnix)} — ${fmt(endUnix)}`
}
