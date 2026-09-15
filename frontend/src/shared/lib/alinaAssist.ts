/** Hardcoded: replay-setup and Vite HMR are for this product only. */
export function isAlinaAssistProject(profile: { project?: { id?: string; name?: string } } | null | undefined): boolean {
  const id = profile?.project?.id?.trim().toLowerCase() || ''
  const name = profile?.project?.name?.trim().toLowerCase() || ''
  return id === 'alina-assist' || name === 'alina assist'
}
