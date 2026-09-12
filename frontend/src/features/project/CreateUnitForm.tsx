import { useMemo, useState } from 'react'
import { Plus } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { api } from '../../shared/api/client'
import { parseJSON } from '../../shared/api/http'
import { Button } from '../../shared/ui/Button'
import { Card } from '../../shared/ui/Card'
import { localized } from '../../shared/lib/localized'
import type { ProjectProfile } from '../../shared/types/api'

const KINDS = ['api', 'worker', 'bot', 'web', 'mobile', 'bridge', 'ops', 'workspace'] as const

interface CreateUnitFormProps {
  profile: ProjectProfile
  onCreated: () => void
}

type Draft = {
  name: string
  group: string
  kind: string
  id: string
  repo: string
  github_repo: string
  purpose_en: string
  purpose_ru: string
}

function deriveFromName(name: string) {
  const words = name
    .trim()
    .split(/[\s_-]+/)
    .filter(Boolean)
  const pascal = words.map((w) => w.charAt(0).toUpperCase() + w.slice(1)).join('')
  const kebab = words
    .map((w) => w.toLowerCase().replace(/[^a-z0-9]+/g, ''))
    .filter(Boolean)
    .join('-')
  const id = kebab.replace(/-/g, '_')
  return { repo: pascal, github_repo: kebab, id }
}

export function CreateUnitForm({ profile, onCreated }: CreateUnitFormProps) {
  const { t, i18n } = useTranslation()
  const canCreate = Boolean(profile.github?.org) && profile.github_live?.auth === 'ok'
  const defaultGroup = profile.groups[0]?.id || ''
  const empty: Draft = {
    name: '',
    group: defaultGroup,
    kind: 'worker',
    id: '',
    repo: '',
    github_repo: '',
    purpose_en: '',
    purpose_ru: '',
  }
  const [open, setOpen] = useState(false)
  const [auto, setAuto] = useState(true)
  const [draft, setDraft] = useState<Draft>(empty)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [message, setMessage] = useState<string | null>(null)

  const org = profile.github?.org || ''

  const hint = useMemo(() => {
    if (!profile.github?.org) return t('project.create.needOrg')
    if (profile.github_live?.auth !== 'ok') return t('project.create.needAuth')
    return t('project.create.subtitle', { org })
  }, [profile.github?.org, profile.github_live?.auth, org, t])

  const update = (patch: Partial<Draft>, fromName = false) => {
    setDraft((prev) => {
      const next = { ...prev, ...patch }
      if (fromName && auto) {
        const derived = deriveFromName(patch.name ?? prev.name)
        next.id = derived.id
        next.repo = derived.repo
        next.github_repo = derived.github_repo
      }
      return next
    })
    setError(null)
    setMessage(null)
  }

  const submit = async () => {
    setSaving(true)
    setError(null)
    setMessage(null)
    try {
      const res = await fetch(api.projectServices, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: draft.id.trim(),
          name: draft.name.trim(),
          group: draft.group,
          kind: draft.kind,
          repo: draft.repo.trim(),
          github_repo: draft.github_repo.trim(),
          purpose: { en: draft.purpose_en.trim(), ru: draft.purpose_ru.trim() },
        }),
      })
      const created = await parseJSON<{ cloned: boolean; html_url: string; pushed?: boolean; warning?: string }>(res)
      if (created.cloned && created.pushed) {
        setMessage(t('project.create.ok', { url: created.html_url }))
      } else if (created.cloned) {
        setMessage(t('project.create.okLocal', { url: created.html_url, warning: created.warning || '' }))
      } else {
        setMessage(t('project.create.okNoClone', { url: created.html_url, warning: created.warning || '' }))
      }
      setDraft({ ...empty, group: defaultGroup })
      setAuto(true)
      onCreated()
    } catch (err) {
      setError(err instanceof Error ? err.message : t('project.create.failed'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <Card className="p-4 bg-white dark:bg-slate-900/70 space-y-3">
      <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3">
        <div>
          <h2 className="text-sm font-semibold text-slate-900 dark:text-white">{t('project.create.title')}</h2>
          <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">{hint}</p>
        </div>
        <Button
          variant={open ? 'outline' : 'default'}
          size="sm"
          onClick={() => setOpen((v) => !v)}
          disabled={!canCreate}
          className="gap-1.5 shrink-0"
        >
          <Plus className="w-3.5 h-3.5" />
          {open ? t('project.create.cancel') : t('project.create.open')}
        </Button>
      </div>

      {open && (
        <div className="space-y-3 pt-1">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            <label className="block md:col-span-2">
              <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('project.create.name')}</span>
              <input
                value={draft.name}
                onChange={(e) => update({ name: e.target.value }, true)}
                placeholder={t('project.create.namePlaceholder')}
                className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm"
              />
            </label>
            <label className="block">
              <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('project.create.group')}</span>
              <select
                value={draft.group}
                onChange={(e) => update({ group: e.target.value })}
                className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm"
              >
                {profile.groups.map((group) => (
                  <option key={group.id} value={group.id}>
                    {localized(group.label, i18n.language)}
                  </option>
                ))}
              </select>
            </label>
            <label className="block">
              <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('project.create.kind')}</span>
              <select
                value={draft.kind}
                onChange={(e) => update({ kind: e.target.value })}
                className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm"
              >
                {KINDS.map((kind) => (
                  <option key={kind} value={kind}>
                    {t(`project.kinds.${kind}`)}
                  </option>
                ))}
              </select>
            </label>
            <label className="block">
              <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('project.create.id')}</span>
              <input
                value={draft.id}
                onChange={(e) => {
                  setAuto(false)
                  update({ id: e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, '') })
                }}
                className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm font-mono"
              />
            </label>
            <label className="block">
              <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('project.create.folder')}</span>
              <input
                value={draft.repo}
                onChange={(e) => {
                  setAuto(false)
                  update({ repo: e.target.value.replace(/[^\w.-]/g, '') })
                }}
                className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm font-mono"
              />
            </label>
            <label className="block md:col-span-2">
              <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('project.create.githubRepo')}</span>
              <div className="mt-1 flex items-center gap-2 text-sm">
                <span className="text-slate-400 shrink-0">{org}/</span>
                <input
                  value={draft.github_repo}
                  onChange={(e) => {
                    setAuto(false)
                    update({ github_repo: e.target.value.replace(/[^A-Za-z0-9._-]/g, '') })
                  }}
                  className="flex-1 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 font-mono"
                />
              </div>
            </label>
            <label className="block">
              <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('project.create.purposeEn')}</span>
              <input
                value={draft.purpose_en}
                onChange={(e) => update({ purpose_en: e.target.value })}
                className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm"
              />
            </label>
            <label className="block">
              <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('project.create.purposeRu')}</span>
              <input
                value={draft.purpose_ru}
                onChange={(e) => update({ purpose_ru: e.target.value })}
                className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm"
              />
            </label>
          </div>
          {(error || message) && (
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
          <div className="flex justify-end">
            <Button
              size="sm"
              onClick={submit}
              disabled={saving || !draft.name.trim() || !draft.id || !draft.github_repo || !draft.repo}
            >
              {saving ? t('project.create.working') : t('project.create.submit')}
            </Button>
          </div>
        </div>
      )}
    </Card>
  )
}
