import { useTranslation } from 'react-i18next'
import { cn } from '../../shared/lib/utils'

const PLACEHOLDER_COUNT = 8

interface TeamPulseLoadingProps {
  compact?: boolean
  className?: string
}

export function TeamPulseLoading({ compact, className }: TeamPulseLoadingProps) {
  const { t } = useTranslation()

  return (
    <div className={cn('space-y-3', className)} role="status" aria-live="polite">
      <div className="flex items-center gap-2 text-sm text-slate-500 dark:text-slate-400">
        <span
          className="inline-block w-4 h-4 rounded-full border-2 border-indigo-500 border-t-transparent animate-spin shrink-0"
          aria-hidden
        />
        <span>{t('pulse.loading')}</span>
      </div>
      <div
        className={
          compact
            ? 'grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3'
            : 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4'
        }
      >
        {Array.from({ length: PLACEHOLDER_COUNT }, (_, i) => (
          <div
            key={i}
            className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950/40 p-3 space-y-2 animate-pulse"
          >
            <div className="flex justify-between gap-2">
              <div className="h-4 bg-slate-200 dark:bg-slate-800 rounded-md w-[70%]" />
              <div className="h-5 w-5 bg-slate-200 dark:bg-slate-800 rounded-full shrink-0" />
            </div>
            <div className="flex items-center gap-2">
              <div className="w-7 h-7 rounded-full bg-slate-200 dark:bg-slate-800 shrink-0" />
              <div className="h-3 bg-slate-200 dark:bg-slate-800 rounded flex-1 max-w-[5rem]" />
            </div>
            <div className="h-3 bg-slate-200 dark:bg-slate-800 rounded w-full" />
          </div>
        ))}
      </div>
    </div>
  )
}
