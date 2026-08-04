---
name: Ethereal Monolith
colors:
  surface: '#131313'
  surface-dim: '#131313'
  surface-bright: '#393939'
  surface-container-lowest: '#0e0e0e'
  surface-container-low: '#1c1b1b'
  surface-container: '#20201f'
  surface-container-high: '#2a2a2a'
  surface-container-highest: '#353535'
  on-surface: '#e5e2e1'
  on-surface-variant: '#b9cacb'
  inverse-surface: '#e5e2e1'
  inverse-on-surface: '#313030'
  outline: '#849495'
  outline-variant: '#3b494b'
  surface-tint: '#00dbe9'
  primary: '#dbfcff'
  on-primary: '#00363a'
  primary-container: '#00f0ff'
  on-primary-container: '#006970'
  inverse-primary: '#006970'
  secondary: '#fff9ef'
  on-secondary: '#3a3000'
  secondary-container: '#ffdb3c'
  on-secondary-container: '#725f00'
  tertiary: '#f5f5f5'
  on-tertiary: '#2f3131'
  tertiary-container: '#d9d9d9'
  on-tertiary-container: '#5d5f5f'
  error: '#ffb4ab'
  on-error: '#690005'
  error-container: '#93000a'
  on-error-container: '#ffdad6'
  primary-fixed: '#7df4ff'
  primary-fixed-dim: '#00dbe9'
  on-primary-fixed: '#002022'
  on-primary-fixed-variant: '#004f54'
  secondary-fixed: '#ffe16d'
  secondary-fixed-dim: '#e9c400'
  on-secondary-fixed: '#221b00'
  on-secondary-fixed-variant: '#544600'
  tertiary-fixed: '#e2e2e2'
  tertiary-fixed-dim: '#c6c6c7'
  on-tertiary-fixed: '#1a1c1c'
  on-tertiary-fixed-variant: '#454747'
  background: '#131313'
  on-background: '#e5e2e1'
  surface-variant: '#353535'
  void-black: '#0A0A0A'
  surface-gray: '#1E1E1E'
  border-glass: rgba(255, 255, 255, 0.12)
  glow-cyan: rgba(0, 240, 255, 0.4)
  glow-gold: rgba(255, 215, 0, 0.3)
typography:
  display-lg:
    fontFamily: Inter
    fontSize: 80px
    fontWeight: '700'
    lineHeight: 88px
    letterSpacing: -0.04em
  display-lg-mobile:
    fontFamily: Inter
    fontSize: 48px
    fontWeight: '700'
    lineHeight: 52px
    letterSpacing: -0.03em
  headline-xl:
    fontFamily: Inter
    fontSize: 48px
    fontWeight: '600'
    lineHeight: 56px
    letterSpacing: -0.02em
  headline-lg:
    fontFamily: Inter
    fontSize: 32px
    fontWeight: '600'
    lineHeight: 40px
    letterSpacing: -0.02em
  body-lg:
    fontFamily: Inter
    fontSize: 18px
    fontWeight: '400'
    lineHeight: 32px
    letterSpacing: 0em
  body-md:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: '400'
    lineHeight: 28px
    letterSpacing: 0em
  label-mono:
    fontFamily: JetBrains Mono
    fontSize: 12px
    fontWeight: '500'
    lineHeight: 16px
    letterSpacing: 0.05em
rounded:
  sm: 0.25rem
  DEFAULT: 0.5rem
  md: 0.75rem
  lg: 1rem
  xl: 1.5rem
  full: 9999px
spacing:
  margin-desktop: 64px
  margin-mobile: 24px
  gutter: 24px
  section-gap: 120px
  container-max: 1280px
---

## Brand & Style

The design system is a fusion of high-concept aerospace aesthetics and editorial precision. It targets a sophisticated audience that values both technical depth and clarity. The personality is "Intelligent Minimalism"—feeling both lightweight and structurally sound.

The visual style combines **Glassmorphism** with **High-Contrast Editorial** layouts. We utilize the deep, atmospheric voids of dark-mode backgrounds typical of research environments, punctuated by the structural clarity and bold color-blocking seen in modern design tooling platforms. The aesthetic response should be one of "effortless power" and "structured discovery."

## Colors

The palette is anchored in a "Void Black" base to create an infinite sense of depth. 

- **Primary (Cyan):** Used for interactive states, progress indicators, and "active data" visualizations. It should possess a subtle outer glow (2-4px) to simulate light emission.
- **Secondary (Gold):** Reserved for high-value insights, premium features, and brand accents. 
- **Neutral/Chrome:** Pure White and Black are used for the navigational skeleton, ensuring the "Figma-esque" structural clarity.
- **Color-Blocking:** Use full-bleed background panels in primary or secondary shades with 10% opacity to create distinct storytelling sections without losing the dark-mode immersion.

## Typography

This design system uses a dual-font approach. **Inter** handles all primary communication with a focus on tight tracking for large display text to create a high-impact, editorial feel. 

**JetBrains Mono** is introduced for labels, metadata, and technical data points to reinforce the "Alpha" technical nature of the brand.

**Fluid Scaling:** Headlines should use a "tight" vertical rhythm. Body copy requires a generous line height (at least 1.6x) to contrast against the dense, bold headlines. For display roles, use negative letter-spacing to create a "locked" visual block.

## Layout & Spacing

The layout follows a **Fixed Grid** model for content containers, centered within a fluid viewport. 

- **Vertical Rhythm:** Adopt a "Color-Block" approach where sections are separated by significant vertical gaps (120px+). 
- **Grid:** A 12-column system is used for desktop. 
- **Reflow:** On mobile, margins shrink to 24px and the 12-column grid collapses to a single-column stack. 
- **Storytelling:** Alternate between full-width immersive dark sections and constrained, white-background "chrome" sections to maintain high visual interest and pacing.

## Elevation & Depth

Depth is achieved through **Tonal Layers** and **Backdrop Blurs** rather than traditional drop shadows.

- **Level 0 (Base):** #0A0A0A (Void).
- **Level 1 (Card):** #1E1E1E with a 1px `border-glass` stroke.
- **Level 2 (Overlay):** Semi-transparent surfaces (rgba 255, 255, 255, 0.05) with a 20px backdrop-blur (saturate 150%).
- **Interactive Depth:** When an element is hovered, use a subtle "glow" effect—a box-shadow with 0px offset, 20px blur, and the `glow-cyan` color at low opacity.

## Shapes

The shape language balances precision with approachability. 

- **Primary Containers:** Use `rounded-lg` (1rem/16px) or larger (up to 24px) for main content cards and imagery to soften the high-contrast aesthetic.
- **Interactive Elements:** Buttons and tags must be **Pill-shaped** (full radius) to contrast against the rigid grid and provide a "tactile" feel.
- **Data Viz:** Use rounded caps on line charts and bar graphs to maintain the language.

## Components

- **Buttons:** Primary buttons are pill-shaped, using a solid `primary_color` (Cyan) with black text. Secondary buttons use a white ghost-border with white text.
- **Chips/Tags:** Small, pill-shaped, using `label-mono` typography. Background should be `border-glass` with a subtle 1px border.
- **Lists:** Clean, borderless entries separated by 1px horizontal lines at 10% white opacity. Use the Gold accent for bullet points or icons.
- **Input Fields:** Bottom-border only or very subtle ghost-box (4% white fill). Use Cyan for the active focus state and cursor.
- **Cards:** Large radius (24px), dark gray fill (#1E1E1E), and a faint top-down linear gradient (White @ 5% to Transparent) to simulate "overhead lighting."
- **Data Visualizations:** Minimalist line graphs using the Cyan glow. No heavy axes or grid lines—use the "monolith" approach where the data floats in the void.