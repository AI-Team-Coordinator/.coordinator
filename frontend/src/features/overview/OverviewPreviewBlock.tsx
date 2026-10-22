import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '../../shared/ui/Button'

export const OVERVIEW_COLLAPSED_ROWS = 3
export const OVERVIEW_EXPANDED_ROWS = 12
/** Compact Team Pulse on overview — more slots per row. */
export const OVERVIEW_PULSE_COLLAPSED_ROWS = 8
export const OVERVIEW_PULSE_EXPANDED_ROWS = 20

interface OverviewPreviewBlockProps {
  title: string
  icon?: React.ReactNode
  total?: number
  empty?: boolean
  showAllHash: string
  /** card = bordered panel; plain = title outside, content on page background (overview) */
  layout?: 'card' | 'plain'
  collapsedRows?: number
  expandedRows?: number
  /** Rows added per “Show more” (default: expandedRows − collapsedRows). */
  expandStep?: number
  /** Items already loaded in memory (tasks/events); triggers onLoadMore when showing more. */
  loadedCount?: number
  onLoadMore?: () => void
  /** e.g. relay API fetch limit — keep ≥ visible rows. */
  onVisibleRowsChange?: (visible: number) => void
  children: (visibleCount: number) => React.ReactNode
}

export function OverviewPreviewBlock({
  title,
  icon,
  total,
  empty,
  showAllHash,
  layout = 'card',
  collapsedRows = OVERVIEW_COLLAPSED_ROWS,
  expandedRows = OVERVIEW_EXPANDED_ROWS,
  expandStep,
  loadedCount,
  onLoadMore,
  onVisibleRowsChange,
  children,
}: OverviewPreviewBlockProps) {
  const { t } = useTranslation()
  const [visibleCount, setVisibleCount] = useState(collapsedRows)
  const step = expandStep ?? Math.max(1, expandedRows - collapsedRows)

  useEffect(() => {
    setVisibleCount(collapsedRows)
  }, [collapsedRows])

  useEffect(() => {
    onVisibleRowsChange?.(visibleCount)
  }, [visibleCount, onVisibleRowsChange])

  const openAll = () => {
    window.location.hash = showAllHash
  }

  const cap =
    total !== undefined
      ? total
      : loadedCount !== undefined
        ? loadedCount
        : expandedRows

  const canShowMore = !empty && visibleCount < cap
  const canShowLess = visibleCount > collapsedRows

  const showMore = () => {
    const next = Math.min(visibleCount + step, total ?? Number.POSITIVE_INFINITY)
    setVisibleCount(next)
    if (onLoadMore && loadedCount !== undefined && next > loadedCount) {
      onLoadMore()
    }
  }

  const showLess = () => {
    setVisibleCount(collapsedRows)
  }

  const shellClass =
    layout === 'plain' ? 'space-y-4' : 'rounded-xl border border-slate-200/80 dark:border-slate-700/80 bg-white dark:bg-slate-950/40 p-5 space-y-4'

  return (
    <div className={shellClass}>
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex items-center gap-2 min-w-0">
          {icon}
          <h3
            className={
              layout === 'plain'
                ? 'text-lg font-semibold text-slate-900 dark:text-white truncate'
                : 'text-base font-semibold text-slate-900 dark:text-white truncate'
            }
          >
            {title}
          </h3>
          {total !== undefined && (
            <span className="text-xs text-slate-500 dark:text-slate-400 shrink-0 font-normal">
              {t('overview.previewTotal', { count: total })}
            </span>
          )}
        </div>
        <Button size="sm" variant="outline" onClick={openAll} className="shrink-0">
          {t('overview.showAll')}
        </Button>
      </div>

      {empty ? (
        <p className="text-sm text-slate-500">{t('overview.previewEmpty')}</p>
      ) : (
        <>
          <div className="relative">{children(visibleCount)}</div>
          {(canShowMore || canShowLess) && (
            <div className="flex flex-wrap justify-center items-center gap-2 pt-1">
              {canShowLess && (
                <Button
                  size="sm"
                  variant="ghost"
                  className="text-slate-600 dark:text-slate-400"
                  onClick={showLess}
                >
                  {t('overview.showLess')}
                </Button>
              )}
              {canShowMore && (
                <Button
                  size="sm"
                  variant="ghost"
                  className="text-indigo-600 dark:text-indigo-400"
                  onClick={showMore}
                >
                  {t('overview.showMore')}
                </Button>
              )}
            </div>
          )}
        </>
      )}
    </div>
  )
}
