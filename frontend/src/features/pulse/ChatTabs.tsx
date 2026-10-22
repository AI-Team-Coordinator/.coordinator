import { MessageSquare } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import type { ChatTab } from '../../shared/types/api'

interface ChatTabsProps {
  chats?: ChatTab[]
}

export function ChatTabs({ chats }: ChatTabsProps) {
  const { t } = useTranslation()
  if (!chats || chats.length === 0) {
    return null
  }
  const label = chats.length > 1 ? t('pulse.cursorTabs') : t('pulse.cursorTab')

  return (
    <div>
      <div className="text-[10px] font-medium uppercase tracking-wider text-violet-500 dark:text-violet-400/80 mb-2">
        {label}
      </div>
      <div className="flex flex-wrap gap-1.5">
        {chats.map((chat, index) => {
          const name = (chat.title || '').trim() || t('pulse.unnamedChat')
          return (
            <span
              key={chat.session_id || `${name}-${index}`}
              className="inline-flex max-w-full items-center gap-1.5 rounded-md border border-violet-200 dark:border-violet-800/70 bg-violet-50 dark:bg-violet-950/60 px-2.5 py-1.5 text-[11px] font-medium leading-snug text-violet-900 dark:text-violet-100"
              title={name}
            >
              <MessageSquare className="w-3 h-3 shrink-0 text-violet-500 dark:text-violet-300" />
              <span className="truncate">{name}</span>
            </span>
          )
        })}
      </div>
    </div>
  )
}
