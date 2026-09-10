import { useTranslation } from 'react-i18next'
import { Badge } from '../../shared/ui/Badge'
import { formatDate, formatTime } from '../../shared/lib/formatters'
import { UsageSpend } from '../../shared/ui/UsageSpend'
import { TaskDocLink } from '../docs/TaskDocLink'
import type { EventItem } from '../../shared/types/api'

interface TimelineItemProps {
  event: EventItem
}

export function TimelineItem({ event }: TimelineItemProps) {
  const { t, i18n } = useTranslation()
  const isStart = event.event === 'task_started'
  const isDeploy = event.event.startsWith('deploy_')
  const isMerged = event.event === 'repo_merged'
  const isResearch = event.event.startsWith('research_')
  const eventLabel = t(`timeline.events.${event.event}`, { defaultValue: event.event })

  return (
    <div className="py-3 flex items-center justify-between gap-4 text-xs">
      <div className="flex items-center space-x-2.5 truncate">
        <span
          className={`w-2 h-2 rounded-full shrink-0 ${
            isStart ? 'bg-emerald-500' : isDeploy ? 'bg-amber-500' : isMerged ? 'bg-sky-500' : isResearch ? 'bg-violet-500' : 'bg-indigo-500'
          }`}
        />
        <span className="font-bold text-slate-700 dark:text-slate-300 font-mono">
          @{event.alias}
        </span>
        <Badge
          variant={isStart ? 'success' : isDeploy ? 'warning' : 'default'}
          className="text-[10px] uppercase font-semibold"
        >
          {eventLabel}
        </Badge>
        {event.service && (
          <span className="text-slate-600 dark:text-slate-300 font-medium">{event.service}</span>
        )}
        {event.summary && !event.task_id && (
          <span className="text-slate-600 dark:text-slate-300 truncate max-w-[18rem]" title={event.summary}>
            {event.summary}
          </span>
        )}
        {event.findings && (
          <span className="text-slate-500 dark:text-slate-400 truncate max-w-[22rem]" title={event.findings}>
            {event.findings}
          </span>
        )}
        <TaskDocLink taskId={event.task_id} className="text-slate-900 dark:text-slate-200">
          {event.task_id}
        </TaskDocLink>
        {event.branch && (
          <span className="text-slate-400 dark:text-slate-500 font-mono hidden md:inline">
            ({event.branch})
          </span>
        )}
        <UsageSpend
          variant="inline"
          budgetUsd={event.budget_usd}
          ondemandUsd={event.ondemand_usd}
          cursorModelsPct={event.cursor_models_pct}
          otherModelsPct={event.other_models_pct}
          spendKind={event.spend_kind}
        />
      </div>

      <div className="text-slate-400 dark:text-slate-500 whitespace-nowrap text-right font-mono text-[11px]">
        {formatDate(event.timestamp, i18n.language)} {formatTime(event.timestamp)}
      </div>
    </div>
  )
}
