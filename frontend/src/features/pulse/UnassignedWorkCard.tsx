import { useTranslation } from 'react-i18next'
import { Card } from '../../shared/ui/Card'
import { Badge } from '../../shared/ui/Badge'
import type { StrayRepo } from '../../shared/types/api'

interface UnassignedWorkCardProps {
  items: StrayRepo[]
}

export function UnassignedWorkCard({ items }: UnassignedWorkCardProps) {
  const { t } = useTranslation()
  if (items.length === 0) {
    return null
  }

  return (
    <Card className="p-5 border-amber-300/70 dark:border-amber-700/50 bg-amber-50/50 dark:bg-amber-950/20 md:col-span-2 lg:col-span-1">
      <div className="flex items-center justify-between gap-2 mb-3">
        <h3 className="font-semibold text-slate-900 dark:text-white text-sm">{t('pulse.strayTitle')}</h3>
        <Badge variant="warning" className="text-[10px] font-medium">
          {t('pulse.strayBadge')}
        </Badge>
      </div>
      <p className="text-xs text-slate-500 dark:text-slate-400 mb-3">{t('pulse.straySubtitle')}</p>
      <div className="space-y-3">
        {items.map((item) => (
          <div key={`${item.repo}-${item.branch}`}>
            <div className="flex items-center justify-between gap-2 text-xs">
              <span className="font-medium text-slate-800 dark:text-slate-200 truncate">{item.name}</span>
              <span className="font-mono text-[10px] text-slate-400 truncate">{item.branch}</span>
            </div>
            <ul className="mt-1 space-y-0.5">
              {item.commits.map((c) => (
                <li key={c.hash} className="text-[11px] text-slate-600 dark:text-slate-400 font-mono truncate" title={c.subject}>
                  {c.hash} {c.subject}
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
    </Card>
  )
}
