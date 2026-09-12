export type LocalizedText = {
  en: string
  ru: string
}

export interface MemberState {
  alias: string
  name: string
  role?: string
  access?: 'admin' | 'member' | string
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
  clock_paused?: boolean
  repos?: RepoWork[]
  cost_usd?: number
  budget_usd?: number
  ondemand_usd?: number
  cursor_models_pct?: number
  other_models_pct?: number
  spend_kind?: 'infra' | 'product' | 'research' | string
  research?: ResearchState
  tasks?: MemberTaskState[]
}

export interface MemberTaskState {
  task_id: string
  task_title?: string
  task_doc?: string
  task_summary?: string
  branch?: string
  services?: string[]
  started_at?: string
  duration_seconds?: number
  clock_paused?: boolean
  repos?: RepoWork[]
  cost_usd?: number
  budget_usd?: number
  ondemand_usd?: number
  cursor_models_pct?: number
  other_models_pct?: number
  spend_kind?: string
  chats?: ChatTab[]
}

export interface ChatTab {
  title?: string
  session_id?: string
}

export interface ResearchState {
  status: 'active' | string
  summary?: string
  started_at?: string
  duration_seconds?: number
  clock_paused?: boolean
  cost_usd?: number
  budget_usd?: number
  ondemand_usd?: number
  cursor_models_pct?: number
  other_models_pct?: number
  chat?: ChatTab
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
  research_completed: number
  research_completed_today: number
  research_completed_week: number
  completed_today: number
  completed_this_week: number
  cost_usd_today: number
  cost_usd_week: number
  cost_usd_cycle: number
  cost_usd_total: number
  cost_usd_avg: number
  cost_tasks: number
  budget_usd_today: number
  budget_usd_week: number
  budget_usd_cycle: number
  budget_usd_total: number
  budget_usd_avg: number
  budget_tasks: number
  budget_usd_product_today: number
  budget_usd_product_week: number
  budget_usd_product_cycle: number
  budget_usd_infra_today: number
  budget_usd_infra_week: number
  budget_usd_infra_cycle: number
  ondemand_usd_today: number
  ondemand_usd_week: number
  ondemand_usd_cycle: number
  budget_usd_open: number
  cost_usd_open: number
  ondemand_usd_open: number
  budget_usd_research_today: number
  budget_usd_research_week: number
  budget_usd_research_cycle: number
  budget_usd_research_open: number
  cost_usd_research_today: number
  cost_usd_research_week: number
  cost_usd_research_cycle: number
  cost_usd_research_open: number
  billing_cycle_start?: string
  plan_price_usd?: number
  cursor_models_pct_today: number
  other_models_pct_today: number
  cursor_models_pct_cycle: number
  other_models_pct_cycle: number
  cursor_models_pct_open: number
  other_models_pct_open: number
  cursor_models_pct_product_cycle: number
  other_models_pct_product_cycle: number
  cursor_models_pct_infra_cycle: number
  other_models_pct_infra_cycle: number
  cursor_models_pct_research_today: number
  other_models_pct_research_today: number
  cursor_models_pct_research_cycle: number
  other_models_pct_research_cycle: number
  cursor_models_pct_research_open: number
  other_models_pct_research_open: number
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
  clock_paused?: boolean
  cost_usd?: number
  budget_usd?: number
  ondemand_usd?: number
  cursor_models_pct?: number
  other_models_pct?: number
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
  spend_kind?: 'infra' | 'product' | 'research' | string
  summary?: string
  findings?: string
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
  access?: 'admin' | 'member' | string
  focus?: string[]
}

export type CollaborationMode = 'solo' | 'team'

export interface SetupState {
  needed: boolean
  coordinator_name?: string
  project_name: string
  alias?: string
  name?: string
  role?: string
  language: string
  layout: 'in-repo' | 'workspace-parent' | 'existing' | string
  docs_dir: string
  data_dir: string
  local: boolean
  collaboration?: CollaborationMode | string
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
