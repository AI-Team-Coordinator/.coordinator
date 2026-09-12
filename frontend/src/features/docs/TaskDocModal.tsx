import { useEffect, useState } from 'react'
import { X } from 'lucide-react'
import Markdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { useTranslation } from 'react-i18next'
import { api } from '../../shared/api/client'
import { parseJSON } from '../../shared/api/http'
import { markdownComponents } from './markdownComponents'

export interface TaskDocPayload {
  task_id: string
  title: string
  rel_path: string
  markdown: string
}

interface TaskDocModalProps {
  taskId: string | null
  onClose: () => void
}

export function TaskDocModal({ taskId, onClose }: TaskDocModalProps) {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [doc, setDoc] = useState<TaskDocPayload | null>(null)

  useEffect(() => {
    if (!taskId) {
      setDoc(null)
      setError(null)
      setLoading(false)
      return
    }
    let cancelled = false
    setLoading(true)
    setError(null)
    setDoc(null)
    fetch(`${api.docs}/${encodeURIComponent(taskId)}`)
      .then((res) => parseJSON<TaskDocPayload>(res))
      .then((data) => {
        if (!cancelled) setDoc(data)
      })
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : t('docs.missing'))
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [taskId, t])

  useEffect(() => {
    if (!taskId) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKey)
    const prev = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => {
      window.removeEventListener('keydown', onKey)
      document.body.style.overflow = prev
    }
  }, [taskId, onClose])

  if (!taskId) return null

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center p-4 sm:p-8">
      <button type="button" className="absolute inset-0 bg-slate-950/50 dark:bg-black/60" aria-label={t('docs.close')} onClick={onClose} />
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="task-doc-title"
        className="relative z-10 w-full max-w-3xl max-h-[88vh] flex flex-col rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-xl"
      >
        <div className="flex items-start justify-between gap-3 px-4 py-3 border-b border-slate-100 dark:border-slate-800">
          <div className="min-w-0">
            <div id="task-doc-title" className="font-mono text-sm text-slate-900 dark:text-slate-100 truncate">
              {taskId}
            </div>
            {doc?.rel_path && (
              <div className="text-[11px] text-slate-500 dark:text-slate-400 truncate mt-0.5">{doc.rel_path}</div>
            )}
          </div>
          <button
            type="button"
            onClick={onClose}
            className="shrink-0 p-1 rounded-md text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800"
            aria-label={t('docs.close')}
          >
            <X className="w-4 h-4" />
          </button>
        </div>
        <div className="overflow-y-auto px-5 py-4 text-sm text-slate-800 dark:text-slate-200">
          {loading && <p className="text-slate-500 dark:text-slate-400">{t('docs.loading')}</p>}
          {!loading && error && <p className="text-slate-500 dark:text-slate-400">{t('docs.missing')}</p>}
          {!loading && doc && (
            <Markdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
              {doc.markdown}
            </Markdown>
          )}
        </div>
      </div>
    </div>
  )
}
