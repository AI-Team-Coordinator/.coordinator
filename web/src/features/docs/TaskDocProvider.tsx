import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from 'react'
import { TaskDocModal } from './TaskDocModal'

type TaskDocContextValue = {
  openDoc: (taskId: string) => void
}

const TaskDocContext = createContext<TaskDocContextValue>({
  openDoc: () => {},
})

export function useTaskDoc() {
  return useContext(TaskDocContext)
}

export function TaskDocProvider({ children }: { children: ReactNode }) {
  const [taskId, setTaskId] = useState<string | null>(null)
  const openDoc = useCallback((id: string) => {
    const next = id.trim()
    if (next) setTaskId(next)
  }, [])
  const value = useMemo(() => ({ openDoc }), [openDoc])

  return (
    <TaskDocContext.Provider value={value}>
      {children}
      <TaskDocModal taskId={taskId} onClose={() => setTaskId(null)} />
    </TaskDocContext.Provider>
  )
}
