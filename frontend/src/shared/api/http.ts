export async function readError(res: Response): Promise<string> {
  try {
    const data = (await res.json()) as { error?: string }
    if (data.error) return data.error.replace(/^E\d+:\s*/, '')
  } catch {
    // ignore
  }
  return res.statusText || `HTTP ${res.status}`
}

export async function parseJSON<T>(res: Response): Promise<T> {
  if (!res.ok) {
    throw new Error(await readError(res))
  }
  const ct = res.headers.get('content-type') || ''
  if (!ct.includes('application/json')) {
    const text = (await res.text()).trim()
    if (text.startsWith('<')) {
      throw new Error('API returned HTML instead of JSON — restart ./utils/backend.sh (old API or wrong URL)')
    }
    throw new Error(text.slice(0, 120) || 'Unexpected response type')
  }
  return JSON.parse(await res.text()) as T
}
