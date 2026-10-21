import { useTranslation } from 'react-i18next'
import { formatDate } from '../../shared/lib/formatters'
import type { ConnectorActivityEntry } from './connectorActivityTypes'
import { formatConnectorActivityText } from './formatConnectorActivity'

interface ConnectorActivityTableProps {
  entries: ConnectorActivityEntry[]
  compact?: boolean
}

function when(createdAt: string, lang: string, withTime: boolean): string {
  if (!createdAt) return ''
  const ms = Date.parse(createdAt)
  if (Number.isNaN(ms)) return createdAt
  return formatDate(ms / 1000, lang, withTime)
}

function levelStyles(level: string): string {
  if (level === 'error') {
    return 'border-rose-200/80 dark:border-rose-900/50 bg-rose-50/40 dark:bg-rose-950/20'
  }
  if (level === 'warn') {
    return 'border-amber-200/80 dark:border-amber-900/50 bg-amber-50/30 dark:bg-amber-950/15'
  }
  return 'border-slate-200/80 dark:border-slate-700/80 bg-slate-50/50 dark:bg-slate-900/30'
}

export function ConnectorActivityTable({ entries, compact }: ConnectorActivityTableProps) {
  const { t, i18n } = useTranslation()

  if (!entries.length) {
    return <p className="text-sm text-slate-500">{t('relay.activity.empty')}</p>
  }

  return (
    <ul className="space-y-2">
      {entries.map((entry) => (
        <li
          key={entry.id}
          className={`rounded-xl border px-3 py-2.5 ${levelStyles(entry.level)}`}
        >
          <div className="flex flex-wrap items-baseline justify-between gap-2">
            <span className="text-sm text-slate-900 dark:text-white">{formatConnectorActivityText(entry, t)}</span>
            <span className="text-[10px] text-slate-500 shrink-0">
              {when(entry.created_at, i18n.language, !compact)}
            </span>
          </div>
          <div className="text-[11px] text-slate-500 mt-0.5">
            <span className="font-medium text-slate-600 dark:text-slate-400">{entry.connector_id}</span>
            <span> · {entry.kind}</span>
          </div>
        </li>
      ))}
    </ul>
  )
}
