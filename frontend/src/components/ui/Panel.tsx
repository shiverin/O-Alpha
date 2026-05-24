import React from 'react'

interface PanelProps {
  children: React.ReactNode
  className?: string
  variant?: 'default' | 'elevated' | 'inset'
}

/**
 * Glass/panel component for consistent card-like surfaces
 *
 * Base styling: bg-surface-container-high/80 border border-outline-variant/50 rounded-2xl
 * Variants:
 *   default: standard glass effect
 *   elevated: higher elevation on hover
 *   inset: inner shadow effect
 */
export const Panel = ({ children, className = '', variant = 'default' }: PanelProps) => {
  const baseClasses = 'bg-surface-container-high/80 border border-outline-variant/50 rounded-2xl'

  let variantClasses = ''
  switch (variant) {
    case 'elevated':
      variantClasses = 'hover:bg-surface-container-highest/80 transition-colors duration-300'
      break
    case 'inset':
      variantClasses = 'inset-shadow-[0px_0px_0px_1px_rgb(0,0,0,0.05)]'
      break
    default:
      variantClasses = ''
  }

  return (
    <div className={`${baseClasses} ${variantClasses} ${className}`}>
      {children}
    </div>
  )
}