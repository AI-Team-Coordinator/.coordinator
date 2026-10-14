import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Card } from '../../shared/ui/Card'
import { formatDate, formatPoolPct, formatUSD } from '../../shared/lib/formatters'
import type { HourParticipant, HourSpend } from '../../shared/types/api'
import { TaskDocLink } from '../docs/TaskDocLink'

interface ConsumptionChartProps {
  grain: 'hour' | 'day'
  rows: HourSpend[]
  title?: string
  subtitle?: string
  boundaries?: { startUnix: number; endUnix: number }
  boundaryHourUnix?: number
  markerUnix?: number[]
}

export function HourlySpendChart({ grain, rows, title, subtitle, boundaries, boundaryHourUnix, markerUnix }: ConsumptionChartProps) {
  const { t, i18n } = useTranslation()
  const [hover, setHover] = useState<number | null>(null)

  const max = useMemo(() => {
    let peak = 0
    for (const hour of rows) {
      peak = Math.max(peak, hour.cursor_models_pct + hour.other_models_pct, hour.ondemand_usd * 10)
    }
    return peak || 1
  }, [rows])

  const now = Date.now() / 1000
  const active = hover != null ? rows[hover] : null

  return (
    <Card
      className="p-4 space-y-3 bg-white dark:bg-slate-900/60 border-slate-200 dark:border-slate-800/80"
      onMouseLeave={() => setHover(null)}
    >
      <div>
        <h2 className="text-[11px] font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
          {title || t('spend.consumption')}
        </h2>
        {subtitle ? <p className="text-[10px] text-slate-400 dark:text-slate-500 mt-0.5">{subtitle}</p> : null}
      </div>
      <div className="flex items-end gap-px sm:gap-1 min-h-[120px]">
        {rows.map((hour, index) => {
          const total = hour.cursor_models_pct + hour.other_models_pct
          const raw = Math.max(total, hour.ondemand_usd * 10)
          const height = raw > 0 ? Math.max(6, Math.round((raw / max) * 96)) : 2
          const cursorH = total > 0 ? Math.round((hour.cursor_models_pct / total) * height) : 0
          const otherH = Math.max(0, height - cursorH)
          const future = grain === 'hour' && hour.hour > now
          const label =
            grain === 'day'
              ? formatDate(hour.hour, i18n.language)
              : new Date(hour.hour * 1000).toLocaleTimeString([], { hour: '2-digit' })
          const showLabel = grain === 'day' ? index % 2 === 0 || index === rows.length - 1 : hourLabelShown(index)
          const selected = hover === index
          const isPeriodStart = boundaries && hour.hour === boundaries.startUnix
          const isPeriodEnd = boundaries && hour.hour === boundaries.endUnix
          const isCycleStartHour = boundaryHourUnix != null && hour.hour === boundaryHourUnix
          const isBillingMarker = markerUnix?.includes(hour.hour) ?? false
          const boundaryRing =
            isPeriodStart || isPeriodEnd || isCycleStartHour || isBillingMarker
              ? 'ring-2 ring-amber-400/80 dark:ring-amber-500/60 ring-offset-1 ring-offset-white dark:ring-offset-slate-900'
              : ''
          return (
            <button
              key={hour.hour}
              type="button"
              aria-pressed={selected}
              className={`flex flex-col items-center gap-1 flex-1 min-w-0 ${future ? 'opacity-40' : ''}`}
              onMouseEnter={() => setHover(index)}
              onFocus={() => setHover(index)}
              onClick={() => setHover(index)}
            >
              <div
                className={`flex flex-col justify-end w-full max-w-[14px] h-24 rounded-sm overflow-hidden bg-slate-100 dark:bg-slate-800/80 ${
                  selected ? 'ring-2 ring-indigo-400 dark:ring-indigo-300 ring-offset-1 ring-offset-white dark:ring-offset-slate-900' : boundaryRing
                }`}
              >
                {otherH > 0 ? (
                  <div className="w-full bg-slate-400 dark:bg-slate-500" style={{ height: otherH }} />
                ) : null}
                {cursorH > 0 ? (
                  <div className="w-full bg-emerald-500 dark:bg-emerald-400" style={{ height: cursorH }} />
                ) : null}
              </div>
              <span
                className={`text-[9px] font-mono h-3 leading-none ${
                  selected ? 'text-indigo-600 dark:text-indigo-300' : 'text-slate-400 dark:text-slate-500'
                }`}
              >
                {showLabel ? label : ''}
              </span>
            </button>
          )
        })}
      </div>
      {boundaries && grain === 'day' && rows.length > 1 ? (
        <p className="text-[10px] text-slate-400 dark:text-slate-500">
          {t('spend.periodBoundaryHint')}
        </p>
      ) : null}
      <HourDetail grain={grain} hour={active} boundaries={boundaries} boundaryHourUnix={boundaryHourUnix} />
    </Card>
  )
}

function hourLabelShown(index: number): boolean {
  return index % 3 === 0 || index === 23
}

function HourDetail({
  grain,
  hour,
  boundaries,
  boundaryHourUnix,
}: {
  grain: 'hour' | 'day'
  hour: HourSpend | null
  boundaries?: { startUnix: number; endUnix: number }
  boundaryHourUnix?: number
}) {
  const { t, i18n } = useTranslation()
  if (!hour) {
    return (
      <div className="rounded-lg border border-dashed border-slate-200 dark:border-slate-700 bg-slate-50/70 dark:bg-slate-800/30 px-3 py-2.5">
        <p className="text-[11px] text-slate-400 dark:text-slate-500">{t('spend.consumptionHover')}</p>
      </div>
    )
  }
  const when =
    grain === 'day'
      ? formatDate(hour.hour, i18n.language)
      : new Date(hour.hour * 1000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  let boundaryNote = ''
  if (boundaries && grain === 'day') {
    if (hour.hour === boundaries.startUnix) boundaryNote = t('spend.periodStart')
    else if (hour.hour === boundaries.endUnix) boundaryNote = t('spend.periodEnd')
  }
  if (boundaryHourUnix != null && grain === 'hour' && hour.hour === boundaryHourUnix) {
    boundaryNote = t('spend.cycleStartHour')
  }
  const people = hour.participants ?? []
  const aliases = hour.aliases?.length ? hour.aliases : hour.alias ? [hour.alias] : []
  const empty =
    (hour.cursor_models_pct ?? 0) <= 0 &&
    (hour.other_models_pct ?? 0) <= 0 &&
    (hour.ondemand_usd ?? 0) <= 0 &&
    (hour.budget_usd ?? 0) <= 0
  return (
    <div className="rounded-lg border border-indigo-200/80 dark:border-indigo-500/25 bg-indigo-50/70 dark:bg-indigo-950/30 px-3 py-2.5 space-y-2 min-h-[4.5rem]">
      <div className="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
        <div className="font-mono text-xs font-semibold text-slate-800 dark:text-slate-100">
          {when}
          {boundaryNote ? <span className="ml-2 text-[10px] font-normal text-amber-600 dark:text-amber-400">{boundaryNote}</span> : null}
        </div>
        {aliases.length > 0 ? (
          <div className="text-[11px] text-slate-500 dark:text-slate-400">{aliases.map((a) => `@${a}`).join(' · ')}</div>
        ) : null}
      </div>
      {empty ? (
        <div className="text-xs text-slate-400 dark:text-slate-500">—</div>
      ) : (
        <div className="text-xs font-medium text-slate-800 dark:text-slate-100">
          {hour.budget_usd != null && hour.budget_usd > 0 ? `${formatUSD(hour.budget_usd)} · ` : null}
          {hour.cursor_models_pct > 0 ? t('usage.cursorShort', { pct: formatPoolPct(hour.cursor_models_pct) }) : null}
          {hour.cursor_models_pct > 0 && hour.other_models_pct > 0 ? ' · ' : null}
          {hour.other_models_pct > 0 ? t('usage.otherShort', { pct: formatPoolPct(hour.other_models_pct) }) : null}
          {hour.ondemand_usd > 0 ? ` · ${t('spend.ondemand')} ${formatUSD(hour.ondemand_usd)}` : null}
        </div>
      )}
      {people.length === 0 && !empty ? (
        <div className="text-[11px] text-slate-500 dark:text-slate-400">{t('spend.unboundChat')}</div>
      ) : people.length > 0 ? (
        <ul className="space-y-1">
          {people.map((p) => (
            <li key={`${p.kind}-${p.id || p.alias}`} className="text-xs leading-snug">
              <ParticipantRow participant={p} />
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  )
}

function ParticipantRow({ participant }: { participant: HourParticipant }) {
  const { t } = useTranslation()
  const kind = participant.kind === 'research' ? t('pulse.research') : t('pulse.task')
  const label = participant.title || participant.id || ''
  const who = participant.alias ? `@${participant.alias}` : ''
  const taskId = participant.kind === 'task' ? participant.id?.trim() : ''
  return (
    <span className="inline-flex flex-wrap items-baseline gap-x-1.5 gap-y-0.5 text-slate-700 dark:text-slate-200">
      <span className="text-[10px] uppercase tracking-wide text-slate-500 dark:text-slate-400">{kind}</span>
      {taskId ? (
        <TaskDocLink taskId={taskId} className="text-xs font-medium text-indigo-700 dark:text-indigo-300">
          {label || taskId}
        </TaskDocLink>
      ) : (
        <span className="font-medium">{label}</span>
      )}
      {who ? <span className="text-[11px] text-slate-500 dark:text-slate-400">{who}</span> : null}
    </span>
  )
}

export { padTodayPersonalHours as padTodayHours } from './personalSpend'
