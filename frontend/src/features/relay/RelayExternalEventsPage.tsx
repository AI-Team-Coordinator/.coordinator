import { ArrowLeft } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { ExternalEvent } from '../../shared/types/api'
import { Button } from '../../shared/ui/Button'
import { Select } from '../../shared/ui/Select'
import { ExternalEventDetailModal } from './ExternalEventDetailModal'
import { ExternalEventsTable } from './ExternalEventsTable'
import { relayMainHash } from './relayNav'
import { useExternalEvents } from './useExternalEvents'

const PAGE_SIZE = 30

interface RelayExternalEventsPageProps {
  backHash?: string
}

export function RelayExternalEventsPage({ backHash }: RelayExternalEventsPageProps = {} as RelayExternalEventsPageProps) {
  const { t } = useTranslation()
  const [source, setSource] = useState('')
  const [offset, setOffset] = useState(0)
  const [selected, setSelected] = useState<ExternalEvent | null>(null)
  const { data, loading, error, reload } = useExternalEvents(PAGE_SIZE, offset, source)

  const sourceOptions = useMemo(() => {
    const opts = [{ value: '', label: t('relay.events.filterAll') }]
    for (const s of data?.sources ?? []) {
      opts.push({ value: s, label: s })
    }
    return opts
  }, [data?.sources, t])

  const total = data?.total ?? 0
  const pageStart = total === 0 ? 0 : offset + 1
  const pageEnd = Math.min(offset + PAGE_SIZE, total)
  const canPrev = offset > 0
  const canNext = offset + PAGE_SIZE < total

  const goBack = () => {
    window.location.hash = backHash ?? relayMainHash()
  }

  const onSourceChange = (value: string) => {
    setSource(value)
    setOffset(0)
  }

  return (
    <div className="space-y-6 max-w-3xl">
      <button
        type="button"
        onClick={goBack}
        className="inline-flex items-center gap-1.5 text-sm text-indigo-600 dark:text-indigo-400 hover:underline"
      >
        <ArrowLeft className="w-4 h-4" />
        {t('relay.events.back')}
      </button>

      <div>
        <h2 className="text-lg font-semibold text-slate-900 dark:text-white">{t('relay.events.allTitle')}</h2>
        <p className="text-sm text-slate-600 dark:text-slate-300 mt-0.5">{t('relay.events.allHelp')}</p>
      </div>

      <div className="flex flex-wrap items-end gap-4">
        <label className="block min-w-[200px]">
          <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('relay.events.filterSource')}</span>
          <Select className="mt-1" value={source} onChange={(e) => onSourceChange(e.target.value)}>
            {sourceOptions.map((opt) => (
              <option key={opt.value || '__all'} value={opt.value}>{opt.label}</option>
            ))}
          </Select>
        </label>
        {total > 0 && (
          <p className="text-xs text-slate-500 pb-1">
            {t('relay.events.pageRange', { start: pageStart, end: pageEnd, total })}
          </p>
        )}
      </div>

      {loading && <p className="text-sm text-slate-500">{t('relay.events.loading')}</p>}
      {error && <p className="text-xs text-rose-600">{t('relay.events.loadError')}</p>}
      {!loading && !error && data && (
        <ExternalEventsTable events={data.events} onSelect={setSelected} />
      )}

      <div className="flex flex-wrap gap-2">
        <Button size="sm" variant="outline" disabled={!canPrev || loading} onClick={() => setOffset((o) => Math.max(0, o - PAGE_SIZE))}>
          {t('relay.events.prev')}
        </Button>
        <Button size="sm" variant="outline" disabled={!canNext || loading} onClick={() => setOffset((o) => o + PAGE_SIZE)}>
          {t('relay.events.next')}
        </Button>
      </div>

      {selected && (
        <ExternalEventDetailModal
          event={selected}
          onClose={() => setSelected(null)}
          onTriageChanged={reload}
        />
      )}
    </div>
  )
}
