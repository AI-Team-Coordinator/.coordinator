import { useTranslation } from 'react-i18next'
import { Card } from '../../shared/ui/Card'
import { InfoTooltip } from '../../shared/ui/InfoTooltip'
import { formatUSD, formatCycleTime, formatDate } from '../../shared/lib/formatters'
import { UsageHeadline } from '../../shared/ui/UsageSpend'
import { cursorSpendFromStats } from './cursorSpend'
import type { Stats } from '../../shared/types/api'

interface MetricsGridProps {
  stats: Stats | null
  onStatusClick?: (status: 'in_progress' | 'completed') => void
  onSpendClick?: () => void
  hideSpend?: boolean
  /** Shown under feat/fix/research counts (e.g. billing cycle). */
  countsScopeLabel?: string
}

export function MetricsGrid({ stats, onStatusClick, onSpendClick, hideSpend, countsScopeLabel }: MetricsGridProps) {
  const { t, i18n } = useTranslation()
  const spend = cursorSpendFromStats(stats)
  const cycleLabel = spend.billingCycleStart
    ? t('metrics.cycleSince', { date: formatDate(spend.billingCycleStart, i18n.language) })
    : t('metrics.thisCycle')
  const spendCardClass = onSpendClick
    ? 'cursor-pointer hover:border-indigo-300 dark:hover:border-indigo-700'
    : ''

  return (
    <section
      className={`grid grid-cols-2 gap-4 ${
        hideSpend ? 'md:grid-cols-4' : 'md:grid-cols-3 xl:grid-cols-6'
      }`}
    >
      <Card
        className={`p-4 bg-white dark:bg-slate-900/70 border-slate-200 dark:border-slate-800/80 ${
          onStatusClick ? 'cursor-pointer hover:border-indigo-300 dark:hover:border-indigo-700' : ''
        }`}
        onClick={onStatusClick ? () => onStatusClick('in_progress') : undefined}
        role={onStatusClick ? 'button' : undefined}
      >
        <div className="text-[11px] font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
          {t('metrics.activeNow')}
        </div>
        <div className="mt-2 text-2xl md:text-3xl font-extrabold text-slate-900 dark:text-white flex items-baseline gap-2">
          {stats ? stats.active_now : 0}
          <span className="text-xs text-emerald-600 dark:text-emerald-400 font-normal">
            {t('metrics.inProgress')}
          </span>
        </div>
      </Card>

      <Card className="p-4 bg-white dark:bg-slate-900/70 border-slate-200 dark:border-slate-800/80">
        <div className="text-[11px] font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
          {t('metrics.avgCycleTime')}
        </div>
        <div className="mt-2 text-2xl md:text-3xl font-extrabold text-slate-900 dark:text-white">
          {formatCycleTime(stats?.avg_cycle_time_minutes)}
        </div>
        <div className="mt-1.5 text-[11px] text-slate-500 dark:text-slate-400 leading-snug">
          {t('metrics.startToMerge')}
        </div>
      </Card>

      <Card
        className={`p-4 bg-white dark:bg-slate-900/70 border-slate-200 dark:border-slate-800/80 ${
          onStatusClick ? 'cursor-pointer hover:border-indigo-300 dark:hover:border-indigo-700' : ''
        }`}
        onClick={onStatusClick ? () => onStatusClick('completed') : undefined}
        role={onStatusClick ? 'button' : undefined}
      >
        <div className="text-[11px] font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
          {t('metrics.completedToday')}
        </div>
        <div className="mt-2 text-2xl md:text-3xl font-extrabold text-slate-900 dark:text-white flex items-baseline gap-2">
          {stats ? stats.completed_today : 0}
          <span className="text-xs text-slate-500 dark:text-slate-400 font-normal">
            ({stats ? stats.completed_this_week : 0} {t('metrics.thisWeek')})
          </span>
        </div>
      </Card>

      <Card className="p-4 bg-white dark:bg-slate-900/70 border-slate-200 dark:border-slate-800/80 min-w-0">
        <div className="xl:hidden">
          <div className="text-[11px] font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
            {t('metrics.featuresFixesResearch')}
          </div>
          <div className="mt-2 flex flex-wrap items-baseline gap-x-1.5 gap-y-0 text-2xl md:text-3xl font-extrabold tabular-nums min-w-0">
            <span className="text-slate-900 dark:text-white">{stats ? stats.features_completed : 0}</span>
            <span className="text-slate-300 dark:text-slate-600 font-semibold text-xl md:text-2xl">/</span>
            <span className="text-amber-500 dark:text-amber-400">{stats ? stats.fixes_completed : 0}</span>
            <span className="text-slate-300 dark:text-slate-600 font-semibold text-xl md:text-2xl">/</span>
            <span className="text-violet-500 dark:text-violet-400">{stats ? stats.research_completed ?? 0 : 0}</span>
          </div>
        </div>
        <div className="hidden xl:flex flex-col gap-1.5 min-w-0">
          <div className="flex items-baseline justify-between gap-2 min-w-0">
            <span className="text-[11px] font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider shrink-0">
              {t('metrics.featShort')}
            </span>
            <span className="text-xl font-extrabold tabular-nums text-slate-900 dark:text-white truncate">
              {stats ? stats.features_completed : 0}
            </span>
          </div>
          <div className="flex items-baseline justify-between gap-2 min-w-0">
            <span className="text-[11px] font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider shrink-0">
              {t('metrics.fixShort')}
            </span>
            <span className="text-xl font-extrabold tabular-nums text-amber-500 dark:text-amber-400 truncate">
              {stats ? stats.fixes_completed : 0}
            </span>
          </div>
          <div className="flex items-baseline justify-between gap-2 min-w-0">
            <span className="text-[11px] font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider shrink-0">
              {t('metrics.researchShort')}
            </span>
            <span className="text-xl font-extrabold tabular-nums text-violet-500 dark:text-violet-400 truncate">
              {stats ? stats.research_completed ?? 0 : 0}
            </span>
          </div>
        </div>
        <div className="mt-1.5 text-[11px] text-slate-500 dark:text-slate-400 leading-snug">
          {countsScopeLabel || t('metrics.researchThisWeek', { count: stats?.research_completed_week ?? 0 })}
        </div>
      </Card>

      {!hideSpend && (
        <>
          <Card
            className={`p-4 bg-white dark:bg-slate-900/70 border-slate-200 dark:border-slate-800/80 overflow-visible ${spendCardClass}`}
        onClick={onSpendClick}
        role={onSpendClick ? 'button' : undefined}
      >
        <div className="flex items-center gap-1 text-[11px] font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
          <span>{spend.hasPlanPrice ? t('metrics.cursorBudget') : t('metrics.cursorPools')}</span>
          <InfoTooltip text={spend.hasPlanPrice ? t('metrics.cursorBudgetTooltip') : t('metrics.cursorPoolsTooltip')} />
        </div>
        {spend.hasPlanPrice ? (
          <div className="mt-2">
            <UsageHeadline usd={spend.planCycle} cursor={spend.cursorCycle} other={spend.otherCycle} hasPrice seatUsd={spend.planPrice} />
          </div>
        ) : (
          <UsageHeadline usd={0} cursor={spend.cursorCycle} other={spend.otherCycle} hasPrice={false} />
        )}
        <div className="mt-1.5 text-[11px] text-slate-500 dark:text-slate-400 leading-snug">
          {spend.hasPlanPrice || spend.hasPools ? cycleLabel : t('metrics.cursorBudgetHintNoPrice')}
        </div>
      </Card>

      <Card
        className={`p-4 bg-white dark:bg-slate-900/70 border-slate-200 dark:border-slate-800/80 overflow-visible ${spendCardClass}`}
        onClick={onSpendClick}
        role={onSpendClick ? 'button' : undefined}
      >
        <div className="flex items-center gap-1 text-[11px] font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
          <span>{t('metrics.cursorSpend')}</span>
          <InfoTooltip text={t('metrics.cursorSpendTooltip')} />
        </div>
        <div className="mt-2 text-2xl md:text-3xl font-extrabold text-slate-900 dark:text-white">
          {spend.hasOnDemand ? formatUSD(spend.ondemandCycle) : '—'}
        </div>
        <div className="mt-1.5 text-[11px] text-slate-500 dark:text-slate-400 leading-snug">
          {spend.hasOnDemand ? cycleLabel : t('metrics.cursorSpendHint')}
        </div>
      </Card>
        </>
      )}
    </section>
  )
}
