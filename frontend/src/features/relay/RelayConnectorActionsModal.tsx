import { Settings, X } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../../shared/api/client'
import { Button } from '../../shared/ui/Button'
import { Card } from '../../shared/ui/Card'
import { ConnectorActivityTable } from './ConnectorActivityTable'
import type { MarketplaceTile } from './RelayMarketplace'
import { RelayTileIcon } from './RelayTileIcon'
import { useConnectorActivity } from './useConnectorActivity'

interface RelayConnectorActionsModalProps {
  tile: MarketplaceTile
  relayBaseUrl?: string
  isAdmin?: boolean
  onClose: () => void
  onOpenSettings: () => void
  onActionDone?: () => void
}

export function RelayConnectorActionsModal({
  tile,
  relayBaseUrl,
  isAdmin,
  onClose,
  onOpenSettings,
  onActionDone,
}: RelayConnectorActionsModalProps) {
  const { t } = useTranslation()
  const [busyAction, setBusyAction] = useState<string | null>(null)
  const [actionMsg, setActionMsg] = useState('')
  const [actionError, setActionError] = useState(false)
  const activity = useConnectorActivity(8, 0, tile.id)

  const actions = useMemo(() => {
    if (tile.actions?.length) {
      return tile.actions
    }
    if (tile.capabilities.includes('poll')) {
      return [{ id: 'poll_now', label: t('relay.marketplace.pollNow') }]
    }
    return []
  }, [tile.actions, tile.capabilities, t])

  const hasWebhook = tile.capabilities.includes('webhook')

  const runAction = async (actionId: string) => {
    if (!isAdmin) return
    setBusyAction(actionId)
    setActionMsg('')
    setActionError(false)
    try {
      const res = await fetch(api.relayConnectorAction(tile.id, actionId), {
        method: 'POST',
        signal: AbortSignal.timeout(130_000),
      })
      if (!res.ok) {
        const { readError } = await import('../../shared/api/http')
        throw new Error(await readError(res))
      }
      await res.json().catch(() => undefined)
      setActionMsg(
        actionId === 'poll_now' ? t('relay.marketplace.pollOk') : t('relay.marketplace.actionOk'),
      )
      if (actionId === 'poll_now') {
        activity.reload()
      }
      onActionDone?.()
    } catch (e) {
      setActionError(true)
      setActionMsg(e instanceof Error ? e.message : t('relay.marketplace.actionError'))
    } finally {
      setBusyAction(null)
    }
  }

  const webhookUrl =
    relayBaseUrl && hasWebhook
      ? `${relayBaseUrl.replace(/\/$/, '')}/connectors/${encodeURIComponent(tile.id)}/webhook`
      : ''

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/45 backdrop-blur-[2px]">
      <Card className="w-full max-w-md shadow-2xl overflow-hidden">
        <div className="flex items-start gap-3 p-4 border-b border-slate-200 dark:border-slate-700">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-200">
            <RelayTileIcon icon={tile.icon} id={tile.id} variant="plain" />
          </div>
          <div className="flex-1 min-w-0">
            <h4 className="text-sm font-semibold text-slate-900 dark:text-white">{tile.name}</h4>
            <p className="text-xs text-slate-500 mt-0.5 line-clamp-2">{tile.detail}</p>
          </div>
          <button
            type="button"
            className="p-1.5 rounded-lg text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800"
            title={t('relay.marketplace.openSettings')}
            onClick={onOpenSettings}
          >
            <Settings className="w-4 h-4" />
          </button>
          <button type="button" className="p-1.5 text-slate-400 hover:text-slate-600" onClick={onClose}>
            <X className="w-4 h-4" />
          </button>
        </div>

        <div className="p-4 space-y-4">
          <div className="flex flex-wrap gap-2 text-[10px] uppercase tracking-wide">
            {tile.enabled && (
              <span className="px-2 py-0.5 rounded-full bg-emerald-100 dark:bg-emerald-950 text-emerald-800 dark:text-emerald-300">
                {t('relay.marketplace.badgeOn')}
              </span>
            )}
            {tile.connected && !tile.enabled && (
              <span className="px-2 py-0.5 rounded-full bg-amber-100 dark:bg-amber-950 text-amber-800 dark:text-amber-300">
                {t('relay.marketplace.badgeConfigured')}
              </span>
            )}
          </div>

          {actions.length > 0 && (
            <div className="space-y-2">
              {actions.some((a) => a.id === 'poll_now') && (
                <p className="text-xs text-slate-500">{t('relay.marketplace.pollHelp')}</p>
              )}
              {isAdmin ? (
                <div className="flex flex-wrap gap-2">
                  {actions.map((a) => (
                    <Button
                      key={a.id}
                      size="sm"
                      disabled={Boolean(busyAction) || !tile.enabled}
                      onClick={() => void runAction(a.id)}
                    >
                      {busyAction === a.id ? t('relay.marketplace.actionRunning') : a.label}
                    </Button>
                  ))}
                </div>
              ) : (
                <p className="text-xs text-amber-700 dark:text-amber-300">{t('relay.marketplace.adminOnly')}</p>
              )}
              {actionMsg && (
                <p className={`text-xs ${actionError ? 'text-rose-600' : 'text-emerald-700 dark:text-emerald-400'}`}>
                  {actionMsg}
                </p>
              )}
            </div>
          )}

          {hasWebhook && webhookUrl && (
            <div className="space-y-1">
              <p className="text-[11px] uppercase tracking-wide text-slate-500">
                {t('relay.marketplace.webhookUrl')}
              </p>
              <code className="block text-[11px] break-all rounded-lg bg-slate-100 dark:bg-slate-900 p-2 text-slate-700 dark:text-slate-300">
                POST {webhookUrl}
              </code>
            </div>
          )}

          {actions.length === 0 && !hasWebhook && (
            <p className="text-sm text-slate-500">{t('relay.marketplace.noActions')}</p>
          )}

          <div className="space-y-2 pt-2 border-t border-slate-200 dark:border-slate-700">
            <p className="text-[11px] uppercase tracking-wide text-slate-500">{t('relay.activity.title')}</p>
            {activity.loading && <p className="text-xs text-slate-500">{t('relay.activity.loading')}</p>}
            {activity.error && <p className="text-xs text-rose-600">{t('relay.activity.loadError')}</p>}
            {!activity.loading && !activity.error && activity.data && (
              <ConnectorActivityTable entries={activity.data.entries} compact />
            )}
          </div>
        </div>

        <div className="p-4 border-t border-slate-200 dark:border-slate-700 flex justify-end">
          <Button size="sm" variant="ghost" onClick={onClose}>{t('relay.marketplace.cancel')}</Button>
        </div>
      </Card>
    </div>
  )
}
