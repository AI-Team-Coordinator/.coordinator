import { Activity, ClipboardList, ListTodo, ScrollText, Users } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ConflictRadar } from '../radar/ConflictRadar'
import { MetricsGrid } from '../metrics/MetricsGrid'
import { NextStepCard } from '../setup/NextStepCard'
import { PersonalConsumptionChart } from '../stats/PersonalConsumptionChart'
import { TaskTable, type TasksFilter } from '../stats/TaskTable'
import { TeamPulseSection } from '../pulse/TeamPulseSection'
import { ActivityTimeline, type EventsFilter } from '../timeline/ActivityTimeline'
import { ExternalSignalsCard } from '../pulse/ExternalSignalsCard'
import { ConnectorActivityTable } from '../relay/ConnectorActivityTable'
import { ExternalEventsTable } from '../relay/ExternalEventsTable'
import { RelayConnectorActivityPage } from '../relay/RelayConnectorActivityPage'
import { RelayExternalEventsPage } from '../relay/RelayExternalEventsPage'
import { useConnectorActivity } from '../relay/useConnectorActivity'
import { useExternalEvents } from '../relay/useExternalEvents'
import { useRelayAccess } from '../relay/useRelayAccess'
import type {
  Conflict,
  EventItem,
  ExternalEvent,
  MemberState,
  ProjectProfile,
  Stats,
  StrayRepo,
  TaskItem,
} from '../../shared/types/api'
import { countPulseCards } from './countPulseCards'
import {
  OverviewPreviewBlock,
  OVERVIEW_EXPANDED_ROWS,
  OVERVIEW_PULSE_COLLAPSED_ROWS,
  OVERVIEW_PULSE_EXPANDED_ROWS,
} from './OverviewPreviewBlock'
import {
  overviewMainHash,
  overviewPulseHash,
  overviewRelayActivityHash,
  overviewRelayEventsHash,
  overviewTasksHash,
  overviewTimelineHash,
  parseOverviewSubview,
} from './overviewNav'
import { OverviewSubpageShell } from './OverviewSubpageShell'

export interface OverviewSectionProps {
  conflicts: Conflict[]
  externalEvents: ExternalEvent[]
  quietBoard: boolean
  stats: Stats | null
  overviewCountsScope: string
  onStatusClick: (status: TasksFilter['status']) => void
  onSpendClick: () => void
  currentUserAlias?: string
  members: MemberState[]
  stray: StrayRepo[]
  serviceNames: Record<string, string>
  tasks: TaskItem[]
  tasksTotal: number
  tasksFilter: TasksFilter
  onTasksFilterChange: (next: TasksFilter) => void
  onTasksLoadMore: () => void
  events: EventItem[]
  eventsTotal: number
  eventsFilter: EventsFilter
  onEventsFilterChange: (next: EventsFilter) => void
  onEventsLoadMore: () => void
  profile: ProjectProfile | null
}

export function OverviewSection(props: OverviewSectionProps) {
  const { t } = useTranslation()
  const [subview, setSubview] = useState(() => parseOverviewSubview(window.location.hash))
  const relay = useRelayAccess()

  useEffect(() => {
    const onHash = () => setSubview(parseOverviewSubview(window.location.hash))
    window.addEventListener('hashchange', onHash)
    return () => window.removeEventListener('hashchange', onHash)
  }, [])

  const [relayEventsFetchLimit, setRelayEventsFetchLimit] = useState(OVERVIEW_EXPANDED_ROWS)
  const [relayActivityFetchLimit, setRelayActivityFetchLimit] = useState(OVERVIEW_EXPANDED_ROWS)
  const bumpRelayEventsLimit = useCallback((visible: number) => {
    setRelayEventsFetchLimit((prev) => Math.max(prev, visible))
  }, [])
  const bumpRelayActivityLimit = useCallback((visible: number) => {
    setRelayActivityFetchLimit((prev) => Math.max(prev, visible))
  }, [])

  const relayEvents = useExternalEvents(relayEventsFetchLimit, 0, '', relay.connected)
  const relayActivity = useConnectorActivity(relayActivityFetchLimit, 0, '', relay.connected)

  const pulseCardTotal = countPulseCards(props.members, props.stray)

  if (subview === 'pulse') {
    return (
      <OverviewSubpageShell title={t('pulse.title')} subtitle={t('pulse.subtitle')}>
        <TeamPulseSection
          members={props.members}
          stray={props.stray}
          serviceNames={props.serviceNames}
        />
      </OverviewSubpageShell>
    )
  }

  if (subview === 'tasks') {
    return (
      <OverviewSubpageShell title={t('tasks.title')} subtitle={t('tasks.subtitle')}>
        <TaskTable
          tasks={props.tasks}
          total={props.tasksTotal}
          members={props.members}
          filter={props.tasksFilter}
          onFilterChange={props.onTasksFilterChange}
          onLoadMore={props.onTasksLoadMore}
        />
      </OverviewSubpageShell>
    )
  }

  if (subview === 'timeline') {
    return (
      <OverviewSubpageShell title={t('timeline.title')} subtitle={t('timeline.subtitle')}>
        <ActivityTimeline
          events={props.events}
          total={props.eventsTotal}
          members={props.members}
          services={props.profile?.services || []}
          filter={props.eventsFilter}
          onFilterChange={props.onEventsFilterChange}
          onLoadMore={props.onEventsLoadMore}
        />
      </OverviewSubpageShell>
    )
  }

  if (subview === 'relay-events') {
    return <RelayExternalEventsPage backHash={overviewMainHash()} />
  }

  if (subview === 'relay-activity') {
    return <RelayConnectorActivityPage backHash={overviewMainHash()} />
  }

  const showTasks = !props.quietBoard || props.tasksTotal > 0
  const showTimeline = !props.quietBoard || props.eventsTotal > 0

  return (
    <div className="space-y-8">
      <ConflictRadar conflicts={props.conflicts} />

      {!relay.connected && props.externalEvents.length > 0 && (
        <ExternalSignalsCard events={props.externalEvents} />
      )}

      {props.quietBoard ? (
        <NextStepCard />
      ) : (
        <MetricsGrid
          stats={props.stats}
          countsScopeLabel={props.overviewCountsScope}
          onStatusClick={(status) => props.onStatusClick(status)}
          onSpendClick={props.onSpendClick}
        />
      )}

      <PersonalConsumptionChart stats={props.stats} alias={props.currentUserAlias} />

      <OverviewPreviewBlock
        layout="plain"
        collapsedRows={OVERVIEW_PULSE_COLLAPSED_ROWS}
        expandedRows={OVERVIEW_PULSE_EXPANDED_ROWS}
        title={t('pulse.title')}
        icon={<Users className="w-5 h-5 text-indigo-500 shrink-0" />}
        total={pulseCardTotal}
        empty={pulseCardTotal === 0}
        showAllHash={overviewPulseHash()}
      >
        {(visibleCount) => (
          <TeamPulseSection
            compact
            members={props.members}
            stray={props.stray}
            serviceNames={props.serviceNames}
            maxCards={visibleCount}
            hideTitle
          />
        )}
      </OverviewPreviewBlock>

      {showTasks && (
        <OverviewPreviewBlock
          title={t('tasks.title')}
          icon={<ListTodo className="w-4 h-4 text-emerald-600" />}
          total={props.tasksTotal}
          empty={props.tasksTotal === 0}
          showAllHash={overviewTasksHash()}
          loadedCount={props.tasks.length}
          onLoadMore={props.onTasksLoadMore}
        >
          {(visibleCount) => (
            <TaskTable
              preview
              tasks={props.tasks}
              total={props.tasksTotal}
              members={props.members}
              filter={props.tasksFilter}
              onFilterChange={props.onTasksFilterChange}
              onLoadMore={props.onTasksLoadMore}
              maxRows={visibleCount}
              hideLoadMore
            />
          )}
        </OverviewPreviewBlock>
      )}

      {relay.connected && (
        <>
          <OverviewPreviewBlock
            title={t('relay.events.title')}
            icon={<ScrollText className="w-4 h-4 text-violet-600" />}
            total={relayEvents.data?.total}
            empty={(relayEvents.data?.total ?? 0) === 0 && !relayEvents.loading}
            showAllHash={overviewRelayEventsHash()}
            loadedCount={relayEvents.data?.events.length}
            onVisibleRowsChange={bumpRelayEventsLimit}
          >
            {(visibleCount) =>
              relayEvents.loading ? (
                <p className="text-sm text-slate-500">{t('relay.events.loading')}</p>
              ) : relayEvents.error ? (
                <p className="text-xs text-rose-600">{t('relay.events.loadError')}</p>
              ) : (
                <ExternalEventsTable
                  events={relayEvents.data?.events ?? []}
                  onTriageChanged={relayEvents.reload}
                  compact
                  maxRows={visibleCount}
                />
              )
            }
          </OverviewPreviewBlock>

          <OverviewPreviewBlock
            title={t('relay.activity.title')}
            icon={<ClipboardList className="w-4 h-4 text-sky-600" />}
            total={relayActivity.data?.total}
            empty={(relayActivity.data?.total ?? 0) === 0 && !relayActivity.loading}
            showAllHash={overviewRelayActivityHash()}
            loadedCount={relayActivity.data?.entries.length}
            onVisibleRowsChange={bumpRelayActivityLimit}
          >
            {(visibleCount) =>
              relayActivity.loading ? (
                <p className="text-sm text-slate-500">{t('relay.activity.loading')}</p>
              ) : relayActivity.error ? (
                <p className="text-xs text-rose-600">{t('relay.activity.loadError')}</p>
              ) : (
                <ConnectorActivityTable
                  entries={(relayActivity.data?.entries ?? []).slice(0, visibleCount)}
                  compact
                />
              )
            }
          </OverviewPreviewBlock>
        </>
      )}

      {showTimeline && (
        <OverviewPreviewBlock
          title={t('timeline.title')}
          icon={<Activity className="w-4 h-4 text-slate-500" />}
          total={props.eventsTotal}
          empty={props.eventsTotal === 0}
          showAllHash={overviewTimelineHash()}
          loadedCount={props.events.length}
          onLoadMore={props.onEventsLoadMore}
        >
          {(visibleCount) => (
            <ActivityTimeline
              preview
              events={props.events}
              total={props.eventsTotal}
              members={props.members}
              services={props.profile?.services || []}
              filter={props.eventsFilter}
              onFilterChange={props.onEventsFilterChange}
              onLoadMore={props.onEventsLoadMore}
              maxRows={visibleCount}
              hideLoadMore
            />
          )}
        </OverviewPreviewBlock>
      )}
    </div>
  )
}
