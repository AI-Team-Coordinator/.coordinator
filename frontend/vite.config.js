// Vite loads vite.config.js before vite.config.ts — keep this file delegating to TS
// so proxy timeouts and UI_PORT/PORT stay in one place.
export { default } from './vite.config.ts'
