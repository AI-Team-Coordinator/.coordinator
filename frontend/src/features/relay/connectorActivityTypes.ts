export type ConnectorActivityEntry = {
  id: number
  connector_id: string
  kind: string
  level: string
  summary: string
  detail?: Record<string, unknown>
  created_at: string
}

export type ConnectorActivityPage = {
  entries: ConnectorActivityEntry[]
  total: number
  limit: number
  offset: number
}
