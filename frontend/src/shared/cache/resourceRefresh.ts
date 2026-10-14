/**
 * Client-side refresh policy for Coordinator API reads.
 * SSE still pushes live pulse fields; this map limits polling cost.
 *
 * | Resource   | TTL    | Notes |
 * |------------|--------|-------|
 * | project    | 1h     | Name, services tree — settings bus |
 * | sync       | 30s    | coordinator-state sync line |
 * | pulse      | 15s    | Slots + git (heavy); SSE fills gaps |
 * | workItems  | 10s    | work_tasks/*.json + light slots |
 * | stats      | 60s    | Aggregates / billing windows |
 * | events     | 30s    | Activity log (filter changes bypass cache key) |
 * | tasks      | 60s    | Stats task table (filter changes bypass cache key) |
 * | health     | 5s     | API boot id |
 *
 * Hard browser reload (navigation type `reload`) forces a fresh fetch once per resource on startup.
 */

export type CacheResource =
  | 'project'
  | 'sync'
  | 'pulse'
  | 'stats'
  | 'events'
  | 'tasks'
  | 'workItems'
  | 'health'

export const REFRESH_POLICY: Record<CacheResource, { ttlMs: number; description: string }> = {
  project: { ttlMs: 60 * 60 * 1000, description: 'Project profile and service map' },
  sync: { ttlMs: 30_000, description: 'Settings / coordinator-state sync status' },
  pulse: { ttlMs: 15_000, description: 'Team pulse, conflicts, stray repos' },
  workItems: { ttlMs: 10_000, description: 'Work backlog files + slot linkage' },
  stats: { ttlMs: 60_000, description: 'Stats envelope and billing periods' },
  events: { ttlMs: 30_000, description: 'Activity timeline page' },
  tasks: { ttlMs: 60_000, description: 'Historical tasks table' },
  health: { ttlMs: 5_000, description: 'API health / started_at' },
}

type CacheEntry = { data: unknown; fetchedAt: number }

const store = new Map<string, CacheEntry>()

let forceOnceAfterReload = isHardReload()

export function isHardReload(): boolean {
  if (typeof performance === 'undefined') return false
  const nav = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming | undefined
  return nav?.type === 'reload'
}

/** True once after a hard reload; cleared on first consume. */
export function consumeForceRefresh(): boolean {
  if (!forceOnceAfterReload) return false
  forceOnceAfterReload = false
  return true
}

export function invalidateAllResources(): void {
  store.clear()
}

export function invalidateResource(key: CacheResource, urlKey = ''): void {
  store.delete(cacheKey(key, urlKey))
}

function cacheKey(key: CacheResource, urlKey: string): string {
  return urlKey ? `${key}:${urlKey}` : key
}

export function getCached<T>(key: CacheResource, urlKey = ''): T | undefined {
  const entry = store.get(cacheKey(key, urlKey))
  return entry?.data as T | undefined
}

export function isFresh(key: CacheResource, urlKey = ''): boolean {
  const entry = store.get(cacheKey(key, urlKey))
  if (!entry) return false
  return Date.now() - entry.fetchedAt < REFRESH_POLICY[key].ttlMs
}

export function setCached(key: CacheResource, data: unknown, urlKey = ''): void {
  store.set(cacheKey(key, urlKey), { data, fetchedAt: Date.now() })
}

export async function fetchJSONCached<T>(
  key: CacheResource,
  url: string,
  opts?: { force?: boolean; urlKey?: string }
): Promise<T | null> {
  const urlKey = opts?.urlKey ?? ''
  const force = opts?.force ?? false
  if (!force && isFresh(key, urlKey)) {
    return getCached<T>(key, urlKey) ?? null
  }
  const res = await fetch(url)
  if (!res.ok) {
    return getCached<T>(key, urlKey) ?? null
  }
  const data = (await res.json()) as T
  setCached(key, data, urlKey)
  return data
}

/** Poll tick: returns true if this resource is due for refresh. */
export function shouldPoll(key: CacheResource, urlKey = ''): boolean {
  return !isFresh(key, urlKey)
}
