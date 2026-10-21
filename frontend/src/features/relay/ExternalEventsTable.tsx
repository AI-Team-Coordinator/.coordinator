import { useTranslation } from 'react-i18next'
import { formatDate } from '../../shared/lib/formatters'
import type { ExternalEvent } from '../../shared/types/api'
import { externalEventClassStyles } from './externalEventStyles'

interface ExternalEventsTableProps {
  events: ExternalEvent[]
  onSelect: (ev: ExternalEvent) => void
  compact?: boolean
}

function eventWhen(ev: ExternalEvent, lang: string, withTime: boolean): string {
  const raw = ev.received_at || ev.occurred_at
  if (!raw) return ''
  const ms = Date.parse(raw)
  if (Number.isNaN(ms)) return raw
  return formatDate(ms / 1000, lang, withTime)
}

export function ExternalEventsTable({ events, onSelect, compact }: ExternalEventsTableProps) {
  const { t, i18n } = useTranslation()

  if (!events.length) {
    return <p className="text-sm text-slate-500">{t('relay.events.empty')}</p>
  }

  return (
    <ul className="space-y-2">
      {events.map((ev) => (
        <li key={ev.id}>
          <button
            type="button"
            onClick={() => onSelect(ev)}
            className={`w-full text-left rounded-xl border px-3 py-2.5 transition-colors hover:ring-1 hover:ring-indigo-300/60 dark:hover:ring-indigo-600/50 ${externalEventClassStyles(ev.class)}`}
          >
            <div className="flex flex-wrap items-baseline justify-between gap-2">
              <span className="text-sm font-medium text-slate-900 dark:text-white">
                {ev.title || ev.kind}
              </span>
              <span className="text-[10px] text-slate-500 shrink-0">
                {eventWhen(ev, i18n.language, !compact)}
              </span>
            </div>
            <div className="text-[11px] text-slate-500 mt-0.5">
              <span className="font-medium text-slate-600 dark:text-slate-400">{ev.source}</span>
              <span> · {ev.kind}</span>
              {ev.class && <span> · {ev.class}</span>}
            </div>
            {!compact && ev.summary && (
              <p className="text-xs text-slate-600 dark:text-slate-300 mt-1 line-clamp-2">{ev.summary}</p>
            )}
          </button>
        </li>
      ))}
    </ul>
  )
}
