import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { ServiceGroupCard } from './ServiceGroupCard'
import { GitHubOrgCard } from './GitHubOrgCard'
import { CreateUnitForm } from './CreateUnitForm'
import type { ProjectProfile } from '../../shared/types/api'

interface ServiceMapProps {
  profile: ProjectProfile | null
  onCreated?: () => void
}

export function ServiceMap({ profile, onCreated }: ServiceMapProps) {
  const { t } = useTranslation()

  const grouped = useMemo(() => {
    if (!profile) return []
    return profile.groups
      .map((group) => ({
        group,
        services: profile.services.filter((svc) => svc.group === group.id),
      }))
      .filter((item) => item.services.length > 0)
  }, [profile])

  if (!profile) {
    return null
  }

  const bound = Boolean(profile.github?.org)

  return (
    <section className="space-y-4">
      {bound ? (
        <>
          <GitHubOrgCard github={profile.github} live={profile.github_live} />
          <CreateUnitForm profile={profile} onCreated={onCreated || (() => {})} />
        </>
      ) : grouped.length === 0 ? (
        <p className="text-sm text-slate-500 dark:text-slate-400">{t('project.emptyBoard')}</p>
      ) : null}
      {grouped.length > 0 && (
        <>
          <div>
            <h2 className="text-lg font-semibold text-slate-900 dark:text-white">{t('project.mapTitle')}</h2>
            <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              {t('project.mapSubtitle', { edges: profile.edges.length })}
            </p>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-4">
            {grouped.map(({ group, services }) => (
              <ServiceGroupCard
                key={group.id}
                group={group}
                services={services}
                githubOrg={profile.github?.org}
              />
            ))}
          </div>
        </>
      )}
    </section>
  )
}
