import { Clock, Pause } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Badge } from '../../shared/ui/Badge'
import { cn } from '../../shared/lib/utils'
import type { MemberState } from '../../shared/types/api'
import type { MemberCardState } from './memberCardState'

interface MemberCardHeaderProps {
  member: MemberState
  state: MemberCardState
  /** Overview compact cards: icon-only in progress / parked badges. */
  compact?: boolean
}

/** First row of a pulse member card: title + status badge (unchanged across compact/full). */
export function MemberCardHeader({ member, state, compact }: MemberCardHeaderProps) {
  const { t } = useTranslation()
  const { isParked, isActive, taskTitle, clockPaused } = state
  const parkedLabel = t('pulse.parked')
  const inProgressLabel = t('pulse.inProgress')

  return (
    <div className="flex items-start justify-between gap-2">
      <div className="min-w-0 flex-1">
        {isActive && taskTitle ? (
          <h3
            className="font-semibold text-slate-900 dark:text-white text-sm leading-snug truncate"
            title={taskTitle}
          >
            {taskTitle}
          </h3>
        ) : (
          <h3 className="font-semibold text-slate-900 dark:text-white text-sm truncate">{member.name}</h3>
        )}
      </div>
      <div className="flex items-center gap-1.5 shrink-0">
        {isParked ? (
          <Badge
            variant="warning"
            className={compact ? 'px-1 py-0.5' : 'font-bold text-[10px]'}
            title={parkedLabel}
            aria-label={parkedLabel}
          >
            <Pause className={cn('w-3 h-3', !compact && 'mr-1')} />
            {!compact && parkedLabel}
          </Badge>
        ) : isActive ? (
          <Badge
            variant="success"
            className={compact ? 'px-1.5 py-0.5' : 'font-bold text-[10px]'}
            title={inProgressLabel}
            aria-label={inProgressLabel}
          >
            {compact ? (
              <span className="w-2 h-2 rounded-full bg-emerald-600 dark:bg-emerald-400 block" aria-hidden />
            ) : (
              <>● {inProgressLabel}</>
            )}
          </Badge>
        ) : (
          <Badge variant="neutral" className="font-bold text-[10px]">
            ○ {t('pulse.idle')}
          </Badge>
        )}
        {isActive && clockPaused ? (
          <span title={t('pulse.clockPaused')}>
            <Clock
              className="w-3.5 h-3.5 text-amber-500 dark:text-amber-400"
              aria-label={t('pulse.clockPaused')}
            />
          </span>
        ) : null}
      </div>
    </div>
  )
}

export function MemberCardIdleFooter({ isBusy }: { isBusy: boolean }) {
  const { t } = useTranslation()
  if (isBusy) return null
  return (
    <div className="pt-4 border-t border-slate-100 dark:border-slate-800/40 text-xs text-slate-400 dark:text-slate-500 italic">
      {t('pulse.noActiveTask')}
    </div>
  )
}
