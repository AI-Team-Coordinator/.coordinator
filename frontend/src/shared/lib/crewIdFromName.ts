const CYRILLIC: Record<string, string> = {
  а: 'a',
  б: 'b',
  в: 'v',
  г: 'g',
  д: 'd',
  е: 'e',
  ё: 'e',
  ж: 'zh',
  з: 'z',
  и: 'i',
  й: 'y',
  к: 'k',
  л: 'l',
  м: 'm',
  н: 'n',
  о: 'o',
  п: 'p',
  р: 'r',
  с: 's',
  т: 't',
  у: 'u',
  ф: 'f',
  х: 'h',
  ц: 'ts',
  ч: 'ch',
  ш: 'sh',
  щ: 'sch',
  ъ: '',
  ы: 'y',
  ь: '',
  э: 'e',
  ю: 'yu',
  я: 'ya',
}

function transliterate(text: string): string {
  let out = ''
  for (const ch of text.toLowerCase()) {
    if (CYRILLIC[ch] !== undefined) {
      out += CYRILLIC[ch]
    } else {
      out += ch
    }
  }
  return out
}

/** Stable slug for org_topology crew id (matches backend slugIDPattern). */
export function crewIdFromName(name: string): string {
  const slug = transliterate(name)
    .trim()
    .toLowerCase()
    .replace(/\s+/g, '-')
    .replace(/[^a-z0-9_-]/g, '')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
  if (!slug) return ''
  if (/^[0-9]/.test(slug)) return `c-${slug}`.slice(0, 64)
  return slug.slice(0, 64)
}
