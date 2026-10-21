import { useCallback, useEffect, useState } from 'react'
import { api } from '../../shared/api/client'
import type { ExternalEvent } from '../../shared/types/api'

export type ExternalEventsPage = {
  events: ExternalEvent[]
  total: number
  limit: number
  offset: number
  sources: string[]
}

export function useExternalEvents(limit: number, offset: number, source: string) {
  const [data, setData] = useState<ExternalEventsPage | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = useCallback(() => {
    setLoading(true)
    setError('')
    const params = new URLSearchParams()
    params.set('limit', String(limit))
    params.set('offset', String(offset))
    if (source.trim()) params.set('source', source.trim())
    fetch(`${api.externalEvents}?${params}`)
      .then((r) => (r.ok ? r.json() : Promise.reject(new Error('load'))))
      .then((body: ExternalEventsPage) => setData(body))
      .catch(() => setError('load'))
      .finally(() => setLoading(false))
  }, [limit, offset, source])

  useEffect(() => {
    load()
  }, [load])

  return { data, loading, error, reload: load }
}
