import { X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Card } from '../../shared/ui/Card'
import { Button } from '../../shared/ui/Button'
import { Z_MODAL_OVERLAY } from '../../shared/ui/modalLayers'

interface MemberCardDetailModalShellProps {
  onClose: () => void
  children: React.ReactNode
}

export function MemberCardDetailModalShell({ onClose, children }: MemberCardDetailModalShellProps) {
  const { t } = useTranslation()

  return (
    <div
      className="fixed inset-0 flex items-center justify-center p-4 sm:p-6 bg-black/45 backdrop-blur-[2px]"
      style={{ zIndex: Z_MODAL_OVERLAY }}
      role="dialog"
      aria-modal="true"
      onClick={onClose}
    >
      <Card
        className="w-full max-w-lg max-h-[min(90vh,720px)] overflow-hidden flex flex-col shadow-2xl border-slate-200/90 dark:border-slate-700"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between gap-2 px-4 py-3 border-b border-slate-200 dark:border-slate-700 shrink-0">
          <h4 className="text-sm font-semibold text-slate-900 dark:text-white truncate">
            {t('pulse.cardDetailTitle')}
          </h4>
          <Button type="button" size="sm" variant="ghost" className="shrink-0 px-2" onClick={onClose}>
            <X className="w-4 h-4" />
          </Button>
        </div>
        <div className="overflow-y-auto px-4 py-4 sm:px-5 sm:py-4 min-h-0">{children}</div>
      </Card>
    </div>
  )
}
