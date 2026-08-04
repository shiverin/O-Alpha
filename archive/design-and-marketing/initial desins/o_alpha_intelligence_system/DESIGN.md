---
name: O(Alpha) Intelligence System
colors:
  surface: '#051424'
  surface-dim: '#051424'
  surface-bright: '#2c3a4c'
  surface-container-lowest: '#010f1f'
  surface-container-low: '#0d1c2d'
  surface-container: '#122131'
  surface-container-high: '#1c2b3c'
  surface-container-highest: '#273647'
  on-surface: '#d4e4fa'
  on-surface-variant: '#bac9cc'
  inverse-surface: '#d4e4fa'
  inverse-on-surface: '#233143'
  outline: '#849396'
  outline-variant: '#3b494c'
  surface-tint: '#00daf3'
  primary: '#c3f5ff'
  on-primary: '#00363d'
  primary-container: '#00e5ff'
  on-primary-container: '#00626e'
  inverse-primary: '#006875'
  secondary: '#fff9ef'
  on-secondary: '#3a3000'
  secondary-container: '#ffdb3c'
  on-secondary-container: '#725f00'
  tertiary: '#e6ecff'
  on-tertiary: '#283041'
  tertiary-container: '#c8d0e6'
  on-tertiary-container: '#51596b'
  error: '#ffb4ab'
  on-error: '#690005'
  error-container: '#93000a'
  on-error-container: '#ffdad6'
  primary-fixed: '#9cf0ff'
  primary-fixed-dim: '#00daf3'
  on-primary-fixed: '#001f24'
  on-primary-fixed-variant: '#004f58'
  secondary-fixed: '#ffe16d'
  secondary-fixed-dim: '#e9c400'
  on-secondary-fixed: '#221b00'
  on-secondary-fixed-variant: '#544600'
  tertiary-fixed: '#dbe2f8'
  tertiary-fixed-dim: '#bec6dc'
  on-tertiary-fixed: '#131c2b'
  on-tertiary-fixed-variant: '#3f4758'
  background: '#051424'
  on-background: '#d4e4fa'
  surface-variant: '#273647'
typography:
  headline-xl:
    fontFamily: Inter
    fontSize: 48px
    fontWeight: '700'
    lineHeight: 56px
    letterSpacing: -0.02em
  headline-lg:
    fontFamily: Inter
    fontSize: 32px
    fontWeight: '600'
    lineHeight: 40px
    letterSpacing: -0.01em
  headline-lg-mobile:
    fontFamily: Inter
    fontSize: 24px
    fontWeight: '600'
    lineHeight: 32px
  body-md:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: '400'
    lineHeight: 24px
  data-lg:
    fontFamily: JetBrains Mono
    fontSize: 20px
    fontWeight: '500'
    lineHeight: 28px
  data-sm:
    fontFamily: JetBrains Mono
    fontSize: 12px
    fontWeight: '400'
    lineHeight: 16px
    letterSpacing: 0.05em
  label-caps:
    fontFamily: Inter
    fontSize: 11px
    fontWeight: '700'
    lineHeight: 16px
    letterSpacing: 0.1em
rounded:
  sm: 0.125rem
  DEFAULT: 0.25rem
  md: 0.375rem
  lg: 0.5rem
  xl: 0.75rem
  full: 9999px
spacing:
  base: 8px
  gutter: 24px
  margin-mobile: 16px
  margin-desktop: 64px
  grid-size: 32px
---

## Brand & Style

The visual identity is anchored in **Quant-Minimalism**, a style that fuses the clinical precision of high-frequency trading interfaces with a modern, sophisticated aesthetic. It targets high-net-worth individuals and tech-literate investors who value speed, logic, and automated precision.

The emotional response is one of **calculated confidence**. By utilizing deep, nocturnal backgrounds paired with vibrant, light-emitting data points, the UI evokes the feeling of a command center. Key design pillars include:

*   **Precision & Accuracy:** Sharp corners, monospaced typography for metrics, and thin-gauge lines.
*   **Technological Sophistication:** Subtle background grids and circular orbital motifs suggest a continuous AI "radar" scanning the markets.
*   **Trust through Transparency:** Glassmorphic surfaces provide a sense of layered information architecture, making complex data feel organized and accessible.

## Colors

The palette is designed for high legibility in low-light environments, using light as a functional signifier for data importance.

*   **Core Background:** The deepest navy (`#020817`) serves as the "void," ensuring that all primary actions and data visualizations appear to emit light.
*   **Electric Cyan (Primary):** Used for interactive elements, progress indicators, and active "live" states. It represents the "energy" of the AI agent.
*   **Alpha Gold (Signature):** Reserved exclusively for the brand mark, critical "Alpha" generation metrics, and elite status indicators.
*   **Functional Slate (Neutral):** A desaturated palette of greys used for secondary text, borders, and inactive UI states to prevent visual clutter.
*   **Semantic Accents:** Utilize a strictly controlled red for risk/drawdown and emerald for profit, both desaturated to maintain the professional tone.

## Typography

This design system uses a dual-font strategy to separate narrative from data.

*   **Inter:** Used for all prose, headings, and interface controls. It provides a human-centric, readable foundation that balances the technical nature of the product.
*   **JetBrains Mono:** Used for all "hard" numbers, ticker symbols, code snippets, and technical labels. The monospaced nature ensures that columns of changing numbers do not "jump" visually and maintain perfect vertical alignment.
*   **Stylistic Note:** Use "Label-Caps" for section headers and metadata to create a "heads-up display" (HUD) feel.

## Layout & Spacing

The layout is built on a **12-column fluid grid** that locks to a maximum width of 1440px on desktop. 

*   **Grid Infrastructure:** A subtle background grid (visible at 32px increments) should be rendered in a very low-opacity cyan (`#00E5FF10`) to reinforce the "data-driven" theme.
*   **Density:** The spacing is generous between major sections to provide "air," but tight within data cards to allow for information density.
*   **Breakpoints:**
    *   **Desktop (1280px+):** Full 12-column layout with 24px gutters.
    *   **Tablet (768px - 1279px):** 8-column layout; glassmorphic cards stack vertically where necessary.
    *   **Mobile (below 768px):** 4-column layout with 16px margins. Headlines scale down to `headline-lg-mobile` for better fit.

## Elevation & Depth

Depth is created through **Glassmorphism and Tonal Layering** rather than traditional shadows.

1.  **Level 0 (Floor):** The deep navy background with the faint structural grid.
2.  **Level 1 (Surface):** Semi-transparent panels with a 12px backdrop blur. These "glass" cards use a 1px solid border in a slightly lighter navy or low-opacity white to define their edges.
3.  **Level 2 (Active/Hover):** When an element is focused or hovered, it gains a subtle outer glow (cyan) and increased border opacity.
4.  **Floating Elements:** Tooltips and dropdowns utilize a higher backdrop-blur (20px) and a distinct "Alpha Gold" or "Electric Cyan" top-border (2px) to denote importance.

## Shapes

The design system uses a **Soft (Level 1)** roundedness profile. This specific choice balances the "sharpness" of financial data with the modern "software" feel.

*   **Primary Elements:** Buttons and Input fields use a 0.25rem (4px) radius.
*   **Containers:** Larger data cards and glassmorphism panels use 0.75rem (12px) to create a clear container-to-content relationship.
*   **Special Shapes:** The "Alpha" logo and certain status indicators may use perfect circles to represent "orbits" and "cycles," contrasting against the otherwise rectangular grid.

## Components

*   **Primary Buttons:** Solid Electric Cyan background with black text for maximum contrast. No rounded-pill shapes; stick to the 4px radius.
*   **Data Chips:** Small, monospaced labels with a low-opacity cyan fill and a 1px solid cyan border.
*   **Cards:** Glassmorphic treatment with a 1px border (`rgba(255, 255, 255, 0.1)`). Header areas within cards should be separated by a thin horizontal rule.
*   **Inputs:** Darker than the card surface, with an "active" state that glows Cyan. Use JetBrains Mono for the input text.
*   **Status Indicators:** Use a "pulsing" dot animation for "Live" status. Cyan for active trading, Gold for Alpha-generating events, and desaturated Slate for idle.
*   **Charts:** Vector-based line charts with a gradient "area" fill below the line, using Primary Cyan. The line itself should have a subtle neon glow effect.