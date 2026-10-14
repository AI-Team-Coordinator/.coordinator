import { useTranslation } from 'react-i18next'
import type { StatsPeriodPreset } from './statsPeriod'

interface StatsPeriodToolbarProps {
  preset: StatsPeriodPreset
  onPresetChange: (preset: StatsPeriodPreset) => void
  customFrom: string
  customTo: string
  onCustomFromChange: (v: string) => void
  onCustomToChange: (v: string) => void
}

const PRESETS: StatsPeriodPreset[] = ['billing', '7d', '30d', 'custom']

export function StatsPeriodToolbar({
  preset,
  onPresetChange,
  customFrom,
  customTo,
  onCustomFromChange,
  onCustomToChange,
}: StatsPeriodToolbarProps) {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col sm:flex-row sm:items-end gap-3 flex-wrap">
      <div className="inline-flex rounded-lg border border-slate-200 dark:border-slate-700 p-0.5 bg-slate-50 dark:bg-slate-900/80">
        {PRESETS.map((id) => (
          <button
            key={id}
            type="button"
            onClick={() => onPresetChange(id)}
            className={`px-2.5 py-1.5 text-[11px] font-medium rounded-md transition ${
              preset === id
                ? 'bg-white dark:bg-slate-800 text-indigo-700 dark:text-indigo-300 shadow-sm'
                : 'text-slate-500 dark:text-slate-400'
            }`}
          >
            {t(`statsPeriod.${id}`)}
          </button>
        ))}
      </div>
      {preset === 'custom' && (
        <div className="flex flex-wrap items-center gap-2 text-[11px]">
          <label className="flex flex-col gap-0.5 text-slate-500">
            {t('statsPeriod.from')}
            <input
              type="date"
              value={customFrom}
              onChange={(e) => onCustomFromChange(e.target.value)}
              className="rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2 py-1"
            />
          </label>
          <label className="flex flex-col gap-0.5 text-slate-500">
            {t('statsPeriod.to')}
            <input
              type="date"
              value={customTo}
              onChange={(e) => onCustomToChange(e.target.value)}
              className="rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2 py-1"
            />
          </label>
        </div>
      )}
    </div>
  )
}
