import { Info } from 'lucide-react'
import { cn } from '../lib/utils'

interface InfoTooltipProps {
  text: string
  className?: string
}

export function InfoTooltip({ text, className }: InfoTooltipProps) {
  return (
    <button
      type="button"
      aria-label={text}
      className={cn(
        'relative inline-flex items-center justify-center rounded-sm text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500 group/tip shrink-0',
        className
      )}
    >
      <Info className="w-3.5 h-3.5" aria-hidden />
      <span
        role="tooltip"
        className="pointer-events-none absolute z-40 right-0 top-full mt-1.5 w-64 rounded-md border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 px-2.5 py-2 text-left text-[11px] font-normal normal-case tracking-normal leading-snug text-slate-600 dark:text-slate-300 shadow-lg opacity-0 group-hover/tip:opacity-100 group-focus/tip:opacity-100 transition-opacity"
      >
        {text}
      </span>
    </button>
  )
}
