import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../../shared/api/client'
import { Button } from '../../shared/ui/Button'
import { Card } from '../../shared/ui/Card'
import { Select } from '../../shared/ui/Select'
import type { TeamPerson } from '../../shared/types/api'

type IssuedToken = {
  id: string
  alias: string
  label?: string
  created_at: string
  created_by: string
}

interface RelayTokensAdminProps {
  currentAlias?: string
  onUseSecret?: (secret: string) => void
}

export function RelayTokensAdmin({ currentAlias, onUseSecret }: RelayTokensAdminProps) {
  const { t } = useTranslation()
  const [tokens, setTokens] = useState<IssuedToken[]>([])
  const [team, setTeam] = useState<TeamPerson[]>([])
  const [alias, setAlias] = useState(currentAlias || '')
  const [label, setLabel] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [freshSecret, setFreshSecret] = useState('')

  const load = () => {
    fetch(api.relayTokens)
      .then((r) => (r.ok ? r.json() : null))
      .then((data: { tokens?: IssuedToken[] } | null) => setTokens(data?.tokens ?? []))
      .catch(() => setTokens([]))
  }

  useEffect(() => {
    load()
    fetch(api.team)
      .then((r) => (r.ok ? r.json() : null))
      .then((data: { members?: TeamPerson[] } | null) => {
        const members = data?.members ?? []
        setTeam(members)
        if (!alias && currentAlias) setAlias(currentAlias)
        else if (!alias && members[0]?.alias) setAlias(members[0].alias)
      })
      .catch(() => {})
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const issue = async (forAlias: string) => {
    setBusy(true)
    setError('')
    setFreshSecret('')
    try {
      const res = await fetch(api.relayTokens, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ alias: forAlias, label: label.trim() }),
      })
      const data = await res.json().catch(() => ({}))
      if (!res.ok) {
        throw new Error(typeof data.message === 'string' ? data.message : t('relay.tokensIssueError'))
      }
      setFreshSecret((data as { secret?: string }).secret || '')
      load()
    } catch (e) {
      setError(e instanceof Error ? e.message : t('relay.tokensIssueError'))
    } finally {
      setBusy(false)
    }
  }

  const revoke = async (id: string) => {
    setBusy(true)
    setError('')
    try {
      const res = await fetch(`${api.relayTokens}/${id}`, { method: 'DELETE' })
      if (!res.ok) throw new Error('revoke')
      const data = (await res.json()) as { tokens?: IssuedToken[] }
      setTokens(data.tokens ?? [])
    } catch {
      setError(t('relay.tokensRevokeError'))
    } finally {
      setBusy(false)
    }
  }

  return (
    <Card className="p-4 space-y-3">
      <div>
        <h3 className="text-sm font-medium text-slate-800 dark:text-slate-200">{t('relay.tokensTitle')}</h3>
        <p className="text-xs text-slate-500 dark:text-slate-400 mt-1">{t('relay.tokensHelp')}</p>
      </div>
      {error && <p className="text-xs text-rose-600 dark:text-rose-400">{error}</p>}
      {freshSecret && (
        <div className="rounded-lg border border-amber-200 dark:border-amber-800 bg-amber-50/80 dark:bg-amber-950/30 p-3 space-y-2">
          <p className="text-xs text-amber-900 dark:text-amber-200">{t('relay.tokensCopyOnce')}</p>
          <code className="block text-xs break-all text-slate-800 dark:text-slate-200">{freshSecret}</code>
          <div className="flex flex-wrap gap-2">
            {onUseSecret && (
              <Button size="sm" variant="secondary" onClick={() => onUseSecret(freshSecret)}>
                {t('relay.tokensUseForConnect')}
              </Button>
            )}
            <Button size="sm" variant="ghost" onClick={() => setFreshSecret('')}>
              {t('relay.tokensDismissSecret')}
            </Button>
          </div>
        </div>
      )}
      <div className="flex flex-wrap items-end gap-3">
        <label className="block min-w-[8rem]">
          <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('relay.tokensMember')}</span>
          <Select className="mt-1" value={alias} onChange={(e) => setAlias(e.target.value)} disabled={busy}>
            {team.map((m) => (
              <option key={m.alias} value={m.alias}>
                {m.name || m.alias} ({m.alias})
              </option>
            ))}
          </Select>
        </label>
        <label className="block min-w-[10rem] flex-1">
          <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('relay.tokensLabel')}</span>
          <input
            value={label}
            onChange={(e) => setLabel(e.target.value)}
            disabled={busy}
            className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm"
          />
        </label>
        <Button size="sm" disabled={busy || !alias} onClick={() => void issue(alias)}>
          {t('relay.tokensIssue')}
        </Button>
      </div>
      {tokens.length === 0 ? (
        <p className="text-xs text-slate-500">{t('relay.tokensEmpty')}</p>
      ) : (
        <ul className="divide-y divide-slate-200 dark:divide-slate-800 text-sm">
          {tokens.map((tok) => (
            <li key={tok.id} className="py-2 flex items-center justify-between gap-3">
              <div className="min-w-0">
                <span className="font-medium text-slate-800 dark:text-slate-200">{tok.alias}</span>
                {tok.label && <span className="text-slate-500"> · {tok.label}</span>}
                <p className="text-[11px] text-slate-400 truncate">
                  {t('relay.tokensMeta', { by: tok.created_by, at: tok.created_at })}
                </p>
              </div>
              <Button size="sm" variant="ghost" disabled={busy} onClick={() => void revoke(tok.id)}>
                {t('relay.tokensRevoke')}
              </Button>
            </li>
          ))}
        </ul>
      )}
    </Card>
  )
}
