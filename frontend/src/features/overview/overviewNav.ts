export type OverviewSubview = 'main' | 'pulse' | 'tasks' | 'timeline' | 'relay-events' | 'relay-activity'

export function parseOverviewSubview(hash: string): OverviewSubview {
  const value = hash.replace(/^#/, '').trim().toLowerCase()
  if (value === 'overview/pulse') return 'pulse'
  if (value === 'overview/tasks') return 'tasks'
  if (value === 'overview/timeline') return 'timeline'
  if (value === 'overview/relay-events') return 'relay-events'
  if (value === 'overview/relay-activity') return 'relay-activity'
  return 'main'
}

export function overviewMainHash(): string {
  return 'overview'
}

export function overviewPulseHash(): string {
  return 'overview/pulse'
}

export function overviewTasksHash(): string {
  return 'overview/tasks'
}

export function overviewTimelineHash(): string {
  return 'overview/timeline'
}

export function overviewRelayEventsHash(): string {
  return 'overview/relay-events'
}

export function overviewRelayActivityHash(): string {
  return 'overview/relay-activity'
}
