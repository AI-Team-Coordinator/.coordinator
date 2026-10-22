import type { MarketplaceTile } from './RelayMarketplace'

export type RelayConnectorConfig = {
  connector_id?: string
  enabled?: boolean
  connected?: boolean
  non_secret?: Record<string, unknown>
  secrets?: Record<string, string>
}

export function nonSecretDisplayValue(v: unknown): string {
  if (v === null || v === undefined) return ''
  if (typeof v === 'boolean') return v ? 'true' : 'false'
  if (typeof v === 'number') return String(v)
  return String(v)
}

export function applyRelayConfigToForm(tile: MarketplaceTile, cfg: RelayConnectorConfig) {
  const next: Record<string, string> = {}
  const savedSecrets: string[] = []
  const schema = tile.config_schema ?? []
  for (const field of schema) {
    if (field.type === 'secret') {
      if (cfg.secrets?.[field.id]) {
        savedSecrets.push(field.id)
      }
      next[field.id] = ''
    } else {
      const v = cfg.non_secret?.[field.id]
      if (v !== undefined) {
        next[field.id] = nonSecretDisplayValue(v)
      } else {
        next[field.id] = ''
      }
    }
  }
  return { next, savedSecrets }
}

export function buildConnectorSavePayload(
  tile: MarketplaceTile,
  form: Record<string, string>,
  loadedNonSecret: Record<string, unknown>,
  enable: boolean,
) {
  const filteredSecrets: Record<string, string> = {}
  const nonSecret: Record<string, string> = {}
  const schema = tile.config_schema ?? []
  for (const field of schema) {
    const raw = (form[field.id] ?? '').trim()
    if (field.type === 'secret') {
      if (raw) filteredSecrets[field.id] = raw
      continue
    }
    if (raw) {
      nonSecret[field.id] = raw
    } else {
      const prev = loadedNonSecret[field.id]
      if (prev !== undefined && prev !== null) {
        nonSecret[field.id] = nonSecretDisplayValue(prev)
      }
    }
  }
  const payload: {
    secrets?: Record<string, string>
    non_secret: Record<string, string>
    enable: boolean
  } = {
    non_secret: nonSecret,
    enable,
  }
  if (Object.keys(filteredSecrets).length > 0) {
    payload.secrets = filteredSecrets
  }
  return payload
}
