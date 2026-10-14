import { Loader2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useLoadingStatus } from './LoadingStatusContext'

export function LoadingStatusOverlay() {
  const { activities } = useLoadingStatus()
  const { t } = useTranslation()

  if (activities.length === 0) return null

  const labels = activities.map((a) => a.message)
  const detail =
    labels.length === 1
      ? labels[0]
      : t('loading.combined', { items: labels.join(t('loading.separator')) })

  return (
    <>
      <div className="fixed top-0 inset-x-0 z-30 h-px overflow-hidden bg-slate-200/40 dark:bg-slate-700/30">
        <div className="coordinator-indeterminate h-full w-1/4 bg-slate-400/70 dark:bg-slate-500/60" />
      </div>
      <div
        className="fixed z-40 left-3 bottom-[4.75rem] sm:bottom-3 max-w-[min(20rem,calc(100vw-1.5rem))] rounded-md px-2.5 py-1.5 bg-slate-100/90 dark:bg-slate-900/90 border border-slate-200/50 dark:border-slate-600/40"
        role="status"
        aria-live="polite"
      >
        <p className="flex items-center gap-2 text-xs text-slate-600/90 dark:text-slate-400/90 leading-snug">
          <Loader2 className="h-3.5 w-3.5 shrink-0 animate-spin text-slate-500/80 dark:text-slate-400/80" aria-hidden />
          <span className="min-w-0">{detail}</span>
        </p>
      </div>
    </>
  )
}
