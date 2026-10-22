import { ArrowLeft } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { overviewMainHash } from './overviewNav'

interface OverviewSubpageShellProps {
  title: string
  subtitle?: string
  children: React.ReactNode
}

export function OverviewSubpageShell({ title, subtitle, children }: OverviewSubpageShellProps) {
  const { t } = useTranslation()
  const goBack = () => {
    window.location.hash = overviewMainHash()
  }

  return (
    <div className="space-y-6">
      <button
        type="button"
        onClick={goBack}
        className="inline-flex items-center gap-1.5 text-sm text-indigo-600 dark:text-indigo-400 hover:underline"
      >
        <ArrowLeft className="w-4 h-4" />
        {t('overview.back')}
      </button>
      <div>
        <h2 className="text-lg font-semibold text-slate-900 dark:text-white">{title}</h2>
        {subtitle && <p className="text-sm text-slate-600 dark:text-slate-300 mt-0.5">{subtitle}</p>}
      </div>
      {children}
    </div>
  )
}
