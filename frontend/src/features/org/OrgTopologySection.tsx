import { useEffect, useMemo, useState } from 'react'
import { Link2, Plus, Save, Trash2, Building2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { api } from '../../shared/api/client'
import { parseJSON } from '../../shared/api/http'
import { Button } from '../../shared/ui/Button'
import { Card } from '../../shared/ui/Card'
import type { Crew, CrewConnector, OrgTopology } from '../../shared/types/api'

const CREW_KINDS = ['internal', 'affiliated', 'external'] as const

type DraftCrew = Crew & { key: string }
type DraftConnector = CrewConnector & { key: string }

interface OrgTopologySectionProps {
  onChanged: () => void
  onDirtyChange: (dirty: boolean) => void
}

export function OrgTopologySection({ onChanged, onDirtyChange }: OrgTopologySectionProps) {
  const { t } = useTranslation()
  const [crews, setCrews] = useState<DraftCrew[]>([])
  const [connectors, setConnectors] = useState<DraftConnector[]>([])
  const [savedPayload, setSavedPayload] = useState('')
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const load = async () => {
    const res = await fetch(api.orgTopology)
    const data = await parseJSON<OrgTopology>(res)
    const nextCrews = (data.crews || []).map((c) => ({ ...c, key: c.id }))
    const nextLinks = (data.crew_connectors || []).map((l) => ({ ...l, key: l.id }))
    setCrews(nextCrews)
    setConnectors(nextLinks)
    setSavedPayload(JSON.stringify({ crews: nextCrews, connectors: nextLinks }))
  }

  useEffect(() => {
    load().catch((err: Error) => setError(err.message))
  }, [])

  const dirtyForm = useMemo(
    () => JSON.stringify({ crews, connectors }) !== savedPayload,
    [crews, connectors, savedPayload]
  )

  useEffect(() => {
    onDirtyChange(dirtyForm)
  }, [dirtyForm, onDirtyChange])

  const crewLabel = (id: string) => {
    const crew = crews.find((c) => c.id === id)
    if (!crew) return id
    return crew.label.ru || crew.label.en || id
  }

  const save = async () => {
    setSaving(true)
    setError(null)
    setMessage(null)
    try {
      const res = await fetch(api.orgTopology, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          version: 1,
          crews: crews.map((c) => ({
            id: c.id.trim().toLowerCase(),
            label: c.label,
            kind: c.kind || 'internal',
            interest_brief: c.interest_brief || '',
            notes: c.notes || '',
          })),
          crew_connectors: connectors.map((l) => ({
            id: l.id.trim().toLowerCase(),
            from_crew_id: l.from_crew_id,
            to_crew_id: l.to_crew_id,
            summary: l.summary || '',
            notes: l.notes || '',
            interests: (l.interests || []).map((item) => ({
              title: item.title,
              detail: item.detail || '',
              tags: item.tags || [],
            })),
          })),
        }),
      })
      const data = await parseJSON<OrgTopology & { pushed?: boolean; warning?: string }>(res)
      const nextCrews = (data.crews || []).map((c) => ({ ...c, key: c.id }))
      const nextLinks = (data.crew_connectors || []).map((l) => ({ ...l, key: l.id }))
      setCrews(nextCrews)
      setConnectors(nextLinks)
      setSavedPayload(JSON.stringify({ crews: nextCrews, connectors: nextLinks }))
      setMessage(data.pushed ? t('org.savedPushed') : t('org.savedLocal', { warning: data.warning || '' }))
      onChanged()
    } catch (err) {
      setError(err instanceof Error ? err.message : t('org.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  const addCrew = () => {
    setCrews((prev) => [
      ...prev,
      {
        key: `new-${Date.now()}`,
        id: '',
        label: { en: '', ru: '' },
        kind: 'internal',
        interest_brief: '',
        notes: '',
      },
    ])
  }

  const addConnector = () => {
    setConnectors((prev) => [
      ...prev,
      {
        key: `new-${Date.now()}`,
        id: '',
        from_crew_id: crews[0]?.id || '',
        to_crew_id: crews[1]?.id || crews[0]?.id || '',
        summary: '',
        notes: '',
        interests: [{ title: '', detail: '', tags: [] }],
      },
    ])
  }

  return (
    <section className="space-y-8">
      <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center gap-2">
            <Building2 className="w-4 h-4 text-indigo-500" />
            {t('org.title')}
          </h2>
          <p className="text-xs text-slate-500 dark:text-slate-400 mt-0.5">{t('org.subtitle')}</p>
        </div>
        <Button size="sm" onClick={save} disabled={saving || !dirtyForm} className="gap-1.5 self-end">
          <Save className="w-3.5 h-3.5" />
          {t('org.save')}
        </Button>
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
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-medium text-slate-800 dark:text-slate-200">{t('org.crews')}</h3>
          <Button variant="outline" size="sm" onClick={addCrew} className="gap-1.5">
            <Plus className="w-3.5 h-3.5" />
            {t('org.addCrew')}
          </Button>
        </div>
        {crews.map((crew) => (
          <Card key={crew.key} className="p-4 space-y-3">
            <div className="grid grid-cols-1 md:grid-cols-12 gap-3">
              <label className="md:col-span-3 block">
                <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('org.crewId')}</span>
                <input
                  value={crew.id}
                  onChange={(e) =>
                    setCrews((prev) =>
                      prev.map((c) => (c.key === crew.key ? { ...c, id: e.target.value.toLowerCase() } : c))
                    )
                  }
                  className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm font-mono"
                />
              </label>
              <label className="md:col-span-3 block">
                <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('org.labelRu')}</span>
                <input
                  value={crew.label.ru}
                  onChange={(e) =>
                    setCrews((prev) =>
                      prev.map((c) =>
                        c.key === crew.key ? { ...c, label: { ...c.label, ru: e.target.value } } : c
                      )
                    )
                  }
                  className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm"
                />
              </label>
              <label className="md:col-span-3 block">
                <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('org.labelEn')}</span>
                <input
                  value={crew.label.en}
                  onChange={(e) =>
                    setCrews((prev) =>
                      prev.map((c) =>
                        c.key === crew.key ? { ...c, label: { ...c.label, en: e.target.value } } : c
                      )
                    )
                  }
                  className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm"
                />
              </label>
              <label className="md:col-span-2 block">
                <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('org.kind')}</span>
                <select
                  value={crew.kind}
                  onChange={(e) =>
                    setCrews((prev) =>
                      prev.map((c) => (c.key === crew.key ? { ...c, kind: e.target.value } : c))
                    )
                  }
                  className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm"
                >
                  {CREW_KINDS.map((k) => (
                    <option key={k} value={k}>{t(`org.kinds.${k}`)}</option>
                  ))}
                </select>
              </label>
              <div className="md:col-span-1 flex md:justify-end md:pt-6">
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => setCrews((prev) => prev.filter((c) => c.key !== crew.key))}
                  className="text-slate-400 hover:text-rose-600"
                >
                  <Trash2 className="w-4 h-4" />
                </Button>
              </div>
            </div>
            <label className="block">
              <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('org.interestBrief')}</span>
              <textarea
                value={crew.interest_brief || ''}
                onChange={(e) =>
                  setCrews((prev) =>
                    prev.map((c) => (c.key === crew.key ? { ...c, interest_brief: e.target.value } : c))
                  )
                }
                rows={2}
                className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm"
              />
            </label>
          </Card>
        ))}
      </div>

      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-medium text-slate-800 dark:text-slate-200 flex items-center gap-2">
            <Link2 className="w-4 h-4" />
            {t('org.connectors')}
          </h3>
          <Button variant="outline" size="sm" onClick={addConnector} disabled={crews.length < 2} className="gap-1.5">
            <Plus className="w-3.5 h-3.5" />
            {t('org.addConnector')}
          </Button>
        </div>
        {connectors.map((link) => (
          <Card key={link.key} className="p-4 space-y-3">
            <div className="grid grid-cols-1 md:grid-cols-12 gap-3 items-start">
              <label className="md:col-span-3 block">
                <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('org.connectorId')}</span>
                <input
                  value={link.id}
                  onChange={(e) =>
                    setConnectors((prev) =>
                      prev.map((l) => (l.key === link.key ? { ...l, id: e.target.value.toLowerCase() } : l))
                    )
                  }
                  className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm font-mono"
                />
              </label>
              <label className="md:col-span-3 block">
                <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('org.fromCrew')}</span>
                <select
                  value={link.from_crew_id}
                  onChange={(e) =>
                    setConnectors((prev) =>
                      prev.map((l) => (l.key === link.key ? { ...l, from_crew_id: e.target.value } : l))
                    )
                  }
                  className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm"
                >
                  {crews.map((c) => (
                    <option key={c.key} value={c.id}>{crewLabel(c.id)}</option>
                  ))}
                </select>
              </label>
              <label className="md:col-span-3 block">
                <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('org.toCrew')}</span>
                <select
                  value={link.to_crew_id}
                  onChange={(e) =>
                    setConnectors((prev) =>
                      prev.map((l) => (l.key === link.key ? { ...l, to_crew_id: e.target.value } : l))
                    )
                  }
                  className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm"
                >
                  {crews.map((c) => (
                    <option key={c.key} value={c.id}>{crewLabel(c.id)}</option>
                  ))}
                </select>
              </label>
              <div className="md:col-span-3 flex md:justify-end md:pt-6">
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => setConnectors((prev) => prev.filter((l) => l.key !== link.key))}
                  className="text-slate-400 hover:text-rose-600"
                >
                  <Trash2 className="w-4 h-4" />
                </Button>
              </div>
            </div>
            <label className="block">
              <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('org.summary')}</span>
              <input
                value={link.summary || ''}
                onChange={(e) =>
                  setConnectors((prev) =>
                    prev.map((l) => (l.key === link.key ? { ...l, summary: e.target.value } : l))
                  )
                }
                className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm"
              />
            </label>
            {(link.interests || []).map((item, idx) => (
              <div key={idx} className="grid grid-cols-1 md:grid-cols-2 gap-2">
                <input
                  placeholder={t('org.interestTitle')}
                  value={item.title}
                  onChange={(e) =>
                    setConnectors((prev) =>
                      prev.map((l) => {
                        if (l.key !== link.key) return l
                        const interests = [...(l.interests || [])]
                        interests[idx] = { ...interests[idx], title: e.target.value }
                        return { ...l, interests }
                      })
                    )
                  }
                  className="rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm"
                />
                <input
                  placeholder={t('org.interestDetail')}
                  value={item.detail || ''}
                  onChange={(e) =>
                    setConnectors((prev) =>
                      prev.map((l) => {
                        if (l.key !== link.key) return l
                        const interests = [...(l.interests || [])]
                        interests[idx] = { ...interests[idx], detail: e.target.value }
                        return { ...l, interests }
                      })
                    )
                  }
                  className="rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm"
                />
              </div>
            ))}
          </Card>
        ))}
      </div>
    </section>
  )
}
