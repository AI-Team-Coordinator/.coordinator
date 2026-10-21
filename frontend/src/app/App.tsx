import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Header } from '../features/header/Header'
import { ConflictRadar } from '../features/radar/ConflictRadar'
import { MetricsGrid } from '../features/metrics/MetricsGrid'
import { TeamPulseSection } from '../features/pulse/TeamPulseSection'
import { ExternalSignalsCard } from '../features/pulse/ExternalSignalsCard'
import { ActivityTimeline, type EventsFilter } from '../features/timeline/ActivityTimeline'
import { TaskTable, type TasksFilter } from '../features/stats/TaskTable'
import { SpendBreakdown } from '../features/stats/SpendBreakdown'
import { HourlySpendChart } from '../features/stats/HourlySpendChart'
import { PersonalConsumptionChart } from '../features/stats/PersonalConsumptionChart'
import { StatsPeriodToolbar } from '../features/stats/StatsPeriodToolbar'
import {
  billingPeriodMarkersOnDays,
  periodRangeForPreset,
  statsURL,
  type StatsPeriodPreset,
} from '../features/stats/statsPeriod'
import { cursorSpendFromStats } from '../features/metrics/cursorSpend'
import { formatDate } from '../shared/lib/formatters'
import { ServiceMap } from '../features/project/ServiceMap'
import { TeamRosterSection } from '../features/team/TeamRosterSection'
import { OrgTopologySection } from '../features/org/OrgTopologySection'
import { RelaySection } from '../features/relay/RelaySection'
import { WorkItemsSection } from '../features/work/WorkItemsSection'
import { SettingsSection } from '../features/settings/SettingsSection'
import { SectionNav } from '../features/nav/SectionNav'
import { parseSection, type SectionId } from '../features/nav/sections'
import { SetupScreen } from '../features/setup/SetupScreen'
import { NextStepCard } from '../features/setup/NextStepCard'
import { CoordinatorLogo } from '../shared/ui/Logo'
import { isAlinaAssistProject } from '../shared/lib/alinaAssist'
import type {
  BillingPeriod,
  Conflict,
  EventItem,
  MemberState,
  ExternalEvent,
  ProjectProfile,
  SetupState,
  Stats,
  StatsEnvelope,
  StrayRepo,
  SyncStatus,
  TaskItem,
  UserSettings,
  WorkItem,
} from '../shared/types/api'
import { api } from '../shared/api/client'
import {
  consumeForceRefresh,
  fetchJSONCached,
  getCached,
  isFresh,
  shouldPoll,
  type CacheResource,
} from '../shared/cache/resourceRefresh'
import i18n from '../shared/i18n/config'
import { useLoadingStatus } from '../shared/loading/LoadingStatusContext'
import { LoadingStatusOverlay } from '../shared/loading/LoadingStatusOverlay'

export function App() {
  const { t } = useTranslation()
  const [section, setSection] = useState<SectionId>(() => parseSection(window.location.hash))
  const [visitedSections, setVisitedSections] = useState<Set<SectionId>>(() =>
    new Set([parseSection(window.location.hash)])
  )
  const [members, setMembers] = useState<MemberState[]>([])
  const [stray, setStray] = useState<StrayRepo[]>([])
  const [externalEvents, setExternalEvents] = useState<ExternalEvent[]>([])
  const [conflicts, setConflicts] = useState<Conflict[]>([])
  const [stats, setStats] = useState<Stats | null>(null)
  const [events, setEvents] = useState<EventItem[]>([])
  const [eventsTotal, setEventsTotal] = useState(0)
  const [eventsFilter, setEventsFilter] = useState<EventsFilter>({ alias: '', event: '', service: '' })
  const eventsFilterRef = useRef(eventsFilter)
  eventsFilterRef.current = eventsFilter
  const [tasks, setTasks] = useState<TaskItem[]>([])
  const [tasksTotal, setTasksTotal] = useState(0)
  const [workItems, setWorkItems] = useState<WorkItem[]>([])
  const [workItemsReady, setWorkItemsReady] = useState(false)
  const [tasksFilter, setTasksFilter] = useState<TasksFilter>({ status: '', alias: '', kind: '' })
  const tasksFilterRef = useRef(tasksFilter)
  tasksFilterRef.current = tasksFilter
  const [profile, setProfile] = useState<ProjectProfile | null>(null)
  const [syncStatus, setSyncStatus] = useState<SyncStatus | null>(null)
  const [connected, setConnected] = useState(false)
  const { track } = useLoadingStatus()
  const [formDirty, setFormDirty] = useState(false)
  const [setupNeeded, setSetupNeeded] = useState<boolean | null>(null)
  const [setupPrefill, setSetupPrefill] = useState<SetupState | null>(null)
  const [setupReplay, setSetupReplay] = useState(false)
  const [statsPeriodPreset, setStatsPeriodPreset] = useState<StatsPeriodPreset>('billing')
  const [statsCustomFrom, setStatsCustomFrom] = useState('')
  const [statsCustomTo, setStatsCustomTo] = useState('')
  const [statsSectionStats, setStatsSectionStats] = useState<Stats | null>(null)
  const [billingPeriods, setBillingPeriods] = useState<BillingPeriod[]>([])
  const [statsPeriodFrom, setStatsPeriodFrom] = useState(0)
  const [userSettings, setUserSettings] = useState<UserSettings | null>(null)
  const [userSettingsLoading, setUserSettingsLoading] = useState(true)
  const [userSettingsLoadError, setUserSettingsLoadError] = useState(false)
  const [avatarBust, setAvatarBust] = useState(0)

  const serviceNames = useMemo(() => {
    const map: Record<string, string> = {}
    for (const svc of profile?.services || []) {
      if (svc.id) map[svc.id] = svc.name
      if (svc.name) map[svc.name] = svc.name
      if (svc.repo) map[svc.repo] = svc.name
    }
    return map
  }, [profile])

  const overviewCountsScope = useMemo(() => {
    if (statsPeriodFrom > 0) {
      return t('metrics.countsSinceCycle', {
        date: formatDate(statsPeriodFrom, i18n.language, true),
      })
    }
    const spend = cursorSpendFromStats(stats)
    if (spend.billingCycleStart) {
      return t('metrics.countsSinceCycle', {
        date: formatDate(spend.billingCycleStart, i18n.language, true),
      })
    }
    return t('metrics.countsAllTime')
  }, [stats, statsPeriodFrom, t, i18n.language])

  const statsView = statsSectionStats ?? stats
  const billingDayMarkers = useMemo(
    () => billingPeriodMarkersOnDays(statsView?.days ?? [], billingPeriods),
    [statsView?.days, billingPeriods]
  )

  const currentUser = useMemo(() => {
    const alias = syncStatus?.alias
    const member = members.find((m) => m.alias === alias)
    return {
      alias,
      name: userSettings?.name || member?.name,
      role: member?.role || userSettings?.role,
      access: member?.access || userSettings?.access,
      avatarUrl: userSettings?.avatar_url,
      avatarBust,
    }
  }, [syncStatus?.alias, members, userSettings, avatarBust])

  const onUserSettingsSaved = useCallback((next: UserSettings) => {
    setUserSettings(next)
    setAvatarBust(Date.now())
    const lang = next.language?.startsWith('ru') ? 'ru' : 'en'
    void i18n.changeLanguage(lang)
  }, [])

  const changeSection = (next: SectionId) => {
    setSection(next)
    window.location.hash = next
  }

  const eventsURL = (filter: EventsFilter, offset = 0) => {
    const params = new URLSearchParams()
    params.set('limit', '30')
    params.set('offset', String(offset))
    if (filter.alias) params.set('alias', filter.alias)
    if (filter.event) params.set('event', filter.event)
    if (filter.service) params.set('service', filter.service)
    return `${api.events}?${params.toString()}`
  }

  const tasksURL = (filter: TasksFilter, offset = 0) => {
    const params = new URLSearchParams()
    params.set('limit', '50')
    params.set('offset', String(offset))
    if (filter.status) params.set('status', filter.status)
    if (filter.alias) params.set('alias', filter.alias)
    if (filter.kind) params.set('kind', filter.kind)
    return `${api.tasks}?${params.toString()}`
  }

  const applyEventsPayload = (data: { events?: EventItem[]; total?: number }, append: boolean) => {
    setEventsTotal(data.total || 0)
    const next = data.events || []
    setEvents((prev) => (append ? [...prev, ...next] : next))
  }

  const applyTasksPayload = (data: { tasks?: TaskItem[]; total?: number }, append: boolean) => {
    setTasksTotal(data.total || 0)
    const next = data.tasks || []
    setTasks((prev) => (append ? [...prev, ...next] : next))
  }

  const loadAll = useCallback(
    async (opts?: { force?: boolean }) => {
      const force = opts?.force ?? consumeForceRefresh()

      const run = async (
        key: CacheResource,
        trackId: string,
        label: string,
        url: string,
        apply: (data: unknown) => void,
        urlKey = ''
      ) => {
        if (!force && isFresh(key, urlKey)) {
          const cached = getCached(key, urlKey)
          if (cached !== undefined) {
            apply(cached)
            return
          }
        }
        await track(trackId, label, async () => {
          const data = await fetchJSONCached<unknown>(key, url, { force, urlKey })
          if (data !== null) apply(data)
        })
      }

      try {
        const eventsUrl = eventsURL(eventsFilter)
        const tasksUrl = tasksURL(tasksFilter)

        await Promise.all([
          run('project', 'load-project', t('loading.project'), api.project, (data) =>
            setProfile(data as ProjectProfile)
          ),
          run('sync', 'load-sync', t('loading.sync'), api.sync, (data) =>
            setSyncStatus(data as SyncStatus)
          ),
          run('pulse', 'load-pulse', t('loading.pulse'), api.pulse, (data) => {
            const pulse = data as {
              members?: MemberState[]
              conflicts?: Conflict[]
              stray?: StrayRepo[]
              external?: { events?: ExternalEvent[] }
            }
            setMembers(pulse.members || [])
            setConflicts(pulse.conflicts || [])
            setStray(pulse.stray || [])
            setExternalEvents(pulse.external?.events || [])
          }),
          run('stats', 'load-stats', t('loading.stats'), api.stats, (data) => {
            const envelope = data as StatsEnvelope
            if (envelope.stats) setStats(envelope.stats)
            if (envelope.billing_periods) setBillingPeriods(envelope.billing_periods)
            setStatsPeriodFrom(envelope.period_from ?? 0)
          }),
          run('events', 'load-events', t('loading.events'), eventsUrl, (data) => {
            applyEventsPayload(data as { events?: EventItem[]; total?: number }, false)
          }, eventsUrl),
          run('tasks', 'load-tasks', t('loading.tasks'), tasksUrl, (data) => {
            applyTasksPayload(data as { tasks?: TaskItem[]; total?: number }, false)
          }, tasksUrl),
          run('workItems', 'load-work-items', t('work.loading'), api.workItems, (data) => {
            const payload = data as { items?: WorkItem[] }
            setWorkItems(payload.items || [])
            setWorkItemsReady(true)
          }),
        ])
        setConnected(true)
      } catch (err) {
        console.error('Failed to load dashboard data:', err)
        setConnected(false)
      }
    },
    [track, t, eventsFilter, tasksFilter]
  )

  const onEventsFilter = (next: EventsFilter) => {
    setEventsFilter(next)
    const url = eventsURL(next, 0)
    void fetchJSONCached<{ events?: EventItem[]; total?: number }>('events', url, {
      force: true,
      urlKey: url,
    }).then((d) => {
      if (d) applyEventsPayload(d, false)
    })
  }

  const onEventsLoadMore = () => {
    fetch(eventsURL(eventsFilter, events.length))
      .then((r) => r.json())
      .then((d) => applyEventsPayload(d, true))
      .catch(() => {})
  }

  const onTasksFilter = (next: TasksFilter) => {
    setTasksFilter(next)
    const url = tasksURL(next, 0)
    void fetchJSONCached<{ tasks?: TaskItem[]; total?: number }>('tasks', url, {
      force: true,
      urlKey: url,
    }).then((d) => {
      if (d) applyTasksPayload(d, false)
    })
  }

  const onTasksLoadMore = () => {
    fetch(tasksURL(tasksFilter, tasks.length))
      .then((r) => r.json())
      .then((d) => applyTasksPayload(d, true))
      .catch(() => {})
  }

  const onDirtyChange = useCallback((dirty: boolean) => {
    setFormDirty(dirty)
  }, [])

  useEffect(() => {
    if (section !== 'stats') return
    const { from, to } = periodRangeForPreset(
      statsPeriodPreset,
      stats?.billing_cycle_start,
      statsCustomFrom,
      statsCustomTo
    )
    void track('stats-period', i18n.t('loading.statsPeriod'), async () => {
      const r = await fetch(statsURL(api.stats, from, to))
      if (!r.ok) return
      const envelope = (await r.json()) as StatsEnvelope
      if (!envelope.stats) return
      setStatsSectionStats(envelope.stats)
      if (envelope.billing_periods) setBillingPeriods(envelope.billing_periods)
      if (statsPeriodPreset === 'billing' && envelope.period_from) {
        setStatsPeriodFrom(envelope.period_from)
      }
    })
  }, [section, statsPeriodPreset, statsCustomFrom, statsCustomTo, stats?.billing_cycle_start, track])

  const replaySetup = useCallback(() => {
    fetch(api.setup)
      .then((r) => (r.ok ? r.json() : setupPrefill))
      .then((data: SetupState) => {
        setSetupPrefill(data)
        setSetupReplay(true)
      })
      .catch(() => setSetupReplay(true))
  }, [setupPrefill])

  const quietBoard = (profile?.services?.length || 0) === 0 && !profile?.github?.org

  useEffect(() => {
    const onHash = () => setSection(parseSection(window.location.hash))
    window.addEventListener('hashchange', onHash)
    return () => window.removeEventListener('hashchange', onHash)
  }, [])

  useEffect(() => {
    setVisitedSections((prev) => {
      if (prev.has(section)) return prev
      const next = new Set(prev)
      next.add(section)
      return next
    })
  }, [section])

  useEffect(() => {
    fetch(api.setup)
      .then((r) => (r.ok ? r.json() : { needed: false }))
      .then((data: SetupState) => {
        setSetupPrefill(data)
        setSetupNeeded(Boolean(data.needed))
        if (data.language) {
          void i18n.changeLanguage(data.language.startsWith('ru') ? 'ru' : 'en')
        }
      })
      .catch(() => setSetupNeeded(false))
  }, [])

  useEffect(() => {
    if (setupNeeded !== false) return
    setUserSettingsLoadError(false)
    void track('user-settings', i18n.t('loading.profile'), async () => {
      setUserSettingsLoading(true)
      try {
        const r = await fetch(api.meSettings)
        if (!r.ok) throw new Error('settings')
        const data = (await r.json()) as UserSettings
        setUserSettings(data)
        if (data.language) {
          const lang = data.language.startsWith('ru') ? 'ru' : 'en'
          if (!i18n.language.startsWith(lang)) {
            void i18n.changeLanguage(lang)
          }
        }
      } catch {
        setUserSettingsLoadError(true)
      } finally {
        setUserSettingsLoading(false)
      }
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps -- once after setup; not on every language change
  }, [setupNeeded, track])

  useEffect(() => {
    if (setupNeeded !== false) return
    let bootId: string | null = null
    let reloading = false
    let lastBootCheck = 0

    const checkServerBoot = () => {
      const now = Date.now()
      if (reloading || now - lastBootCheck < 1000) return
      if (!shouldPoll('health')) return
      lastBootCheck = now
      void fetchJSONCached<{ started_at?: string }>('health', api.health).then((d) => {
        const id = d?.started_at
        if (!id || reloading) return
        if (bootId && bootId !== id) {
          reloading = true
          window.location.reload()
          return
        }
        bootId = id
      })
    }

    const pollStaleResources = () => {
      if (shouldPoll('pulse')) {
        void fetchJSONCached<{
          members?: MemberState[]
          conflicts?: Conflict[]
          stray?: StrayRepo[]
          external?: { events?: ExternalEvent[] }
        }>(
          'pulse',
          api.pulse
        ).then((d) => {
          if (!d) return
          if (d.members) setMembers(d.members)
          if (d.conflicts) setConflicts(d.conflicts)
          if (Array.isArray(d.stray)) setStray(d.stray)
          if (d.external?.events) setExternalEvents(d.external.events)
        })
      }
      if (shouldPoll('stats')) {
        void fetchJSONCached<StatsEnvelope>('stats', api.stats).then((d) => {
          if (d?.stats) setStats(d.stats)
        })
      }
      const evUrl = eventsURL(eventsFilterRef.current)
      if (shouldPoll('events', evUrl)) {
        void fetchJSONCached<{ events?: EventItem[]; total?: number }>('events', evUrl, { urlKey: evUrl }).then(
          (d) => {
            if (d) applyEventsPayload(d, false)
          }
        )
      }
      const tkUrl = tasksURL(tasksFilterRef.current)
      if (shouldPoll('tasks', tkUrl)) {
        void fetchJSONCached<{ tasks?: TaskItem[]; total?: number }>('tasks', tkUrl, { urlKey: tkUrl }).then((d) => {
          if (d) applyTasksPayload(d, false)
        })
      }
      if (shouldPoll('workItems')) {
        void fetchJSONCached<{ items?: WorkItem[] }>('workItems', api.workItems).then((d) => {
          if (!d) return
          setWorkItems(d.items || [])
          setWorkItemsReady(true)
        })
      }
      if (shouldPoll('project')) {
        void fetchJSONCached<ProjectProfile>('project', api.project).then((d) => {
          if (d) setProfile(d)
        })
      }
      if (shouldPoll('sync')) {
        void fetchJSONCached<SyncStatus>('sync', api.sync).then((d) => {
          if (d) setSyncStatus(d)
        })
      }
    }

    void loadAll()

    const eventSource = new EventSource(api.stream)
    eventSource.onopen = () => setConnected(true)
    eventSource.onmessage = (e) => {
      try {
        const data = JSON.parse(e.data)
        if (data.members) setMembers(data.members)
        if (data.conflicts) setConflicts(data.conflicts)
        if (Array.isArray(data.stray)) setStray(data.stray)
        if (data.external?.events) setExternalEvents(data.external.events)
      } catch (err) {
        console.warn('Failed to parse SSE message:', err)
      }
    }
    eventSource.onerror = () => {
      setConnected(false)
      checkServerBoot()
    }

    const pollInterval = setInterval(() => {
      checkServerBoot()
      pollStaleResources()
    }, 3000)

    const ticker = setInterval(() => {
      setMembers((prev) =>
        prev.map((m) => ({
          ...m,
          duration_seconds: m.clock_paused
            ? m.duration_seconds
            : m.status === 'in_progress'
              ? (m.duration_seconds || 0) + 1
              : m.duration_seconds,
          tasks: m.tasks?.map((task) => ({
            ...task,
            duration_seconds:
              task.clock_paused || task.status === 'parked'
                ? task.duration_seconds
                : (task.duration_seconds || 0) + 1,
          })),
          research:
            m.research?.status === 'active' && !m.research.clock_paused
              ? { ...m.research, duration_seconds: (m.research.duration_seconds || 0) + 1 }
              : m.research,
        }))
      )
      setTasks((prev) =>
        prev.map((task) => {
          if (task.status === 'in_progress' && !task.clock_paused) {
            return { ...task, duration_seconds: (task.duration_seconds || 0) + 1 }
          }
          return task
        })
      )
    }, 1000)

    return () => {
      eventSource.close()
      clearInterval(pollInterval)
      clearInterval(ticker)
    }
  }, [setupNeeded, loadAll])

  if (setupNeeded === null) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-50 dark:bg-slate-950">
        <CoordinatorLogo size={48} />
      </div>
    )
  }

  if (setupNeeded && setupPrefill) {
    return (
      <SetupScreen
        initial={setupPrefill}
        onDone={(next) => {
          setSetupPrefill(next)
          setSetupNeeded(false)
        }}
      />
    )
  }

  if (setupReplay && setupPrefill) {
    return (
      <SetupScreen
        initial={setupPrefill}
        replay
        onCancel={() => setSetupReplay(false)}
        onDone={() => setSetupReplay(false)}
      />
    )
  }

  return (
    <div className="min-h-screen flex flex-col md:flex-row">
      <LoadingStatusOverlay />
      <aside className="hidden md:flex md:flex-col md:sticky md:top-0 md:h-screen md:w-[88px] shrink-0 border-r border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950">
        <SectionNav section={section} onChange={changeSection} teamDirty={formDirty} />
      </aside>

      <div className="flex-1 min-w-0 pb-24 md:pb-0">
        <div className="max-w-7xl mx-auto px-4 py-6 md:px-6 md:py-8 space-y-6">
          <Header
            connected={connected}
            profile={profile}
            currentUser={currentUser}
          />

          {section === 'work' && (
            <WorkItemsSection items={workItems} loading={!workItemsReady} />
          )}

          {section === 'overview' && (
            <div className="space-y-8">
              <ConflictRadar conflicts={conflicts} />
              <ExternalSignalsCard events={externalEvents} />
              {quietBoard ? <NextStepCard /> : (
                <MetricsGrid
                  stats={stats}
                  countsScopeLabel={overviewCountsScope}
                  onStatusClick={(status) => onTasksFilter({ ...tasksFilter, status })}
                  onSpendClick={() => changeSection('stats')}
                />
              )}
              <PersonalConsumptionChart stats={stats} alias={currentUser.alias} />
              <TeamPulseSection members={members} stray={stray} serviceNames={serviceNames} />
              {(!quietBoard || tasksTotal > 0) && (
                <TaskTable
                  tasks={tasks}
                  total={tasksTotal}
                  members={members}
                  filter={tasksFilter}
                  onFilterChange={onTasksFilter}
                  onLoadMore={onTasksLoadMore}
                />
              )}
              {(!quietBoard || eventsTotal > 0) && (
                <ActivityTimeline
                  events={events}
                  total={eventsTotal}
                  members={members}
                  services={profile?.services || []}
                  filter={eventsFilter}
                  onFilterChange={onEventsFilter}
                  onLoadMore={onEventsLoadMore}
                />
              )}
            </div>
          )}

          {visitedSections.has('team') && (
            <div className={section === 'team' ? 'block' : 'hidden'}>
              <TeamRosterSection
                services={profile?.services || []}
                onChanged={() => void loadAll({ force: true })}
                onDirtyChange={onDirtyChange}
              />
            </div>
          )}

          {visitedSections.has('organization') && (
            <div className={section === 'organization' ? 'block' : 'hidden'}>
              <OrgTopologySection
                onChanged={() => void loadAll({ force: true })}
                onDirtyChange={onDirtyChange}
              />
            </div>
          )}

          {section === 'relay' && (
            <RelaySection
              isAdmin={currentUser.access === 'admin'}
              currentAlias={currentUser.alias}
            />
          )}

          {visitedSections.has('settings') && (
            <div className={section === 'settings' ? 'block' : 'hidden'}>
              <SettingsSection
                settings={userSettings}
                loading={userSettingsLoading}
                loadError={userSettingsLoadError}
                avatarBust={avatarBust}
                onSaved={onUserSettingsSaved}
              />
            </div>
          )}

          {section === 'project' && <ServiceMap profile={profile} onCreated={() => void loadAll({ force: true })} />}

          {section === 'stats' && (
            <div className="space-y-6">
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <h2 className="text-lg font-semibold text-slate-900 dark:text-white">{t('sections.stats')}</h2>
                <StatsPeriodToolbar
                  preset={statsPeriodPreset}
                  onPresetChange={setStatsPeriodPreset}
                  customFrom={statsCustomFrom}
                  customTo={statsCustomTo}
                  onCustomFromChange={setStatsCustomFrom}
                  onCustomToChange={setStatsCustomTo}
                />
              </div>
              <SpendBreakdown stats={statsView} />
              <HourlySpendChart
                grain="day"
                rows={statsView?.days ?? []}
                markerUnix={billingDayMarkers}
                subtitle={billingDayMarkers.length > 0 ? t('spend.billingPeriodMarkersHint') : undefined}
              />
              <MetricsGrid
                stats={statsView}
                countsScopeLabel={
                  statsPeriodPreset === 'billing' && statsView?.billing_cycle_start
                    ? t('metrics.countsSinceCycle', {
                        date: formatDate(statsView.billing_cycle_start, i18n.language, true),
                      })
                    : undefined
                }
                onStatusClick={(status) => onTasksFilter({ ...tasksFilter, status })}
                hideSpend
              />
              <TaskTable
                tasks={tasks}
                total={tasksTotal}
                members={members}
                filter={tasksFilter}
                onFilterChange={onTasksFilter}
                onLoadMore={onTasksLoadMore}
              />
              <ActivityTimeline
                events={events}
                total={eventsTotal}
                members={members}
                services={profile?.services || []}
                filter={eventsFilter}
                onFilterChange={onEventsFilter}
                onLoadMore={onEventsLoadMore}
              />
            </div>
          )}

          {section === 'overview' && isAlinaAssistProject(profile) && setupPrefill && (
            <div className="pt-10 mt-4 border-t border-slate-200 dark:border-slate-800">
              <button
                type="button"
                className="text-xs text-slate-400 hover:text-slate-600 dark:hover:text-slate-300"
                onClick={replaySetup}
              >
                {t('setup.testLaunch')}
              </button>
            </div>
          )}
        </div>
      </div>

      <div className="md:hidden fixed bottom-0 inset-x-0 z-20 border-t border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950">
        <SectionNav section={section} onChange={changeSection} teamDirty={formDirty} />
      </div>
    </div>
  )
}
