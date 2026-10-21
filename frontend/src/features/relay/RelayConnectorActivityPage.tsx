import { ArrowLeft } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../../shared/api/client'
import { Button } from '../../shared/ui/Button'
import { Select } from '../../shared/ui/Select'
import { ConnectorActivityTable } from './ConnectorActivityTable'
import type { MarketplaceTile } from './RelayMarketplace'
import { relayMainHash } from './relayNav'
import { useConnectorActivity } from './useConnectorActivity'

const PAGE_SIZE = 30

export function RelayConnectorActivityPage() {
  const { t } = useTranslation()
  const [connectorId, setConnectorId] = useState('')
  const [offset, setOffset] = useState(0)
  const [tiles, setTiles] = useState<MarketplaceTile[]>([])
  const { data, loading, error } = useConnectorActivity(PAGE_SIZE, offset, connectorId)

  useEffect(() => {
    fetch(api.relayMarketplace)
      .then((r) => (r.ok ? r.json() : Promise.reject(new Error('load'))))
      .then((body: { connectors?: MarketplaceTile[] }) => setTiles(body.connectors ?? []))
      .catch(() => setTiles([]))
  }, [])

  const connectorOptions = useMemo(() => {
    const opts = [{ value: '', label: t('relay.activity.filterAll') }]
    for (const tile of tiles) {
      opts.push({ value: tile.id, label: tile.name || tile.id })
    }
    return opts
  }, [tiles, t])

  const total = data?.total ?? 0
  const pageStart = total === 0 ? 0 : offset + 1
  const pageEnd = Math.min(offset + PAGE_SIZE, total)
  const canPrev = offset > 0
  const canNext = offset + PAGE_SIZE < total

  const goBack = () => {
    window.location.hash = relayMainHash()
  }

  const onConnectorChange = (value: string) => {
    setConnectorId(value)
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
        {t('relay.activity.back')}
      </button>

      <div>
        <h2 className="text-lg font-semibold text-slate-900 dark:text-white">{t('relay.activity.allTitle')}</h2>
        <p className="text-sm text-slate-600 dark:text-slate-300 mt-0.5">{t('relay.activity.allHelp')}</p>
      </div>

      <div className="flex flex-wrap items-end gap-4">
        <label className="block min-w-[200px]">
          <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('relay.activity.filterConnector')}</span>
          <Select className="mt-1" value={connectorId} onChange={(e) => onConnectorChange(e.target.value)}>
            {connectorOptions.map((opt) => (
              <option key={opt.value || '__all'} value={opt.value}>{opt.label}</option>
            ))}
          </Select>
        </label>
        {total > 0 && (
          <p className="text-xs text-slate-500 pb-1">
            {t('relay.activity.pageRange', { start: pageStart, end: pageEnd, total })}
          </p>
        )}
      </div>

      {loading && <p className="text-sm text-slate-500">{t('relay.activity.loading')}</p>}
      {error && <p className="text-xs text-rose-600">{t('relay.activity.loadError')}</p>}
      {!loading && !error && data && <ConnectorActivityTable entries={data.entries} />}

      <div className="flex flex-wrap gap-2">
        <Button size="sm" variant="outline" disabled={!canPrev || loading} onClick={() => setOffset((o) => Math.max(0, o - PAGE_SIZE))}>
          {t('relay.activity.prev')}
        </Button>
        <Button size="sm" variant="outline" disabled={!canNext || loading} onClick={() => setOffset((o) => o + PAGE_SIZE)}>
          {t('relay.activity.next')}
        </Button>
      </div>
    </div>
  )
}
