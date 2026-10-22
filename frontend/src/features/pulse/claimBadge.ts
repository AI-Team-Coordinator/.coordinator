export function claimBadge(raw: string, serviceNames: Record<string, string>): string {
  const [head, ...rest] = raw.split('/')
  const name = serviceNames[head] || serviceNames[head.replace(/^\./, '')] || serviceNames[raw] || head
  return rest.length > 0 ? `${name}/${rest.join('/')}` : name
}
