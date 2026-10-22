import { useState } from 'react'
import { cn } from '../../shared/lib/utils'

interface MemberAvatarProps {
  alias: string
  avatarUrl?: string
  className?: string
}

export function MemberAvatar({ alias, avatarUrl, className }: MemberAvatarProps) {
  const [failed, setFailed] = useState(false)
  const showImg = Boolean(avatarUrl) && !failed

  return showImg ? (
    <img
      src={avatarUrl}
      alt=""
      onError={() => setFailed(true)}
      className={cn(
        'rounded-full object-cover border border-slate-200 dark:border-slate-700 shrink-0 bg-slate-100 dark:bg-slate-800',
        className
      )}
    />
  ) : (
    <div
      className={cn(
        'rounded-full bg-indigo-50 dark:bg-slate-800 border border-indigo-200 dark:border-slate-700 flex items-center justify-center font-bold text-indigo-600 dark:text-indigo-300 shrink-0',
        className
      )}
    >
      <span className="text-[10px] leading-none">{alias}</span>
    </div>
  )
}
