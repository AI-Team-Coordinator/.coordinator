import { X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { formatDate } from '../../shared/lib/formatters'
import type { ExternalEvent } from '../../shared/types/api'
import { Card } from '../../shared/ui/Card'
import { Button } from '../../shared/ui/Button'

interface ExternalEventDetailModalProps {
  event: ExternalEvent
  onClose: () => void
}

function formatTs(raw?: string, lang?: string) {
  if (!raw) return '—'
  const ms = Date.parse(raw)
  if (Number.isNaN(ms)) return raw
  return formatDate(ms / 1000, lang || 'en', true)
}

export function ExternalEventDetailModal({ event, onClose }: ExternalEventDetailModalProps) {
  const { t, i18n } = useTranslation()
  const rows: { label: string; value: string }[] = [
    { label: t('relay.events.fieldId'), value: event.id },
    { label: t('relay.events.fieldSource'), value: event.source },
    { label: t('relay.events.fieldKind'), value: event.kind },
    { label: t('relay.events.fieldClass'), value: event.class },
    { label: t('relay.events.fieldExternalId'), value: event.external_id || '—' },
    { label: t('relay.events.fieldReceived'), value: formatTs(event.received_at, i18n.language) },
    { label: t('relay.events.fieldOccurred'), value: formatTs(event.occurred_at, i18n.language) },
    { label: t('relay.events.fieldTitle'), value: event.title || '—' },
    { label: t('relay.events.fieldSummary'), value: event.summary || '—' },
    { label: t('relay.events.fieldUrl'), value: event.url || '—' },
    { label: t('relay.events.fieldTags'), value: event.tags?.length ? event.tags.join(', ') : '—' },
  ]

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/45 backdrop-blur-[2px]">
      <Card className="w-full max-w-lg max-h-[90vh] overflow-hidden flex flex-col shadow-2xl">
        <div className="flex items-start justify-between gap-2 p-4 border-b border-slate-200 dark:border-slate-700">
          <div>
            <h4 className="text-sm font-semibold text-slate-900 dark:text-white">
              {event.title || event.kind}
            </h4>
            <p className="text-xs text-slate-500 mt-0.5">{event.source} · {event.kind}</p>
          </div>
          <button type="button" className="text-slate-400 hover:text-slate-600" onClick={onClose}>
            <X className="w-4 h-4" />
          </button>
        </div>
        <div className="p-4 overflow-y-auto space-y-3 text-sm">
          <dl className="space-y-2">
            {rows.map((row) => (
              <div key={row.label}>
                <dt className="text-[10px] uppercase tracking-wide text-slate-500">{row.label}</dt>
                <dd className="text-slate-800 dark:text-slate-200 break-words whitespace-pre-wrap">
                  {row.label === t('relay.events.fieldUrl') && event.url ? (
                    <a
                      href={event.url}
                      target="_blank"
                      rel="noreferrer"
                      className="text-indigo-600 dark:text-indigo-400 hover:underline"
                    >
                      {event.url}
                    </a>
                  ) : (
                    row.value
                  )}
                </dd>
              </div>
            ))}
          </dl>
          {event.data && Object.keys(event.data).length > 0 && (
            <div>
              <p className="text-[10px] uppercase tracking-wide text-slate-500 mb-1">
                {t('relay.events.fieldData')}
              </p>
              <pre className="text-xs rounded-lg bg-slate-950 text-slate-100 p-3 overflow-x-auto">
                {JSON.stringify(event.data, null, 2)}
              </pre>
            </div>
          )}
        </div>
        <div className="p-4 border-t border-slate-200 dark:border-slate-700">
          <Button size="sm" variant="ghost" onClick={onClose}>{t('relay.marketplace.cancel')}</Button>
        </div>
      </Card>
    </div>
  )
}
