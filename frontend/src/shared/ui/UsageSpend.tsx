import { useTranslation } from 'react-i18next'
import { Layers } from 'lucide-react'
import { formatPoolPct, formatUSD } from '../lib/formatters'

export type UsageSpendValues = {
  budgetUsd?: number | null
  ondemandUsd?: number | null
  cursorModelsPct?: number | null
  otherModelsPct?: number | null
  spendKind?: string
  shared?: boolean
}

export function hasUsageSpend(v: UsageSpendValues): boolean {
  const hasPrice = !v.shared && typeof v.budgetUsd === 'number'
  return (
    (hasPrice && (v.budgetUsd ?? 0) > 0) ||
    (v.cursorModelsPct ?? 0) > 0 ||
    (v.otherModelsPct ?? 0) > 0 ||
    (!v.shared && (v.ondemandUsd ?? 0) > 0)
  )
}

interface UsageSpendProps extends UsageSpendValues {
  variant?: 'block' | 'cell' | 'inline'
}

export function UsageSpend({
  budgetUsd,
  ondemandUsd,
  cursorModelsPct,
  otherModelsPct,
  spendKind,
  shared = false,
  variant = 'block',
}: UsageSpendProps) {
  const { t } = useTranslation()
  const hasPrice = !shared && typeof budgetUsd === 'number'
  const cursor = cursorModelsPct ?? 0
  const other = otherModelsPct ?? 0
  const od = shared ? 0 : ondemandUsd ?? 0
  const infra = spendKind === 'infra' ? ` · ${t('metrics.infraSpend')}` : ''

  if (variant === 'inline') {
    return (
      <>
        {shared ? <SharedQuotaIcon /> : null}
        {hasPrice && (budgetUsd ?? 0) > 0 && (
          <span className="text-emerald-700 dark:text-emerald-400 font-mono shrink-0">
            {t('timeline.planShare', { amount: formatUSD(budgetUsd) })}
          </span>
        )}
        {od > 0 && (
          <span className="text-amber-600 dark:text-amber-400 font-mono shrink-0">
            {t('timeline.onDemand', { amount: formatUSD(od) })}
          </span>
        )}
        {spendKind === 'infra' && (
          <span className="text-slate-500 dark:text-slate-400 font-mono shrink-0">{t('timeline.infra')}</span>
        )}
        {cursor > 0 && (
          <span
            className={`font-mono shrink-0 ${
              hasPrice
                ? 'text-slate-400 dark:text-slate-500 hidden lg:inline'
                : 'text-emerald-700 dark:text-emerald-400'
            }`}
          >
            {t('timeline.cursorModels', { pct: formatPoolPct(cursor) })}
          </span>
        )}
        {other > 0 && (
          <span
            className={`font-mono shrink-0 ${
              hasPrice
                ? 'text-slate-400 dark:text-slate-500 hidden xl:inline'
                : 'text-emerald-700 dark:text-emerald-400'
            }`}
          >
            {t('timeline.otherModels', { pct: formatPoolPct(other) })}
          </span>
        )}
      </>
    )
  }

  if (variant === 'cell') {
    return (
      <div className="text-right space-y-0.5">
        {hasPrice ? (
          <div className="font-mono text-emerald-700 dark:text-emerald-400">{formatUSD(budgetUsd)}</div>
        ) : null}
        {cursor > 0 ? (
          <div className={poolClass(hasPrice)}>{t('usage.cursorShort', { pct: formatPoolPct(cursor) })}</div>
        ) : null}
        {other > 0 ? (
          <div className={poolClass(hasPrice)}>{t('usage.otherShort', { pct: formatPoolPct(other) })}</div>
        ) : null}
        {!hasPrice && cursor <= 0 && other <= 0 ? <span className="text-slate-400 dark:text-slate-500">—</span> : null}
      </div>
    )
  }

  return (
    <div className="space-y-1.5">
      {hasPrice && (
        <div className="flex items-center justify-between text-xs">
          <span className="text-slate-500 dark:text-slate-400">{t('pulse.planShare')}</span>
          <span className="font-mono text-emerald-700 dark:text-emerald-400 font-semibold">
            {formatUSD(budgetUsd)}
            {infra}
          </span>
        </div>
      )}
      {cursor > 0 && (
        <div className="flex items-center justify-between text-xs">
          <span className="text-slate-500 dark:text-slate-400">{t('usage.cursorPool')}</span>
          <span className={hasPrice ? 'font-mono text-slate-500 dark:text-slate-400' : 'font-mono text-emerald-700 dark:text-emerald-400 font-semibold'}>
            {t('usage.deltaPct', { pct: formatPoolPct(cursor) })}
          </span>
        </div>
      )}
      {other > 0 && (
        <div className="flex items-center justify-between text-xs">
          <span className="text-slate-500 dark:text-slate-400">{t('usage.otherPool')}</span>
          <span className={hasPrice ? 'font-mono text-slate-500 dark:text-slate-400' : 'font-mono text-emerald-700 dark:text-emerald-400 font-semibold'}>
            {t('usage.deltaPct', { pct: formatPoolPct(other) })}
          </span>
        </div>
      )}
      {od > 0 && (
        <div className="flex items-center justify-between text-xs">
          <span className="text-slate-500 dark:text-slate-400">{t('pulse.meter')}</span>
          <span className="font-mono text-amber-600 dark:text-amber-400">{formatUSD(od)}</span>
        </div>
      )}
    </div>
  )
}

export function SharedQuotaIcon({ label }: { label?: string }) {
  const { t } = useTranslation()
  const text = label || t('usage.sharedTask')
  return (
    <button
      type="button"
      aria-label={text}
      onClick={(e) => e.stopPropagation()}
      className="relative inline-flex items-center justify-center text-indigo-600 dark:text-indigo-300 hover:text-indigo-500 dark:hover:text-indigo-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500 rounded-sm shrink-0 group/shared"
    >
      <Layers className="w-3.5 h-3.5" aria-hidden />
      <span
        role="tooltip"
        className="pointer-events-none absolute z-40 right-0 top-full mt-1.5 w-56 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 px-2.5 py-2 text-left text-[11px] font-normal normal-case tracking-normal leading-snug text-slate-600 dark:text-slate-300 shadow-lg opacity-0 group-hover/shared:opacity-100 group-focus/shared:opacity-100 transition-opacity"
      >
        {text}
      </span>
    </button>
  )
}

function poolClass(secondary: boolean): string {
  return secondary
    ? 'font-mono text-[11px] text-slate-500 dark:text-slate-400'
    : 'font-mono text-emerald-700 dark:text-emerald-400 font-semibold'
}

export function UsageHeadline({
  usd,
  cursor,
  other,
  hasPrice,
  seatUsd,
}: {
  usd: number
  cursor: number
  other: number
  hasPrice: boolean
  seatUsd?: number
}) {
  const { t } = useTranslation()
  if (hasPrice) {
    if (usd <= 0 && cursor <= 0 && other <= 0 && !(seatUsd && seatUsd > 0)) {
      return <div className="mt-2 text-2xl md:text-3xl font-extrabold text-slate-900 dark:text-white">—</div>
    }
    return (
      <>
        <div className="mt-1 text-2xl md:text-3xl font-extrabold text-slate-900 dark:text-white tabular-nums">
          {formatUSD(usd)}
          {seatUsd && seatUsd > 0 ? (
            <span className="ml-1.5 text-base md:text-lg font-semibold text-slate-400 dark:text-slate-500">
              / {formatUSD(seatUsd)}
            </span>
          ) : null}
        </div>
        {(cursor > 0 || other > 0) && (
          <div className="mt-1 text-[11px] text-slate-500 dark:text-slate-400 font-mono leading-snug">
            {[
              cursor > 0 ? t('usage.cursorShort', { pct: formatPoolPct(cursor) }) : null,
              other > 0 ? t('usage.otherShort', { pct: formatPoolPct(other) }) : null,
            ]
              .filter(Boolean)
              .join(' · ')}
          </div>
        )}
      </>
    )
  }
  if (cursor <= 0 && other <= 0) {
    return <div className="mt-2 text-2xl md:text-3xl font-extrabold text-slate-900 dark:text-white">—</div>
  }
  return (
    <div className="mt-2 space-y-0.5">
      {cursor > 0 && (
        <div className="text-2xl md:text-3xl font-extrabold text-slate-900 dark:text-white tabular-nums">
          {t('usage.cursorShort', { pct: formatPoolPct(cursor) })}
        </div>
      )}
      {other > 0 && (
        <div className="text-lg font-semibold text-slate-700 dark:text-slate-200 tabular-nums">
          {t('usage.otherShort', { pct: formatPoolPct(other) })}
        </div>
      )}
    </div>
  )
}
