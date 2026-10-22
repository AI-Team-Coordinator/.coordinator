import { RadioTower, Settings } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '../../shared/ui/Button'
import { Card } from '../../shared/ui/Card'
import { RelayConnectorActivityPage } from './RelayConnectorActivityPage'
import { RelayConnectorActivityPreview } from './RelayConnectorActivityPreview'
import { RelayExternalEventsPage } from './RelayExternalEventsPage'
import { RelayExternalEventsPreview } from './RelayExternalEventsPreview'
import { RelayMarketplace } from './RelayMarketplace'
import { parseRelaySubview, relaySettingsHash } from './relayNav'
import { RelaySettings } from './RelaySettings'
import { useRelayAccess } from './useRelayAccess'

interface RelaySectionProps {
  isAdmin?: boolean
  currentAlias?: string
}

export function RelaySection({ isAdmin, currentAlias }: RelaySectionProps) {
  const { t } = useTranslation()
  const relay = useRelayAccess()
  const [subview, setSubview] = useState(() => parseRelaySubview(window.location.hash))

  useEffect(() => {
    const onHash = () => setSubview(parseRelaySubview(window.location.hash))
    window.addEventListener('hashchange', onHash)
    return () => window.removeEventListener('hashchange', onHash)
  }, [])

  const openSettings = () => {
    window.location.hash = relaySettingsHash()
  }

  if (subview === 'events') {
    return (
      <section>
        <RelayExternalEventsPage />
      </section>
    )
  }

  if (subview === 'activity') {
    return (
      <section>
        <RelayConnectorActivityPage />
      </section>
    )
  }

  if (subview === 'settings') {
    return (
      <section>
        <RelaySettings
          isAdmin={isAdmin}
          currentAlias={currentAlias}
          access={relay.access}
          url={relay.url}
          setUrl={relay.setUrl}
          token={relay.token}
          setToken={relay.setToken}
          connected={relay.connected}
          loading={relay.loading}
          busy={relay.busy}
          error={relay.error}
          onConnect={relay.connect}
          onDisconnect={relay.disconnect}
        />
      </section>
    )
  }

  return (
    <section className="space-y-8">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center gap-2">
            <RadioTower className="w-4 h-4 text-indigo-500" />
            {t('relay.title')}
          </h2>
          <p className="text-sm text-slate-600 dark:text-slate-300 mt-0.5 max-w-2xl">{t('relay.subtitle')}</p>
        </div>
        <div className="flex items-center gap-2">
          <span
            className={`text-[11px] px-2.5 py-1 rounded-full border ${
              relay.connected
                ? 'border-emerald-200 dark:border-emerald-800 text-emerald-700 dark:text-emerald-300'
                : 'border-amber-200 dark:border-amber-800 text-amber-800 dark:text-amber-200'
            }`}
          >
            {relay.connected ? t('relay.connected') : t('relay.notConnected')}
          </span>
          <Button size="sm" variant="outline" onClick={openSettings} className="gap-1.5">
            <Settings className="w-3.5 h-3.5" />
            {t('relay.settings.open')}
          </Button>
        </div>
      </div>

      {!relay.connected && (
        <Card className="p-4 border-indigo-200/60 dark:border-indigo-900/50 bg-indigo-50/40 dark:bg-indigo-950/20">
          <p className="text-sm text-slate-700 dark:text-slate-300">{t('relay.marketplace.needConnection')}</p>
          <Button size="sm" className="mt-3" onClick={openSettings}>
            {t('relay.settings.open')}
          </Button>
        </Card>
      )}

      <RelayMarketplace
        connected={relay.connected}
        relayBaseUrl={relay.access?.url}
        relayCanConfigure={relay.access?.scope !== 'read'}
        isAdmin={isAdmin}
      />

      <RelayConnectorActivityPreview />

      <RelayExternalEventsPreview />

      <div className="rounded-xl border border-slate-200/80 dark:border-slate-700/80 bg-slate-50/50 dark:bg-slate-900/30 p-5 space-y-2">
        <h3 className="text-sm font-medium text-slate-800 dark:text-slate-200">{t('relay.inboxTitle')}</h3>
        <ul className="text-sm text-slate-600 dark:text-slate-300 space-y-1.5 list-disc pl-4">
          <li>{t('relay.inboxReview')}</li>
          <li>{t('relay.inboxCi')}</li>
          <li>{t('relay.inboxLater')}</li>
        </ul>
        <p className="text-xs text-slate-500 dark:text-slate-400 pt-1">{t('relay.inboxHidden')}</p>
      </div>
    </section>
  )
}
