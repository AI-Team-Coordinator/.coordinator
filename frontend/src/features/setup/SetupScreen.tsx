import { useEffect, useState, type FormEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../../shared/api/client'
import { parseJSON } from '../../shared/api/http'
import { cn } from '../../shared/lib/utils'
import { Button } from '../../shared/ui/Button'
import { LanguageSwitcher } from '../../shared/ui/LanguageSwitcher'
import { ThemeToggle } from '../../shared/ui/ThemeToggle'
import { aliasFromName, sanitizeAlias } from '../../shared/lib/aliasFromName'
import type { CollaborationMode, SetupState } from '../../shared/types/api'
import { Typewriter } from './Typewriter'

interface SetupScreenProps {
  initial: SetupState
  onDone: (next: SetupState) => void
  onCancel?: () => void
  replay?: boolean
}

export function SetupScreen({ initial, onDone, onCancel, replay }: SetupScreenProps) {
  const { t, i18n } = useTranslation()
  const [projectName, setProjectName] = useState(initial.project_name || '')
  const [name, setName] = useState(initial.name || '')
  const [alias, setAlias] = useState((initial.alias || aliasFromName(initial.name || '')).toUpperCase())
  const [aliasTouched, setAliasTouched] = useState(Boolean(initial.alias))
  const [docsDir, setDocsDir] = useState(initial.docs_dir || 'docs')
  const [dataDir, setDataDir] = useState(initial.data_dir || 'coordinator-data')
  const [collaboration, setCollaboration] = useState<CollaborationMode>(
    initial.collaboration === 'team' ? 'team' : 'solo'
  )
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const coordinatorName = initial.coordinator_name?.trim() || t('setup.coordinatorFallback')
  const inRepo = initial.layout === 'in-repo'

  useEffect(() => {
    if (initial.language) {
      void i18n.changeLanguage(initial.language)
    }
    // Only prefill once — do not fight the language button after that.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    setSaving(true)
    setError(null)
    try {
      const language = i18n.language?.startsWith('ru') ? 'ru' : 'en'
      if (replay) {
        onDone(initial)
        return
      }
      const res = await fetch(api.setup, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          project_name: projectName.trim(),
          alias: alias.trim().toUpperCase(),
          name: name.trim(),
          role: initial.role || 'founder',
          language,
          docs_dir: docsDir.trim(),
          data_dir: dataDir.trim(),
          collaboration,
        }),
      })
      const next = await parseJSON<SetupState>(res)
      void i18n.changeLanguage(next.language || language)
      onDone(next)
    } catch (err) {
      setError(err instanceof Error ? err.message : t('setup.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="min-h-screen flex flex-col">
      <div className="flex items-center justify-end px-4 py-4 md:px-8 gap-2">
        <LanguageSwitcher />
        <ThemeToggle />
      </div>

      <div className="flex-1 flex items-start justify-center px-4 pb-16">
        <form onSubmit={submit} className="w-full max-w-3xl space-y-6 pt-4 md:pt-8">
          <div className="flex flex-col items-center text-center space-y-4">
            <img
              src="/coordinator-octopus-purple-front-512.png"
              alt=""
              width={256}
              height={256}
              className="w-56 h-56 md:w-64 md:h-64 object-contain select-none"
            />
            <Typewriter
              className="space-y-2 max-w-md"
              lines={[
                {
                  text: t('setup.helloLine1'),
                  heading: true,
                  className: 'text-2xl font-bold tracking-tight text-slate-900 dark:text-white',
                },
                {
                  text: t('setup.helloLine2', { name: coordinatorName }),
                  className: 'text-sm text-slate-600 dark:text-slate-300',
                },
              ]}
            />
          </div>

          <div className="space-y-2">
            <p className="text-[11px] uppercase tracking-wide text-slate-500 dark:text-slate-400 text-center sm:text-left">
              {t('setup.modeLabel')}
            </p>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <ModeCard
                selected={collaboration === 'solo'}
                title={t('setup.soloTitle')}
                lead={t('setup.soloLead')}
                points={[t('setup.soloPointLocal'), t('setup.soloPointChats'), t('setup.soloPointPeople')]}
                onClick={() => setCollaboration('solo')}
              />
              <ModeCard
                selected={collaboration === 'team'}
                title={t('setup.teamTitle')}
                lead={t('setup.teamLead')}
                points={[
                  t('setup.teamPointAgents'),
                  inRepo ? t('setup.teamPointMono') : t('setup.teamPointRepo'),
                ]}
                onClick={() => setCollaboration('team')}
              />
            </div>
          </div>

          <label className="block space-y-1.5">
            <span className="text-[11px] uppercase tracking-wide text-slate-500 dark:text-slate-400">
              {t('setup.projectName')}
            </span>
            <input
              value={projectName}
              onChange={(e) => setProjectName(e.target.value)}
              required
              autoFocus
              className="w-full rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-3 py-2.5 text-sm"
            />
          </label>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <label className="sm:col-span-2 block space-y-1.5">
              <span className="text-[11px] uppercase tracking-wide text-slate-500 dark:text-slate-400">
                {t('setup.yourName')}
              </span>
              <input
                value={name}
                onChange={(e) => {
                  const next = e.target.value
                  setName(next)
                  if (!aliasTouched) {
                    setAlias(aliasFromName(next))
                  }
                }}
                required
                className="w-full rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-3 py-2.5 text-sm"
              />
            </label>
            <label className="block space-y-1.5">
              <span className="text-[11px] uppercase tracking-wide text-slate-500 dark:text-slate-400">
                {t('setup.alias')}
              </span>
              <input
                value={alias}
                onChange={(e) => {
                  setAliasTouched(true)
                  setAlias(sanitizeAlias(e.target.value))
                }}
                required
                minLength={2}
                placeholder="EK"
                className="w-full rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-3 py-2.5 text-sm font-mono"
              />
              <span className="block text-xs text-slate-500 dark:text-slate-400">{t('setup.aliasHint')}</span>
            </label>
          </div>

          <label className="block space-y-1.5">
            <span className="text-[11px] uppercase tracking-wide text-slate-500 dark:text-slate-400">
              {t('setup.docsDir')}
            </span>
            <input
              value={docsDir}
              onChange={(e) => setDocsDir(e.target.value)}
              required
              placeholder="docs"
              className="w-full rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-3 py-2.5 text-sm font-mono"
            />
            <span className="block text-xs text-slate-500 dark:text-slate-400">{t('setup.docsDirHint')}</span>
          </label>

          <label className="block space-y-1.5">
            <span className="text-[11px] uppercase tracking-wide text-slate-500 dark:text-slate-400">
              {t('setup.dataDir')}
            </span>
            <input
              value={dataDir}
              onChange={(e) => setDataDir(e.target.value)}
              required
              placeholder="coordinator-data"
              className="w-full rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-3 py-2.5 text-sm font-mono"
            />
            <span className="block text-xs text-slate-500 dark:text-slate-400">{t('setup.dataDirHint')}</span>
          </label>

          {error && (
            <p className="text-sm text-rose-600 dark:text-rose-300 bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-900 rounded-lg px-3 py-2">
              {error}
            </p>
          )}

          <div className="space-y-2">
            <Button type="submit" disabled={saving} className="w-full h-11 text-sm">
              {saving ? t('setup.saving') : t('setup.start')}
            </Button>
            {onCancel && (
              <Button type="button" variant="ghost" className="w-full h-10 text-sm" onClick={onCancel}>
                {t('setup.back')}
              </Button>
            )}
          </div>
        </form>
      </div>
    </div>
  )
}

function ModeCard({
  selected,
  title,
  lead,
  points,
  onClick,
}: {
  selected: boolean
  title: string
  lead: string
  points: string[]
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={selected}
      className={cn(
        'rounded-2xl border-2 p-4 text-left transition-colors h-full',
        selected
          ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-950/40 shadow-sm'
          : 'border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 hover:border-slate-400 dark:hover:border-slate-500'
      )}
    >
      <p className="text-base font-semibold text-slate-900 dark:text-white">{title}</p>
      <p className="mt-1 text-sm text-slate-600 dark:text-slate-300 leading-relaxed">{lead}</p>
      <ul className="mt-3 space-y-1.5 text-sm text-slate-600 dark:text-slate-300">
        {points.map((point) => (
          <li key={point} className="flex gap-2">
            <span className="mt-2 h-1.5 w-1.5 shrink-0 rounded-full bg-indigo-500" />
            <span>{point}</span>
          </li>
        ))}
      </ul>
    </button>
  )
}
