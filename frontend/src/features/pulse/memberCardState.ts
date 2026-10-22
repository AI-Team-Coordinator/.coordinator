import type { MemberState, MemberTaskState } from '../../shared/types/api'

export interface MemberCardState {
  fromTask: boolean
  isParked: boolean
  isActive: boolean
  isBusy: boolean
  taskId?: string
  branch?: string
  summary?: string
  isFix: boolean
  claimed: string[]
  taskTitle: string
  taskDoc?: string
  hasDoc: boolean
  clockPaused: boolean
}

export function memberCardState(member: MemberState, task?: MemberTaskState): MemberCardState {
  const fromTask = Boolean(task)
  const isParked = fromTask ? task?.status === 'parked' : member.status === 'parked'
  const isActive =
    Boolean(task) || (!task && (member.status === 'in_progress' || member.status === 'parked'))
  const taskId = fromTask ? task?.task_id : member.task_id
  const branch = fromTask ? task?.branch : member.branch
  const summary = fromTask ? task?.task_summary : member.task_summary
  const isFix = Boolean(taskId && (taskId.startsWith('FIX-') || (branch && branch.startsWith('fix/'))))
  const claimed = (fromTask ? task?.services : member.services) || []
  const taskTitle = fromTask
    ? task?.task_title || task?.task_id || ''
    : member.task_title || member.task_id || ''
  const taskDoc = fromTask ? task?.task_doc : member.task_doc
  const hasDoc = fromTask ? Boolean(task?.has_doc) : false
  const clockPaused = fromTask ? Boolean(task?.clock_paused) : Boolean(member.clock_paused)

  return {
    fromTask,
    isParked,
    isActive,
    isBusy: isActive,
    taskId,
    branch,
    summary,
    isFix,
    claimed,
    taskTitle,
    taskDoc,
    hasDoc,
    clockPaused,
  }
}

/** Compact card row under Fix/Feature: summary, or doc title when there is a markdown file. */
export function compactDetailLine(state: MemberCardState): { text: string; link: boolean } {
  const summary = (state.summary || '').trim()
  const docTitle = (state.taskTitle || '').trim()
  if (state.hasDoc && state.taskId) {
    const text = summary || docTitle
    return text ? { text, link: true } : { text: '', link: false }
  }
  return summary ? { text: summary, link: false } : { text: '', link: false }
}
