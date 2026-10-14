import { useMemo, useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import type { Stats } from '../../shared/types/api'
import { HourlySpendChart } from './HourlySpendChart'
import {
  buildPersonalCycleDays,
  formatPeriodRange,
  padTodayPersonalHours,
} from './personalSpend'
import { startOfLocalDay, unixFromISO } from '../metrics/cursorSpend'

type Grain = 'hour' | 'day'

interface PersonalConsumptionChartProps {
  stats: Stats | null
  alias?: string
}

export function PersonalConsumptionChart({ stats, alias }: PersonalConsumptionChartProps) {
  const { t, i18n } = useTranslation()
  const [grain, setGrain] = useState<Grain>('hour')

  const cycleStart = unixFromISO(stats?.billing_cycle_start)
  const daySeries = useMemo(
    () => buildPersonalCycleDays(stats?.hours, alias, stats?.billing_cycle_start),
    [stats?.hours, stats?.billing_cycle_start, alias]
  )

  const hourRows = useMemo(
    () => padTodayPersonalHours(stats?.hours, alias),
    [stats?.hours, alias]
  )

  const periodCaption = formatPeriodRange(daySeries.periodStart, daySeries.periodEnd, i18n.language)
  const multiDayCycle = daySeries.periodEnd > daySeries.periodStart

  const boundaries = multiDayCycle
    ? { startUnix: daySeries.periodStart, endUnix: daySeries.periodEnd }
    : undefined

  const cycleStartHourToday = useMemo(() => {
    if (cycleStart <= 0) return undefined
    const todayStart = startOfLocalDay()
    if (cycleStart < todayStart || cycleStart >= todayStart + 86400) return undefined
    return cycleStart
  }, [cycleStart])

  return (
    <div className="space-y-2">
      <div className="flex flex-wrap items-center justify-between gap-2 px-1">
        <p className="text-[11px] text-slate-500 dark:text-slate-400">
          {t('spend.billingPeriod')}: <span className="font-medium text-slate-700 dark:text-slate-300">{periodCaption}</span>
        </p>
        <div className="inline-flex rounded-lg border border-slate-200 dark:border-slate-700 p-0.5 bg-slate-50 dark:bg-slate-900/80">
          <GrainButton active={grain === 'hour'} onClick={() => setGrain('hour')}>
            {t('spend.personalByHour')}
          </GrainButton>
          <GrainButton active={grain === 'day'} onClick={() => setGrain('day')}>
            {t('spend.personalByDay')}
          </GrainButton>
        </div>
      </div>
      <HourlySpendChart
        grain={grain}
        rows={grain === 'hour' ? hourRows : daySeries.rows}
        title={t('spend.consumptionPersonal')}
        boundaries={boundaries}
        boundaryHourUnix={grain === 'hour' ? cycleStartHourToday : undefined}
        subtitle={grain === 'hour' && multiDayCycle ? t('spend.personalHoursTodayInCycle') : undefined}
      />
    </div>
  )
}

function GrainButton({
  active,
  onClick,
  children,
}: {
  active: boolean
  onClick: () => void
  children: ReactNode
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`px-2.5 py-1 text-[11px] font-medium rounded-md transition ${
        active
          ? 'bg-white dark:bg-slate-800 text-indigo-700 dark:text-indigo-300 shadow-sm'
          : 'text-slate-500 dark:text-slate-400 hover:text-slate-800 dark:hover:text-slate-200'
      }`}
    >
      {children}
    </button>
  )
}
