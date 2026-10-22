import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import type { MemberState, WorkItem } from '../../shared/types/api'
import { cn } from '../../shared/lib/utils'
import { WorkItemKanbanCard } from './WorkItemKanbanCard'

const COLUMNS = ['backlog', 'open', 'done'] as const
type ColumnKey = (typeof COLUMNS)[number]

const COLUMN_STYLES: Record<ColumnKey, string> = {
  backlog: 'border-slate-200 dark:border-slate-700 bg-slate-50/80 dark:bg-slate-900/30',
  open: 'border-emerald-200 dark:border-emerald-900/50 bg-emerald-50/50 dark:bg-emerald-950/20',
  done: 'border-slate-200 dark:border-slate-800 bg-slate-100/80 dark:bg-slate-900/40',
}

const COLUMN_HEADER: Record<ColumnKey, string> = {
  backlog: 'text-slate-600 dark:text-slate-400',
  open: 'text-emerald-800 dark:text-emerald-300',
  done: 'text-slate-500 dark:text-slate-500',
}

interface WorkItemsSectionProps {
  items: WorkItem[]
  members?: MemberState[]
  loading?: boolean
}

export function WorkItemsSection({ items: allItems, members = [], loading = false }: WorkItemsSectionProps) {
  const { t } = useTranslation()

  const avatarByAlias = useMemo(() => {
    const m = new Map<string, string | undefined>()
    for (const member of members) {
      m.set(member.alias, member.avatar_url)
    }
    return m
  }, [members])

  const columns = useMemo(() => {
    const buckets: Record<ColumnKey, WorkItem[]> = { backlog: [], open: [], done: [] }
    for (const item of allItems) {
      const key: ColumnKey =
        item.status === 'open' || item.status === 'done' ? (item.status as ColumnKey) : 'backlog'
      buckets[key].push(item)
    }
    return COLUMNS.map((key) => ({ key, rows: buckets[key] }))
  }, [allItems])

  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-lg font-semibold text-slate-900 dark:text-slate-100">{t('work.title')}</h2>
        <p className="text-sm text-slate-500 dark:text-slate-400">{t('work.subtitle')}</p>
      </div>

      {loading ? (
        <p className="text-sm text-slate-500">{t('work.loading')}</p>
      ) : allItems.length === 0 ? (
        <p className="text-sm text-slate-500">{t('work.empty')}</p>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4 items-start min-w-0">
          {columns.map(({ key, rows }) => (
            <section
              key={key}
              className={cn(
                'rounded-xl border flex flex-col min-h-[12rem] max-h-[calc(100vh-12rem)] min-w-0 overflow-hidden',
                COLUMN_STYLES[key]
              )}
            >
              <header
                className={cn(
                  'shrink-0 px-3 py-2.5 border-b border-inherit flex items-center justify-between gap-2',
                  COLUMN_HEADER[key]
                )}
              >
                <h3 className="text-xs font-semibold uppercase tracking-wide">{t(`work.status.${key}`)}</h3>
                <span className="text-xs font-medium tabular-nums">{rows.length}</span>
              </header>
              <div className="flex-1 min-w-0 overflow-y-auto overflow-x-hidden p-2 space-y-2">
                {rows.length === 0 ? (
                  <p className="text-xs text-slate-400 dark:text-slate-500 px-1 py-4 text-center">
                    {t('work.columnEmpty')}
                  </p>
                ) : (
                  rows.map((item) => (
                    <WorkItemKanbanCard key={item.task_id} item={item} avatarByAlias={avatarByAlias} />
                  ))
                )}
              </div>
            </section>
          ))}
        </div>
      )}
    </div>
  )
}
