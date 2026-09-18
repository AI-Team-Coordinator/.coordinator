import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { RelatedLinks } from '../docs/RelatedLinks'
import type { TaskItem } from '../../shared/types/api'

interface EvolutionForestProps {
  tasks: TaskItem[]
}

interface ForestNode {
  task: TaskItem
  children: TaskItem[]
}

export function EvolutionForest({ tasks }: EvolutionForestProps) {
  const { t } = useTranslation()
  const titles = useMemo(() => {
    const map: Record<string, string> = {}
    for (const task of tasks) {
      if (task.task_id) map[task.task_id] = task.title || task.task_id
    }
    return map
  }, [tasks])

  const forest = useMemo(() => buildForest(tasks), [tasks])
  if (forest.length === 0) {
    return null
  }

  return (
    <div className="pt-2 space-y-3">
      <div>
        <h3 className="text-sm font-semibold text-slate-900 dark:text-white">{t('tasks.evolution')}</h3>
        <p className="text-[11px] text-slate-500 dark:text-slate-400">{t('tasks.evolutionHint')}</p>
      </div>
      <ul className="space-y-2">
        {forest.map((node) => (
          <li key={node.task.task_id} className="text-xs">
            <div className="font-medium text-slate-800 dark:text-slate-100">{node.task.title || node.task.task_id}</div>
            <RelatedLinks taskIds={node.task.related_tasks} docs={node.task.related_docs} titles={titles} />
            {node.children.length > 0 ? (
              <ul className="mt-1 ml-3 border-l border-slate-200 dark:border-slate-700 pl-3 space-y-1.5">
                {node.children.map((child) => (
                  <li key={`${node.task.task_id}-${child.task_id}`}>
                    <div className="text-slate-700 dark:text-slate-200">
                      {child.kind === 'fix' ? `${t('pulse.fix')}: ` : ''}
                      {child.title || child.task_id}
                    </div>
                    <RelatedLinks taskIds={child.related_tasks} docs={child.related_docs} titles={titles} />
                  </li>
                ))}
              </ul>
            ) : null}
          </li>
        ))}
      </ul>
    </div>
  )
}

function buildForest(tasks: TaskItem[]): ForestNode[] {
  const byId = new Map<string, TaskItem>()
  for (const task of tasks) {
    if (task.task_id) byId.set(task.task_id, task)
  }
  const children = new Map<string, TaskItem[]>()
  const childIds = new Set<string>()
  for (const task of tasks) {
    const parents = (task.related_tasks || []).filter((id) => byId.has(id) && id !== task.task_id)
    if (parents.length === 0) continue
    childIds.add(task.task_id)
    for (const parentId of parents) {
      const list = children.get(parentId) || []
      list.push(task)
      children.set(parentId, list)
    }
  }
  const roots: ForestNode[] = []
  const seen = new Set<string>()
  for (const task of tasks) {
    const kids = children.get(task.task_id) || []
    const isRoot = kids.length > 0 && !childIds.has(task.task_id)
    if (!isRoot) continue
    if (seen.has(task.task_id)) continue
    seen.add(task.task_id)
    roots.push({ task, children: kids })
  }
  return roots
}
