import type { Stats } from '../../shared/types/api'

export function cursorSpendFromStats(stats: Stats | null) {
  const taskCycle = stats?.budget_usd_cycle ?? 0
  const researchCycle = stats?.budget_usd_research_cycle ?? 0
  const planCycle = taskCycle + researchCycle
  const ondemandTaskCycle = stats?.cost_usd_cycle ?? 0
  const ondemandResearchCycle = stats?.cost_usd_research_cycle ?? 0
  const ondemandCycle = ondemandTaskCycle + ondemandResearchCycle

  return {
    hasPlan: Boolean(
      stats &&
        (stats.budget_tasks > 0 || planCycle > 0 || (stats.research_completed ?? 0) > 0)
    ),
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
  }
}
