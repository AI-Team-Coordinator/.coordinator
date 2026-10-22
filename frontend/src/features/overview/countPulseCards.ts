import type { MemberState, MemberTaskState, StrayRepo } from '../../shared/types/api'

function memberTasks(member: MemberState): MemberTaskState[] {
  if (member.tasks && member.tasks.length > 0) return member.tasks
  if ((member.status === 'in_progress' || member.status === 'parked') && member.task_id) {
    return [{ task_id: member.task_id }]
  }
  return []
}

/** Matches card count in TeamPulseSection (including UnassignedWorkCard). */
export function countPulseCards(members: MemberState[], _stray: StrayRepo[]): number {
  let n = 0
  for (const member of members) {
    const tasks = memberTasks(member)
    if (tasks.length > 0) n += tasks.length
    else n += 1
    if (member.research?.status === 'active') n += 1
  }
  return n + 1
}
