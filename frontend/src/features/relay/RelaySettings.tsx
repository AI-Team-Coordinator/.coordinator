import { ArrowLeft } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '../../shared/ui/Button'
import { Card } from '../../shared/ui/Card'
import { relayMainHash } from './relayNav'
import { RelayTokensAdmin } from './RelayTokensAdmin'
import { DEFAULT_RELAY_LOCAL, type RelayAccess } from './useRelayAccess'

interface RelaySettingsProps {
  isAdmin?: boolean
  currentAlias?: string
  access: RelayAccess | null
  url: string
  setUrl: (v: string) => void
  token: string
  setToken: (v: string) => void
  connected: boolean
  loading: boolean
  busy: boolean
  error: string
  onConnect: () => void
  onDisconnect: () => void
}

export function RelaySettings({
  isAdmin,
  currentAlias,
  access,
  url,
  setUrl,
  token,
  setToken,
  connected,
  loading,
  busy,
  error,
  onConnect,
  onDisconnect,
}: RelaySettingsProps) {
  const { t } = useTranslation()

  const goBack = () => {
    window.location.hash = relayMainHash()
  }

  return (
    <div className="space-y-6 max-w-2xl">
      <div className="flex items-center gap-3">
        <button
          type="button"
          onClick={goBack}
          className="inline-flex items-center gap-1.5 text-sm text-indigo-600 dark:text-indigo-400 hover:underline"
        >
          <ArrowLeft className="w-4 h-4" />
          {t('relay.settings.back')}
        </button>
      </div>

      <div>
        <h2 className="text-lg font-semibold text-slate-900 dark:text-white">{t('relay.settings.title')}</h2>
        <p className="text-sm text-slate-600 dark:text-slate-300 mt-0.5">{t('relay.settings.subtitle')}</p>
      </div>

      <Card className="p-5 space-y-4 border-slate-200/80 dark:border-slate-700/80 shadow-sm">
        <div className="flex items-center justify-between gap-3">
          <h3 className="text-sm font-medium text-slate-800 dark:text-slate-200">{t('relay.pairTitle')}</h3>
          <span
            className={`text-[11px] px-2.5 py-0.5 rounded-full border font-medium ${
              connected
                ? 'border-emerald-200 dark:border-emerald-800 bg-emerald-50 dark:bg-emerald-950/40 text-emerald-700 dark:text-emerald-300'
                : 'border-slate-200 dark:border-slate-700 text-slate-500'
            }`}
          >
            {connected ? t('relay.connected') : t('relay.notConnected')}
          </span>
        </div>
        <p className="text-xs text-slate-500 dark:text-slate-400 leading-relaxed">{t('relay.pairHelp')}</p>
        <p className="text-xs text-slate-500 dark:text-slate-400">{t('relay.localHint')}</p>
        {connected && access?.scope && (
          <p className="text-xs text-slate-500">
            {t('relay.settings.scope', { scope: access.scope })}
          </p>
        )}
        {connected && access?.scope === 'read' && (
          <p className="text-xs text-amber-700 dark:text-amber-300 leading-relaxed">
            {t('relay.settings.readOnlyToken')}
          </p>
        )}
        {error && <p className="text-xs text-rose-600 dark:text-rose-400">{error}</p>}
        <label className="block">
          <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('relay.url')}</span>
          <input
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder={DEFAULT_RELAY_LOCAL}
            disabled={connected || busy}
            className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-3 py-2 text-sm disabled:opacity-60"
          />
        </label>
        <label className="block">
          <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('relay.token')}</span>
          <input
            type="password"
            value={token}
            onChange={(e) => setToken(e.target.value)}
            autoComplete="off"
            disabled={connected || busy}
            className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-3 py-2 text-sm disabled:opacity-60"
          />
        </label>
        <div className="flex flex-wrap items-center gap-3 pt-1">
          {connected ? (
            <Button size="sm" variant="outline" disabled={busy} onClick={() => void onDisconnect()}>
              {t('relay.disconnect')}
            </Button>
          ) : (
            <Button
              size="sm"
              disabled={busy || loading || !url.trim() || !token.trim()}
              onClick={() => void onConnect()}
            >
              {busy ? t('relay.connecting') : t('relay.connect')}
            </Button>
          )}
          {!connected && !busy && (
            <button
              type="button"
              className="text-[11px] text-indigo-600 dark:text-indigo-400 hover:underline"
              onClick={() => setUrl(DEFAULT_RELAY_LOCAL)}
            >
              {t('relay.useLocalDefault')}
            </button>
          )}
        </div>
      </Card>

      {isAdmin && (
        <RelayTokensAdmin currentAlias={currentAlias} onUseSecret={(secret) => setToken(secret)} />
      )}
    </div>
  )
}
