import { BarChart3, Boxes, LayoutDashboard, Users } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { CoordinatorLogo } from '../../shared/ui/Logo'
import { cn } from '../../shared/lib/utils'
import { SECTIONS, type SectionId } from './sections'

const ICONS = {
  overview: LayoutDashboard,
  team: Users,
  project: Boxes,
  stats: BarChart3,
} as const

interface SectionNavProps {
  section: SectionId
  onChange: (section: SectionId) => void
  teamDirty?: boolean
}

export function SectionNav({ section, onChange, teamDirty }: SectionNavProps) {
  const { t } = useTranslation()

  return (
    <nav className="flex md:flex-col items-center justify-around md:justify-start gap-1 md:gap-2 p-2 md:p-3">
      <div className="hidden md:flex pb-2 mb-1 border-b border-slate-200 dark:border-slate-800">
        <CoordinatorLogo size={40} />
      </div>
      {SECTIONS.map((id) => {
        const Icon = ICONS[id]
        const active = section === id
        return (
          <button
            key={id}
            type="button"
            onClick={() => onChange(id)}
            className={cn(
              'relative w-16 h-16 rounded-xl flex flex-col items-center justify-center gap-1 transition-colors',
              active
                ? 'bg-indigo-600 text-white shadow-sm shadow-indigo-500/30'
                : 'text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 hover:text-slate-800 dark:hover:text-slate-200'
            )}
          >
            <Icon className="w-5 h-5" strokeWidth={1.75} />
            <span className="text-[10px] leading-none font-medium">{t(`sections.${id}`)}</span>
            {id === 'team' && teamDirty && (
              <span className="absolute top-1.5 right-1.5 w-1.5 h-1.5 rounded-full bg-amber-400" />
            )}
          </button>
        )
      })}
    </nav>
  )
}
