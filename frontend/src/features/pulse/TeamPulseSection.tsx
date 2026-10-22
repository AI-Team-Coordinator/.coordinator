import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { MemberCard } from './MemberCard'
import { MemberCardCompact } from './MemberCardCompact'
import { ResearchCard } from './ResearchCard'
import { ResearchCardCompact } from './ResearchCardCompact'
import { UnassignedWorkCard } from './UnassignedWorkCard'
import type { MemberState, MemberTaskState, StrayRepo } from '../../shared/types/api'

interface TeamPulseSectionProps {
  members: MemberState[]
  stray: StrayRepo[]
  serviceNames: Record<string, string>
  maxCards?: number
  hideTitle?: boolean
  /** Overview preview: compact cards + modal for full detail. */
  compact?: boolean
}

function memberTasks(member: MemberState): MemberTaskState[] {
  if (member.tasks && member.tasks.length > 0) {
    return member.tasks
  }
  if ((member.status === 'in_progress' || member.status === 'parked') && member.task_id) {
    return [
      {
        task_id: member.task_id,
        task_title: member.task_title,
        task_doc: member.task_doc,
        task_summary: member.task_summary,
        branch: member.branch,
        services: member.services,
        duration_seconds: member.duration_seconds,
        repos: member.repos,
        cost_usd: member.cost_usd,
        budget_usd: member.budget_usd,
        ondemand_usd: member.ondemand_usd,
        cursor_models_pct: member.cursor_models_pct,
        other_models_pct: member.other_models_pct,
        spend_kind: member.spend_kind,
        spend_shared: member.spend_shared,
      },
    ]
  }
  return []
}

export function TeamPulseSection({
  members,
  stray,
  serviceNames,
  maxCards,
  hideTitle,
  compact,
}: TeamPulseSectionProps) {
  const { t } = useTranslation()

  const memberCards = members.flatMap((member) => {
    const tasks = memberTasks(member)
    const cards: ReactNode[] = tasks.map((task) =>
      compact ? (
        <MemberCardCompact
          key={`${member.alias}-${task.task_id}`}
          member={member}
          task={task}
          serviceNames={serviceNames}
        />
      ) : (
        <MemberCard
          key={`${member.alias}-${task.task_id}`}
          member={member}
          task={task}
          serviceNames={serviceNames}
        />
      )
    )
    if (member.research?.status === 'active') {
      cards.push(
        compact ? (
          <ResearchCardCompact key={`${member.alias}-research`} member={member} />
        ) : (
          <ResearchCard key={`${member.alias}-research`} member={member} />
        )
      )
    }
    if (cards.length === 0) {
      cards.push(
        compact ? (
          <MemberCardCompact key={member.alias} member={member} serviceNames={serviceNames} />
        ) : (
          <MemberCard key={member.alias} member={member} serviceNames={serviceNames} />
        )
      )
    }
    return cards
  })

  const hasStray = stray.length > 0
  const memberLimit =
    maxCards == null ? memberCards.length : Math.min(memberCards.length, maxCards)
  const visibleMembers = memberCards.slice(0, memberLimit)
  const showStrayFullWidth = hasStray && (maxCards == null || maxCards > memberCards.length)

  return (
    <section className="space-y-4">
      {!hideTitle && (
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center gap-2">
            {t('pulse.title')}
            <span className="text-xs text-slate-500 dark:text-slate-400 font-normal">
              ({t('pulse.trackedCount', { count: members.length })})
            </span>
          </h2>
        </div>
      )}

      <div className="space-y-3">
        <div
          className={
            compact
              ? 'grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3'
              : 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4'
          }
        >
          {visibleMembers}
        </div>
        {showStrayFullWidth ? <UnassignedWorkCard items={stray} fullWidth /> : null}
      </div>
    </section>
  )
}
