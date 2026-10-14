import { Settings } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '../../shared/api/client'
import i18n from '../../shared/i18n/config'
import { Button } from '../../shared/ui/Button'
import { Card } from '../../shared/ui/Card'
import { Select } from '../../shared/ui/Select'
import type { UserSettings } from '../../shared/types/api'

function avatarSrc(url: string | undefined, bust: number) {
  if (!url) return undefined
  return `${url}?v=${bust}`
}

async function fileToJpegBlob(file: File): Promise<Blob> {
  const bitmap = await createImageBitmap(file)
  const size = 256
  const scale = Math.min(size / bitmap.width, size / bitmap.height, 1)
  const w = Math.max(1, Math.round(bitmap.width * scale))
  const h = Math.max(1, Math.round(bitmap.height * scale))
  const canvas = document.createElement('canvas')
  canvas.width = w
  canvas.height = h
  const ctx = canvas.getContext('2d')
  if (!ctx) throw new Error('canvas')
  ctx.drawImage(bitmap, 0, 0, w, h)
  bitmap.close()
  return new Promise((resolve, reject) => {
    canvas.toBlob((blob) => (blob ? resolve(blob) : reject(new Error('encode'))), 'image/jpeg', 0.88)
  })
}

interface SettingsSectionProps {
  settings: UserSettings | null
  loading: boolean
  loadError?: boolean
  avatarBust?: number
  onSaved?: (settings: UserSettings) => void
}

export function SettingsSection({
  settings,
  loading,
  loadError,
  avatarBust = 0,
  onSaved,
}: SettingsSectionProps) {
  const { t } = useTranslation()
  const fileRef = useRef<HTMLInputElement>(null)
  const [saving, setSaving] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [error, setError] = useState('')
  const [name, setName] = useState('')
  const [language, setLanguage] = useState('en')

  useEffect(() => {
    if (!settings) return
    setName(settings.name)
    setLanguage(settings.language?.startsWith('ru') ? 'ru' : 'en')
  }, [settings])

  const saveProfile = async () => {
    setSaving(true)
    setError('')
    try {
      const res = await fetch(api.meSettings, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: name.trim(), language }),
      })
      if (!res.ok) throw new Error('save')
      const data = (await res.json()) as UserSettings
      void i18n.changeLanguage(language)
      onSaved?.(data)
    } catch {
      setError(t('settings.saveError'))
    } finally {
      setSaving(false)
    }
  }

  const onPickAvatar = async (file: File | undefined) => {
    if (!file || !file.type.startsWith('image/')) return
    setUploading(true)
    setError('')
    try {
      const blob = await fileToJpegBlob(file)
      const form = new FormData()
      form.append('file', blob, 'avatar.jpg')
      const res = await fetch(api.meAvatar, { method: 'POST', body: form })
      if (!res.ok) throw new Error('upload')
      const data = (await res.json()) as UserSettings
      onSaved?.(data)
    } catch {
      setError(t('settings.avatarError'))
    } finally {
      setUploading(false)
      if (fileRef.current) fileRef.current.value = ''
    }
  }

  const removeAvatar = async () => {
    setUploading(true)
    setError('')
    try {
      const res = await fetch(api.meAvatar, { method: 'DELETE' })
      if (!res.ok) throw new Error('delete')
      const data = (await res.json()) as UserSettings
      onSaved?.(data)
    } catch {
      setError(t('settings.avatarError'))
    } finally {
      setUploading(false)
    }
  }

  const roleLabel =
    settings?.role ? t(`pulse.roles.${settings.role}`, { defaultValue: settings.role }) : ''
  const accessLabel = settings?.access === 'admin' ? t('setup.accessAdmin') : ''

  return (
    <section className="space-y-8 max-w-xl">
      <div>
        <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center gap-2">
          <Settings className="w-4 h-4 text-indigo-500" />
          {t('settings.title')}
        </h2>
      </div>

      {(error || loadError) && (
        <p className="text-sm text-rose-600 dark:text-rose-400">
          {error || t('settings.loadError')}
        </p>
      )}

      {loading ? (
        <p className="text-sm text-slate-500">{t('settings.loading')}</p>
      ) : settings ? (
        <>
          <Card className="p-4 space-y-4">
            <h3 className="text-sm font-medium text-slate-800 dark:text-slate-200">{t('settings.profile')}</h3>
            <div className="flex items-start gap-4">
              <div className="shrink-0">
                {settings.avatar_url ? (
                  <img
                    src={avatarSrc(settings.avatar_url, avatarBust)}
                    alt=""
                    className="w-16 h-16 rounded-xl object-cover border border-slate-200 dark:border-slate-700"
                  />
                ) : (
                  <div className="w-16 h-16 rounded-xl bg-indigo-50 dark:bg-slate-800 border border-indigo-200 dark:border-slate-700 flex items-center justify-center font-bold text-sm text-indigo-600 dark:text-indigo-300">
                    {settings.alias}
                  </div>
                )}
              </div>
              <div className="space-y-2 min-w-0 flex-1">
                <p className="text-xs text-slate-500 dark:text-slate-400">
                  @{settings.alias}
                  {roleLabel ? ` · ${roleLabel}` : ''}
                  {accessLabel ? ` · ${accessLabel}` : ''}
                </p>
                <input
                  ref={fileRef}
                  type="file"
                  accept="image/*"
                  className="hidden"
                  onChange={(e) => onPickAvatar(e.target.files?.[0])}
                />
                <div className="flex flex-wrap gap-2">
                  <Button
                    size="sm"
                    variant="secondary"
                    disabled={uploading}
                    onClick={() => fileRef.current?.click()}
                  >
                    {uploading ? t('settings.uploading') : t('settings.avatarUpload')}
                  </Button>
                  {settings.avatar_url && (
                    <Button size="sm" variant="ghost" disabled={uploading} onClick={() => void removeAvatar()}>
                      {t('settings.avatarRemove')}
                    </Button>
                  )}
                </div>
              </div>
            </div>
            <label className="block">
              <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('settings.displayName')}</span>
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="mt-1 w-full rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-950 px-2.5 py-1.5 text-sm"
              />
            </label>
            <Button size="sm" disabled={saving || !name.trim()} onClick={() => void saveProfile()}>
              {saving ? t('settings.saving') : t('settings.save')}
            </Button>
          </Card>

          <Card className="p-4 space-y-3">
            <label className="block">
              <span className="text-[11px] uppercase tracking-wide text-slate-500">{t('settings.language')}</span>
              <Select
                className="mt-1 max-w-xs"
                value={language}
                onChange={(e) => setLanguage(e.target.value)}
              >
                <option value="en">{t('settings.langEn')}</option>
                <option value="ru">{t('settings.langRu')}</option>
              </Select>
            </label>
            <Button size="sm" disabled={saving} onClick={() => void saveProfile()}>
              {saving ? t('settings.saving') : t('settings.save')}
            </Button>
          </Card>
        </>
      ) : null}
    </section>
  )
}
