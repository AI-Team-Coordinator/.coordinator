export const SECTIONS = ['overview', 'work', 'team', 'organization', 'relay', 'project', 'stats', 'settings'] as const

export type SectionId = (typeof SECTIONS)[number]

export function parseSection(raw: string): SectionId {
  const value = raw.replace(/^#/, '') as SectionId
  return SECTIONS.includes(value) ? value : 'overview'
}
