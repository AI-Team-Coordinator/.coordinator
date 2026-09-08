import { useTranslation } from 'react-i18next'
import { MemberCard } from './MemberCard'
import { UnassignedWorkCard } from './UnassignedWorkCard'
import type { MemberState, StrayRepo } from '../../shared/types/api'

interface TeamPulseSectionProps {
  members: MemberState[]
  stray: StrayRepo[]
  serviceNames: Record<string, string>
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
        {members.map((member) => (
          <MemberCard key={member.alias} member={member} serviceNames={serviceNames} />
        ))}
        <UnassignedWorkCard items={stray} />
      </div>
    </section>
  )
}
