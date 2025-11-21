'use client'

import { ReactNode } from 'react'

interface ProvidersProps {
  children: ReactNode
}

export function Providers({ children }: ProvidersProps) {
  return (
    <div className="flex h-screen bg-white dark:bg-dark-900">
      {children}
    </div>
  )
}

