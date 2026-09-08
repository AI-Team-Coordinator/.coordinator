export type LocalizedText = {
  en: string
  ru: string
}

export interface MemberState {
  alias: string
  name: string
  role?: string
  focus?: string[]
  services?: string[]
  status: 'in_progress' | 'idle'
  task_id?: string
  task_title?: string
  task_doc?: string
  task_summary?: string
  branch?: string
  updated_at: string
  duration_seconds?: number
  repos?: RepoWork[]
  cost_usd?: number
  budget_usd?: number
  ondemand_usd?: number
  cursor_models_pct?: number
  other_models_pct?: number
  spend_kind?: 'infra' | 'product' | string
}

export interface RepoWork {
  id: string
  name: string
  repo: string
  kind?: 'workspace' | string
  state: 'local' | 'pushed' | 'merged' | 'deployed' | 'infra'
  ahead?: number
  dirty?: boolean
  deployed?: boolean
}

export interface StrayCommit {
  hash: string
  subject: string
}

export interface StrayRepo {
  id: string
  name: string
  repo: string
  branch: string
  commits: StrayCommit[]
}

export interface Conflict {
  severity: 'warning' | 'critical'
  title: string
  description: string
  affected_aliases: string[]
}

export interface Stats {
  total_completed: number
  active_now: number
  avg_cycle_time_minutes: number
  features_completed: number
  fixes_completed: number
  completed_today: number
  completed_this_week: number
  cost_usd_today: number
  cost_usd_week: number
  cost_usd_total: number
  cost_usd_avg: number
  cost_tasks: number
  budget_usd_today: number
  budget_usd_week: number
  budget_usd_total: number
  budget_usd_avg: number
  budget_tasks: number
  budget_usd_product_today: number
  budget_usd_product_week: number
  budget_usd_infra_today: number
  budget_usd_infra_week: number
  ondemand_usd_today: number
  ondemand_usd_week: number
  budget_usd_open: number
  cost_usd_open: number
  ondemand_usd_open: number
}

export interface TaskItem {
  task_id: string
  title?: string
  status: 'in_progress' | 'completed' | string
  kind?: 'feature' | 'fix' | string
  alias?: string
  branch?: string
  services?: string[]
  started_at?: number
  completed_at?: number
  duration_seconds?: number
  cost_usd?: number
  budget_usd?: number
  ondemand_usd?: number
  spend_kind?: 'infra' | 'product' | string
}

export interface EventItem {
  timestamp: number
  event: string
  task_id: string
  branch?: string
  alias?: string
  repo?: string
  service?: string
  status?: string
  cost_usd?: number
  budget_usd?: number
  ondemand_usd?: number
  cursor_models_pct?: number
  other_models_pct?: number
  usage_plan?: string
  plan_price_usd?: number
  spend_kind?: 'infra' | 'product' | string
}

export interface ProjectMeta {
  id: string
  name: string
  tagline: LocalizedText
}

export interface ServiceGroup {
  id: string
  label: LocalizedText
}

export interface GitHubBinding {
  host: string
  org: string
  ssh_host?: string
  html_url: string
}

export interface GitHubLive {
  auth: 'ok' | 'missing' | 'error' | string
  login?: string
  role?: string
  detail?: string
}

export interface ServiceNode {
  id: string
  name: string
  group: string
  kind: string
  repo: string
  github_repo?: string
  github_org?: string
  html_url?: string
  purpose: LocalizedText
}

export interface ServiceEdge {
  from: string
  to: string
  via: string
  kind: string
}

export interface ProjectProfile {
  version: number
  project: ProjectMeta
  github?: GitHubBinding
  github_live?: GitHubLive
  groups: ServiceGroup[]
  services: ServiceNode[]
  edges: ServiceEdge[]
}

export interface TeamPerson {
  alias: string
  name: string
  role?: string
  focus?: string[]
}

export interface SyncStatus {
  alias: string
  branch: string
  dirty_files: string[]
}

export interface SyncResult {
  alias: string
  committed: boolean
  pushed: boolean
  files: string[]
  message: string
}
