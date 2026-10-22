import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Card } from '../../shared/ui/Card'
import { Select } from '../../shared/ui/Select'
import { Button } from '../../shared/ui/Button'
import { TimelineItem } from './TimelineItem'
import type { EventItem, MemberState, ServiceNode } from '../../shared/types/api'

export interface EventsFilter {
  alias: string
  event: string
  service: string
}

interface ActivityTimelineProps {
  events: EventItem[]
  total: number
  members: MemberState[]
  services: ServiceNode[]
  filter: EventsFilter
  onFilterChange: (next: EventsFilter) => void
  onLoadMore: () => void
  preview?: boolean
  maxRows?: number
  hideLoadMore?: boolean
}

type KindFilter = 'all' | 'features' | 'fixes'

const EVENT_TYPES = [
  'task_started',
  'task_completed',
  'task_parked',
  'repo_merged',
  'deploy_started',
  'deploy_finished',
  'deploy_failed',
  'coordinator_warning',
  'coordinator_stop',
  'handoff',
  'decision',
] as const

export function ActivityTimeline({
  events,
  total,
  members,
  services,
  filter,
  onFilterChange,
  onLoadMore,
  preview,
  maxRows,
  hideLoadMore,
}: ActivityTimelineProps) {
  const { t } = useTranslation()
  const [kind, setKind] = useState<KindFilter>('all')

  const aliases = useMemo(() => {
    const set = new Set<string>()
    for (const member of members) {
      if (member.alias) set.add(member.alias)
    }
    for (const ev of events) {
      if (ev.alias) set.add(ev.alias)
    }
    return Array.from(set).sort()
  }, [members, events])

  const filteredEvents = useMemo(() => {
    if (kind === 'features') {
      return events.filter(
        (e) => !e.task_id?.startsWith('FIX-') && !(e.branch && e.branch.startsWith('fix/'))
      )
    }
    if (kind === 'fixes') {
      return events.filter(
        (e) => e.task_id?.startsWith('FIX-') || (e.branch && e.branch.startsWith('fix/'))
      )
    }
    return events
  }, [events, kind])

  const patch = (part: Partial<EventsFilter>) => onFilterChange({ ...filter, ...part })

  const visibleEvents = useMemo(() => {
    const base = filteredEvents
    return maxRows != null ? base.slice(0, maxRows) : base
  }, [filteredEvents, maxRows])

  const body = (
    <>
      {!preview && (
      <div className="flex flex-col gap-3">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div>
            <h2 className="text-lg font-semibold text-slate-900 dark:text-white">
              {t('timeline.title')}
            </h2>
            <p className="text-xs text-slate-500 dark:text-slate-400">
              {t('timeline.subtitle')}
            </p>
          </div>

          <div className="flex items-center space-x-1 p-1 rounded-lg bg-slate-100 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 text-xs self-start sm:self-auto">
            {(['all', 'features', 'fixes'] as const).map((mode) => (
              <button
                key={mode}
                onClick={() => setKind(mode)}
                className={`px-3 py-1 rounded-md transition-colors ${
                  kind === mode
                    ? 'bg-indigo-600 text-white font-medium shadow-xs'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'
                }`}
              >
                {mode === 'all'
                  ? t('timeline.filterAll')
                  : mode === 'features'
                    ? t('timeline.filterFeatures')
                    : t('timeline.filterFixes')}
              </button>
            ))}
          </div>
        </div>

        <div className="flex flex-wrap gap-2">
          <Select
            density="sm"
            className="inline-block w-auto"
            selectClassName="w-auto min-w-[7rem] bg-slate-50 dark:bg-slate-900"
            value={filter.alias}
            onChange={(e) => patch({ alias: e.target.value })}
          >
            <option value="">{t('timeline.filterAlias')}</option>
            {aliases.map((alias) => (
              <option key={alias} value={alias}>
                {alias}
              </option>
            ))}
          </Select>
          <Select
            density="sm"
            className="inline-block w-auto"
            selectClassName="w-auto min-w-[7rem] bg-slate-50 dark:bg-slate-900"
            value={filter.event}
            onChange={(e) => patch({ event: e.target.value })}
          >
            <option value="">{t('timeline.filterEvent')}</option>
            {EVENT_TYPES.map((type) => (
              <option key={type} value={type}>
                {t(`timeline.events.${type}`, { defaultValue: type })}
              </option>
            ))}
          </Select>
          <Select
            density="sm"
            className="inline-block w-auto"
            selectClassName="w-auto min-w-[7rem] bg-slate-50 dark:bg-slate-900"
            value={filter.service}
            onChange={(e) => patch({ service: e.target.value })}
          >
            <option value="">{t('timeline.filterService')}</option>
            {services.map((svc) => (
              <option key={svc.id} value={svc.id}>
                {svc.name}
              </option>
            ))}
          </Select>
        </div>
      </div>
      )}

      <div className="divide-y divide-slate-100 dark:divide-slate-800/60 pt-2">
        {visibleEvents.length === 0 ? (
          <div className="py-8 text-center text-sm text-slate-400 dark:text-slate-500">
            {t('timeline.empty')}
          </div>
        ) : (
          visibleEvents.map((ev, idx) => <TimelineItem key={`${ev.timestamp}-${idx}`} event={ev} />)
        )}
      </div>

      {!preview && !hideLoadMore && events.length < total && (
        <div className="pt-2">
          <Button variant="secondary" size="sm" onClick={onLoadMore}>
            {t('timeline.loadMore', { shown: events.length, total })}
          </Button>
        </div>
      )}
    </>
  )

  if (preview) {
    return body
  }

  return (
    <Card className="p-6 space-y-4 bg-white dark:bg-slate-900/60 border-slate-200 dark:border-slate-800/80">
      {body}
    </Card>
  )
}
