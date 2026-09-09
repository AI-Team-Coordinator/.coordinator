import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Header } from '../features/header/Header'
import { ConflictRadar } from '../features/radar/ConflictRadar'
import { MetricsGrid } from '../features/metrics/MetricsGrid'
import { TeamPulseSection } from '../features/pulse/TeamPulseSection'
import { ActivityTimeline, type EventsFilter } from '../features/timeline/ActivityTimeline'
import { TaskTable, type TasksFilter } from '../features/stats/TaskTable'
import { SpendBreakdown } from '../features/stats/SpendBreakdown'
import { ServiceMap } from '../features/project/ServiceMap'
import { TeamRosterSection } from '../features/team/TeamRosterSection'
import { SectionNav } from '../features/nav/SectionNav'
import { parseSection, type SectionId } from '../features/nav/sections'
import type { Conflict, EventItem, MemberState, ProjectProfile, Stats, StrayRepo, SyncStatus, TaskItem } from '../shared/types/api'
import { api } from '../shared/api/client'

export function App() {
  const { t } = useTranslation()
  const [section, setSection] = useState<SectionId>(() => parseSection(window.location.hash))
  const [members, setMembers] = useState<MemberState[]>([])
  const [stray, setStray] = useState<StrayRepo[]>([])
  const [conflicts, setConflicts] = useState<Conflict[]>([])
  const [stats, setStats] = useState<Stats | null>(null)
  const [events, setEvents] = useState<EventItem[]>([])
  const [eventsTotal, setEventsTotal] = useState(0)
  const [eventsFilter, setEventsFilter] = useState<EventsFilter>({ alias: '', event: '', service: '' })
  const eventsFilterRef = useRef(eventsFilter)
  eventsFilterRef.current = eventsFilter
  const [tasks, setTasks] = useState<TaskItem[]>([])
  const [tasksTotal, setTasksTotal] = useState(0)
  const [tasksFilter, setTasksFilter] = useState<TasksFilter>({ status: '', alias: '', kind: '' })
  const tasksFilterRef = useRef(tasksFilter)
  tasksFilterRef.current = tasksFilter
  const [profile, setProfile] = useState<ProjectProfile | null>(null)
  const [syncStatus, setSyncStatus] = useState<SyncStatus | null>(null)
  const [connected, setConnected] = useState(false)
  const [loading, setLoading] = useState(true)
  const [formDirty, setFormDirty] = useState(false)

  const serviceNames = useMemo(() => {
    const map: Record<string, string> = {}
    for (const svc of profile?.services || []) {
      map[svc.id] = svc.name
    }
    return map
  }, [profile])

  const currentUser = useMemo(() => {
    const alias = syncStatus?.alias
    const member = members.find((m) => m.alias === alias)
    return {
      alias,
      name: member?.name,
      role: member?.role,
    }
  }, [syncStatus?.alias, members])

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

  const loadAll = async () => {
    setLoading(true)
    try {
      const apply = async (input: Promise<Response>, onOk: (data: unknown) => void) => {
        const res = await input
        if (res.ok) {
          onOk(await res.json())
        }
      }

      await Promise.all([
        apply(fetch(api.project), (data) => setProfile(data as ProjectProfile)),
        apply(fetch(api.sync), (data) => setSyncStatus(data as SyncStatus)),
        apply(fetch(api.pulse), (data) => {
          const pulse = data as { members?: MemberState[]; conflicts?: Conflict[]; stray?: StrayRepo[] }
          setMembers(pulse.members || [])
          setConflicts(pulse.conflicts || [])
          setStray(pulse.stray || [])
        }),
        apply(fetch(api.stats), (data) => {
          const envelope = data as { stats?: Stats }
          if (envelope.stats) setStats(envelope.stats)
        }),
        apply(fetch(eventsURL(eventsFilter)), (data) => {
          applyEventsPayload(data as { events?: EventItem[]; total?: number }, false)
        }),
        apply(fetch(tasksURL(tasksFilter)), (data) => {
          applyTasksPayload(data as { tasks?: TaskItem[]; total?: number }, false)
        }),
      ])
      setConnected(true)
    } catch (err) {
      console.error('Failed to load dashboard data:', err)
      setConnected(false)
    } finally {
      setLoading(false)
    }
  }

  const onEventsFilter = (next: EventsFilter) => {
    setEventsFilter(next)
    fetch(eventsURL(next, 0))
      .then((r) => r.json())
      .then((d) => applyEventsPayload(d, false))
      .catch(() => {})
  }

  const onEventsLoadMore = () => {
    fetch(eventsURL(eventsFilter, events.length))
      .then((r) => r.json())
      .then((d) => applyEventsPayload(d, true))
      .catch(() => {})
  }

  const onTasksFilter = (next: TasksFilter) => {
    setTasksFilter(next)
    fetch(tasksURL(next, 0))
      .then((r) => r.json())
      .then((d) => applyTasksPayload(d, false))
      .catch(() => {})
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
    const onHash = () => setSection(parseSection(window.location.hash))
    window.addEventListener('hashchange', onHash)
    return () => window.removeEventListener('hashchange', onHash)
  }, [])

  useEffect(() => {
    let bootId: string | null = null
    let reloading = false
    let lastBootCheck = 0

    const checkServerBoot = () => {
      const now = Date.now()
      if (reloading || now - lastBootCheck < 1000) return
      lastBootCheck = now
      fetch(api.health)
        .then((r) => (r.ok ? r.json() : null))
        .then((d: { started_at?: string } | null) => {
          const id = d?.started_at
          if (!id || reloading) return
          if (bootId && bootId !== id) {
            reloading = true
            window.location.reload()
            return
          }
          bootId = id
        })
        .catch(() => {})
    }

    loadAll()
    checkServerBoot()

    const eventSource = new EventSource(api.stream)
    eventSource.onopen = () => setConnected(true)
    eventSource.onmessage = (e) => {
      try {
        const data = JSON.parse(e.data)
        if (data.members) setMembers(data.members)
        if (data.conflicts) setConflicts(data.conflicts)
        if (Array.isArray(data.stray)) setStray(data.stray)
      } catch (err) {
        console.warn('Failed to parse SSE message:', err)
      }
    }
    eventSource.onerror = () => {
      setConnected(false)
      checkServerBoot()
    }

    const statsInterval = setInterval(() => {
      checkServerBoot()
      fetch(api.pulse)
        .then((r) => r.json())
        .then((d) => {
          if (d.members) setMembers(d.members)
          if (d.conflicts) setConflicts(d.conflicts)
          if (Array.isArray(d.stray)) setStray(d.stray)
        })
        .catch(() => {})

      fetch(api.stats)
        .then((r) => r.json())
        .then((d) => d.stats && setStats(d.stats))
        .catch(() => {})

      fetch(eventsURL(eventsFilterRef.current))
        .then((r) => r.json())
        .then((d) => applyEventsPayload(d, false))
        .catch(() => {})

      fetch(tasksURL(tasksFilterRef.current))
        .then((r) => r.json())
        .then((d) => applyTasksPayload(d, false))
        .catch(() => {})

      fetch(api.project)
        .then((r) => r.json())
        .then((d) => d && setProfile(d as ProjectProfile))
        .catch(() => {})

      fetch(api.sync)
        .then((r) => (r.ok ? r.json() : null))
        .then((d) => d && setSyncStatus(d))
        .catch(() => {})
    }, 5000)

    const ticker = setInterval(() => {
      setMembers((prev) =>
        prev.map((m) => ({
          ...m,
          duration_seconds:
            m.status === 'in_progress' ? (m.duration_seconds || 0) + 1 : m.duration_seconds,
          tasks: m.tasks?.map((task) => ({
            ...task,
            duration_seconds: (task.duration_seconds || 0) + 1,
          })),
          research:
            m.research?.status === 'active'
              ? { ...m.research, duration_seconds: (m.research.duration_seconds || 0) + 1 }
              : m.research,
        }))
      )
      setTasks((prev) =>
        prev.map((task) => {
          if (task.status === 'in_progress') {
            return { ...task, duration_seconds: (task.duration_seconds || 0) + 1 }
          }
          return task
        })
      )
    }, 1000)

    return () => {
      eventSource.close()
      clearInterval(statsInterval)
      clearInterval(ticker)
    }
  }, [])

  return (
    <div className="min-h-screen flex flex-col md:flex-row">
      {loading && (
        <div className="fixed top-0 inset-x-0 z-30 h-0.5 overflow-hidden bg-slate-200/80 dark:bg-slate-800/80">
          <div className="coordinator-indeterminate h-full w-1/3 bg-indigo-600" />
        </div>
      )}
      <aside className="hidden md:flex md:flex-col md:sticky md:top-0 md:h-screen md:w-[88px] shrink-0 border-r border-slate-200 dark:border-slate-800 bg-white/80 dark:bg-slate-950/80 backdrop-blur-sm">
        <SectionNav section={section} onChange={changeSection} teamDirty={formDirty} />
      </aside>

      <div className="flex-1 min-w-0 pb-24 md:pb-0">
        <div className="max-w-7xl mx-auto px-4 py-6 md:px-6 md:py-8 space-y-6">
          <Header connected={connected} profile={profile} currentUser={currentUser} />

          {section === 'overview' && (
            <div className="space-y-8">
              <ConflictRadar conflicts={conflicts} />
              <MetricsGrid
                stats={stats}
                onStatusClick={(status) => onTasksFilter({ ...tasksFilter, status })}
                onSpendClick={() => changeSection('stats')}
              />
              <TeamPulseSection members={members} stray={stray} serviceNames={serviceNames} />
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

          <div className={section === 'team' ? 'block' : 'hidden'}>
            <TeamRosterSection
              services={profile?.services || []}
              onChanged={loadAll}
              onDirtyChange={onDirtyChange}
            />
          </div>

          {section === 'project' && <ServiceMap profile={profile} onCreated={loadAll} />}

          {section === 'stats' && (
            <div className="space-y-6">
              <h2 className="text-lg font-semibold text-slate-900 dark:text-white">{t('sections.stats')}</h2>
              <SpendBreakdown stats={stats} />
              <MetricsGrid
                stats={stats}
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
        </div>
      </div>

      <div className="md:hidden fixed bottom-0 inset-x-0 z-20 border-t border-slate-200 dark:border-slate-800 bg-white/95 dark:bg-slate-950/95 backdrop-blur-sm">
        <SectionNav section={section} onChange={changeSection} teamDirty={formDirty} />
      </div>
    </div>
  )
}
