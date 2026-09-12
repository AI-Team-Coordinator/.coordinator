import { useMemo } from 'react'
import { Clock } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Card } from '../../shared/ui/Card'
import { Badge } from '../../shared/ui/Badge'
import { Button } from '../../shared/ui/Button'
import { formatDate, formatDuration, formatTime, formatUSD } from '../../shared/lib/formatters'
import { UsageSpend } from '../../shared/ui/UsageSpend'
import { TaskDocLink } from '../docs/TaskDocLink'
import type { MemberState, TaskItem } from '../../shared/types/api'

export interface TasksFilter {
  status: '' | 'in_progress' | 'completed'
  alias: string
  kind: '' | 'feature' | 'fix'
}

interface TaskTableProps {
  tasks: TaskItem[]
  total: number
  members: MemberState[]
  filter: TasksFilter
  onFilterChange: (next: TasksFilter) => void
  onLoadMore: () => void
}

const selectClass =
  'px-2 py-1 rounded-md bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 text-xs text-slate-700 dark:text-slate-200'

const STATUS_MODES = ['all', 'in_progress', 'completed'] as const

export function TaskTable({ tasks, total, members, filter, onFilterChange, onLoadMore }: TaskTableProps) {
  const { t, i18n } = useTranslation()

  const namesByAlias = useMemo(() => {
    const map: Record<string, string> = {}
    for (const member of members) {
      if (!member.alias) continue
      map[member.alias] = member.name?.trim() || member.alias
    }
    return map
  }, [members])

  const people = useMemo(() => {
    const set = new Set<string>()
    for (const member of members) {
      if (member.alias) set.add(member.alias)
    }
    for (const task of tasks) {
      if (task.alias) set.add(task.alias)
    }
    return Array.from(set).sort((a, b) => {
      const left = namesByAlias[a] || a
      const right = namesByAlias[b] || b
      return left.localeCompare(right)
    })
  }, [members, tasks, namesByAlias])

  const personName = (alias?: string) => {
    if (!alias) return ''
    return namesByAlias[alias] || alias
  }

  const patch = (part: Partial<TasksFilter>) => onFilterChange({ ...filter, ...part })
  const statusMode = filter.status === '' ? 'all' : filter.status

  return (
    <Card className="p-6 space-y-4 bg-white dark:bg-slate-900/60 border-slate-200 dark:border-slate-800/80">
      <div className="flex flex-col gap-3">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div>
            <h2 className="text-lg font-semibold text-slate-900 dark:text-white">{t('tasks.title')}</h2>
            <p className="text-xs text-slate-500 dark:text-slate-400">{t('tasks.subtitle')}</p>
          </div>
          <div className="flex items-center space-x-1 p-1 rounded-lg bg-slate-100 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 text-xs self-start sm:self-auto">
            {STATUS_MODES.map((mode) => (
              <button
                key={mode}
                type="button"
                onClick={() => patch({ status: mode === 'all' ? '' : mode })}
                className={`px-3 py-1 rounded-md transition-colors ${
                  statusMode === mode
                    ? 'bg-indigo-600 text-white font-medium shadow-xs'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'
                }`}
              >
                {t(`tasks.status.${mode}`)}
              </button>
            ))}
          </div>
        </div>

        <div className="flex flex-wrap gap-2">
          <select
            className={selectClass}
            value={filter.alias}
            onChange={(e) => patch({ alias: e.target.value })}
          >
            <option value="">{t('timeline.filterAlias')}</option>
            {people.map((alias) => (
              <option key={alias} value={alias}>
                {personName(alias)}
              </option>
            ))}
          </select>
          <select
            className={selectClass}
            value={filter.kind}
            onChange={(e) => patch({ kind: e.target.value as TasksFilter['kind'] })}
          >
            <option value="">{t('timeline.filterAll')}</option>
            <option value="feature">{t('timeline.filterFeatures')}</option>
            <option value="fix">{t('timeline.filterFixes')}</option>
          </select>
        </div>
      </div>

      <div className="overflow-x-auto -mx-2 sm:mx-0">
        <table className="w-full min-w-[720px] text-left text-xs">
          <thead>
            <tr className="border-b border-slate-200 dark:border-slate-800 text-[11px] uppercase tracking-wider text-slate-500 dark:text-slate-400">
              <th className="py-2 pr-3 font-medium">{t('tasks.columns.status')}</th>
              <th className="py-2 pr-3 font-medium">{t('tasks.columns.task')}</th>
              <th className="py-2 pr-3 font-medium">{t('tasks.columns.person')}</th>
              <th className="py-2 pr-3 font-medium">{t('tasks.columns.when')}</th>
              <th className="py-2 pr-3 font-medium">{t('tasks.columns.duration')}</th>
              <th className="py-2 pr-3 font-medium text-right">{t('tasks.columns.plan')}</th>
              <th className="py-2 font-medium text-right">{t('tasks.columns.ondemand')}</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100 dark:divide-slate-800/60">
            {tasks.length === 0 ? (
              <tr>
                <td colSpan={7} className="py-8 text-center text-sm text-slate-400 dark:text-slate-500">
                  {t('tasks.empty')}
                </td>
              </tr>
            ) : (
              tasks.map((task) => {
                const open = task.status === 'in_progress'
                const whenTs = open ? task.started_at : task.completed_at || task.started_at
                return (
                  <tr key={`${task.task_id}-${task.status}-${task.started_at}`} className="align-top">
                    <td className="py-3 pr-3 whitespace-nowrap">
                      <Badge variant={open ? 'success' : 'default'} className="text-[10px] uppercase">
                        {t(`tasks.status.${task.status}`)}
                      </Badge>
                      {task.kind === 'fix' && (
                        <span className="ml-1.5 text-amber-600 dark:text-amber-400">{t('pulse.fix')}</span>
                      )}
                      {task.spend_kind === 'infra' && (
                        <span className="ml-1.5 text-slate-400">{t('timeline.infra')}</span>
                      )}
                    </td>
                    <td className="py-3 pr-3 min-w-0">
                      <div className="font-medium text-slate-900 dark:text-slate-100 truncate">
                        {task.title || task.task_id}
                      </div>
                      <TaskDocLink
                        taskId={task.task_id}
                        className="text-[11px] text-slate-500 dark:text-slate-400"
                      >
                        {task.task_id}
                      </TaskDocLink>
                      {task.branch && (
                        <div className="font-mono text-[11px] text-slate-400 dark:text-slate-500 truncate">
                          {task.branch}
                        </div>
                      )}
                    </td>
                    <td className="py-3 pr-3 text-slate-700 dark:text-slate-300 whitespace-nowrap">
                      {task.alias ? personName(task.alias) : '—'}
                    </td>
                    <td className="py-3 pr-3 text-slate-500 dark:text-slate-400 whitespace-nowrap font-mono text-[11px]">
                      {whenTs ? (
                        <>
                          <div>
                            {formatDate(whenTs, i18n.language)} {formatTime(whenTs)}
                          </div>
                          {open ? t('tasks.started') : t('tasks.finished')}
                        </>
                      ) : (
                        '—'
                      )}
                    </td>
                    <td className="py-3 pr-3 font-mono text-slate-700 dark:text-slate-300 whitespace-nowrap">
                      <span className="inline-flex items-center gap-1.5">
                        {formatDuration(task.duration_seconds)}
                        {open && task.clock_paused ? (
                          <Clock
                            className="w-3.5 h-3.5 text-amber-500 dark:text-amber-400"
                            aria-label={t('pulse.clockPaused')}
                            title={t('pulse.clockPaused')}
                          />
                        ) : null}
                      </span>
                    </td>
                    <td className="py-3 pr-3 text-right whitespace-nowrap">
                      <UsageSpend
                        variant="cell"
                        budgetUsd={task.budget_usd}
                        cursorModelsPct={task.cursor_models_pct}
                        otherModelsPct={task.other_models_pct}
                      />
                    </td>
                    <td className="py-3 text-right font-mono text-amber-600 dark:text-amber-400 whitespace-nowrap">
                      {task.ondemand_usd != null && task.ondemand_usd > 0
                        ? formatUSD(task.ondemand_usd)
                        : task.cost_usd != null && task.cost_usd > 0
                          ? formatUSD(task.cost_usd)
                          : '—'}
                    </td>
                  </tr>
                )
              })
            )}
          </tbody>
        </table>
      </div>

      {tasks.length < total && (
        <div className="pt-2">
          <Button variant="secondary" size="sm" onClick={onLoadMore}>
            {t('timeline.loadMore', { shown: tasks.length, total })}
          </Button>
        </div>
      )}
    </Card>
  )
}
