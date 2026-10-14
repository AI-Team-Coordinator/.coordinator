import { Component, type ErrorInfo, type ReactNode } from 'react'

interface Props {
  children: ReactNode
}

interface State {
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State {
    return { error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('Coordinator UI error:', error, info.componentStack)
  }

  render() {
    if (this.state.error) {
      return (
        <div className="min-h-screen flex items-center justify-center p-6 bg-slate-50 dark:bg-slate-950">
          <div className="max-w-md space-y-3 text-sm text-slate-700 dark:text-slate-300">
            <p className="font-semibold text-slate-900 dark:text-white">Something went wrong</p>
            <p className="text-xs text-slate-500 dark:text-slate-400">
              If you use Cursor&apos;s built-in browser, try a hard reload or open{' '}
              <code className="text-indigo-600 dark:text-indigo-400">http://127.0.0.1:5175</code>{' '}
              in Chrome or Safari.
            </p>
            <pre className="text-xs overflow-auto rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 p-3 text-rose-700 dark:text-rose-300">
              {this.state.error.message}
            </pre>
            <button
              type="button"
              className="text-xs text-indigo-600 dark:text-indigo-400 hover:underline"
              onClick={() => window.location.reload()}
            >
              Reload page
            </button>
          </div>
        </div>
      )
    }
    return this.props.children
  }
}
