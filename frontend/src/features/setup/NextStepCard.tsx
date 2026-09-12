import { useTranslation } from 'react-i18next'
import { Card } from '../../shared/ui/Card'

export function NextStepCard() {
  const { t } = useTranslation()

  return (
    <Card className="p-5 border-indigo-200 dark:border-indigo-900/60 bg-indigo-50/70 dark:bg-indigo-950/30">
      <p className="text-sm font-semibold text-slate-900 dark:text-white">{t('setup.nextTitle')}</p>
      <p className="mt-1 text-sm text-slate-600 dark:text-slate-300 leading-relaxed">{t('setup.nextBody')}</p>
    </Card>
  )
}
