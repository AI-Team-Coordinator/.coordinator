import { Radio } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { ThemeToggle } from '../../shared/ui/ThemeToggle'
import { LanguageSwitcher } from '../../shared/ui/LanguageSwitcher'
import { ProjectBanner } from '../project/ProjectBanner'
import { CurrentUser } from './CurrentUser'
import type { ProjectProfile } from '../../shared/types/api'

interface HeaderProps {
  connected: boolean
  profile: ProjectProfile | null
  currentUser: {
    alias?: string
    name?: string
    role?: string
  }
}

export function Header({ connected, profile, currentUser }: HeaderProps) {
  const { t } = useTranslation()

  return (
    <header className="flex flex-wrap items-center justify-between pb-6 border-b border-slate-200 dark:border-slate-800 gap-3">
      <div className="flex items-center gap-4 min-w-0">
        <ProjectBanner profile={profile} />
        <div className="hidden sm:block w-px h-8 bg-slate-200 dark:bg-slate-800 shrink-0" />
        <CurrentUser alias={currentUser.alias} name={currentUser.name} role={currentUser.role} />
      </div>

      <div className="flex items-center space-x-2 ml-auto">
        <div className="flex items-center space-x-2 bg-slate-100 dark:bg-slate-900 px-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-800 text-xs text-slate-600 dark:text-slate-300">
          <span
            className={`w-2 h-2 rounded-full ${
              connected ? 'bg-emerald-500 animate-pulse-slow' : 'bg-rose-500'
            }`}
          />
          <span className="hidden sm:inline">
            {connected ? t('nav.liveSync') : t('nav.connecting')}
          </span>
          <Radio className="w-3.5 h-3.5 text-slate-400" />
        </div>

        <LanguageSwitcher />
        <ThemeToggle />
      </div>
    </header>
  )
}
