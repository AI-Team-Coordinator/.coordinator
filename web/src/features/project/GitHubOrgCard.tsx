import { useTranslation } from 'react-i18next'
import { ExternalLink } from 'lucide-react'
import { Card } from '../../shared/ui/Card'
import { Badge } from '../../shared/ui/Badge'
import type { GitHubBinding, GitHubLive } from '../../shared/types/api'

interface GitHubOrgCardProps {
  github?: GitHubBinding
  live?: GitHubLive
}

export function GitHubOrgCard({ github, live }: GitHubOrgCardProps) {
  const { t } = useTranslation()

  if (!github?.org) {
    return (
      <Card className="p-4 bg-white dark:bg-slate-900/70">
        <p className="text-sm font-semibold text-slate-900 dark:text-white">{t('project.github.title')}</p>
        <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">{t('project.github.unbound')}</p>
      </Card>
    )
  }

  const authVariant = live?.auth === 'ok' ? 'success' : live?.auth ? 'warning' : 'neutral'
  const authLabel =
    live?.auth === 'ok'
      ? t('project.github.authOk', { login: live.login || '—', role: live.role || 'member' })
      : live?.auth === 'missing'
        ? t('project.github.authMissing')
        : live?.detail
          ? t('project.github.authError', { detail: live.detail })
          : t('project.github.authUnknown')

  return (
    <Card className="p-4 bg-white dark:bg-slate-900/70">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-xs uppercase tracking-wide text-slate-400">{t('project.github.title')}</p>
          <a
            href={github.html_url}
            target="_blank"
            rel="noreferrer"
            className="mt-1 inline-flex items-center gap-1.5 text-sm font-semibold text-slate-900 dark:text-white hover:text-indigo-600 dark:hover:text-indigo-300"
          >
            {github.org}
            <ExternalLink className="w-3.5 h-3.5 shrink-0" />
          </a>
          {github.ssh_host && (
            <p className="text-[11px] text-slate-400 mt-1">
              {t('project.github.sshHost', { host: github.ssh_host })}
            </p>
          )}
        </div>
        <Badge variant={authVariant} className="max-w-full whitespace-normal text-left font-medium">
          {authLabel}
        </Badge>
      </div>
    </Card>
  )
}
