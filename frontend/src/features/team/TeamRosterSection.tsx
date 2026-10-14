import { useEffect, useMemo, useState } from 'react'
import { ChevronDown, Plus, Save, Trash2, Users } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { api } from '../../shared/api/client'
import { parseJSON } from '../../shared/api/http'
import { Button } from '../../shared/ui/Button'
import { Card } from '../../shared/ui/Card'
import { Select } from '../../shared/ui/Select'
import { aliasFromName, sanitizeAlias } from '../../shared/lib/aliasFromName'
import i18n from '../../shared/i18n/config'
import { localized } from '../../shared/lib/localized'
import { cn } from '../../shared/lib/utils'
import type { Crew, ServiceNode, TeamPerson } from '../../shared/types/api'

const ROLES = ['founder', 'engineer'] as const

type DraftMember = {
  key: string
  alias: string
  name: string
  role: string
  access: string
  focus: string[]
  crews: string[]
  primaryCrew: string
  pulseScope: string
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
  const [crewOptions, setCrewOptions] = useState<Crew[]>([])
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const load = async () => {
    const teamRes = await fetch(api.team)
    const team = await parseJSON<{ members: TeamPerson[] }>(teamRes)
    let crews: Crew[] = []
    try {
      const orgRes = await fetch(api.orgTopology)
      const org = await parseJSON<{ crews: Crew[] }>(orgRes)
      crews = org.crews || []
    } catch {
      crews = []
    }
    setCrewOptions(crews)
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
      {
        key: `new-${Date.now()}`,
        alias: '',
        name: '',
        role: '',
        access: 'member',
        focus: [],
        crews: [],
        primaryCrew: '',
        pulseScope: 'member',
      },
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
            crews: m.crews,
            primary_crew: m.primaryCrew,
            pulse_scope: m.pulseScope,
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
          <p className="text-sm text-slate-600 dark:text-slate-300 mt-0.5">{t('team.subtitle')}</p>
          <p className="text-[11px] text-slate-400 dark:text-slate-500 mt-1">{t('team.saveFootnote')}</p>
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
          <TeamMemberCard
            key={member.key}
            member={member}
            services={services}
            crewOptions={crewOptions}
            canRemove={draft.length > 1}
            onUpdate={(patch) => updateMember(member.key, patch)}
            onRemove={() => removeMember(member.key)}
            onToggleFocus={(serviceId) => toggleFocus(member.key, serviceId)}
          />
        ))}
      </div>
    </section>
  )
}

interface TeamMemberCardProps {
  member: DraftMember
  services: ServiceNode[]
  crewOptions: Crew[]
  canRemove: boolean
  onUpdate: (patch: Partial<DraftMember>) => void
  onRemove: () => void
  onToggleFocus: (serviceId: string) => void
}

function TeamMemberCard({
  member,
  services,
  crewOptions,
  canRemove,
  onUpdate,
  onRemove,
  onToggleFocus,
}: TeamMemberCardProps) {
  const { t } = useTranslation()
  const [focusOpen, setFocusOpen] = useState(false)
  const focusCount = member.focus.length

  return (
    <Card className="p-4">
      <div className="grid grid-cols-1 md:grid-cols-12 gap-3 items-start">
        <label className="md:col-span-2 block">
          <span className="text-[11px] uppercase tracking-wide text-slate-500 dark:text-slate-400">
            {t('team.alias')}
          </span>
          <input
            value={member.alias}
            onChange={(e) => onUpdate({ alias: sanitizeAlias(e.target.value) })}
            placeholder="EK"
            className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm font-mono"
          />
          <span className="mt-1 block text-[10px] text-slate-500 dark:text-slate-400">{t('team.aliasHint')}</span>
        </label>
        <label className="md:col-span-5 block">
          <span className="text-[11px] uppercase tracking-wide text-slate-500 dark:text-slate-400">
            {t('team.name')}
          </span>
          <input
            value={member.name}
            onChange={(e) => {
              const nextName = e.target.value
              const prevAuto = aliasFromName(member.name)
              const nextAuto = aliasFromName(nextName)
              const patch: Partial<DraftMember> = { name: nextName }
              if (!member.alias || member.alias === prevAuto) {
                patch.alias = nextAuto
              }
              onUpdate(patch)
            }}
            className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm"
          />
        </label>
        <label className="md:col-span-4 block">
          <span className="text-[11px] uppercase tracking-wide text-slate-500 dark:text-slate-400">
            {t('team.role')}
          </span>
          <Select className="mt-1" value={member.role} onChange={(e) => onUpdate({ role: e.target.value })}>
            <option value="">{t('team.roleNone')}</option>
            {ROLES.map((role) => (
              <option key={role} value={role}>
                {t(`pulse.roles.${role}`)}
              </option>
            ))}
          </Select>
        </label>
        <div className="md:col-span-1 flex md:justify-end md:pt-6">
          <Button
            variant="ghost"
            size="icon"
            onClick={onRemove}
            disabled={!canRemove}
            title={t('team.remove')}
            className="text-slate-400 hover:text-rose-600"
          >
            <Trash2 className="w-4 h-4" />
          </Button>
        </div>
      </div>

      {crewOptions.length > 0 && (
        <div className="mt-3 grid grid-cols-1 md:grid-cols-2 gap-3">
          <div>
            <div className="text-[11px] uppercase tracking-wide text-slate-500 dark:text-slate-400 mb-1.5">
              {t('team.crews')}
            </div>
            <div className="flex flex-wrap gap-1.5">
              {crewOptions.map((crew) => {
                const active = member.crews.includes(crew.id)
                return (
                  <button
                    key={crew.id}
                    type="button"
                    onClick={() => {
                      onUpdate({
                        crews: active
                          ? member.crews.filter((id) => id !== crew.id)
                          : [...member.crews, crew.id],
                      })
                    }}
                    className={`text-[11px] px-2 py-1 rounded-full border transition ${
                      active
                        ? 'bg-violet-50 dark:bg-violet-950/60 text-violet-700 dark:text-violet-300 border-violet-200 dark:border-violet-800'
                        : 'bg-transparent text-slate-500 dark:text-slate-400 border-slate-200 dark:border-slate-700 hover:border-slate-400'
                    }`}
                  >
                    {localized(crew.label, i18n.language) || crew.id}
                  </button>
                )
              })}
            </div>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
            <label className="block">
              <span className="text-[11px] uppercase tracking-wide text-slate-500 dark:text-slate-400">
                {t('team.primaryCrew')}
              </span>
              <Select
                className="mt-1"
                value={member.primaryCrew}
                onChange={(e) => onUpdate({ primaryCrew: e.target.value })}
              >
                <option value="">{t('team.roleNone')}</option>
                {crewOptions.map((crew) => (
                  <option key={crew.id} value={crew.id}>
                    {localized(crew.label, i18n.language) || crew.id}
                  </option>
                ))}
              </Select>
            </label>
            <label className="block">
              <span className="text-[11px] uppercase tracking-wide text-slate-500 dark:text-slate-400">
                {t('team.pulseScope')}
              </span>
              <Select
                className="mt-1"
                value={member.pulseScope || 'member'}
                onChange={(e) => onUpdate({ pulseScope: e.target.value })}
              >
                <option value="member">{t('team.pulseScopeMember')}</option>
                <option value="executive">{t('team.pulseScopeExecutive')}</option>
              </Select>
            </label>
          </div>
        </div>
      )}

      {services.length > 0 && (
        <div className="mt-3 border-t border-slate-100 dark:border-slate-800/80 pt-2">
          <button
            type="button"
            onClick={() => setFocusOpen((v) => !v)}
            className="flex w-full items-center gap-2 rounded-lg py-1.5 text-left text-sm text-slate-700 dark:text-slate-200 hover:bg-slate-50 dark:hover:bg-slate-900/50 px-1 -mx-1"
            aria-expanded={focusOpen}
          >
            <ChevronDown
              className={cn('w-4 h-4 shrink-0 text-slate-400 transition-transform', focusOpen && 'rotate-180')}
            />
            <span className="font-medium">{t('team.focus')}</span>
            <span className="text-xs text-slate-400 dark:text-slate-500 truncate">
              {focusCount > 0 ? t('team.focusSummary', { count: focusCount }) : t('team.focusSummaryEmpty')}
            </span>
          </button>
          {focusOpen && (
            <div className="mt-2 pl-1">
              <p className="text-[11px] text-slate-500 dark:text-slate-400 mb-2">{t('team.focusHint')}</p>
              <div className="flex flex-wrap gap-1.5">
                {services.map((svc) => {
                  const active = member.focus.includes(svc.id)
                  return (
                    <button
                      key={svc.id}
                      type="button"
                      onClick={() => onToggleFocus(svc.id)}
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
          )}
        </div>
      )}
    </Card>
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
    crews: person.crews || [],
    primaryCrew: person.primary_crew || '',
    pulseScope: person.pulse_scope || 'member',
  }
}
