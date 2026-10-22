import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Card } from '../../shared/ui/Card'
import { Badge } from '../../shared/ui/Badge'
import { TaskDocLink } from '../docs/TaskDocLink'
import { claimBadge } from './claimBadge'
import { MemberAvatar } from './MemberAvatar'
import { MemberCardDetailModal } from './MemberCardDetailModal'
import { MemberCardHeader } from './MemberCardHeader'
import { compactDetailLine, memberCardState } from './memberCardState'
import type { MemberState, MemberTaskState } from '../../shared/types/api'

interface MemberCardCompactProps {
  member: MemberState
  serviceNames: Record<string, string>
  task?: MemberTaskState
}

export function MemberCardCompact({ member, serviceNames, task }: MemberCardCompactProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const state = memberCardState(member, task)
  const claimedLabels = state.claimed.map((id) => ({ id, label: claimBadge(id, serviceNames) }))
  const detail = compactDetailLine(state)

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
        <MemberCardHeader member={member} state={state} />

        <div className="flex items-center gap-2 min-w-0">
          <MemberAvatar alias={member.alias} avatarUrl={member.avatar_url} className="w-7 h-7" />
          <span className="text-xs font-medium text-slate-600 dark:text-slate-300 shrink-0">@{member.alias}</span>
          {claimedLabels.length > 0 ? (
            <div className="flex flex-wrap gap-1 min-w-0 justify-end flex-1">
              {claimedLabels.map(({ id, label }) => (
                <Badge key={id} variant="neutral" className="text-[10px] font-medium truncate max-w-[8rem]">
                  {label}
                </Badge>
              ))}
            </div>
          ) : null}
        </div>

        {state.isBusy ? (
          <div className="flex items-center gap-1.5 min-w-0 overflow-hidden">
            <Badge variant={state.isFix ? 'warning' : 'default'} className="text-[10px] uppercase shrink-0">
              {state.isFix ? t('pulse.fix') : t('pulse.feature')}
            </Badge>
            {detail.text ? (
              detail.link && state.taskId ? (
                <TaskDocLink
                  taskId={state.taskId}
                  title={detail.text}
                  className="text-xs text-slate-700 dark:text-slate-200 min-w-0 flex-1 truncate !font-sans"
                >
                  {detail.text}
                </TaskDocLink>
              ) : (
                <span
                  className="text-xs text-slate-700 dark:text-slate-200 min-w-0 flex-1 truncate"
                  title={detail.text}
                >
                  {detail.text}
                </span>
              )
            ) : null}
          </div>
        ) : null}
      </Card>

      {open ? (
        <MemberCardDetailModal
          member={member}
          task={task}
          serviceNames={serviceNames}
          onClose={() => setOpen(false)}
        />
      ) : null}
    </>
  )
}
