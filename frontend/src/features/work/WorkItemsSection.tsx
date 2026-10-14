import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { WorkItem } from '../../shared/types/api'
import { cn } from '../../shared/lib/utils'

type StatusFilter = '' | 'backlog' | 'open' | 'done'

interface WorkItemsSectionProps {
  items: WorkItem[]
  loading?: boolean
}

export function WorkItemsSection({ items: allItems, loading = false }: WorkItemsSectionProps) {
  const { t } = useTranslation()
  const [status, setStatus] = useState<StatusFilter>('')

  const items = useMemo(() => {
    if (!status) return allItems
    return allItems.filter((item) => item.status === status)
  }, [allItems, status])

  const grouped = useMemo(() => {
    const order = ['open', 'backlog', 'done'] as const
    const buckets: Record<string, WorkItem[]> = { open: [], backlog: [], done: [] }
    for (const item of items) {
      const key = item.status in buckets ? item.status : 'backlog'
      buckets[key].push(item)
    }
    return order.map((k) => ({ key: k, rows: buckets[k] }))
  }, [items])

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold text-slate-900 dark:text-slate-100">{t('work.title')}</h2>
          <p className="text-sm text-slate-500 dark:text-slate-400">{t('work.subtitle')}</p>
        </div>
        <div className="flex gap-1 rounded-lg border border-slate-200 dark:border-slate-700 p-1">
          {(['', 'open', 'backlog', 'done'] as StatusFilter[]).map((s) => (
            <button
              key={s || 'all'}
              type="button"
              onClick={() => setStatus(s)}
              className={cn(
                'px-3 py-1 text-xs rounded-md font-medium transition-colors',
                status === s
                  ? 'bg-indigo-600 text-white'
                  : 'text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800'
              )}
            >
              {s ? t(`work.status.${s}`) : t('work.status.all')}
            </button>
          ))}
        </div>
      </div>

      {loading ? (
        <p className="text-sm text-slate-500">{t('work.loading')}</p>
      ) : items.length === 0 ? (
        <p className="text-sm text-slate-500">{t('work.empty')}</p>
      ) : (
        <div className="space-y-6">
          {grouped.map(
            ({ key, rows }) =>
              rows.length > 0 && (
                <section key={key}>
                  <h3 className="text-xs font-semibold uppercase tracking-wide text-slate-400 mb-2">
                    {t(`work.status.${key}`)} ({rows.length})
                  </h3>
                  <ul className="space-y-2">
                    {rows.map((item) => (
                      <li
                        key={item.task_id}
                        className="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/50 p-4"
                      >
                        <div className="flex flex-wrap items-start justify-between gap-2">
                          <div>
                            <p className="font-medium text-slate-900 dark:text-slate-100">
                              {item.title || item.summary || item.task_id}
                            </p>
                            <p className="text-xs text-slate-500 font-mono mt-0.5">{item.task_id}</p>
                          </div>
                          <span
                            className={cn(
                              'text-[10px] uppercase font-semibold px-2 py-0.5 rounded-full',
                              item.status === 'open' && 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-200',
                              item.status === 'backlog' && 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300',
                              item.status === 'done' && 'bg-slate-200 text-slate-600 dark:bg-slate-800 dark:text-slate-400'
                            )}
                          >
                            {t(`work.status.${item.status}`)}
                          </span>
                        </div>
                        <dl className="mt-3 grid grid-cols-2 sm:grid-cols-4 gap-2 text-xs text-slate-600 dark:text-slate-400">
                          {item.assignee && (
                            <div>
                              <dt className="text-slate-400">{t('work.assignee')}</dt>
                              <dd className="font-medium">{item.assignee}</dd>
                            </div>
                          )}
                          {item.branch && (
                            <div>
                              <dt className="text-slate-400">{t('work.branch')}</dt>
                              <dd className="font-mono">{item.branch}</dd>
                            </div>
                          )}
                          {item.services && item.services.length > 0 && (
                            <div className="col-span-2">
                              <dt className="text-slate-400">{t('work.services')}</dt>
                              <dd>{item.services.join(', ')}</dd>
                            </div>
                          )}
                        </dl>
                        {item.slots && item.slots.length > 0 && (
                          <p className="mt-2 text-xs text-slate-500">
                            {t('work.slots')}:{' '}
                            {item.slots.map((s) => `${s.alias} (${s.status})`).join(' · ')}
                          </p>
                        )}
                      </li>
                    ))}
                  </ul>
                </section>
              )
          )}
        </div>
      )}
    </div>
  )
}
