import { Radio } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import type { ExternalEvent } from '../../shared/types/api'

interface ExternalSignalsCardProps {
  events: ExternalEvent[]
}

function classStyles(className: string): string {
  switch (className) {
    case 'warning':
      return 'border-amber-200 dark:border-amber-500/40 bg-amber-50 dark:bg-amber-950/30'
    case 'reminder':
      return 'border-sky-200 dark:border-sky-500/40 bg-sky-50 dark:bg-sky-950/30'
    default:
      return 'border-slate-200 dark:border-slate-600 bg-slate-50 dark:bg-slate-900/40'
  }
}

export function ExternalSignalsCard({ events }: ExternalSignalsCardProps) {
  const { t } = useTranslation()
  if (!events?.length) {
    return null
  }

  return (
    <div className="rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900/50 p-4 shadow-sm space-y-3">
      <div className="flex items-center gap-2">
        <Radio className="w-4 h-4 text-violet-600 dark:text-violet-400" />
        <h3 className="text-sm font-semibold text-slate-900 dark:text-white">
          {t('pulse.externalTitle')}
        </h3>
        <span className="text-xs text-slate-500 dark:text-slate-400">({events.length})</span>
      </div>
      <ul className="space-y-2">
        {events.map((ev) => (
          <li
            key={ev.id}
            className={`text-xs rounded-lg border px-3 py-2 ${classStyles(ev.class)}`}
          >
            <div className="font-medium text-slate-900 dark:text-slate-100">
              {ev.title || ev.kind}
              <span className="text-slate-500 dark:text-slate-400 font-normal"> · {ev.source}</span>
            </div>
            {ev.summary && (
              <p className="mt-1 text-slate-700 dark:text-slate-300 leading-relaxed">{ev.summary}</p>
            )}
            {ev.url && (
              <a
                href={ev.url}
                target="_blank"
                rel="noreferrer"
                className="mt-1 inline-block text-violet-700 dark:text-violet-300 hover:underline"
              >
                {t('pulse.externalLink')}
              </a>
            )}
          </li>
        ))}
      </ul>
    </div>
  )
}
