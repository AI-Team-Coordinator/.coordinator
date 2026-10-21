export type RelaySubview = 'main' | 'settings' | 'events' | 'activity'

export function parseRelaySubview(hash: string): RelaySubview {
  const value = hash.replace(/^#/, '').trim().toLowerCase()
  if (value === 'relay/settings') return 'settings'
  if (value === 'relay/events') return 'events'
  if (value === 'relay/activity') return 'activity'
  return 'main'
}

export function relaySettingsHash(): string {
  return 'relay/settings'
}

export function relayMainHash(): string {
  return 'relay'
}

export function relayEventsHash(): string {
  return 'relay/events'
}

export function relayActivityHash(): string {
  return 'relay/activity'
}
