import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { ExternalEvent } from '../../shared/types/api'
import { ExternalEventDetailModal } from './ExternalEventDetailModal'
import { externalEventNeedsAction, formatExternalEventShort } from './formatExternalEvent'

interface ExternalEventsTableProps {
  events: ExternalEvent[]
  onSelect?: (ev: ExternalEvent) => void
  onTriageChanged?: () => void
  compact?: boolean
  maxRows?: number
}

function formatDateTime(raw: string | undefined, locale: string): string {
  if (!raw) return '—'
  const ms = Date.parse(raw)
  if (Number.isNaN(ms)) return raw
  return new Date(ms).toLocaleString(locale, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  })
}

function kindBadgeClass(ev: ExternalEvent): string {
  if (externalEventNeedsAction(ev)) {
    return 'bg-amber-100 text-amber-900 dark:bg-amber-950 dark:text-amber-200 ring-1 ring-amber-300/60 dark:ring-amber-700/50'
  }
  switch (ev.class) {
    case 'warning':
      return 'bg-amber-50 text-amber-800 dark:bg-amber-950/80 dark:text-amber-200'
    case 'reminder':
      return 'bg-sky-50 text-sky-800 dark:bg-sky-950/80 dark:text-sky-200'
    default:
      return 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300'
  }
}

export function ExternalEventsTable({
  events,
  onSelect,
  onTriageChanged,
  compact,
  maxRows,
}: ExternalEventsTableProps) {
  const { t, i18n } = useTranslation()
  const [internalSelected, setInternalSelected] = useState<ExternalEvent | null>(null)

  const rows = maxRows != null ? events.slice(0, maxRows) : events

  if (!rows?.length) {
    return <p className="text-sm text-slate-500">{t('relay.events.empty')}</p>
  }

  const pick = (ev: ExternalEvent) => {
    if (onSelect) onSelect(ev)
    else setInternalSelected(ev)
  }

  return (
    <>
      <div className="overflow-x-auto rounded-xl border border-slate-200/80 dark:border-slate-700/80">
        <table className="w-full text-left text-xs border-collapse">
          <thead>
            <tr className="border-b border-slate-200 dark:border-slate-700 bg-slate-50/80 dark:bg-slate-900/60">
              <th className="px-2.5 py-2 font-medium text-slate-500 whitespace-nowrap">{t('relay.events.colTime')}</th>
              <th className="px-2.5 py-2 font-medium text-slate-500 whitespace-nowrap">{t('relay.events.colSource')}</th>
              <th className="px-2.5 py-2 font-medium text-slate-500 whitespace-nowrap">{t('relay.events.colKind')}</th>
              {!compact && (
                <th className="px-2.5 py-2 font-medium text-slate-500 whitespace-nowrap">{t('relay.events.colClass')}</th>
              )}
              <th className="px-2.5 py-2 font-medium text-slate-500 w-24">{t('relay.events.colAction')}</th>
              <th className="px-2.5 py-2 font-medium text-slate-500">{t('relay.events.colResult')}</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((ev) => {
              const needs = externalEventNeedsAction(ev)
              const result = formatExternalEventShort(ev)
              return (
                <tr
                  key={ev.id}
                  className={`border-b border-slate-100 dark:border-slate-800 last:border-0 cursor-pointer hover:bg-slate-50/80 dark:hover:bg-slate-800/40 ${needs ? 'bg-amber-50/40 dark:bg-amber-950/15' : ''}`}
                  onClick={() => pick(ev)}
                >
                  <td className="px-2.5 py-2 font-mono text-[11px] text-slate-600 dark:text-slate-400 whitespace-nowrap align-top">
                    {formatDateTime(ev.received_at || ev.occurred_at, i18n.language)}
                  </td>
                  <td className="px-2.5 py-2 font-mono text-[11px] text-slate-700 dark:text-slate-300 whitespace-nowrap align-top">
                    {ev.source}
                  </td>
                  <td className="px-2.5 py-2 align-top">
                    <span className={`inline-block px-1.5 py-0.5 rounded text-[10px] font-medium font-mono ${kindBadgeClass(ev)}`}>
                      {ev.kind}
                    </span>
                  </td>
                  {!compact && (
                    <td className="px-2.5 py-2 font-mono text-[10px] text-slate-500 align-top">{ev.class || '—'}</td>
                  )}
                  <td className="px-2.5 py-2 align-top whitespace-nowrap">
                    {needs ? (
                      <span className="text-[10px] font-semibold uppercase text-amber-700 dark:text-amber-300">
                        {t('relay.events.actionRequired')}
                      </span>
                    ) : (
                      <span className="text-[10px] text-slate-400">—</span>
                    )}
                  </td>
                  <td className="px-2.5 py-2 align-top min-w-0 max-w-[280px] sm:max-w-md font-mono text-[11px] text-slate-800 dark:text-slate-200 truncate" title={result}>
                    {result}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
      {internalSelected && !onSelect && (
        <ExternalEventDetailModal
          event={internalSelected}
          onClose={() => setInternalSelected(null)}
          onTriageChanged={onTriageChanged}
        />
      )}
    </>
  )
}
