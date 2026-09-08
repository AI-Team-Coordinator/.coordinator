import { useTranslation } from 'react-i18next'
import { Badge } from '../../shared/ui/Badge'
import { formatDate, formatTime, formatUSD } from '../../shared/lib/formatters'
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
  const hasBudget = typeof event.budget_usd === 'number' && event.budget_usd > 0
  const hasOnDemand = typeof event.ondemand_usd === 'number' && event.ondemand_usd > 0
  const hasCursorPct = event.cursor_models_pct !== undefined && event.cursor_models_pct > 0
  const hasOtherPct = event.other_models_pct !== undefined && event.other_models_pct > 0

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
        {hasBudget && (
          <span className="text-emerald-700 dark:text-emerald-400 font-mono shrink-0">
            {t('timeline.planShare', { amount: formatUSD(event.budget_usd) })}
          </span>
        )}
        {hasOnDemand && (
          <span className="text-amber-600 dark:text-amber-400 font-mono shrink-0">
            {t('timeline.onDemand', { amount: formatUSD(event.ondemand_usd) })}
          </span>
        )}
        {event.spend_kind === 'infra' && (
          <span className="text-slate-500 dark:text-slate-400 font-mono shrink-0">
            {t('timeline.infra')}
          </span>
        )}
        {hasCursorPct && (
          <span className="text-slate-400 dark:text-slate-500 font-mono hidden lg:inline shrink-0">
            {t('timeline.cursorModels', { pct: event.cursor_models_pct?.toFixed(1) })}
          </span>
        )}
        {hasOtherPct && (
          <span className="text-slate-400 dark:text-slate-500 font-mono hidden xl:inline shrink-0">
            {t('timeline.otherModels', { pct: event.other_models_pct?.toFixed(1) })}
          </span>
        )}
      </div>

      <div className="text-slate-400 dark:text-slate-500 whitespace-nowrap text-right font-mono text-[11px]">
        {formatDate(event.timestamp, i18n.language)} {formatTime(event.timestamp)}
      </div>
    </div>
  )
}
