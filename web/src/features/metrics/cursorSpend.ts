import type { Stats } from '../../shared/types/api'

export function cursorSpendFromStats(stats: Stats | null) {
  const taskCycle = stats?.budget_usd_cycle ?? 0
  const researchCycle = stats?.budget_usd_research_cycle ?? 0
  const planCycle = taskCycle + researchCycle
  const ondemandTaskCycle = stats?.cost_usd_cycle ?? 0
  const ondemandResearchCycle = stats?.cost_usd_research_cycle ?? 0
  const ondemandCycle = ondemandTaskCycle + ondemandResearchCycle
  const cursorCycle = (stats?.cursor_models_pct_cycle ?? 0) + (stats?.cursor_models_pct_research_cycle ?? 0)
  const otherCycle = (stats?.other_models_pct_cycle ?? 0) + (stats?.other_models_pct_research_cycle ?? 0)
  const cursorToday = (stats?.cursor_models_pct_today ?? 0) + (stats?.cursor_models_pct_research_today ?? 0)
  const otherToday = (stats?.other_models_pct_today ?? 0) + (stats?.other_models_pct_research_today ?? 0)
  const cursorOpen = (stats?.cursor_models_pct_open ?? 0) + (stats?.cursor_models_pct_research_open ?? 0)
  const otherOpen = (stats?.other_models_pct_open ?? 0) + (stats?.other_models_pct_research_open ?? 0)

  return {
    hasPlanPrice: Boolean(stats?.plan_price_usd && stats.plan_price_usd > 0),
    hasPlan: Boolean(stats && (stats.budget_tasks > 0 || planCycle > 0)),
    hasPools: cursorCycle > 0 || otherCycle > 0 || cursorToday > 0 || otherToday > 0,
    hasOnDemand: Boolean(stats && (stats.cost_tasks > 0 || ondemandResearchCycle > 0)),
    planCycle,
    ondemandCycle,
    taskCycle,
    researchCycle,
    productCycle: stats?.budget_usd_product_cycle ?? 0,
    infraCycle: stats?.budget_usd_infra_cycle ?? 0,
    planToday: (stats?.budget_usd_today ?? 0) + (stats?.budget_usd_research_today ?? 0),
    ondemandToday: (stats?.cost_usd_today ?? 0) + (stats?.cost_usd_research_today ?? 0),
    planOpen: stats?.budget_usd_open ?? 0,
    ondemandOpen: (stats?.cost_usd_open ?? 0) + (stats?.cost_usd_research_open ?? 0),
    ondemandTaskCycle,
    ondemandResearchCycle,
    billingCycleStart: stats?.billing_cycle_start,
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
