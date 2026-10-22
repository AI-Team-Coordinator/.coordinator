import { X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { formatConnectorActivityText } from './formatConnectorActivity'
import type { ConnectorActivityEntry } from './connectorActivityTypes'
import { Card } from '../../shared/ui/Card'
import { Button } from '../../shared/ui/Button'

interface ConnectorActivityDetailModalProps {
  entry: ConnectorActivityEntry
  onClose: () => void
}

function formatDateTime(raw: string, locale: string): string {
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

export function ConnectorActivityDetailModal({ entry, onClose }: ConnectorActivityDetailModalProps) {
  const { t, i18n } = useTranslation()
  const narrative = formatConnectorActivityText(entry, t)
  const detailJson =
    entry.detail && Object.keys(entry.detail).length > 0
      ? JSON.stringify(entry.detail, null, 2)
      : '—'

  const rows: { label: string; value: string }[] = [
    { label: t('relay.activity.fieldId'), value: String(entry.id) },
    { label: t('relay.activity.fieldConnector'), value: entry.connector_id },
    { label: t('relay.activity.fieldKind'), value: entry.kind },
    { label: t('relay.activity.fieldLevel'), value: entry.level },
    { label: t('relay.activity.fieldCreated'), value: formatDateTime(entry.created_at, i18n.language) },
    { label: t('relay.activity.fieldSummary'), value: entry.summary || '—' },
    { label: t('relay.activity.fieldNarrative'), value: narrative },
  ]

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center p-4 bg-black/45 backdrop-blur-[2px]">
      <Card className="w-full max-w-lg max-h-[85vh] overflow-hidden flex flex-col shadow-2xl">
        <div className="flex items-start justify-between gap-2 p-4 border-b border-slate-200 dark:border-slate-700">
          <h4 className="text-sm font-semibold text-slate-900 dark:text-white">{t('relay.activity.detailTitle')}</h4>
          <button type="button" className="text-slate-400 hover:text-slate-600" onClick={onClose}>
            <X className="w-4 h-4" />
          </button>
        </div>
        <div className="p-4 overflow-y-auto space-y-4 text-sm">
          <dl className="space-y-2">
            {rows.map((row) => (
              <div key={row.label}>
                <dt className="text-[10px] uppercase tracking-wide text-slate-500">{row.label}</dt>
                <dd className="text-slate-800 dark:text-slate-200 mt-0.5 whitespace-pre-wrap break-words font-mono text-xs">
                  {row.value}
                </dd>
              </div>
            ))}
          </dl>
          <div>
            <p className="text-[10px] uppercase tracking-wide text-slate-500 mb-1">{t('relay.activity.fieldDetail')}</p>
            <pre className="text-[11px] rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-900 p-3 overflow-x-auto text-slate-800 dark:text-slate-200">
              {detailJson}
            </pre>
          </div>
        </div>
        <div className="p-4 border-t border-slate-200 dark:border-slate-700 flex justify-end">
          <Button size="sm" variant="ghost" onClick={onClose}>{t('relay.activity.close')}</Button>
        </div>
      </Card>
    </div>
  )
}
