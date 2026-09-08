import React from 'react'
import ReactDOM from 'react-dom/client'
import { App } from './app/App'
import { ThemeProvider } from './shared/theme/theme-context'
import { TaskDocProvider } from './features/docs/TaskDocProvider'
import './shared/i18n/config'
import './index.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ThemeProvider>
      <TaskDocProvider>
        <App />
      </TaskDocProvider>
    </ThemeProvider>
  </React.StrictMode>
)
