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

  return (
    <div className="bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-500/40 rounded-xl p-4 shadow-sm animate-pulse-slow">
      <div className="flex items-start space-x-3">
        <AlertTriangle className="w-5 h-5 text-rose-600 dark:text-rose-400 shrink-0 mt-0.5" />
        <div className="flex-1 space-y-1">
          <h3 className="font-semibold text-rose-800 dark:text-rose-300 text-sm">
            {t('radar.warning')}
          </h3>
          {conflicts.map((c, i) => (
            <div key={i} className="text-xs text-rose-700 dark:text-rose-200/90 leading-relaxed">
              <strong>{c.title}:</strong> {c.description} (
              {t('radar.affected')}: {c.affected_aliases.join(', ')})
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
