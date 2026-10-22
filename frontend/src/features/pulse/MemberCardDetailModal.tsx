import { MemberCard } from './MemberCard'
import { MemberCardDetailModalShell } from './MemberCardDetailModalShell'
import type { MemberState, MemberTaskState } from '../../shared/types/api'

interface MemberCardDetailModalProps {
  member: MemberState
  task?: MemberTaskState
  serviceNames: Record<string, string>
  onClose: () => void
}

export function MemberCardDetailModal({
  member,
  task,
  serviceNames,
  onClose,
}: MemberCardDetailModalProps) {
  return (
    <MemberCardDetailModalShell onClose={onClose}>
      <MemberCard member={member} task={task} serviceNames={serviceNames} embedded />
    </MemberCardDetailModalShell>
  )
}
