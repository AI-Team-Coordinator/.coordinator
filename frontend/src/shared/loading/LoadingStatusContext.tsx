import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from 'react'

export type LoadingActivity = {
  id: string
  message: string
}

type LoadingStatusContextValue = {
  activities: LoadingActivity[]
  track: <T>(id: string, message: string, work: () => Promise<T>) => Promise<T>
  start: (id: string, message: string) => void
  end: (id: string) => void
}

const LoadingStatusContext = createContext<LoadingStatusContextValue | null>(null)

export function LoadingStatusProvider({ children }: { children: ReactNode }) {
  const [activities, setActivities] = useState<LoadingActivity[]>([])
  const start = useCallback((id: string, message: string) => {
    setActivities((prev) => {
      const rest = prev.filter((a) => a.id !== id)
      return [...rest, { id, message }]
    })
  }, [])

  const end = useCallback((id: string) => {
    setActivities((prev) => prev.filter((a) => a.id !== id))
  }, [])

  const track = useCallback(
    async <T,>(id: string, message: string, work: () => Promise<T>): Promise<T> => {
      start(id, message)
      try {
        return await work()
      } finally {
        end(id)
      }
    },
    [end, start]
  )

  const value = useMemo(
    () => ({ activities, track, start, end }),
    [activities, track, start, end]
  )

  return <LoadingStatusContext.Provider value={value}>{children}</LoadingStatusContext.Provider>
}

export function useLoadingStatus() {
  const ctx = useContext(LoadingStatusContext)
  if (!ctx) {
    throw new Error('useLoadingStatus must be used within LoadingStatusProvider')
  }
  return ctx
}
