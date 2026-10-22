import { AlertTriangle } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { TaskDocLink } from '../docs/TaskDocLink'
import type { Conflict } from '../../shared/types/api'

interface ConflictRadarProps {
  conflicts: Conflict[]
}

function hasTaskDetail(c: Conflict): boolean {
  return (c.tasks?.length ?? 0) >= 2
}

function zoneLabel(c: Conflict): string | undefined {
  if (c.service) return c.service
  if (c.description.includes('claiming ')) {
    return c.description.split('claiming ').pop()?.trim()
  }
  return undefined
}

export function ConflictRadar({ conflicts }: ConflictRadarProps) {
  const { t } = useTranslation()

  if (!conflicts || conflicts.length === 0) {
    return null
  }

  const critical = conflicts.some((c) => c.severity === 'critical')
  const icon = critical ? 'text-rose-600 dark:text-rose-400' : 'text-amber-600 dark:text-amber-400'
  const titleClass = critical
    ? 'font-semibold text-sm bg-gradient-to-r from-rose-950 via-rose-500 to-rose-950 dark:from-rose-100 dark:via-rose-400 dark:to-rose-100 bg-[length:200%_auto] bg-clip-text text-transparent animate-conflict-title-shimmer'
    : 'font-semibold text-sm text-amber-900 dark:text-amber-200'

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2">
        <AlertTriangle className={`w-5 h-5 ${icon} shrink-0`} />
        <h3 className={titleClass}>
          {critical ? t('radar.warning') : t('radar.notice')}
        </h3>
      </div>

      {conflicts.map((c, i) => {
        const zone = zoneLabel(c)
        const isCritical = c.severity === 'critical'
        const cardShell = isCritical
          ? [
              'border bg-gradient-to-br from-rose-50 via-rose-100/85 to-slate-100',
              'border-rose-300/80 shadow-sm ring-1 ring-rose-200/60',
              'dark:from-rose-950 dark:via-[#3b0f14] dark:to-slate-950 dark:border-rose-800/55',
              'dark:ring-0 dark:shadow-[inset_0_1px_0_0_rgba(251,113,133,0.12)]',
            ].join(' ')
          : [
              'border bg-gradient-to-br from-amber-50 via-amber-100/80 to-slate-100',
              'border-amber-300/70 shadow-sm ring-1 ring-amber-200/50',
              'dark:from-amber-950 dark:via-[#3d2810] dark:to-slate-950 dark:border-amber-800/45',
              'dark:ring-0 dark:shadow-[inset_0_1px_0_0_rgba(251,191,36,0.1)]',
            ].join(' ')
        const textMain = isCritical
          ? 'text-rose-950/90 dark:text-rose-50/95'
          : 'text-amber-950/90 dark:text-amber-50/95'
        const textStrong = isCritical ? 'text-rose-950 dark:text-rose-50' : 'text-amber-950 dark:text-amber-50'
        const textMuted = isCritical
          ? 'text-rose-800/75 dark:text-rose-200/75'
          : 'text-amber-900/75 dark:text-amber-200/75'
        const textCell = isCritical
          ? 'text-rose-900 dark:text-rose-50/90'
          : 'text-amber-900 dark:text-amber-50/90'
        const tableBorder = isCritical
          ? 'border-rose-200/90 dark:border-rose-900/60'
          : 'border-amber-200/90 dark:border-amber-900/50'
        const rowBorder = isCritical
          ? 'border-rose-200/70 dark:border-rose-900/40'
          : 'border-amber-200/70 dark:border-amber-900/35'
        const zoneBadge = isCritical
          ? 'bg-rose-950/10 text-rose-800/80 dark:bg-white/10 dark:text-rose-200/80'
          : 'bg-amber-950/10 text-amber-900/80 dark:bg-white/10 dark:text-amber-200/80'
        const tableShell = isCritical
          ? 'bg-white/95 dark:bg-black/20'
          : 'bg-white/95 dark:bg-black/20'

        return (
          <div
            key={`${c.title}-${c.service ?? ''}-${i}`}
            className={`text-xs ${textMain} space-y-2 rounded-lg p-2.5 ${cardShell}`}
          >
            <div className="flex flex-wrap items-center gap-x-2 gap-y-1 leading-relaxed">
              <strong className={textStrong}>{c.title}</strong>
              {zone && (
                <span className={`font-mono text-[10px] px-1.5 py-0.5 rounded ${zoneBadge}`}>
                  {t('radar.zone')}: {zone}
                </span>
              )}
            </div>
            {hasTaskDetail(c) ? (
              <div className={`overflow-x-auto rounded-md border ${tableBorder} ${tableShell}`}>
                <table className="w-full table-fixed text-left border-collapse text-[11px]">
                  <colgroup>
                    <col style={{ width: '48%' }} />
                    <col style={{ width: '52%' }} />
                  </colgroup>
                  <thead>
                    <tr className={`border-b ${rowBorder} text-[10px] uppercase tracking-wide font-mono ${textMuted}`}>
                      <th className="px-2 py-1.5 font-medium text-left">{t('radar.colTask')}</th>
                      <th className="px-2 py-1.5 font-medium text-left">{t('radar.colBranch')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {c.tasks!.map((row) => {
                      const hasDoc = Boolean(row.has_doc)
                      const headline = (hasDoc ? (row.doc_title ?? '') : (row.title ?? '')).trim()
                      return (
                        <tr key={row.task_id} className={`border-b ${rowBorder} last:border-0`}>
                          <td className={`px-2 py-1.5 align-top max-w-0 overflow-hidden ${textCell}`}>
                            <div className="min-w-0 space-y-0.5">
                              {headline ? (
                                <div
                                  className={`font-sans text-[12px] font-medium leading-snug truncate ${textStrong}`}
                                  title={headline}
                                >
                                  {row.alias && row.alias !== c.affected_aliases?.[0] ? (
                                    <span className={`font-normal opacity-70 ${textMuted}`}>
                                      [{row.alias}]{' '}
                                    </span>
                                  ) : null}
                                  {headline}
                                </div>
                              ) : row.alias && row.alias !== c.affected_aliases?.[0] ? (
                                <div className={`font-sans text-[12px] ${textMuted}`}>[{row.alias}]</div>
                              ) : null}
                              {hasDoc ? (
                                <TaskDocLink
                                  taskId={row.task_id}
                                  title={row.task_id}
                                  className={`block truncate max-w-full text-[10px] ${textMuted}`}
                                />
                              ) : null}
                            </div>
                          </td>
                          <td
                            className={`px-2 py-1.5 align-top max-w-0 overflow-hidden font-mono truncate ${textMuted}`}
                            title={row.branch || undefined}
                          >
                            {row.branch || '—'}
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              </div>
            ) : (
              <p className={`leading-relaxed ${textMuted}`}>
                {c.description}
                {c.affected_aliases?.length ? (
                  <>
                    {' '}
                    ({t('radar.affected')}: {c.affected_aliases.join(', ')})
                  </>
                ) : null}
              </p>
            )}
          </div>
        )
      })}
    </div>
  )
}
