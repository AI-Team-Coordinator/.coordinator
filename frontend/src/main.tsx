import React from 'react'
import ReactDOM from 'react-dom/client'
import { App } from './app/App'
import { ThemeProvider } from './shared/theme/theme-context'
import { TaskDocProvider } from './features/docs/TaskDocProvider'
import { LoadingStatusProvider } from './shared/loading/LoadingStatusContext'
import { ErrorBoundary } from './shared/ui/ErrorBoundary'
import './shared/i18n/config'
import './index.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ErrorBoundary>
      <ThemeProvider>
        <TaskDocProvider>
          <LoadingStatusProvider>
            <App />
          </LoadingStatusProvider>
        </TaskDocProvider>
      </ThemeProvider>
    </ErrorBoundary>
  </React.StrictMode>
)
