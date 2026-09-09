import { useTranslation } from 'react-i18next'
import { Card } from '../../shared/ui/Card'
import { formatDate, formatUSD } from '../../shared/lib/formatters'
import { cursorSpendFromStats } from '../metrics/cursorSpend'
import type { Stats } from '../../shared/types/api'

interface SpendBreakdownProps {
  stats: Stats | null
}

function Money({ value, empty }: { value: number; empty?: boolean }) {
  if (empty) return <span className="text-slate-400 dark:text-slate-500">—</span>
  return <span className="tabular-nums text-slate-800 dark:text-slate-100">{formatUSD(value)}</span>
}

export function SpendBreakdown({ stats }: SpendBreakdownProps) {
  const { t, i18n } = useTranslation()
  const spend = cursorSpendFromStats(stats)
  const cycleLabel = spend.billingCycleStart
    ? t('spend.cycleSince', { date: formatDate(spend.billingCycleStart, i18n.language) })
    : t('spend.thisCycle')

  const rows: { key: string; plan: number; ondemand: number | null; hint?: string }[] = [
    {
      key: 'product',
      plan: spend.productCycle,
      ondemand: null,
      hint: t('spend.productHint'),
    },
    {
      key: 'infra',
      plan: spend.infraCycle,
      ondemand: null,
      hint: t('spend.infraHint'),
    },
    { key: 'tasks', plan: spend.taskCycle, ondemand: spend.ondemandTaskCycle },
    { key: 'research', plan: spend.researchCycle, ondemand: spend.ondemandResearchCycle },
    { key: 'open', plan: spend.planOpen, ondemand: spend.ondemandOpen },
    { key: 'today', plan: spend.planToday, ondemand: spend.ondemandToday },
  ]

  return (
    <Card className="p-6 space-y-4 bg-white dark:bg-slate-900/60 border-slate-200 dark:border-slate-800/80">
      <div>
        <h2 className="text-lg font-semibold text-slate-900 dark:text-white">{t('spend.title')}</h2>
        <p className="mt-1 text-xs text-slate-500 dark:text-slate-400">{cycleLabel}</p>
        <p className="mt-1 text-xs text-slate-500 dark:text-slate-400 leading-snug">{t('spend.hint')}</p>
      </div>

      <div className="grid grid-cols-2 gap-3">
        <div className="rounded-lg border border-slate-200 dark:border-slate-800 px-3 py-2">
          <div className="text-[11px] font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
            {t('spend.plan')}
          </div>
          <div className="mt-1 text-xl font-extrabold text-slate-900 dark:text-white tabular-nums">
            {spend.hasPlan ? formatUSD(spend.planCycle) : '—'}
          </div>
        </div>
        <div className="rounded-lg border border-slate-200 dark:border-slate-800 px-3 py-2">
          <div className="text-[11px] font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
            {t('spend.ondemand')}
          </div>
          <div className="mt-1 text-xl font-extrabold text-slate-900 dark:text-white tabular-nums">
            {spend.hasOnDemand ? formatUSD(spend.ondemandCycle) : '—'}
          </div>
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="text-left text-[11px] font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
              <th className="pb-2 pr-3 font-medium">{t('spend.row')}</th>
              <th className="pb-2 pr-3 font-medium text-right">{t('spend.plan')}</th>
              <th className="pb-2 font-medium text-right">{t('spend.ondemand')}</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
            {rows.map((row) => (
              <tr key={row.key}>
                <td className="py-2 pr-3 text-slate-700 dark:text-slate-200">
                  <div>{t(`spend.${row.key}`)}</div>
                  {row.hint ? (
                    <div className="text-[11px] text-slate-400 dark:text-slate-500">{row.hint}</div>
                  ) : null}
                </td>
                <td className="py-2 pr-3 text-right">
                  <Money value={row.plan} />
                </td>
                <td className="py-2 text-right">
                  <Money value={row.ondemand ?? 0} empty={row.ondemand == null} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Card>
  )
}
