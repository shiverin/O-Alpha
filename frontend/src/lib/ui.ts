/**
 * UI and styling utility functions
 */

/**
 * Returns accent-based styling classes for hover/active states
 * Useful for creating theme-aware interactive elements
 *
 * @param accent - The accent color to use (primary, secondary, tertiary, etc.)
 * @returns string of Tailwind classes for accent-based styling
 */
export const getAccentStyle = (accent: 'primary' | 'secondary' | 'tertiary'): string => {
  switch (accent) {
    case 'primary':
      return 'hover:bg-primary-container/10 active:bg-primary-container/20'
    case 'secondary':
      return 'hover:bg-secondary-container/10 active:bg-secondary-container/20'
    case 'tertiary':
      return 'hover:bg-tertiary-container/10 active:bg-tertiary-container/20'
    default:
      return ''
  }
}

/**
 * Returns border styling classes for different intensities
 *
 * @param intensity - Border intensity ('light', 'medium', 'strong', or number 0-100)
 * @returns string of Tailwind classes for border styling
 */
export const getBorderStyle = (
  intensity: 'light' | 'medium' | 'strong' | number = 'medium'
): string => {
  if (typeof intensity === 'number') {
    // Clamp between 0-100
    const clamped = Math.max(0, Math.min(100, intensity))
    return `border border-outline-variant/${clamped}`
  }

  switch (intensity) {
    case 'light':
      return 'border border-outline-variant/30'
    case 'medium':
      return 'border border-outline-variant/50'
    case 'strong':
      return 'border border-outline-variant/70'
    default:
      return 'border border-outline-variant/50'
  }
}

/**
 * Returns glass/panel styling classes with optional hover effect
 *
 * @param elevated - whether to include hover elevation effect
 * @returns string of Tailwind classes for glass panel styling
 */
export const getPanelStyle = (elevated: boolean = false): string => {
  const base = 'bg-surface-container-high/80 border border-outline-variant/50 rounded-2xl'
  return elevated ? `${base} hover:bg-surface-container-highest/80 transition-colors duration-300` : base
}

/**
 * Returns container layout classes
 *
 * @param fluid - if true, removes max-width constraint
 * @param px - horizontal padding ('mobile', 'desktop', or 'none')
 * @param maxWidth - custom max-width (when not fluid)
 * @returns string of Tailwind classes for container layout
 */
export const getContainerStyle = (
  fluid: boolean = false,
  px: 'mobile' | 'desktop' | 'none' = 'desktop',
  maxWidth: string | number = '[1440px]'
): string => {
  // Padding
  let pxClass = ''
  switch (px) {
    case 'mobile':
      pxClass = 'px-margin-mobile'
      break
    case 'desktop':
      pxClass = 'px-margin-desktop'
      break
    case 'none':
      pxClass = 'px-0'
      break
  }

  // Width
  let widthClass = 'w-full'
  if (!fluid && maxWidth) {
    widthClass += ` max-w-${typeof maxWidth === 'number' ? `[${maxWidth}px]` : maxWidth}`
  }

  // Centering
  const mxClass = !fluid ? 'mx-auto' : ''

  return `${pxClass} ${widthClass} ${mxClass}`.trim()
}