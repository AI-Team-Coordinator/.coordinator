import { Languages } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '../lib/utils'

export function LanguageSwitcher() {
  const { i18n } = useTranslation()

  const currentLang = i18n.language && i18n.language.startsWith('ru') ? 'ru' : 'en'
  const nextLang = currentLang === 'en' ? 'ru' : 'en'

  const toggleLanguage = () => {
    i18n.changeLanguage(nextLang)
  }

  return (
    <button
      onClick={toggleLanguage}
      title={nextLang === 'ru' ? 'Переключить на русский' : 'Switch to English'}
      className={cn(
        'flex items-center space-x-1 px-2.5 py-1.5 rounded-lg border text-xs font-semibold uppercase tracking-wider transition-colors',
        'bg-slate-100 dark:bg-slate-800/80 border-slate-200 dark:border-slate-700/60',
        'text-slate-700 dark:text-slate-300 hover:text-indigo-600 dark:hover:text-indigo-400'
      )}
    >
      <Languages className="w-3.5 h-3.5" />
      <span>{nextLang.toUpperCase()}</span>
    </button>
  )
}
