import { useEffect, useState } from 'react'

export interface TypedLine {
  text: string
  className?: string
  heading?: boolean
}

interface TypewriterProps {
  lines: TypedLine[]
  speedMs?: number
  linePauseMs?: number
  startDelayMs?: number
  className?: string
}

function prefersReducedMotion() {
  return typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

/**
 * Types the lines out one character at a time. Each line keeps a hidden copy of
 * its full text so the form below never jumps while the text grows.
 */
export function Typewriter({
  lines,
  speedMs = 26,
  linePauseMs = 520,
  startDelayMs = 240,
  className,
}: TypewriterProps) {
  const joined = lines.map((line) => line.text).join('\n')
  const [typed, setTyped] = useState(() => (prefersReducedMotion() ? joined.length : 0))

  useEffect(() => {
    if (prefersReducedMotion()) {
      setTyped(joined.length)
      return
    }
    setTyped(0)
    let count = 0
    let timer = 0
    const tick = () => {
      count += 1
      setTyped(count)
      if (count >= joined.length) return
      timer = window.setTimeout(tick, joined[count - 1] === '\n' ? linePauseMs : speedMs)
    }
    timer = window.setTimeout(tick, startDelayMs)
    return () => window.clearTimeout(timer)
  }, [joined, speedMs, linePauseMs, startDelayMs])

  const shown = joined.slice(0, typed).split('\n')
  const done = typed >= joined.length

  return (
    <div className={className}>
      {lines.map((line, index) => {
        const Tag = line.heading ? 'h1' : 'p'
        const visible = shown[index] ?? ''
        return (
          <Tag key={index} className={`relative ${line.className ?? ''}`}>
            <span className="opacity-0">{line.text}</span>
            <span className="absolute inset-0" aria-hidden="true">
              {visible}
              {!done && index === shown.length - 1 && (
                <span className="ml-0.5 inline-block w-[2px] h-[1em] align-[-0.1em] bg-violet-500 animate-pulse" />
              )}
            </span>
          </Tag>
        )
      })}
    </div>
  )
}
