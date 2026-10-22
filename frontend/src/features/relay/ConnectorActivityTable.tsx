import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ConnectorActivityDetailModal } from './ConnectorActivityDetailModal'
import type { ConnectorActivityEntry } from './connectorActivityTypes'
import { formatConnectorActivityShort } from './formatConnectorActivity'

interface ConnectorActivityTableProps {
  entries: ConnectorActivityEntry[]
  compact?: boolean
}

function formatDateTime(raw: string, locale: string): string {
  if (!raw) return ''
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

function levelBadge(level: string): string {
  if (level === 'error') {
    return 'bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-200'
  }
  if (level === 'warn') {
    return 'bg-amber-100 text-amber-900 dark:bg-amber-950 dark:text-amber-200'
  }
  return 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300'
}

export function ConnectorActivityTable({ entries, compact }: ConnectorActivityTableProps) {
  const { t, i18n } = useTranslation()
  const [selected, setSelected] = useState<ConnectorActivityEntry | null>(null)

  if (!entries?.length) {
    return <p className="text-sm text-slate-500">{t('relay.activity.empty')}</p>
  }

  return (
    <>
      <div className="overflow-x-auto rounded-xl border border-slate-200/80 dark:border-slate-700/80">
        <table className="w-full text-left text-xs border-collapse">
          <thead>
            <tr className="border-b border-slate-200 dark:border-slate-700 bg-slate-50/80 dark:bg-slate-900/60">
              <th className="px-2.5 py-2 font-medium text-slate-500 whitespace-nowrap">{t('relay.activity.colTime')}</th>
              <th className="px-2.5 py-2 font-medium text-slate-500 whitespace-nowrap">{t('relay.activity.colConnector')}</th>
              {!compact && (
                <th className="px-2.5 py-2 font-medium text-slate-500 whitespace-nowrap">{t('relay.activity.colKind')}</th>
              )}
              <th className="px-2.5 py-2 font-medium text-slate-500 w-16">{t('relay.activity.colLevel')}</th>
              <th className="px-2.5 py-2 font-medium text-slate-500">{t('relay.activity.colResult')}</th>
            </tr>
          </thead>
          <tbody>
            {entries.map((entry) => {
              const result = formatConnectorActivityShort(entry)
              return (
                <tr
                  key={entry.id}
                  className="border-b border-slate-100 dark:border-slate-800 last:border-0 cursor-pointer hover:bg-slate-50/80 dark:hover:bg-slate-800/40"
                  onClick={() => setSelected(entry)}
                >
                  <td className="px-2.5 py-2 font-mono text-[11px] text-slate-600 dark:text-slate-400 whitespace-nowrap align-top">
                    {formatDateTime(entry.created_at, i18n.language)}
                  </td>
                  <td className="px-2.5 py-2 font-mono text-[11px] text-slate-700 dark:text-slate-300 whitespace-nowrap align-top">
                    {entry.connector_id}
                  </td>
                  {!compact && (
                    <td className="px-2.5 py-2 font-mono text-[11px] text-slate-600 dark:text-slate-400 whitespace-nowrap align-top">
                      {entry.kind}
                    </td>
                  )}
                  <td className="px-2.5 py-2 align-top">
                    <span className={`inline-block px-1.5 py-0.5 rounded text-[10px] font-medium uppercase ${levelBadge(entry.level)}`}>
                      {entry.level}
                    </span>
                  </td>
                  <td
                    className="px-2.5 py-2 align-top min-w-0 max-w-[280px] sm:max-w-md font-mono text-[11px] text-slate-800 dark:text-slate-200 truncate"
                    title={result}
                  >
                    {result}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
      {selected && (
        <ConnectorActivityDetailModal entry={selected} onClose={() => setSelected(null)} />
      )}
    </>
  )
}
