import { useTranslation } from 'react-i18next'
import { Card } from '../../shared/ui/Card'
import { formatDate, formatPoolPct, formatUSD } from '../../shared/lib/formatters'
import { UsageHeadline } from '../../shared/ui/UsageSpend'
import { cursorSpendFromStats } from '../metrics/cursorSpend'
import type { Stats } from '../../shared/types/api'

interface SpendBreakdownProps {
  stats: Stats | null
}

function Money({ value, empty }: { value: number; empty?: boolean }) {
  if (empty) return <span className="text-slate-400 dark:text-slate-500">—</span>
  return <span className="tabular-nums text-slate-800 dark:text-slate-100">{formatUSD(value)}</span>
}

function PlanCell({
  usd,
  cursor,
  other,
  hasPrice,
}: {
  usd: number
  cursor: number
  other: number
  hasPrice: boolean
}) {
  const { t } = useTranslation()
  if (hasPrice) {
    return (
      <div className="text-right">
        <Money value={usd} />
        {cursor > 0 || other > 0 ? (
          <div className="text-[11px] text-slate-500 dark:text-slate-400 font-mono">
            {[
              cursor > 0 ? t('usage.cursorShort', { pct: formatPoolPct(cursor) }) : null,
              other > 0 ? t('usage.otherShort', { pct: formatPoolPct(other) }) : null,
            ]
              .filter(Boolean)
              .join(' · ')}
          </div>
        ) : null}
      </div>
    )
  }
  if (cursor <= 0 && other <= 0) {
    return <span className="text-slate-400 dark:text-slate-500">—</span>
  }
  return (
    <div className="text-right font-mono font-semibold text-slate-800 dark:text-slate-100 space-y-0.5">
      {cursor > 0 ? <div>{t('usage.cursorShort', { pct: formatPoolPct(cursor) })}</div> : null}
      {other > 0 ? <div>{t('usage.otherShort', { pct: formatPoolPct(other) })}</div> : null}
    </div>
  )
}

export function SpendBreakdown({ stats }: SpendBreakdownProps) {
  const { t, i18n } = useTranslation()
  const spend = cursorSpendFromStats(stats)
  const cycleLabel = spend.billingCycleStart
    ? t('spend.cycleSince', { date: formatDate(spend.billingCycleStart, i18n.language) })
    : t('spend.thisCycle')
  const planLabel = spend.hasPlanPrice ? t('spend.plan') : t('spend.pools')

  const rows: {
    key: string
    plan: number
    cursor: number
    other: number
    ondemand: number | null
    hint?: string
  }[] = [
    {
      key: 'product',
      plan: spend.productCycle,
      cursor: spend.cursorProductCycle,
      other: spend.otherProductCycle,
      ondemand: null,
      hint: t('spend.productHint'),
    },
    {
      key: 'infra',
      plan: spend.infraCycle,
      cursor: spend.cursorInfraCycle,
      other: spend.otherInfraCycle,
      ondemand: null,
      hint: t('spend.infraHint'),
    },
    {
      key: 'tasks',
      plan: spend.taskCycle,
      cursor: spend.cursorTaskCycle,
      other: spend.otherTaskCycle,
      ondemand: spend.ondemandTaskCycle,
    },
    {
      key: 'research',
      plan: spend.researchCycle,
      cursor: spend.cursorResearchCycle,
      other: spend.otherResearchCycle,
      ondemand: spend.ondemandResearchCycle,
    },
    {
      key: 'open',
      plan: spend.planOpen,
      cursor: spend.cursorOpen,
      other: spend.otherOpen,
      ondemand: spend.ondemandOpen,
    },
    {
      key: 'today',
      plan: spend.planToday,
      cursor: spend.cursorToday,
      other: spend.otherToday,
      ondemand: spend.ondemandToday,
    },
  ]

  return (
    <Card className="p-6 space-y-4 bg-white dark:bg-slate-900/60 border-slate-200 dark:border-slate-800/80">
      <div>
        <h2 className="text-lg font-semibold text-slate-900 dark:text-white">{t('spend.title')}</h2>
        <p className="mt-1 text-xs text-slate-500 dark:text-slate-400">{cycleLabel}</p>
      </div>

      <div className="grid grid-cols-2 gap-3">
        <div className="rounded-lg border border-slate-200 dark:border-slate-800 px-3 py-2">
          <div className="text-[11px] font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
            {planLabel}
          </div>
          <UsageHeadline
            usd={spend.planCycle}
            cursor={spend.cursorCycle}
            other={spend.otherCycle}
            hasPrice={spend.hasPlanPrice}
            seatUsd={spend.planPrice}
          />
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
              <th className="pb-2 pr-3 font-medium text-right">{planLabel}</th>
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
                <td className="py-2 pr-3">
                  <PlanCell usd={row.plan} cursor={row.cursor} other={row.other} hasPrice={spend.hasPlanPrice} />
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
