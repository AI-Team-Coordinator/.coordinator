import { forwardRef, type ButtonHTMLAttributes } from 'react'
import { cn } from '../lib/utils'

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'default' | 'outline' | 'ghost' | 'secondary'
  size?: 'sm' | 'md' | 'icon'
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'default', size = 'md', ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={cn(
          'inline-flex items-center justify-center font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500 disabled:pointer-events-none disabled:opacity-50 rounded-lg',
          {
            'bg-indigo-600 text-white hover:bg-indigo-500 shadow-sm': variant === 'default',
            'border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800':
              variant === 'outline',
            'text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800':
              variant === 'ghost',
            'bg-slate-100 dark:bg-slate-800 text-slate-900 dark:text-slate-100 hover:bg-slate-200 dark:hover:bg-slate-700':
              variant === 'secondary',
            'text-xs px-2.5 py-1.5 h-8': size === 'sm',
            'text-sm px-3.5 py-2 h-9': size === 'md',
            'h-9 w-9 p-0': size === 'icon',
          },
          className
        )}
        {...props}
      />
    )
  }
)
Button.displayName = 'Button'
