import { Plug, Webhook, X } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../../shared/api/client'
import { Button } from '../../shared/ui/Button'
import { Card } from '../../shared/ui/Card'
import { RelayConnectorActionsModal } from './RelayConnectorActionsModal'
import {
  applyRelayConfigToForm,
  buildConnectorSavePayload,
  type RelayConnectorConfig,
} from './relayConfigForm'
import { RelayTileIcon } from './RelayTileIcon'

export type MarketplaceAction = {
  id: string
  label: string
}

export type MarketplaceTile = {
  id: string
  name: string
  detail: string
  icon: string
  capabilities: string[]
  actions?: MarketplaceAction[]
  connectable: boolean
  config_schema: ConfigField[]
  enabled: boolean
  connected: boolean
}

type ConfigField = {
  id: string
  label: string
  type: string
  required: boolean
  help?: string
}

function capabilityLabel(cap: string, t: (k: string) => string): string {
  if (cap === 'webhook') return t('relay.marketplace.capWebhook')
  if (cap === 'poll') return t('relay.marketplace.capPoll')
  return cap
}

interface RelayMarketplaceProps {
  connected: boolean
  relayBaseUrl?: string
  relayCanConfigure?: boolean
  isAdmin?: boolean
  onChanged?: () => void
}

export function RelayMarketplace({ connected, relayBaseUrl, relayCanConfigure, isAdmin, onChanged }: RelayMarketplaceProps) {
  const { t } = useTranslation()
  const [tiles, setTiles] = useState<MarketplaceTile[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [active, setActive] = useState<MarketplaceTile | null>(null)
  const [actionsTile, setActionsTile] = useState<MarketplaceTile | null>(null)
  const [secrets, setSecrets] = useState<Record<string, string>>({})
  const [busy, setBusy] = useState(false)
  const [confirmDisable, setConfirmDisable] = useState(false)
  const [configLoading, setConfigLoading] = useState(false)
  const [savedSecretFields, setSavedSecretFields] = useState<string[]>([])
  const [loadedNonSecret, setLoadedNonSecret] = useState<Record<string, unknown>>({})

  const load = useCallback(() => {
    if (!connected) {
      setTiles([])
      return
    }
    setLoading(true)
    setError('')
    fetch(api.relayMarketplace)
      .then((r) => (r.ok ? r.json() : Promise.reject(new Error('load'))))
      .then((data: { connectors?: MarketplaceTile[] }) => {
        setTiles(data.connectors ?? [])
      })
      .catch(() => setError(t('relay.marketplace.loadError')))
      .finally(() => setLoading(false))
  }, [connected, t])

  useEffect(() => {
    load()
  }, [load])

  const openConfig = (tile: MarketplaceTile) => {
    setActive(tile)
    setSecrets({})
    setSavedSecretFields([])
    setLoadedNonSecret({})
    setError('')
    setConfirmDisable(false)
    if (!connected) {
      return
    }
    setConfigLoading(true)
    fetch(`${api.relayConnectorConfig(tile.id)}`)
      .then((r) => (r.ok ? r.json() : Promise.reject(new Error('config'))))
      .then((cfg: RelayConnectorConfig) => {
        setLoadedNonSecret(cfg.non_secret ?? {})
        const { next, savedSecrets } = applyRelayConfigToForm(tile, cfg)
        setSecrets(next)
        setSavedSecretFields(savedSecrets)
      })
      .catch(() => setError(t('relay.marketplace.configLoadError')))
      .finally(() => setConfigLoading(false))
  }

  const openTile = (tile: MarketplaceTile) => {
    if (!tile.connectable) {
      return
    }
    if (tile.connected || tile.enabled) {
      setActionsTile(tile)
      return
    }
    openConfig(tile)
  }

  const closeModal = () => {
    setActive(null)
    setSecrets({})
    setSavedSecretFields([])
    setConfigLoading(false)
    setConfirmDisable(false)
  }

  const save = async (enable: boolean) => {
    if (!active || !isAdmin) return
    if (relayCanConfigure === false) {
      setError(t('relay.marketplace.needMasterToken'))
      return
    }
    setBusy(true)
    setError('')
    try {
      const payload = buildConnectorSavePayload(active, secrets, loadedNonSecret, enable)
      const res = await fetch(api.relayConnectorConfig(active.id), {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
        signal: AbortSignal.timeout(95_000),
      })
      if (!res.ok) {
        const { readError } = await import('../../shared/api/http')
        const msg = await readError(res)
        if (res.status === 403) {
          throw new Error(t('relay.marketplace.needMasterToken'))
        }
        throw new Error(msg)
      }
      const saved = (await res.json().catch(() => null)) as RelayConnectorConfig | null
      if (saved) {
        setLoadedNonSecret(saved.non_secret ?? {})
        const { next, savedSecrets } = applyRelayConfigToForm(active, saved)
        setSecrets(next)
        setSavedSecretFields(savedSecrets)
        if (enable && !saved.enabled && !saved.connected) {
          throw new Error(t('relay.marketplace.saveNotPersisted'))
        }
      }
      closeModal()
      load()
      onChanged?.()
    } catch (e) {
      if (e instanceof Error && e.name === 'TimeoutError') {
        setError(t('relay.marketplace.timeoutError'))
      } else {
        setError(e instanceof Error ? e.message : t('relay.marketplace.saveError'))
      }
    } finally {
      setBusy(false)
    }
  }

  const disable = async (clearSecrets: boolean) => {
    if (!active || !isAdmin) return
    setBusy(true)
    setError('')
    try {
      const res = await fetch(api.relayConnectorDisable(active.id), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ clear_secrets: clearSecrets }),
        signal: AbortSignal.timeout(95_000),
      })
      if (!res.ok) {
        const { readError } = await import('../../shared/api/http')
        throw new Error(await readError(res))
      }
      closeModal()
      load()
      onChanged?.()
    } catch (e) {
      setError(e instanceof Error ? e.message : t('relay.marketplace.disableError'))
    } finally {
      setBusy(false)
      setConfirmDisable(false)
    }
  }

  const statusBadge = (tile: MarketplaceTile) => {
    if (!tile.connectable) {
      return (
        <span className="text-[10px] uppercase tracking-wide text-slate-400">{t('relay.marketplace.comingSoon')}</span>
      )
    }
    if (tile.enabled) {
      return (
        <span className="text-[10px] font-medium uppercase tracking-wide text-emerald-700 dark:text-emerald-300">
          {t('relay.marketplace.badgeOn')}
        </span>
      )
    }
    if (tile.connected) {
      return (
        <span className="text-[10px] font-medium uppercase tracking-wide text-amber-700 dark:text-amber-300">
          {t('relay.marketplace.badgeConfigured')}
        </span>
      )
    }
    return (
      <span className="text-[10px] font-medium uppercase tracking-wide text-slate-400">{t('relay.marketplace.badgeOff')}</span>
    )
  }

  return (
    <div className="space-y-4">
      <div>
        <h3 className="text-base font-semibold text-slate-900 dark:text-white">{t('relay.marketplace.title')}</h3>
        <p className="text-sm text-slate-500 dark:text-slate-400 mt-1 max-w-2xl">{t('relay.marketplace.help')}</p>
      </div>
      {error && !active && <p className="text-xs text-rose-600 dark:text-rose-400">{error}</p>}
      {!connected ? null : loading ? (
        <p className="text-sm text-slate-500">{t('relay.marketplace.loading')}</p>
      ) : tiles.length === 0 ? (
        <p className="text-sm text-slate-500">{t('relay.marketplace.empty')}</p>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {tiles.map((tile) => (
            <button
              key={tile.id}
              type="button"
              disabled={!tile.connectable}
              onClick={() => openTile(tile)}
              className={`group text-left rounded-2xl border p-5 transition-all duration-200 ${
                tile.enabled
                  ? 'border-emerald-300/80 dark:border-emerald-800 bg-gradient-to-br from-emerald-50/90 to-white dark:from-emerald-950/40 dark:to-slate-950 shadow-sm'
                  : 'border-slate-200/90 dark:border-slate-700/90 bg-white dark:bg-slate-950 hover:border-indigo-300 dark:hover:border-indigo-600 hover:shadow-md'
              } ${!tile.connectable ? 'opacity-55 cursor-not-allowed' : 'cursor-pointer'}`}
            >
              <div className="flex items-start justify-between gap-3">
                <RelayTileIcon icon={tile.icon} id={tile.id} />
                {statusBadge(tile)}
              </div>
              <div className="mt-4 text-base font-semibold text-slate-900 dark:text-white group-hover:text-indigo-700 dark:group-hover:text-indigo-300 transition-colors">
                {tile.name}
              </div>
              <p className="text-sm text-slate-500 dark:text-slate-400 mt-1.5 line-clamp-3 leading-relaxed">{tile.detail}</p>
              {(tile.capabilities?.length ?? 0) > 0 && (
                <div className="flex flex-wrap gap-1.5 mt-4">
                  {(tile.capabilities ?? []).map((cap) => (
                    <span
                      key={cap}
                      className="inline-flex items-center gap-1 text-[10px] px-2 py-0.5 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300"
                    >
                      {cap === 'webhook' ? <Webhook className="w-3 h-3" /> : null}
                      {capabilityLabel(cap, t)}
                    </span>
                  ))}
                </div>
              )}
            </button>
          ))}
        </div>
      )}

      {actionsTile && (
        <RelayConnectorActionsModal
          tile={actionsTile}
          relayBaseUrl={relayBaseUrl}
          isAdmin={isAdmin}
          onClose={() => setActionsTile(null)}
          onOpenSettings={() => {
            const tile = actionsTile
            setActionsTile(null)
            if (tile) openConfig(tile)
          }}
          onActionDone={() => load()}
        />
      )}

      {active && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/45 backdrop-blur-[2px]">
          <Card className="w-full max-w-md p-5 space-y-3 shadow-2xl border-slate-200 dark:border-slate-700">
            <div className="flex items-start justify-between gap-2">
              <div className="flex items-center gap-2 min-w-0">
                <Plug className="w-4 h-4 text-indigo-500 shrink-0" />
                <h4 className="text-sm font-semibold text-slate-900 dark:text-white truncate">
                  {active.connected ? t('relay.marketplace.modalSettings') : t('relay.marketplace.modalConnect')}
                  {' — '}
                  {active.name}
                </h4>
              </div>
              <button type="button" className="text-slate-400 hover:text-slate-600 shrink-0" onClick={closeModal}>
                <X className="w-4 h-4" />
              </button>
            </div>
            {error && <p className="text-xs text-rose-600 dark:text-rose-400">{error}</p>}
            {!isAdmin && <p className="text-xs text-amber-700 dark:text-amber-300">{t('relay.marketplace.adminOnly')}</p>}
            {configLoading && (
              <p className="text-xs text-slate-500 dark:text-slate-400">{t('relay.marketplace.configLoading')}</p>
            )}
            {(active.config_schema ?? []).map((field) => (
              <label key={field.id} className="block">
                <span className="text-[11px] uppercase tracking-wide text-slate-500">{field.label}</span>
                <input
                  type={field.type === 'secret' ? 'password' : 'text'}
                  value={secrets[field.id] ?? ''}
                  disabled={!isAdmin || busy || configLoading}
                  placeholder={
                    field.type === 'secret' && savedSecretFields.includes(field.id)
                      ? t('relay.marketplace.secretSavedPlaceholder')
                      : undefined
                  }
                  onChange={(e) => setSecrets((s) => ({ ...s, [field.id]: e.target.value }))}
                  className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm disabled:opacity-60"
                />
                {field.type === 'secret' && savedSecretFields.includes(field.id) && (
                  <p className="text-[11px] text-emerald-700 dark:text-emerald-400 mt-0.5">
                    {t('relay.marketplace.secretSavedHint')}
                  </p>
                )}
                {field.type !== 'secret' && (secrets[field.id] ?? '') !== '' && !configLoading && (
                  <p className="text-[11px] text-slate-500 dark:text-slate-400 mt-0.5">
                    {t('relay.marketplace.nonSecretOnRelay')}
                  </p>
                )}
                {field.help && <p className="text-[11px] text-slate-500 mt-0.5">{field.help}</p>}
              </label>
            ))}
            {confirmDisable && (
              <div
                className="rounded-lg border border-amber-200 dark:border-amber-900/80 bg-amber-50/90 dark:bg-amber-950/40 px-3 py-2.5 text-sm text-amber-950 dark:text-amber-100"
                role="alert"
              >
                <p className="font-medium">{t('relay.marketplace.disableConfirmTitle')}</p>
                <p className="mt-1 text-xs text-amber-900/90 dark:text-amber-200/90 leading-relaxed">
                  {t('relay.marketplace.disableConfirmBody', { name: active.name })}
                </p>
              </div>
            )}
            <div className="flex flex-wrap gap-2 pt-1">
              {isAdmin && !confirmDisable && (
                <>
                  <Button size="sm" disabled={busy} onClick={() => void save(true)}>
                    {active.connected ? t('relay.marketplace.save') : t('relay.marketplace.connect')}
                  </Button>
                  {active.connected && (
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={busy}
                      onClick={() => {
                        setError('')
                        setConfirmDisable(true)
                      }}
                    >
                      {t('relay.marketplace.disable')}
                    </Button>
                  )}
                </>
              )}
              {isAdmin && confirmDisable && (
                <>
                  <Button
                    size="sm"
                    variant="outline"
                    disabled={busy}
                    className="border-rose-300 text-rose-700 hover:bg-rose-50 dark:border-rose-800 dark:text-rose-300 dark:hover:bg-rose-950/50"
                    onClick={() => void disable(false)}
                  >
                    {t('relay.marketplace.disableConfirmAction')}
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    disabled={busy}
                    onClick={() => {
                      setConfirmDisable(false)
                      setError('')
                    }}
                  >
                    {t('relay.marketplace.disableConfirmBack')}
                  </Button>
                </>
              )}
              <Button size="sm" variant="ghost" disabled={busy} onClick={closeModal}>
                {t('relay.marketplace.cancel')}
              </Button>
            </div>
          </Card>
        </div>
      )}
    </div>
  )
}
