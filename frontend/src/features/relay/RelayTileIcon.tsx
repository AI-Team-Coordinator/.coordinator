import type { ReactNode } from 'react'
import { Brain, Cloud, Github } from 'lucide-react'

type Props = {
  icon: string
  id: string
  /** Card tile: framed badge. Modal header: icon only. */
  variant?: 'tile' | 'plain'
}

export function RelayTileIcon({ icon, id, variant = 'tile' }: Props) {
  const key = (icon || id).toLowerCase()

  let inner: ReactNode
  let frameClass =
    'flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br from-indigo-500 to-violet-600 text-white shadow-sm'

  if (key === 'github') {
    frameClass =
      'flex h-11 w-11 items-center justify-center rounded-xl bg-slate-900 text-white shadow-sm'
    inner = <Github className="w-5 h-5" />
  } else if (key === 'llm' || key === 'llm_model_lifecycle') {
    frameClass =
      'flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br from-emerald-500 to-cyan-600 text-white shadow-sm'
    inner = <Brain className="w-5 h-5" />
  } else {
    inner = <Cloud className="w-5 h-5" />
  }

  if (variant === 'plain') {
    return <>{inner}</>
  }
  return <div className={frameClass}>{inner}</div>
}
