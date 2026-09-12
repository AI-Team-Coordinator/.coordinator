import { AlertTriangle } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import type { Conflict } from '../../shared/types/api'

interface ConflictRadarProps {
  conflicts: Conflict[]
}

export function ConflictRadar({ conflicts }: ConflictRadarProps) {
  const { t } = useTranslation()

  if (!conflicts || conflicts.length === 0) {
    return null
  }

  const critical = conflicts.some((c) => c.severity === 'critical')
  const box = critical
    ? 'bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-500/40 animate-pulse-slow'
    : 'bg-amber-50 dark:bg-amber-950/40 border border-amber-200 dark:border-amber-500/40'
  const icon = critical ? 'text-rose-600 dark:text-rose-400' : 'text-amber-600 dark:text-amber-400'
  const heading = critical ? 'text-rose-800 dark:text-rose-300' : 'text-amber-800 dark:text-amber-300'
  const body = critical ? 'text-rose-700 dark:text-rose-200/90' : 'text-amber-800 dark:text-amber-200/90'

  return (
    <div className={`${box} rounded-xl p-4 shadow-sm`}>
      <div className="flex items-start space-x-3">
        <AlertTriangle className={`w-5 h-5 ${icon} shrink-0 mt-0.5`} />
        <div className="flex-1 space-y-1">
          <h3 className={`font-semibold ${heading} text-sm`}>
            {critical ? t('radar.warning') : t('radar.notice')}
          </h3>
          {conflicts.map((c, i) => (
            <div key={i} className={`text-xs ${body} leading-relaxed`}>
              <strong>{c.title}:</strong> {c.description} (
              {t('radar.affected')}: {c.affected_aliases.join(', ')})
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
