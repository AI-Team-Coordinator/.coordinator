export function externalEventClassStyles(className: string): string {
  switch (className) {
    case 'warning':
      return 'border-amber-200 dark:border-amber-500/40 bg-amber-50 dark:bg-amber-950/30'
    case 'reminder':
      return 'border-sky-200 dark:border-sky-500/40 bg-sky-50 dark:bg-sky-950/30'
    default:
      return 'border-slate-200 dark:border-slate-600 bg-slate-50 dark:bg-slate-900/40'
  }
}
