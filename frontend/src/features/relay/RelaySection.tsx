import { RadioTower } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../../shared/api/client'
import i18n from '../../shared/i18n/config'
import { useLoadingStatus } from '../../shared/loading/LoadingStatusContext'
import { Button } from '../../shared/ui/Button'
import { Card } from '../../shared/ui/Card'
import { RelayTokensAdmin } from './RelayTokensAdmin'

export type RelaySourceView = {
  id: string
  name: string
  detail: string
  enabled: boolean
}

type RelayAccess = {
  connected: boolean
  url?: string
  scope?: string
}

const DEFAULT_LOCAL = 'http://127.0.0.1:4322'

interface RelaySectionProps {
  isAdmin?: boolean
  currentAlias?: string
}

export function RelaySection({ isAdmin, currentAlias }: RelaySectionProps) {
  const { t } = useTranslation()
  const { track } = useLoadingStatus()
  const [url, setUrl] = useState(DEFAULT_LOCAL)
  const [token, setToken] = useState('')
  const [access, setAccess] = useState<RelayAccess | null>(null)
  const [sources, setSources] = useState<RelaySourceView[]>([])
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')

  const loadSources = useCallback(() => {
    if (!access?.connected) {
      setSources([])
      return
    }
    fetch(api.relaySources)
      .then((r) => (r.ok ? r.json() : null))
      .then((data: { sources?: RelaySourceView[] } | null) => {
        setSources(data?.sources ?? [])
      })
      .catch(() => setSources([]))
  }, [access?.connected])

  useEffect(() => {
    setError('')
    void track('relay-access', i18n.t('loading.relay'), async () => {
      setLoading(true)
      try {
        const r = await fetch(api.relayAccess)
        const data = r.ok ? ((await r.json()) as RelayAccess) : null
        const next = data ?? { connected: false }
        setAccess(next)
        if (next.connected && next.url) {
          setUrl(next.url)
        }
      } catch {
        setError(t('relay.loadError'))
      } finally {
        setLoading(false)
      }
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    loadSources()
  }, [loadSources])

  const connect = async () => {
    setBusy(true)
    setError('')
    try {
      await track('relay-connect', t('loading.relayConnect'), async () => {
        const res = await fetch(api.relayConnect, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ url: url.trim(), token: token.trim() }),
        })
        const data = await res.json().catch(() => ({}))
        if (!res.ok) {
          const msg =
            typeof data.error === 'string'
              ? data.error
              : typeof data.message === 'string'
                ? data.message
                : t('relay.connectError')
          throw new Error(msg)
        }
        setAccess(data as RelayAccess)
        const srcRes = await fetch(api.relaySources)
        if (srcRes.ok) {
          const srcData = (await srcRes.json()) as { sources?: RelaySourceView[] }
          setSources(srcData.sources ?? [])
        }
      })
    } catch (e) {
      setError(e instanceof Error ? e.message : t('relay.connectError'))
    } finally {
      setBusy(false)
    }
  }

  const disconnect = async () => {
    setBusy(true)
    setError('')
    try {
      const res = await fetch(api.relayDisconnect, { method: 'POST' })
      if (!res.ok) throw new Error('disconnect')
      const data = (await res.json()) as RelayAccess
      setAccess(data)
      setSources([])
    } catch {
      setError(t('relay.disconnectError'))
    } finally {
      setBusy(false)
    }
  }

  const connected = Boolean(access?.connected)

  return (
    <section className="space-y-8">
      <div>
        <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center gap-2">
          <RadioTower className="w-4 h-4 text-indigo-500" />
          {t('relay.title')}
        </h2>
        <p className="text-sm text-slate-600 dark:text-slate-300 mt-0.5">{t('relay.subtitle')}</p>
      </div>

      <Card className="p-4 space-y-3">
        <div className="flex items-center justify-between gap-3">
          <h3 className="text-sm font-medium text-slate-800 dark:text-slate-200">{t('relay.pairTitle')}</h3>
          <span
            className={`text-[11px] px-2 py-0.5 rounded-full border ${
              connected
                ? 'border-emerald-200 dark:border-emerald-800 text-emerald-700 dark:text-emerald-300'
                : 'border-slate-200 dark:border-slate-700 text-slate-500'
            }`}
          >
            {connected ? t('relay.connected') : t('relay.notConnected')}
          </span>
        </div>
        <p className="text-xs text-slate-500 dark:text-slate-400">{t('relay.pairHelp')}</p>
        <p className="text-xs text-slate-500 dark:text-slate-400">{t('relay.localHint')}</p>
        {error && <p className="text-xs text-rose-600 dark:text-rose-400">{error}</p>}
        <label className="block max-w-lg">
          <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('relay.url')}</span>
          <input
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder={DEFAULT_LOCAL}
            disabled={connected || busy}
            className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm disabled:opacity-60"
          />
        </label>
        <label className="block max-w-lg">
          <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('relay.token')}</span>
          <input
            type="password"
            value={token}
            onChange={(e) => setToken(e.target.value)}
            autoComplete="off"
            disabled={connected || busy}
            className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm disabled:opacity-60"
          />
        </label>
        <div className="flex flex-wrap items-center gap-3">
          {connected ? (
            <Button size="sm" variant="outline" disabled={busy} onClick={() => void disconnect()}>
              {t('relay.disconnect')}
            </Button>
          ) : (
            <Button
              size="sm"
              disabled={busy || loading || !url.trim() || !token.trim()}
              onClick={() => void connect()}
            >
              {busy ? t('relay.connecting') : t('relay.connect')}
            </Button>
          )}
          {!connected && !busy && (
            <button
              type="button"
              className="text-[11px] text-indigo-600 dark:text-indigo-400 hover:underline"
              onClick={() => setUrl(DEFAULT_LOCAL)}
            >
              {t('relay.useLocalDefault')}
            </button>
          )}
        </div>
      </Card>

      {isAdmin && (
        <RelayTokensAdmin
          currentAlias={currentAlias}
          onUseSecret={(secret) => setToken(secret)}
        />
      )}

      <div className="space-y-2">
        <h3 className="text-sm font-medium text-slate-800 dark:text-slate-200">{t('relay.inboxTitle')}</h3>
        <ul className="text-sm text-slate-600 dark:text-slate-300 space-y-1.5 list-disc pl-4">
          <li>{t('relay.inboxReview')}</li>
          <li>{t('relay.inboxCi')}</li>
          <li>{t('relay.inboxLater')}</li>
        </ul>
        <p className="text-xs text-slate-500 dark:text-slate-400">{t('relay.inboxHidden')}</p>
      </div>

      <div className="space-y-3">
        <div>
          <h3 className="text-sm font-medium text-slate-800 dark:text-slate-200">{t('relay.adaptersTitle')}</h3>
          <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">{t('relay.adaptersHelp')}</p>
        </div>
        {!connected ? (
          <p className="text-sm text-slate-500 dark:text-slate-400">{t('relay.adaptersEmpty')}</p>
        ) : sources.length === 0 ? (
          <p className="text-sm text-slate-500 dark:text-slate-400">{t('relay.sourcesLoading')}</p>
        ) : (
          sources.map((src) => (
            <Card key={src.id} className="p-4 flex items-start justify-between gap-3">
              <div>
                <div className="text-sm font-medium text-slate-900 dark:text-white">{src.name}</div>
                <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">{src.detail}</p>
              </div>
              <span className="shrink-0 text-[11px] text-slate-400">
                {src.enabled ? t('relay.adapterOn') : t('relay.adapterOff')}
              </span>
            </Card>
          ))
        )}
      </div>
    </section>
  )
}
