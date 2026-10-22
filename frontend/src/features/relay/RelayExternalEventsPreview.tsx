import { ScrollText } from 'lucide-react'
import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { OverviewPreviewBlock, OVERVIEW_EXPANDED_ROWS } from '../overview/OverviewPreviewBlock'
import { ExternalEventsTable } from './ExternalEventsTable'
import { relayEventsHash } from './relayNav'
import { useExternalEvents } from './useExternalEvents'

export function RelayExternalEventsPreview() {
  const { t } = useTranslation()
  const [fetchLimit, setFetchLimit] = useState(OVERVIEW_EXPANDED_ROWS)
  const bumpLimit = useCallback((visible: number) => {
    setFetchLimit((prev) => Math.max(prev, visible))
  }, [])
  const { data, loading, error, reload } = useExternalEvents(fetchLimit, 0, '', true)

  return (
    <OverviewPreviewBlock
      title={t('relay.events.title')}
      icon={<ScrollText className="w-4 h-4 text-violet-600 dark:text-violet-400" />}
      total={data?.total}
      empty={(data?.total ?? 0) === 0 && !loading}
      showAllHash={relayEventsHash()}
      loadedCount={data?.events.length}
      onVisibleRowsChange={bumpLimit}
    >
      {(visibleCount) => (
        <>
          {loading && <p className="text-sm text-slate-500">{t('relay.events.loading')}</p>}
          {error && <p className="text-xs text-rose-600">{t('relay.events.loadError')}</p>}
          {!loading && !error && data && (
            <ExternalEventsTable
              events={data.events}
              onTriageChanged={reload}
              compact
              maxRows={visibleCount}
            />
          )}
        </>
      )}
    </OverviewPreviewBlock>
  )
}
