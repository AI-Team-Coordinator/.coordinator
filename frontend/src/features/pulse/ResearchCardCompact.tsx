import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Card } from '../../shared/ui/Card'
import { Badge } from '../../shared/ui/Badge'
import { MemberAvatar } from './MemberAvatar'
import { ResearchCard } from './ResearchCard'
import { MemberCardDetailModalShell } from './MemberCardDetailModalShell'
import type { MemberState } from '../../shared/types/api'

interface ResearchCardCompactProps {
  member: MemberState
}

export function ResearchCardCompact({ member }: ResearchCardCompactProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const summary = member.research?.summary || ''

  const openDetail = () => setOpen(true)

  return (
    <>
      <Card
        role="button"
        tabIndex={0}
        className="p-3 space-y-2 cursor-pointer hover:bg-slate-50 dark:hover:bg-slate-900/70 transition-colors focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500"
        onClick={openDetail}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            openDetail()
          }
        }}
      >
        <div className="flex items-start justify-between gap-3">
          <h3
            className="font-semibold text-slate-900 dark:text-white text-sm leading-snug line-clamp-2 min-w-0 flex-1"
            title={summary || member.name}
          >
            {summary || t('pulse.research')}
          </h3>
          <Badge variant="default" className="font-bold text-[10px] shrink-0">
            ● {t('pulse.research')}
          </Badge>
        </div>
        <div className="flex items-center gap-2">
          <MemberAvatar alias={member.alias} avatarUrl={member.avatar_url} className="w-7 h-7" />
          <span className="text-xs font-medium text-slate-600 dark:text-slate-300">@{member.alias}</span>
        </div>
      </Card>
      {open ? (
        <MemberCardDetailModalShell onClose={() => setOpen(false)}>
          <ResearchCard member={member} embedded />
        </MemberCardDetailModalShell>
      ) : null}
    </>
  )
}
