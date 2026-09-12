import { createContext, useEffect, useState, type ReactNode } from 'react'
import type { Theme, ThemeContextValue } from './theme-types'
import {
  applyTheme,
  getStoredTheme,
  getSystemTheme,
  setStoredTheme,
  watchSystemTheme,
} from './theme-utils'

export const ThemeContext = createContext<ThemeContextValue | null>(null)

interface ThemeProviderProps {
  children: ReactNode
  defaultTheme?: Theme
}

export function ThemeProvider({ children, defaultTheme }: ThemeProviderProps) {
  const [theme, setThemeState] = useState<Theme>(() => defaultTheme || getStoredTheme())
  const [effectiveTheme, setEffectiveTheme] = useState<'light' | 'dark'>(() => {
    const initial = defaultTheme || getStoredTheme()
    return initial === 'system' ? getSystemTheme() : initial
  })

  useEffect(() => {
    applyTheme(theme)
    setEffectiveTheme(theme === 'system' ? getSystemTheme() : theme)
  }, [theme])

  useEffect(() => {
    if (theme !== 'system') return
    const unsubscribe = watchSystemTheme((systemTheme) => {
      setEffectiveTheme(systemTheme)
      applyTheme('system')
    })
    return unsubscribe
  }, [theme])

  const setTheme = (newTheme: Theme) => {
    setThemeState(newTheme)
    setStoredTheme(newTheme)
    applyTheme(newTheme)
    setEffectiveTheme(newTheme === 'system' ? getSystemTheme() : newTheme)
  }

  return (
    <ThemeContext.Provider value={{ theme, effectiveTheme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  )
}
