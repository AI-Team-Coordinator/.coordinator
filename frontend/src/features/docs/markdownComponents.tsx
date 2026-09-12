import type { Components } from 'react-markdown'

export const markdownComponents: Components = {
  h1: ({ children }) => <h1 className="text-xl font-bold tracking-tight mb-3 text-slate-900 dark:text-white">{children}</h1>,
  h2: ({ children }) => <h2 className="text-lg font-semibold mt-6 mb-2 text-slate-900 dark:text-white">{children}</h2>,
  h3: ({ children }) => <h3 className="text-base font-semibold mt-4 mb-1.5 text-slate-900 dark:text-white">{children}</h3>,
  p: ({ children }) => <p className="mb-3 leading-relaxed">{children}</p>,
  ul: ({ children }) => <ul className="mb-3 ml-5 list-disc space-y-1">{children}</ul>,
  ol: ({ children }) => <ol className="mb-3 ml-5 list-decimal space-y-1">{children}</ol>,
  li: ({ children }) => <li className="leading-relaxed">{children}</li>,
  a: ({ href, children }) => (
    <a href={href} target="_blank" rel="noreferrer" className="text-indigo-600 dark:text-indigo-400 underline underline-offset-2">
      {children}
    </a>
  ),
  code: ({ className, children }) => {
    const block = className?.includes('language-') || String(children).includes('\n')
    if (block) {
      return (
        <code className="block font-mono text-[12px] bg-slate-100 dark:bg-slate-950 rounded-md px-3 py-2 overflow-x-auto mb-3">
          {children}
        </code>
      )
    }
    return <code className="font-mono text-[12px] bg-slate-100 dark:bg-slate-800 rounded px-1 py-0.5">{children}</code>
  },
  pre: ({ children }) => <pre className="mb-3 overflow-x-auto">{children}</pre>,
  blockquote: ({ children }) => (
    <blockquote className="border-l-4 border-slate-300 dark:border-slate-600 pl-3 my-3 text-slate-600 dark:text-slate-400">
      {children}
    </blockquote>
  ),
  table: ({ children }) => (
    <div className="overflow-x-auto mb-3">
      <table className="min-w-full text-left text-xs border-collapse">{children}</table>
    </div>
  ),
  th: ({ children }) => (
    <th className="border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 px-2 py-1.5 font-semibold">
      {children}
    </th>
  ),
  td: ({ children }) => <td className="border border-slate-200 dark:border-slate-700 px-2 py-1.5 align-top">{children}</td>,
  hr: () => <hr className="my-4 border-slate-200 dark:border-slate-800" />,
}
