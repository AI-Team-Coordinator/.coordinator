export function formatDuration(seconds?: number): string {
  if (!seconds || seconds < 0) return '0s'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  if (h >= 24) {
    const d = Math.floor(h / 24)
    const remH = h % 24
    return remH > 0 ? `${d}d ${remH}h` : `${d}d`
  }
  if (h > 0) return `${h}h ${m}m ${s}s`
  if (m > 0) return `${m}m ${s}s`
  return `${s}s`
}

/** Average cycle time from stats (minutes). Scales m → h → d. */
export function formatCycleTime(minutes?: number | null): string {
  if (minutes === undefined || minutes === null || Number.isNaN(minutes) || minutes <= 0) {
    return '—'
  }
  const totalMin = Math.round(minutes)
  if (totalMin < 60) {
    return `${totalMin}m`
  }
  const days = Math.floor(totalMin / (60 * 24))
  const hours = Math.floor((totalMin % (60 * 24)) / 60)
  const mins = totalMin % 60
  if (days > 0) {
    return hours > 0 ? `${days}d ${hours}h` : `${days}d`
  }
  return mins > 0 ? `${hours}h ${mins}m` : `${hours}h`
}

export function formatTime(isoOrTimestamp?: string | number): string {
  if (!isoOrTimestamp) return ''
  const date = typeof isoOrTimestamp === 'number'
    ? new Date(isoOrTimestamp * 1000)
    : new Date(isoOrTimestamp)
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

export function formatDate(isoOrTimestamp?: string | number, locale: string = 'en'): string {
  if (!isoOrTimestamp) return ''
  const date = typeof isoOrTimestamp === 'number'
    ? new Date(isoOrTimestamp * 1000)
    : new Date(isoOrTimestamp)
  return date.toLocaleDateString(locale, { month: 'short', day: 'numeric' })
}

export function formatUSD(amount?: number | null): string {
  if (amount === undefined || amount === null || Number.isNaN(amount)) return '—'
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount)
}
