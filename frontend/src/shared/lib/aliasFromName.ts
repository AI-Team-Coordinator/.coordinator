const CYR_INITIAL: Record<string, string> = {
  А: 'A',
  Б: 'B',
  В: 'V',
  Г: 'G',
  Д: 'D',
  Е: 'E',
  Ё: 'E',
  Ж: 'Z',
  З: 'Z',
  И: 'I',
  Й: 'I',
  К: 'K',
  Л: 'L',
  М: 'M',
  Н: 'N',
  О: 'O',
  П: 'P',
  Р: 'R',
  С: 'S',
  Т: 'T',
  У: 'U',
  Ф: 'F',
  Х: 'H',
  Ц: 'C',
  Ч: 'C',
  Ш: 'S',
  Щ: 'S',
  Ы: 'Y',
  Э: 'E',
  Ю: 'U',
  Я: 'Y',
}

function latinInitial(ch: string): string {
  const upper = ch.toUpperCase()
  if (/[A-Z]/.test(upper)) {
    return upper
  }
  return CYR_INITIAL[upper] || ''
}

function firstLetter(word: string): string {
  for (const ch of word) {
    const letter = latinInitial(ch)
    if (letter) {
      return letter
    }
  }
  return ''
}

/** 2–4 Latin letters: first+last initials, or first two letters of a single name. */
export function aliasFromName(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean)
  if (parts.length >= 2) {
    return (firstLetter(parts[0]) + firstLetter(parts[parts.length - 1])).slice(0, 4)
  }
  if (parts.length === 1) {
    const letters: string[] = []
    for (const ch of parts[0]) {
      const letter = latinInitial(ch)
      if (letter) {
        letters.push(letter)
      }
      if (letters.length >= 2) {
        break
      }
    }
    return letters.join('')
  }
  return ''
}

export function sanitizeAlias(raw: string): string {
  return raw.toUpperCase().replace(/[^A-Z]/g, '').slice(0, 4)
}
