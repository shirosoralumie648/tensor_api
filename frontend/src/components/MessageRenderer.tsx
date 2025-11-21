'use client'

import { useMemo } from 'react'
import markdownIt from 'markdown-it'
import hljs from 'highlight.js'

interface MessageRendererProps {
  content: string
}

export function MessageRenderer({ content }: MessageRendererProps) {
  // 配置 markdown-it
  const md = useMemo(() => {
    const instance = markdownIt({
      html: false,
      breaks: true,
      highlight: (code: string, lang: string) => {
        if (lang && hljs.getLanguage(lang)) {
          try {
            return (
              '<pre class="hljs"><code>' +
              hljs.highlight(code, { language: lang, ignoreIllegals: true })
                .value +
              '</code></pre>'
            )
          } catch {
            // fallback
          }
        }

        return (
          '<pre class="hljs"><code>' +
          md.utils.escapeHtml(code) +
          '</code></pre>'
        )
      },
    })

    return instance
  }, [])

  // 渲染 HTML
  const html = useMemo(() => {
    return md.render(content)
  }, [content, md])

  return (
    <div
      className="prose dark:prose-invert max-w-none text-sm"
      dangerouslySetInnerHTML={{ __html: html }}
    />
  )
}

