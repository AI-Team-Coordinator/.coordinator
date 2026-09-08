import type { ButtonHTMLAttributes } from 'react'
import { cn } from '../../shared/lib/utils'
import { useTaskDoc } from './TaskDocProvider'

interface TaskDocLinkProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  taskId?: string
}

export function TaskDocLink({ taskId, className, children, ...props }: TaskDocLinkProps) {
  const { openDoc } = useTaskDoc()
  if (!taskId) {
    return <span className={className}>{children}</span>
  }
  return (
    <button
      type="button"
      onClick={(e) => {
        e.stopPropagation()
        openDoc(taskId)
      }}
      title={taskId}
      className={cn(
        'font-mono truncate text-left min-w-0 hover:text-indigo-600 dark:hover:text-indigo-300 hover:underline underline-offset-2',
        className
      )}
      {...props}
    >
      {children ?? taskId}
    </button>
  )
}
