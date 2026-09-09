import { useTranslation } from 'react-i18next'
import { MemberCard } from './MemberCard'
import { ResearchCard } from './ResearchCard'
import { UnassignedWorkCard } from './UnassignedWorkCard'
import type { MemberState, MemberTaskState, StrayRepo } from '../../shared/types/api'

interface TeamPulseSectionProps {
  members: MemberState[]
  stray: StrayRepo[]
  serviceNames: Record<string, string>
}

function memberTasks(member: MemberState): MemberTaskState[] {
  if (member.tasks && member.tasks.length > 0) {
    return member.tasks
  }
  if (member.status === 'in_progress' && member.task_id) {
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
        spend_kind: member.spend_kind,
      },
    ]
  }
  return []
}

export function TeamPulseSection({ members, stray, serviceNames }: TeamPulseSectionProps) {
  const { t } = useTranslation()

  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center gap-2">
          {t('pulse.title')}
          <span className="text-xs text-slate-500 dark:text-slate-400 font-normal">
            ({t('pulse.trackedCount', { count: members.length })})
          </span>
        </h2>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {members.flatMap((member) => {
          const tasks = memberTasks(member)
          const cards = tasks.map((task) => (
            <MemberCard
              key={`${member.alias}-${task.task_id}`}
              member={member}
              task={task}
              serviceNames={serviceNames}
            />
          ))
          if (member.research?.status === 'active') {
            cards.push(<ResearchCard key={`${member.alias}-research`} member={member} />)
          }
          if (cards.length === 0) {
            cards.push(<MemberCard key={member.alias} member={member} serviceNames={serviceNames} />)
          }
          return cards
        })}
        <UnassignedWorkCard items={stray} />
      </div>
    </section>
  )
}
