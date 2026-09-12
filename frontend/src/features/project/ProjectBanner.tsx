import { useTranslation } from 'react-i18next'
import { CoordinatorLogo } from '../../shared/ui/Logo'
import type { ProjectProfile } from '../../shared/types/api'

interface ProjectBannerProps {
  profile: ProjectProfile | null
}

export function ProjectBanner({ profile }: ProjectBannerProps) {
  const { t } = useTranslation()
  const name = profile?.project?.name

  return (
    <div className="flex items-center gap-3 min-w-0">
      <CoordinatorLogo size={36} className="shrink-0 md:hidden" />
      {name ? (
        <h1 className="text-xl md:text-2xl font-bold tracking-tight text-slate-900 dark:text-white truncate">
          {name}
        </h1>
      ) : (
        <div className="min-w-0">
          <div className="h-7 w-40 rounded-md bg-slate-200 dark:bg-slate-800" />
          <p className="text-xs text-slate-500 dark:text-slate-400 mt-1.5">{t('nav.gathering')}</p>
        </div>
      )}
    </div>
  )
}
