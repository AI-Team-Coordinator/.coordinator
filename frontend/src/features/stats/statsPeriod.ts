import { unixFromISO, startOfLocalDay } from '../metrics/cursorSpend'
import type { BillingPeriod } from '../../shared/types/api'

export type StatsPeriodPreset = 'billing' | '7d' | '30d' | 'custom'

export function statsURL(base: string, from?: number, to?: number): string {
  const params = new URLSearchParams()
  if (from && from > 0) params.set('from', String(from))
  if (to && to > 0) params.set('to', String(to))
  const q = params.toString()
  return q ? `${base}?${q}` : base
}

export function periodRangeForPreset(
  preset: StatsPeriodPreset,
  billingCycleStartISO?: string,
  customFrom?: string,
  customTo?: string,
  now = new Date()
): { from?: number; to?: number } {
  const today = startOfLocalDay(now)
  if (preset === 'billing') {
    const from = unixFromISO(billingCycleStartISO)
    return from > 0 ? { from } : {}
  }
  if (preset === '7d') {
    return { from: today - 6 * 86400 }
  }
  if (preset === '30d') {
    return { from: today - 29 * 86400 }
  }
  const from = customFrom ? Math.floor(new Date(customFrom).getTime() / 1000) : 0
  const to = customTo ? Math.floor(new Date(customTo).getTime() / 1000) : 0
  return { from: from > 0 ? from : undefined, to: to > 0 ? to : undefined }
}

export function billingPeriodMarkersOnDays(dayRows: { hour: number }[], periods: BillingPeriod[]): number[] {
  if (!periods.length || !dayRows.length) return []
  const starts = new Set(periods.map((p) => startOfLocalDay(new Date(p.start_unix * 1000))))
  const markers: number[] = []
  for (const row of dayRows) {
    if (starts.has(row.hour)) markers.push(row.hour)
  }
  return markers
}
