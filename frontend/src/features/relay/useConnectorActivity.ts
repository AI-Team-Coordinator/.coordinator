import { useCallback, useEffect, useState } from 'react'
import { api } from '../../shared/api/client'
import type { ConnectorActivityPage } from './connectorActivityTypes'

export function useConnectorActivity(limit: number, offset: number, connectorId: string) {
  const [data, setData] = useState<ConnectorActivityPage | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = useCallback(() => {
    setLoading(true)
    setError('')
    const params = new URLSearchParams()
    params.set('limit', String(limit))
    params.set('offset', String(offset))
    if (connectorId.trim()) params.set('connector_id', connectorId.trim())
    fetch(`${api.relayConnectorActivity}?${params}`)
      .then((r) => (r.ok ? r.json() : Promise.reject(new Error('load'))))
      .then((body: ConnectorActivityPage) => setData(body))
      .catch(() => setError('load'))
      .finally(() => setLoading(false))
  }, [limit, offset, connectorId])

  useEffect(() => {
    load()
  }, [load])

  return { data, loading, error, reload: load }
}
