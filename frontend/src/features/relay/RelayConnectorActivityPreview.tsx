import { ClipboardList } from 'lucide-react'
import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { OverviewPreviewBlock, OVERVIEW_EXPANDED_ROWS } from '../overview/OverviewPreviewBlock'
import { ConnectorActivityTable } from './ConnectorActivityTable'
import { relayActivityHash } from './relayNav'
import { useConnectorActivity } from './useConnectorActivity'

export function RelayConnectorActivityPreview() {
  const { t } = useTranslation()
  const [fetchLimit, setFetchLimit] = useState(OVERVIEW_EXPANDED_ROWS)
  const bumpLimit = useCallback((visible: number) => {
    setFetchLimit((prev) => Math.max(prev, visible))
  }, [])
  const { data, loading, error } = useConnectorActivity(fetchLimit, 0, '', true)

  return (
    <OverviewPreviewBlock
      title={t('relay.activity.title')}
      icon={<ClipboardList className="w-4 h-4 text-sky-600 dark:text-sky-400" />}
      total={data?.total}
      empty={(data?.total ?? 0) === 0 && !loading}
      showAllHash={relayActivityHash()}
      loadedCount={data?.entries.length}
      onVisibleRowsChange={bumpLimit}
    >
      {(visibleCount) => (
        <>
          {loading && <p className="text-sm text-slate-500">{t('relay.activity.loading')}</p>}
          {error && <p className="text-xs text-rose-600">{t('relay.activity.loadError')}</p>}
          {!loading && !error && data && (
            <ConnectorActivityTable entries={data.entries.slice(0, visibleCount)} compact />
          )}
        </>
      )}
    </OverviewPreviewBlock>
  )
}
