export const SECTIONS = ['overview', 'work', 'team', 'organization', 'relay', 'project', 'stats', 'settings'] as const

export type SectionId = (typeof SECTIONS)[number]

export function parseSection(raw: string): SectionId {
  const segment = raw.replace(/^#/, '').split('/')[0] as SectionId
  return SECTIONS.includes(segment) ? segment : 'overview'
}
