import { useTranslation } from 'react-i18next'
import { Card } from '../../shared/ui/Card'
import { localized } from '../../shared/lib/localized'
import type { ServiceGroup, ServiceNode } from '../../shared/types/api'

interface ServiceGroupCardProps {
  group: ServiceGroup
  services: ServiceNode[]
  githubOrg?: string
}

export function ServiceGroupCard({ group, services, githubOrg }: ServiceGroupCardProps) {
  const { i18n } = useTranslation()

  return (
    <Card className="p-4 space-y-3 bg-white dark:bg-slate-900/70">
      <div className="flex items-baseline justify-between gap-2">
        <h3 className="text-sm font-semibold text-slate-900 dark:text-white">
          {localized(group.label, i18n.language)}
        </h3>
        <span className="text-[11px] text-slate-400">{services.length}</span>
      </div>
      <ul className="space-y-2.5">
        {services.map((svc) => (
          <li key={svc.id}>
            <div className="text-sm font-medium text-slate-800 dark:text-slate-100">{svc.name}</div>
            {svc.html_url && svc.github_repo && (
              <a
                href={svc.html_url}
                target="_blank"
                rel="noreferrer"
                className="text-[11px] text-indigo-600 dark:text-indigo-300 hover:underline"
              >
                {githubOrg ? `${githubOrg}/${svc.github_repo}` : svc.github_repo}
              </a>
            )}
            <div className="text-xs text-slate-500 dark:text-slate-400 leading-snug">
              {localized(svc.purpose, i18n.language)}
            </div>
          </li>
        ))}
      </ul>
    </Card>
  )
}
