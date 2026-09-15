import type { HourSpend, Stats } from '../../shared/types/api'

export function unixFromISO(raw?: string): number {
  if (!raw) return 0
  const ms = Date.parse(raw)
  return Number.isFinite(ms) ? Math.floor(ms / 1000) : 0
}

export function startOfLocalDay(now = new Date()): number {
  return Math.floor(new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime() / 1000)
}

/** Account-meter totals for hours that overlap [fromUnix, …]. */
export function sumHourlySpend(
  hours: HourSpend[] | undefined,
  fromUnix = 0
): { budget: number; cursor: number; other: number; ondemand: number; n: number } {
  let budget = 0
  let cursor = 0
  let other = 0
  let ondemand = 0
  let n = 0
  for (const hour of hours ?? []) {
    if (fromUnix > 0 && hour.hour + 3600 <= fromUnix) continue
    n += 1
    budget += hour.budget_usd ?? 0
    cursor += hour.cursor_models_pct ?? 0
    other += hour.other_models_pct ?? 0
    ondemand += hour.ondemand_usd ?? 0
  }
  return { budget, cursor, other, ondemand, n }
}

export function cursorSpendFromStats(stats: Stats | null) {
  const taskCycle = stats?.budget_usd_cycle ?? 0
  const researchCycle = stats?.budget_usd_research_cycle ?? 0
  const eventPlanCycle = taskCycle + researchCycle
  const ondemandTaskCycle = stats?.cost_usd_cycle ?? 0
  const ondemandResearchCycle = stats?.cost_usd_research_cycle ?? 0
  const eventOndemandCycle = ondemandTaskCycle + ondemandResearchCycle
  const eventCursorCycle = (stats?.cursor_models_pct_cycle ?? 0) + (stats?.cursor_models_pct_research_cycle ?? 0)
  const eventOtherCycle = (stats?.other_models_pct_cycle ?? 0) + (stats?.other_models_pct_research_cycle ?? 0)
  const eventCursorToday = (stats?.cursor_models_pct_today ?? 0) + (stats?.cursor_models_pct_research_today ?? 0)
  const eventOtherToday = (stats?.other_models_pct_today ?? 0) + (stats?.other_models_pct_research_today ?? 0)
  const eventPlanToday = (stats?.budget_usd_today ?? 0) + (stats?.budget_usd_research_today ?? 0)
  const eventOndemandToday = (stats?.cost_usd_today ?? 0) + (stats?.cost_usd_research_today ?? 0)
  const cursorOpen = (stats?.cursor_models_pct_open ?? 0) + (stats?.cursor_models_pct_research_open ?? 0)
  const otherOpen = (stats?.other_models_pct_open ?? 0) + (stats?.other_models_pct_research_open ?? 0)

  const meterCycle = sumHourlySpend(stats?.hours, unixFromISO(stats?.billing_cycle_start))
  const meterToday = sumHourlySpend(stats?.hours, startOfLocalDay())
  const useMeter = meterCycle.n > 0
  const planCycle = useMeter ? meterCycle.budget : eventPlanCycle
  const ondemandCycle = useMeter ? meterCycle.ondemand : eventOndemandCycle
  const cursorCycle = useMeter ? meterCycle.cursor : eventCursorCycle
  const otherCycle = useMeter ? meterCycle.other : eventOtherCycle
  const planToday = useMeter ? meterToday.budget : eventPlanToday
  const ondemandToday = useMeter ? meterToday.ondemand : eventOndemandToday
  const cursorToday = useMeter ? meterToday.cursor : eventCursorToday
  const otherToday = useMeter ? meterToday.other : eventOtherToday

  return {
    hasPlanPrice: Boolean(stats?.plan_price_usd && stats.plan_price_usd > 0),
    hasPlan: Boolean(stats && (stats.budget_tasks > 0 || planCycle > 0)),
    hasPools: cursorCycle > 0 || otherCycle > 0 || cursorToday > 0 || otherToday > 0,
    hasOnDemand: Boolean(
      (useMeter && meterCycle.ondemand > 0) || (stats && (stats.cost_tasks > 0 || ondemandResearchCycle > 0))
    ),
    planCycle,
    ondemandCycle,
    taskCycle,
    researchCycle,
    productCycle: stats?.budget_usd_product_cycle ?? 0,
    infraCycle: stats?.budget_usd_infra_cycle ?? 0,
    planToday,
    ondemandToday,
    planOpen: stats?.budget_usd_open ?? 0,
    ondemandOpen: (stats?.cost_usd_open ?? 0) + (stats?.cost_usd_research_open ?? 0),
    ondemandTaskCycle,
    ondemandResearchCycle,
    billingCycleStart: stats?.billing_cycle_start,
    planPrice: stats?.plan_price_usd && stats.plan_price_usd > 0 ? stats.plan_price_usd : 0,
    cursorCycle,
    otherCycle,
    cursorToday,
    otherToday,
    cursorOpen,
    otherOpen,
    cursorTaskCycle: stats?.cursor_models_pct_cycle ?? 0,
    otherTaskCycle: stats?.other_models_pct_cycle ?? 0,
    cursorResearchCycle: stats?.cursor_models_pct_research_cycle ?? 0,
    otherResearchCycle: stats?.other_models_pct_research_cycle ?? 0,
    cursorProductCycle: stats?.cursor_models_pct_product_cycle ?? 0,
    otherProductCycle: stats?.other_models_pct_product_cycle ?? 0,
    cursorInfraCycle: stats?.cursor_models_pct_infra_cycle ?? 0,
    otherInfraCycle: stats?.other_models_pct_infra_cycle ?? 0,
  }
}
