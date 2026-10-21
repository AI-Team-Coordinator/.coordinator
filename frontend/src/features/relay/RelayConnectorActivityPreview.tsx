import { ClipboardList } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '../../shared/ui/Button'
import { ConnectorActivityTable } from './ConnectorActivityTable'
import { relayActivityHash } from './relayNav'
import { useConnectorActivity } from './useConnectorActivity'

const PREVIEW_LIMIT = 15

export function RelayConnectorActivityPreview() {
  const { t } = useTranslation()
  const { data, loading, error } = useConnectorActivity(PREVIEW_LIMIT, 0, '')

  const openAll = () => {
    window.location.hash = relayActivityHash()
  }

  return (
    <div className="rounded-xl border border-slate-200/80 dark:border-slate-700/80 bg-white dark:bg-slate-950/40 p-5 space-y-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex items-center gap-2">
          <ClipboardList className="w-4 h-4 text-sky-600 dark:text-sky-400" />
          <h3 className="text-base font-semibold text-slate-900 dark:text-white">{t('relay.activity.title')}</h3>
          {data && (
            <span className="text-xs text-slate-500">{t('relay.activity.total', { count: data.total })}</span>
          )}
        </div>
        <Button size="sm" variant="outline" onClick={openAll}>
          {t('relay.activity.viewAll')}
        </Button>
      </div>
      <p className="text-sm text-slate-500 dark:text-slate-400">{t('relay.activity.previewHelp')}</p>
      {loading && <p className="text-sm text-slate-500">{t('relay.activity.loading')}</p>}
      {error && <p className="text-xs text-rose-600">{t('relay.activity.loadError')}</p>}
      {!loading && !error && data && <ConnectorActivityTable entries={data.entries} compact />}
    </div>
  )
}
