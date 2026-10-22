import { useTranslation } from 'react-i18next'
import { TaskDocLink } from '../docs/TaskDocLink'
import { MemberAvatar } from '../pulse/MemberAvatar'
import { Badge } from '../../shared/ui/Badge'
import type { WorkItem } from '../../shared/types/api'
import { cn } from '../../shared/lib/utils'

interface WorkItemKanbanCardProps {
  item: WorkItem
  avatarByAlias: Map<string, string | undefined>
}

function collectPeople(item: WorkItem): string[] {
  const seen = new Set<string>()
  const out: string[] = []
  const add = (alias?: string) => {
    const a = alias?.trim()
    if (!a || seen.has(a)) return
    seen.add(a)
    out.push(a)
  }
  add(item.assignee)
  for (const s of item.slots ?? []) {
    add(s.alias)
  }
  add(item.created_by)
  return out
}

export function WorkItemKanbanCard({ item, avatarByAlias }: WorkItemKanbanCardProps) {
  const { t } = useTranslation()
  const people = collectPeople(item)
  const headline = item.title || item.summary || item.task_id

  return (
    <article
      className={cn(
        'rounded-lg border border-slate-200 dark:border-slate-800',
        'bg-white dark:bg-slate-900/60 p-3 shadow-sm space-y-2 min-w-0 max-w-full overflow-hidden'
      )}
    >
      {people.length > 0 && (
        <div className="flex items-center gap-1 flex-wrap">
          {people.map((alias) => (
            <span key={alias} className="inline-flex items-center gap-1" title={`@${alias}`}>
              <MemberAvatar alias={alias} avatarUrl={avatarByAlias.get(alias)} className="w-6 h-6" />
            </span>
          ))}
        </div>
      )}

      <div className="min-w-0 max-w-full overflow-hidden">
        <TaskDocLink
          taskId={item.task_id}
          className="text-sm font-medium text-slate-900 dark:text-slate-100 !font-sans block w-full max-w-full text-left line-clamp-3 break-words [overflow-wrap:anywhere] !whitespace-normal"
          title={headline}
        >
          {headline}
        </TaskDocLink>
        <p className="text-[10px] text-slate-500 font-mono mt-0.5 truncate" title={item.task_id}>
          {item.task_id}
        </p>
      </div>

      {item.branch && (
        <p className="text-[11px] font-mono text-slate-500 dark:text-slate-400 truncate" title={item.branch}>
          {item.branch}
        </p>
      )}

      {item.services && item.services.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {item.services.map((svc) => (
            <Badge key={svc} variant="neutral" className="text-[10px] font-medium max-w-full truncate">
              {svc}
            </Badge>
          ))}
        </div>
      )}

      {item.slots && item.slots.length > 0 && (
        <p
          className="text-[10px] text-slate-500 dark:text-slate-400 line-clamp-2 break-words [overflow-wrap:anywhere]"
          title={item.slots.map((s) => `${s.alias} (${s.status})`).join(' · ')}
        >
          {t('work.slots')}:{' '}
          {item.slots.map((s) => `${s.alias} (${s.status})`).join(' · ')}
        </p>
      )}
    </article>
  )
}
