import { useState } from 'react'
import { Copy, Check, Clock, GitBranch } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Card } from '../../shared/ui/Card'
import { Badge } from '../../shared/ui/Badge'
import { formatDuration } from '../../shared/lib/formatters'
import { hasUsageSpend, SharedQuotaIcon, UsageSpend } from '../../shared/ui/UsageSpend'
import { TaskDocLink } from '../docs/TaskDocLink'
import { ChatTabs } from './ChatTabs'
import type { MemberState, MemberTaskState } from '../../shared/types/api'

interface MemberCardProps {
  member: MemberState
  serviceNames: Record<string, string>
  task?: MemberTaskState
}

export function MemberCard({ member, serviceNames, task }: MemberCardProps) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)

  const isActive = Boolean(task) || (!task && member.status === 'in_progress')
  const isBusy = isActive
  const taskId = task?.task_id || member.task_id
  const branch = task?.branch || member.branch
  const summary = task?.task_summary || member.task_summary
  const isFix = Boolean(taskId && (taskId.startsWith('FIX-') || (branch && branch.startsWith('fix/'))))
  const roleLabel = member.role ? t(`pulse.roles.${member.role}`, { defaultValue: member.role }) : ''
  const accessLabel = member.access === 'admin' ? t('setup.accessAdmin') : ''
  const claimed = task?.services || member.services || []
  const repos = task?.repos || member.repos || []
  const workspaceRepos = repos.filter((repo) => repo.kind === 'workspace')
  const productRepos = repos.filter((repo) => repo.kind !== 'workspace')
  const showInfraHint = productRepos.length === 0 && (workspaceRepos.length > 0 || (isActive && claimed.length > 0))
  const claimedLabels = claimed.map((id) => serviceNames[id] || id)
  const taskTitle = task?.task_title || task?.task_id || member.task_title || member.task_id || ''
  const duration = task?.duration_seconds ?? member.duration_seconds
  const budget = task?.budget_usd ?? member.budget_usd
  const ondemand = task?.ondemand_usd ?? member.ondemand_usd
  const cursorModelsPct = task?.cursor_models_pct ?? member.cursor_models_pct
  const otherModelsPct = task?.other_models_pct ?? member.other_models_pct
  const spendKind = task?.spend_kind || member.spend_kind
  const spendShared = task?.spend_shared ?? member.spend_shared

  const copyBranch = () => {
    if (branch) {
      navigator.clipboard.writeText(branch)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    }
  }

  return (
    <Card
      className={`p-5 transition-all duration-200 ${
        isBusy
          ? 'border-indigo-400/60 dark:border-indigo-500/40 bg-white/90 dark:bg-slate-900/90 shadow-md shadow-indigo-500/5'
          : 'bg-slate-50/60 dark:bg-slate-900/40 border-slate-200 dark:border-slate-800/60 opacity-80 hover:opacity-100'
      }`}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          {isActive && taskTitle ? (
            <h3 className="font-semibold text-slate-900 dark:text-white text-sm leading-snug truncate" title={taskTitle}>
              {taskTitle}
            </h3>
          ) : (
            <h3 className="font-semibold text-slate-900 dark:text-white text-sm">{member.name}</h3>
          )}
          {isActive && summary ? (
            <p className="mt-1 text-xs text-slate-500 dark:text-slate-400 leading-snug line-clamp-2" title={summary}>
              {summary}
            </p>
          ) : null}
        </div>
        <div className="flex items-center gap-1.5 shrink-0">
          {isActive ? (
            <Badge variant="success" className="font-bold text-[10px]">
              ● {t('pulse.inProgress')}
            </Badge>
          ) : (
            <Badge variant="neutral" className="font-bold text-[10px]">
              ○ {t('pulse.idle')}
            </Badge>
          )}
          {isActive && (task?.clock_paused ?? member.clock_paused) ? (
            <Clock
              className="w-3.5 h-3.5 text-amber-500 dark:text-amber-400"
              aria-label={t('pulse.clockPaused')}
              title={t('pulse.clockPaused')}
            />
          ) : null}
        </div>
      </div>

      <div className="mt-2 flex items-center space-x-2">
        <div className="w-7 h-7 rounded-full bg-indigo-50 dark:bg-slate-800 border border-indigo-200 dark:border-slate-700 flex items-center justify-center font-bold text-[11px] text-indigo-600 dark:text-indigo-300 shrink-0">
          {member.alias}
        </div>
        <div className="text-xs text-slate-500 dark:text-slate-400 truncate">
          {member.name}
          {' · '}@{member.alias}
          {roleLabel ? ` · ${roleLabel}` : ''}
          {accessLabel ? ` · ${accessLabel}` : ''}
        </div>
      </div>

      {isActive ? (
        <div className="mt-3">
          <ChatTabs chats={task?.chats} />
        </div>
      ) : null}

      {repos.length > 0 ? (
        <div className="mt-3 space-y-1.5">
          {showInfraHint && (
            <div className="text-[11px] text-slate-500 dark:text-slate-400">{t('pulse.infraHint')}</div>
          )}
          {repos.map((repo) => {
            const workspace = repo.kind === 'workspace'
            return (
              <div key={repo.repo} className="flex items-center justify-between gap-2">
                <span className="text-xs text-slate-700 dark:text-slate-300 truncate">{repo.name}</span>
                <div className="flex items-center gap-1 shrink-0">
                  <Badge
                    variant={
                      repo.state === 'merged' || repo.state === 'deployed' || repo.state === 'infra'
                        ? 'success'
                        : repo.state === 'pushed'
                          ? 'default'
                          : 'warning'
                    }
                    className="text-[10px] font-medium"
                  >
                    {t(`pulse.repoState.${repo.state}`, { defaultValue: repo.state })}
                    {repo.ahead ? ` +${repo.ahead}` : ''}
                  </Badge>
                  {repo.dirty && (
                    <Badge variant="warning" className="text-[10px] font-medium">
                      {t('pulse.dirty')}
                    </Badge>
                  )}
                  {!workspace && repo.state === 'merged' && (
                    <Badge variant={repo.deployed ? 'success' : 'neutral'} className="text-[10px] font-medium">
                      {repo.deployed ? t('pulse.repoState.deployed') : t('pulse.notDeployed')}
                    </Badge>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      ) : isActive && claimedLabels.length > 0 ? (
        <div className="mt-3 space-y-1.5">
          {showInfraHint && (
            <div className="text-[11px] text-slate-500 dark:text-slate-400">{t('pulse.infraHint')}</div>
          )}
          <div className="flex flex-wrap gap-1">
            {claimedLabels.map((label) => (
              <Badge key={label} variant="neutral" className="text-[10px] font-medium">
                {label}
              </Badge>
            ))}
          </div>
        </div>
      ) : null}

      {isActive ? (
        <div className="mt-4 pt-4 border-t border-slate-100 dark:border-slate-800/80 space-y-3">
          <div className="flex items-center justify-between text-xs">
            <span className="text-slate-500 dark:text-slate-400">{t('pulse.duration')}:</span>
            <span className="inline-flex items-center gap-1.5">
              <span
                className={`font-mono font-semibold ${
                  task?.clock_paused ?? member.clock_paused
                    ? 'text-amber-600 dark:text-amber-400'
                    : 'text-emerald-600 dark:text-emerald-400'
                }`}
              >
                {formatDuration(duration)}
              </span>
              {spendShared ? <SharedQuotaIcon /> : null}
            </span>
          </div>

          {hasUsageSpend({
            budgetUsd: budget,
            ondemandUsd: ondemand,
            cursorModelsPct,
            otherModelsPct,
            shared: spendShared,
          }) && (
            <UsageSpend
              budgetUsd={budget}
              ondemandUsd={ondemand}
              cursorModelsPct={cursorModelsPct}
              otherModelsPct={otherModelsPct}
              spendKind={spendKind}
              shared={spendShared}
            />
          )}

          <div>
            <div className="text-[11px] text-slate-500 dark:text-slate-400 mb-1 flex items-center gap-1">
              <GitBranch className="w-3 h-3" />
              <span>{t('pulse.branch')}:</span>
            </div>
            <div className="flex items-center justify-between bg-slate-100 dark:bg-slate-950 px-2.5 py-1.5 rounded-md border border-slate-200 dark:border-slate-800 group">
              <span className="font-mono text-xs text-indigo-700 dark:text-indigo-300 truncate select-all">
                {branch || 'unknown'}
              </span>
              {branch && (
                <button
                  onClick={copyBranch}
                  title="Copy branch name"
                  className="ml-2 text-slate-400 hover:text-indigo-600 dark:hover:text-indigo-300 transition shrink-0"
                >
                  {copied ? <Check className="w-3.5 h-3.5 text-emerald-500" /> : <Copy className="w-3.5 h-3.5" />}
                </button>
              )}
            </div>
          </div>

          <div>
            <div className="text-[11px] text-slate-500 dark:text-slate-400 mb-1">{t('pulse.task')}:</div>
            <div className="flex items-center gap-1.5">
              <Badge variant={isFix ? 'warning' : 'default'} className="text-[10px] uppercase">
                {isFix ? t('pulse.fix') : t('pulse.feature')}
              </Badge>
              <TaskDocLink
                taskId={taskId}
                className="text-xs text-slate-800 dark:text-slate-200"
              >
                {taskId}
              </TaskDocLink>
            </div>
          </div>
        </div>
      ) : null}

      {!isBusy ? (
        <div className="mt-4 pt-4 border-t border-slate-100 dark:border-slate-800/40 text-xs text-slate-400 dark:text-slate-500 italic">
          {t('pulse.noActiveTask')}
        </div>
      ) : null}
    </Card>
  )
}
