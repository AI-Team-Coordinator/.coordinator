import { useTranslation } from 'react-i18next'
import { CoordinatorLogo } from '../../shared/ui/Logo'
import { Badge } from '../../shared/ui/Badge'
import { localized } from '../../shared/lib/localized'
import type { ProjectProfile } from '../../shared/types/api'

interface ProjectBannerProps {
  profile: ProjectProfile | null
}

function GatheringBar() {
  return (
    <div className="mt-2 h-0.5 w-48 max-w-full overflow-hidden rounded-full bg-slate-200 dark:bg-slate-800">
      <div className="coordinator-indeterminate h-full w-1/3 rounded-full bg-indigo-600" />
    </div>
  )
}

export function ProjectBanner({ profile }: ProjectBannerProps) {
  const { t, i18n } = useTranslation()

  if (!profile?.project?.name) {
    return (
      <div className="flex items-center gap-3 min-w-0">
        <CoordinatorLogo size={42} className="shrink-0 md:hidden" />
        <div className="min-w-0">
          <div className="h-7 w-40 rounded-md bg-slate-200 dark:bg-slate-800" />
          <GatheringBar />
          <p className="text-xs text-slate-500 dark:text-slate-400 mt-1.5">
            {t('nav.gathering')}
          </p>
        </div>
      </div>
    )
  }

  const name = profile.project.name
  const tagline = localized(profile.project.tagline, i18n.language)
  const serviceCount = profile.services?.length ?? 0

  return (
    <div className="flex items-center gap-3 min-w-0">
      <CoordinatorLogo size={42} className="shrink-0 md:hidden" />
      <div className="min-w-0">
        <div className="flex items-center gap-2 flex-wrap">
          <h1 className="text-xl md:text-2xl font-bold tracking-tight text-slate-900 dark:text-white truncate">
            {name}
          </h1>
          <Badge variant="default" className="text-[11px]">
            {t('nav.badge')}
          </Badge>
          {profile.github?.org && (
            <a
              href={profile.github.html_url}
              target="_blank"
              rel="noreferrer"
              className="text-[11px] font-medium text-slate-500 dark:text-slate-400 hover:text-indigo-600 dark:hover:text-indigo-300 truncate"
            >
              {profile.github.org}
            </a>
          )}
        </div>
        <p className="text-xs md:text-sm text-slate-500 dark:text-slate-400 mt-0.5 truncate">
          {tagline || t('nav.subtitle')}
          {serviceCount > 0 && (
            <span className="text-slate-400 dark:text-slate-500">
              {' · '}
              {t('project.serviceCount', { count: serviceCount })}
            </span>
          )}
        </p>
      </div>
    </div>
  )
}
