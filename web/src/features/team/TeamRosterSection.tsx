import { useEffect, useMemo, useState } from 'react'
import { Plus, Save, Trash2, Users } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { api } from '../../shared/api/client'
import { parseJSON } from '../../shared/api/http'
import { Button } from '../../shared/ui/Button'
import { Card } from '../../shared/ui/Card'
import type { ServiceNode, TeamPerson } from '../../shared/types/api'

const ROLES = ['founder', 'engineer'] as const

type DraftMember = {
  key: string
  alias: string
  name: string
  role: string
  access: string
  focus: string[]
}

interface TeamRosterSectionProps {
  services: ServiceNode[]
  onChanged: () => void
  onDirtyChange: (dirty: boolean) => void
}

export function TeamRosterSection({ services, onChanged, onDirtyChange }: TeamRosterSectionProps) {
  const { t } = useTranslation()
  const [draft, setDraft] = useState<DraftMember[]>([])
  const [saved, setSaved] = useState<DraftMember[]>([])
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const load = async () => {
    const teamRes = await fetch(api.team)
    const team = await parseJSON<{ members: TeamPerson[] }>(teamRes)
    const next = (team.members || []).map(toDraft)
    setDraft(next)
    setSaved(next)
  }

  useEffect(() => {
    load().catch((err: Error) => setError(err.message))
  }, [])

  const dirtyForm = useMemo(() => JSON.stringify(draft) !== JSON.stringify(saved), [draft, saved])

  useEffect(() => {
    onDirtyChange(dirtyForm)
  }, [dirtyForm, onDirtyChange])

  const updateMember = (key: string, patch: Partial<DraftMember>) => {
    setDraft((prev) => prev.map((m) => (m.key === key ? { ...m, ...patch } : m)))
    setMessage(null)
    setError(null)
  }

  const addMember = () => {
    setDraft((prev) => [
      ...prev,
      { key: `new-${Date.now()}`, alias: '', name: '', role: '', access: 'member', focus: [] },
    ])
    setMessage(null)
  }

  const removeMember = (key: string) => {
    if (draft.length <= 1) return
    setDraft((prev) => prev.filter((m) => m.key !== key))
    setMessage(null)
  }

  const toggleFocus = (key: string, serviceId: string) => {
    setDraft((prev) =>
      prev.map((m) => {
        if (m.key !== key) return m
        const has = m.focus.includes(serviceId)
        return {
          ...m,
          focus: has ? m.focus.filter((id) => id !== serviceId) : [...m.focus, serviceId],
        }
      })
    )
    setMessage(null)
  }

  const save = async () => {
    setSaving(true)
    setError(null)
    setMessage(null)
    try {
      const res = await fetch(api.team, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          members: draft.map((m) => ({
            alias: m.alias.trim().toUpperCase(),
            name: m.name.trim(),
            role: m.role,
            access: m.access,
            focus: m.focus,
          })),
        }),
      })
      const team = await parseJSON<{ members: TeamPerson[]; pushed?: boolean; warning?: string }>(res)
      const next = (team.members || []).map(toDraft)
      setDraft(next)
      setSaved(next)
      setMessage(team.pushed ? t('team.savedPushed') : t('team.savedLocal', { warning: team.warning || '' }))
      onChanged()
    } catch (err) {
      setError(err instanceof Error ? err.message : t('team.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <section className="space-y-4">
      <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center gap-2">
            <Users className="w-4 h-4 text-indigo-500" />
            {t('team.title')}
          </h2>
          <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">{t('team.subtitle')}</p>
        </div>
        <div className="flex items-center gap-2 self-end">
          <Button variant="outline" size="sm" onClick={addMember} className="gap-1.5">
            <Plus className="w-3.5 h-3.5" />
            {t('team.add')}
          </Button>
          <Button size="sm" onClick={save} disabled={saving || !dirtyForm} className="gap-1.5">
            <Save className="w-3.5 h-3.5" />
            {t('team.save')}
          </Button>
        </div>
      </div>

      {(message || error) && (
        <div
          className={`text-xs rounded-lg px-3 py-2 border ${
            error
              ? 'border-rose-200 dark:border-rose-900 text-rose-700 dark:text-rose-300 bg-rose-50 dark:bg-rose-950/40'
              : 'border-slate-200 dark:border-slate-800 text-slate-600 dark:text-slate-300 bg-slate-50 dark:bg-slate-900/60'
          }`}
        >
          {error || message}
        </div>
      )}

      <div className="space-y-3">
        {draft.map((member) => (
          <Card key={member.key} className="p-4">
            <div className="grid grid-cols-1 md:grid-cols-12 gap-3 items-start">
              <label className="md:col-span-2 block">
                <span className="text-[11px] uppercase tracking-wide text-slate-500 dark:text-slate-400">
                  {t('team.alias')}
                </span>
                <input
                  value={member.alias}
                  onChange={(e) =>
                    updateMember(member.key, {
                      alias: e.target.value.toUpperCase().replace(/[^A-Z]/g, '').slice(0, 4),
                    })
                  }
                  placeholder="EK"
                  className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm font-mono"
                />
              </label>
              <label className="md:col-span-5 block">
                <span className="text-[11px] uppercase tracking-wide text-slate-500 dark:text-slate-400">
                  {t('team.name')}
                </span>
                <input
                  value={member.name}
                  onChange={(e) => updateMember(member.key, { name: e.target.value })}
                  className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm"
                />
              </label>
              <label className="md:col-span-4 block">
                <span className="text-[11px] uppercase tracking-wide text-slate-500 dark:text-slate-400">
                  {t('team.role')}
                </span>
                <select
                  value={member.role}
                  onChange={(e) => updateMember(member.key, { role: e.target.value })}
                  className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm"
                >
                  <option value="">{t('team.roleNone')}</option>
                  {ROLES.map((role) => (
                    <option key={role} value={role}>
                      {t(`pulse.roles.${role}`)}
                    </option>
                  ))}
                </select>
              </label>
              <div className="md:col-span-1 flex md:justify-end md:pt-6">
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => removeMember(member.key)}
                  disabled={draft.length <= 1}
                  title={t('team.remove')}
                  className="text-slate-400 hover:text-rose-600"
                >
                  <Trash2 className="w-4 h-4" />
                </Button>
              </div>
            </div>

            <div className="mt-3">
              <div className="text-[11px] uppercase tracking-wide text-slate-500 dark:text-slate-400 mb-1.5">
                {t('team.focus')}
              </div>
              <div className="flex flex-wrap gap-1.5">
                {services.map((svc) => {
                  const active = member.focus.includes(svc.id)
                  return (
                    <button
                      key={svc.id}
                      type="button"
                      onClick={() => toggleFocus(member.key, svc.id)}
                      className={`text-[11px] px-2 py-1 rounded-full border transition ${
                        active
                          ? 'bg-indigo-50 dark:bg-indigo-950/60 text-indigo-700 dark:text-indigo-300 border-indigo-200 dark:border-indigo-800'
                          : 'bg-transparent text-slate-500 dark:text-slate-400 border-slate-200 dark:border-slate-700 hover:border-slate-400'
                      }`}
                    >
                      {svc.name}
                    </button>
                  )
                })}
              </div>
            </div>
          </Card>
        ))}
      </div>
    </section>
  )
}

function toDraft(person: TeamPerson): DraftMember {
  return {
    key: person.alias,
    alias: person.alias,
    name: person.name,
    role: person.role || '',
    access: person.access || '',
    focus: person.focus || [],
  }
}
