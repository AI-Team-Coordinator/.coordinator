import type { LocalizedText } from '../types/api'

export function localized(value: LocalizedText | undefined, language: string): string {
  if (!value) return ''
  if (language.startsWith('ru')) {
    return value.ru || value.en
  }
  return value.en || value.ru
}
