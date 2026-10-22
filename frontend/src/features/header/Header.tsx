import { Radio } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { ThemeToggle } from '../../shared/ui/ThemeToggle'
import { ProjectBanner } from '../project/ProjectBanner'
import { CurrentUser } from './CurrentUser'
import type { ApiSyncStatus } from './apiSyncStatus'
import type { ProjectProfile } from '../../shared/types/api'
import { cn } from '../../shared/lib/utils'

interface HeaderProps {
  apiSync: ApiSyncStatus
  profile: ProjectProfile | null
  currentUser: {
    alias?: string
    name?: string
    role?: string
    access?: string
    avatarUrl?: string
    avatarBust?: number
  }
}

const DOT: Record<ApiSyncStatus, string> = {
  loading: 'bg-amber-400 animate-pulse',
  live: 'bg-emerald-500 animate-pulse-slow',
  offline: 'bg-rose-500',
}

export function Header({ apiSync, profile, currentUser }: HeaderProps) {
  const { t } = useTranslation()
  const liveLabel = t(`nav.apiSync.${apiSync}`)

  return (
    <header className="flex flex-wrap items-center justify-between pb-6 border-b border-slate-200 dark:border-slate-800 gap-3">
      <div className="flex items-center gap-4 min-w-0">
        <ProjectBanner profile={profile} />
        <div className="hidden sm:block w-px h-8 bg-slate-200 dark:bg-slate-800 shrink-0" />
        <CurrentUser
          alias={currentUser.alias}
          name={currentUser.name}
          role={currentUser.role}
          access={currentUser.access}
          avatarUrl={currentUser.avatarUrl}
          avatarBust={currentUser.avatarBust}
        />
      </div>

      <div className="flex items-center gap-2 ml-auto shrink-0 h-9">
        <div
          className={cn(
            'flex items-center justify-center gap-1.5 h-9 min-w-9 px-2.5 rounded-lg border',
            'bg-slate-100 dark:bg-slate-900 border-slate-200 dark:border-slate-800'
          )}
          title={liveLabel}
          aria-label={liveLabel}
        >
          <span className={cn('w-2 h-2 rounded-full shrink-0', DOT[apiSync])} />
          <Radio className="w-4 h-4 text-slate-400 shrink-0" aria-hidden />
        </div>
        <ThemeToggle />
      </div>
    </header>
  )
}
