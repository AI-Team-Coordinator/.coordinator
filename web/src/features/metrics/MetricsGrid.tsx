import { useTranslation } from 'react-i18next'
import { Card } from '../../shared/ui/Card'
import { InfoTooltip } from '../../shared/ui/InfoTooltip'
import { formatUSD, formatCycleTime } from '../../shared/lib/formatters'
import type { Stats } from '../../shared/types/api'

interface MetricsGridProps {
  stats: Stats | null
  onStatusClick?: (status: 'in_progress' | 'completed') => void
}

export function MetricsGrid({ stats, onStatusClick }: MetricsGridProps) {
  const { t } = useTranslation()
  const hasBudget = Boolean(stats && stats.budget_tasks > 0)
  const hasMeter = Boolean(stats && stats.cost_tasks > 0)
  const budgetOpen = stats?.budget_usd_open ?? 0
  const costOpen = stats?.cost_usd_open ?? 0

  return (
    <section className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-4">
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

      <Card className="p-4 bg-white dark:bg-slate-900/70 border-slate-200 dark:border-slate-800/80">
        <div className="text-[11px] font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
          {t('metrics.featuresAndFixes')}
        </div>
        <div className="mt-2 text-2xl md:text-3xl font-extrabold text-slate-900 dark:text-white flex items-baseline gap-2">
          <span>{stats ? stats.features_completed : 0}</span>
          <span className="text-slate-400 font-light">/</span>
          <span className="text-amber-500 dark:text-amber-400">
            {stats ? stats.fixes_completed : 0}
          </span>
        </div>
      </Card>

      <Card className="p-4 bg-white dark:bg-slate-900/70 border-slate-200 dark:border-slate-800/80 overflow-visible">
        <div className="flex items-center gap-1 text-[11px] font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
          <span>{t('metrics.cursorBudget')}</span>
          <InfoTooltip text={t('metrics.cursorBudgetTooltip')} />
        </div>
        <div className="mt-2 text-2xl md:text-3xl font-extrabold text-slate-900 dark:text-white">
          {hasBudget ? formatUSD(stats?.budget_usd_today ?? 0) : '—'}
        </div>
        <div className="mt-1.5 space-y-0.5 text-[11px] text-slate-500 dark:text-slate-400 leading-snug">
          {hasBudget ? (
            <>
              <div>
                {t('metrics.productSpend')} {formatUSD(stats?.budget_usd_product_week ?? 0)}
              </div>
              <div>
                {t('metrics.infraSpend')} {formatUSD(stats?.budget_usd_infra_week ?? 0)}
              </div>
              {budgetOpen > 0 && (
                <div>
                  {t('metrics.includesOpen', { amount: formatUSD(budgetOpen) })}
                </div>
              )}
            </>
          ) : (
            <div>{t('metrics.cursorBudgetHint')}</div>
          )}
        </div>
      </Card>

      <Card className="p-4 bg-white dark:bg-slate-900/70 border-slate-200 dark:border-slate-800/80 overflow-visible">
        <div className="flex items-center gap-1 text-[11px] font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
          <span>{t('metrics.cursorSpend')}</span>
          <InfoTooltip text={t('metrics.cursorSpendTooltip')} />
        </div>
        <div className="mt-2 text-2xl md:text-3xl font-extrabold text-slate-900 dark:text-white">
          {hasMeter ? formatUSD(stats?.cost_usd_today ?? 0) : '—'}
        </div>
        <div className="mt-1.5 space-y-0.5 text-[11px] text-slate-500 dark:text-slate-400 leading-snug">
          {hasMeter ? (
            <>
              <div>
                {formatUSD(stats?.cost_usd_week ?? 0)} {t('metrics.thisWeek')}
              </div>
              {costOpen > 0 && (
                <div>
                  {t('metrics.includesOpen', { amount: formatUSD(costOpen) })}
                </div>
              )}
            </>
          ) : (
            <div>{t('metrics.cursorSpendHint')}</div>
          )}
        </div>
      </Card>
    </section>
  )
}
