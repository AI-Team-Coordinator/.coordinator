import { ScrollText } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '../../shared/ui/Button'
import type { ExternalEvent } from '../../shared/types/api'
import { ExternalEventDetailModal } from './ExternalEventDetailModal'
import { ExternalEventsTable } from './ExternalEventsTable'
import { relayEventsHash } from './relayNav'
import { useExternalEvents } from './useExternalEvents'

const PREVIEW_LIMIT = 20

export function RelayExternalEventsPreview() {
  const { t } = useTranslation()
  const { data, loading, error } = useExternalEvents(PREVIEW_LIMIT, 0, '')
  const [selected, setSelected] = useState<ExternalEvent | null>(null)

  const openAll = () => {
    window.location.hash = relayEventsHash()
  }

  return (
    <div className="rounded-xl border border-slate-200/80 dark:border-slate-700/80 bg-white dark:bg-slate-950/40 p-5 space-y-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex items-center gap-2">
          <ScrollText className="w-4 h-4 text-violet-600 dark:text-violet-400" />
          <h3 className="text-base font-semibold text-slate-900 dark:text-white">{t('relay.events.title')}</h3>
          {data && (
            <span className="text-xs text-slate-500">
              {t('relay.events.total', { count: data.total })}
            </span>
          )}
        </div>
        <Button size="sm" variant="outline" onClick={openAll}>
          {t('relay.events.viewAll')}
        </Button>
      </div>
      <p className="text-sm text-slate-500 dark:text-slate-400">{t('relay.events.previewHelp')}</p>
      {loading && <p className="text-sm text-slate-500">{t('relay.events.loading')}</p>}
      {error && <p className="text-xs text-rose-600">{t('relay.events.loadError')}</p>}
      {!loading && !error && data && (
        <ExternalEventsTable events={data.events} onSelect={setSelected} compact />
      )}
      {selected && <ExternalEventDetailModal event={selected} onClose={() => setSelected(null)} />}
    </div>
  )
}
