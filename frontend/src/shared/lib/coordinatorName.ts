/** Default Coordinator name from UI / chat language. Otto rhymes with octopus. */
export function defaultCoordinatorName(lang: string | undefined): string {
  return (lang || '').toLowerCase().startsWith('ru') ? 'Отто' : 'Otto'
}

export function isStockCoordinatorName(name: string | undefined): boolean {
  const n = (name || '').trim()
  return n === '' || n === 'Otto' || n === 'Отто'
}
