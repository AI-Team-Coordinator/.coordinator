import { ChevronDown } from 'lucide-react'
import type { SelectHTMLAttributes } from 'react'
import { cn } from '../lib/utils'

const fieldBase =
  'appearance-none bg-white dark:bg-slate-950 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200 pr-9'

const densities = {
  sm: 'px-2 py-1 rounded-md text-xs',
  md: 'px-2.5 py-1.5 rounded-lg text-sm',
}

export interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  /** Wrapper (positioning for chevron). */
  className?: string
  /** Extra classes on the native `<select>`. */
  selectClassName?: string
  density?: 'sm' | 'md'
}

export function Select({
  className,
  selectClassName,
  density = 'md',
  disabled,
  ...props
}: SelectProps) {
  return (
    <div className={cn('relative w-full', disabled && 'opacity-60', className)}>
      <select
        disabled={disabled}
        className={cn('w-full', fieldBase, densities[density], selectClassName)}
        {...props}
      />
      <ChevronDown
        aria-hidden
        className="pointer-events-none absolute right-2 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400 dark:text-slate-500"
      />
    </div>
  )
}
