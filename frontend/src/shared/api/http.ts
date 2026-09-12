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
  return (await res.json()) as T
}
