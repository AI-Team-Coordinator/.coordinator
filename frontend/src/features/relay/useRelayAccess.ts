import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../../shared/api/client'
import i18n from '../../shared/i18n/config'
import { useLoadingStatus } from '../../shared/loading/LoadingStatusContext'

export type RelayAccess = {
  connected: boolean
  url?: string
  scope?: string
}

export const DEFAULT_RELAY_LOCAL = 'http://127.0.0.1:4322'

export function useRelayAccess() {
  const { t } = useTranslation()
  const { track } = useLoadingStatus()
  const [url, setUrl] = useState(DEFAULT_RELAY_LOCAL)
  const [token, setToken] = useState('')
  const [access, setAccess] = useState<RelayAccess | null>(null)
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')

  const refresh = useCallback(async () => {
    setError('')
    const r = await fetch(api.relayAccess)
    const data = r.ok ? ((await r.json()) as RelayAccess) : null
    const next = data ?? { connected: false }
    setAccess(next)
    if (next.connected && next.url) {
      setUrl(next.url)
    }
    return next
  }, [])

  useEffect(() => {
    void track('relay-access', i18n.t('loading.relay'), async () => {
      setLoading(true)
      try {
        await refresh()
      } catch {
        setError(t('relay.loadError'))
      } finally {
        setLoading(false)
      }
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const connect = async () => {
    setBusy(true)
    setError('')
    try {
      await track('relay-connect', t('loading.relayConnect'), async () => {
        const res = await fetch(api.relayConnect, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ url: url.trim(), token: token.trim() }),
          signal: AbortSignal.timeout(20_000),
        })
        if (!res.ok) {
          const { readError } = await import('../../shared/api/http')
          throw new Error(await readError(res))
        }
        const data = (await res.json()) as RelayAccess
        setAccess(data)
      })
    } catch (e) {
      if (e instanceof Error && e.name === 'TimeoutError') {
        setError(t('relay.connectTimeout'))
      } else {
        setError(e instanceof Error ? e.message : t('relay.connectError'))
      }
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
    } catch {
      setError(t('relay.disconnectError'))
    } finally {
      setBusy(false)
    }
  }

  const connected = Boolean(access?.connected)

  return {
    url,
    setUrl,
    token,
    setToken,
    access,
    loading,
    busy,
    error,
    setError,
    connected,
    connect,
    disconnect,
    refresh,
  }
}
