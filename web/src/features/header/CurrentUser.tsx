import { useTranslation } from 'react-i18next'

interface CurrentUserProps {
  alias?: string
  name?: string
  role?: string
  access?: string
}

export function CurrentUser({ alias, name, role, access }: CurrentUserProps) {
  const { t } = useTranslation()
  if (!alias) {
    return (
      <div className="text-xs text-slate-400 dark:text-slate-500">{t('nav.noAuthor')}</div>
    )
  }

  const roleLabel = role ? t(`pulse.roles.${role}`, { defaultValue: role }) : ''
  const accessLabel = access === 'admin' ? t('setup.accessAdmin') : ''

  return (
    <div className="flex items-center gap-2 min-w-0">
      <div className="w-8 h-8 rounded-lg bg-indigo-50 dark:bg-slate-800 border border-indigo-200 dark:border-slate-700 flex items-center justify-center font-bold text-[11px] text-indigo-600 dark:text-indigo-300 shrink-0">
        {alias}
      </div>
      <div className="min-w-0">
        <div className="text-sm font-medium text-slate-900 dark:text-white truncate leading-tight">
          {name || alias}
        </div>
        <div className="text-[11px] text-slate-500 dark:text-slate-400 truncate">
          @{alias}
          {roleLabel ? ` · ${roleLabel}` : ''}
          {accessLabel ? ` · ${accessLabel}` : ''}
        </div>
      </div>
    </div>
  )
}
