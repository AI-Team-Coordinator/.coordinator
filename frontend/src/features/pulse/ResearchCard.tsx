import { Clock } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Card } from '../../shared/ui/Card'
import { Badge } from '../../shared/ui/Badge'
import { formatDuration } from '../../shared/lib/formatters'
import { hasUsageSpend, SharedQuotaIcon, UsageSpend } from '../../shared/ui/UsageSpend'
import { ChatTabs } from './ChatTabs'
import type { MemberState } from '../../shared/types/api'

interface ResearchCardProps {
  member: MemberState
}

export function ResearchCard({ member }: ResearchCardProps) {
  const { t } = useTranslation()
  const roleLabel = member.role ? t(`pulse.roles.${member.role}`, { defaultValue: member.role }) : ''
  const summary = member.research?.summary || ''

  return (
    <Card className="p-5 transition-all duration-200 border-indigo-400/60 dark:border-indigo-500/40 bg-white/90 dark:bg-slate-900/90 shadow-md shadow-indigo-500/5">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <h3
            className="font-semibold text-slate-900 dark:text-white text-sm leading-snug line-clamp-2"
            title={summary || member.name}
          >
            {summary || t('pulse.research')}
          </h3>
        </div>
        <div className="flex items-center gap-1.5 shrink-0">
          <Badge variant="default" className="font-bold text-[10px]">
            ● {t('pulse.research')}
          </Badge>
          {member.research?.clock_paused ? (
            <Clock
              className="w-3.5 h-3.5 text-amber-500 dark:text-amber-400"
              aria-label={t('pulse.clockPaused')}
              title={t('pulse.clockPaused')}
            />
          ) : null}
        </div>
      </div>

      <div className="mt-2 flex items-center space-x-2">
        <div className="w-7 h-7 rounded-full bg-indigo-50 dark:bg-slate-800 border border-indigo-200 dark:border-slate-700 flex items-center justify-center font-bold text-[11px] text-indigo-600 dark:text-indigo-300 shrink-0">
          {member.alias}
        </div>
        <div className="text-xs text-slate-500 dark:text-slate-400 truncate">
          {member.name}
          {' · '}@{member.alias}
          {roleLabel ? ` · ${roleLabel}` : ''}
        </div>
      </div>

      {member.research?.chat ? (
        <div className="mt-3">
          <ChatTabs chats={[member.research.chat]} />
        </div>
      ) : null}

      <div className="mt-4 pt-4 border-t border-slate-100 dark:border-slate-800/80 space-y-2">
        <div className="flex items-center justify-between text-xs">
          <span className="text-slate-500 dark:text-slate-400">{t('pulse.duration')}:</span>
          <span className="inline-flex items-center gap-1.5">
            <span
              className={`font-mono font-semibold ${
                member.research?.clock_paused
                  ? 'text-amber-600 dark:text-amber-400'
                  : 'text-indigo-600 dark:text-indigo-300'
              }`}
            >
              {formatDuration(member.research?.duration_seconds)}
            </span>
            {member.research?.spend_shared ? <SharedQuotaIcon label={t('usage.sharedResearch')} /> : null}
          </span>
        </div>
        {hasUsageSpend({
          budgetUsd: member.research?.budget_usd,
          ondemandUsd: member.research?.ondemand_usd,
          cursorModelsPct: member.research?.cursor_models_pct,
          otherModelsPct: member.research?.other_models_pct,
          shared: member.research?.spend_shared,
        }) ? (
          <UsageSpend
            budgetUsd={member.research?.budget_usd}
            ondemandUsd={member.research?.ondemand_usd}
            cursorModelsPct={member.research?.cursor_models_pct}
            otherModelsPct={member.research?.other_models_pct}
            shared={member.research?.spend_shared}
          />
        ) : null}
        <p className="text-[11px] text-slate-400 dark:text-slate-500">{t('pulse.researchNoBranch')}</p>
      </div>
    </Card>
  )
}
