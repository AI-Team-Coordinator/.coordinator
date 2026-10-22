import { useState } from 'react'
import { Copy, Check, GitBranch } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Card } from '../../shared/ui/Card'
import { Badge } from '../../shared/ui/Badge'
import { formatDuration } from '../../shared/lib/formatters'
import { hasUsageSpend, SharedQuotaIcon, UsageSpend } from '../../shared/ui/UsageSpend'
import { TaskDocLink } from '../docs/TaskDocLink'
import { RelatedLinks } from '../docs/RelatedLinks'
import { ChatTabs } from './ChatTabs'
import { claimBadge } from './claimBadge'
import { MemberAvatar } from './MemberAvatar'
import { MemberCardHeader, MemberCardIdleFooter } from './MemberCardHeader'
import { cn } from '../../shared/lib/utils'
import { memberCardState } from './memberCardState'
import type { ChatTab, MemberState, MemberTaskState } from '../../shared/types/api'

interface MemberCardProps {
  member: MemberState
  serviceNames: Record<string, string>
  task?: MemberTaskState
  /** Inside detail modal: no outer Card shell. */
  embedded?: boolean
}

export function MemberCard({ member, serviceNames, task, embedded }: MemberCardProps) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)

  const state = memberCardState(member, task)
  const {
    isParked,
    isActive,
    isBusy,
    taskId,
    branch,
    summary,
    isFix,
    claimed,
    hasDoc,
    clockPaused,
  } = state
  const roleLabel = member.role ? t(`pulse.roles.${member.role}`, { defaultValue: member.role }) : ''
  const accessLabel = member.access === 'admin' ? t('setup.accessAdmin') : ''
  const claimedLabels = claimed.map((id) => ({ id, label: claimBadge(id, serviceNames) }))
  const repos = (task ? task?.repos : member.repos) || []
  const workspaceRepos = repos.filter((repo) => repo.kind === 'workspace')
  const productRepos = repos.filter((repo) => repo.kind !== 'workspace')
  const showInfraHint = productRepos.length === 0 && (workspaceRepos.length > 0 || (isActive && claimed.length > 0))
  const duration = task ? task?.duration_seconds : member.duration_seconds
  const budget = task ? task?.budget_usd : member.budget_usd
  const ondemand = task ? task?.ondemand_usd : member.ondemand_usd
  const cursorModelsPct = task ? task?.cursor_models_pct : member.cursor_models_pct
  const otherModelsPct = task ? task?.other_models_pct : member.other_models_pct
  const spendKind = task ? task?.spend_kind : member.spend_kind
  const spendShared = task ? task?.spend_shared : member.spend_shared

  const copyBranch = () => {
    if (branch) {
      navigator.clipboard.writeText(branch)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    }
  }

  const Wrapper = embedded ? 'div' : Card

  return (
    <Wrapper
      className={cn('min-w-0 flex flex-col', embedded ? 'gap-4' : 'gap-3 p-5')}
    >
      <MemberCardHeader member={member} state={state} />
      {isActive && summary ? (
        <p className="text-xs text-slate-500 dark:text-slate-400 leading-relaxed line-clamp-3" title={summary}>
          {summary}
        </p>
      ) : null}
      {isActive && ((task?.related_tasks && task.related_tasks.length > 0) || (task?.related_docs && task.related_docs.length > 0)) ? (
        <div className="space-y-1">
          <div className="text-[10px] uppercase tracking-wide text-slate-400 dark:text-slate-500">{t('pulse.related')}</div>
          <RelatedLinks taskIds={task?.related_tasks} docs={task?.related_docs} />
        </div>
      ) : null}

      <div className="flex items-center gap-2 min-w-0">
        <MemberAvatar alias={member.alias} avatarUrl={member.avatar_url} className="w-7 h-7" />
        <div className="text-xs text-slate-500 dark:text-slate-400 truncate">
          {member.name}
          {' · '}@{member.alias}
          {roleLabel ? ` · ${roleLabel}` : ''}
          {accessLabel ? ` · ${accessLabel}` : ''}
        </div>
      </div>
      {member.crews && member.crews.length > 0 ? (
        <div className="flex flex-wrap gap-1">
          {member.crews.map((id) => (
            <Badge key={id} variant="neutral" className="text-[10px] font-mono">
              {id}
            </Badge>
          ))}
        </div>
      ) : null}

      {isActive ? <ChatTabs chats={taskChats(task)} /> : null}

      {repos.length > 0 ? (
        <div className="space-y-1.5">
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
        <div className="space-y-1.5">
          {showInfraHint && (
            <div className="text-[11px] text-slate-500 dark:text-slate-400">{t('pulse.infraHint')}</div>
          )}
          <div className="flex flex-wrap gap-1">
            {claimedLabels.map(({ id, label }) => (
              <Badge key={id} variant="neutral" className="text-[10px] font-medium">
                {label}
              </Badge>
            ))}
          </div>
        </div>
      ) : null}

      {isActive ? (
        <div className="pt-4 border-t border-slate-100 dark:border-slate-800/80 space-y-3">
          <div className="flex items-center justify-between text-xs">
            <span className="text-slate-500 dark:text-slate-400">{t('pulse.duration')}:</span>
            <span className="inline-flex items-center gap-1.5">
              <span
                className={`font-mono font-semibold ${
                  clockPaused ? 'text-amber-600 dark:text-amber-400' : 'text-emerald-600 dark:text-emerald-400'
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
              <span>{t(isParked ? 'pulse.parkedBranch' : 'pulse.branch')}:</span>
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
              {hasDoc && taskId ? (
                <TaskDocLink taskId={taskId} className="text-xs text-slate-800 dark:text-slate-200">
                  {taskId}
                </TaskDocLink>
              ) : (
                <span className="font-mono text-xs text-slate-800 dark:text-slate-200 truncate">{taskId}</span>
              )}
            </div>
          </div>
        </div>
      ) : null}

      <MemberCardIdleFooter isBusy={isBusy} />
    </Wrapper>
  )
}

function taskChats(task?: MemberTaskState): ChatTab[] | undefined {
  if (task?.chats && task.chats.length > 0) {
    return task.chats
  }
  if (!task?.agents || task.agents.length === 0) {
    return undefined
  }
  return task.agents.map((agent) => ({
    title: agent.title || agent.id,
    session_id: agent.id,
  }))
}

