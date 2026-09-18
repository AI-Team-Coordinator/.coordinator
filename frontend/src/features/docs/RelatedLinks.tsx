import { TaskDocLink } from './TaskDocLink'

const TASK_DOC_FILE = /(?:^|\/)((?:FIX-)?\d{8}-\d{4}-[A-Z]{2,4}-[A-Z0-9_]+)\.md$/i

export function taskIdFromDocPath(path: string): string | undefined {
  const match = path.trim().match(TASK_DOC_FILE)
  return match?.[1]
}

function labelOf(id: string, titles?: Record<string, string>) {
  return titles?.[id] || id
}

interface RelatedLinksProps {
  taskIds?: string[]
  docs?: string[]
  titles?: Record<string, string>
}

export function RelatedLinks({ taskIds, docs, titles }: RelatedLinksProps) {
  const tasks = (taskIds || []).filter(Boolean)
  const files = (docs || []).filter(Boolean)
  if (tasks.length === 0 && files.length === 0) {
    return null
  }
  return (
    <div className="mt-1.5 space-y-1">
      {tasks.map((id) => (
        <div key={`t-${id}`} className="min-w-0">
          <TaskDocLink taskId={id} className="text-[11px] text-slate-500 dark:text-slate-400">
            {labelOf(id, titles)}
          </TaskDocLink>
        </div>
      ))}
      {files.map((path) => {
        const id = taskIdFromDocPath(path)
        const name = path.replace(/^docs\//, '')
        return (
          <div key={`d-${path}`} className="min-w-0">
            <TaskDocLink taskId={id} className="text-[11px] text-slate-500 dark:text-slate-400">
              {name}
            </TaskDocLink>
          </div>
        )
      })}
    </div>
  )
}
