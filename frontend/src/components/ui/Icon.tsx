import React from "react";

interface IconProps {
  name: string;
  size?: "small" | "medium" | "large" | number;
  color?: string;
  className?: string;
  title?: string;
  /** @deprecated Use color prop instead */
  variant?: "primary" | "secondary" | "tertiary";
}

/**
 * Icon wrapper for Material Symbols Outlined icons
 *
 * Usage:
 *   <Icon name="print" size="medium" color="text-primary-container" />
 *
 * Sizes:
 *   small: 20px
 *   medium: 24px (default)
 *   large: 32px
 *   or pass a number for custom size in pixels
 */
export const Icon = ({
  name,
  size,
  color,
  className = "",
  title,
  variant,
}: IconProps) => {
  // Default size
  let sizeClass = "text-base";
  if (typeof size === "string") {
    switch (size) {
      case "small":
        sizeClass = "text-sm";
        break;
      case "medium":
        sizeClass = "text-base";
        break;
      case "large":
        sizeClass = "text-xl";
        break;
    }
  } else if (typeof size === "number") {
    // Custom size in pixels
    sizeClass = `text-[${size}px]`;
  }

  // Color from variant (deprecated) or color prop
  let colorClass = "";
  if (variant) {
    // Deprecated variant prop
    switch (variant) {
      case "primary":
        colorClass = "text-primary-container";
        break;
      case "secondary":
        colorClass = "text-secondary-container";
        break;
      case "tertiary":
        colorClass = "text-tertiary-container";
        break;
    }
  } else if (color) {
    colorClass = color;
  } else {
    // Default color
    colorClass = "text-primary-container";
  }

  return (
    <span
      className={`material-symbols-outlined ${sizeClass} ${colorClass} ${className}`}
      aria-hidden="true"
      role="img"
      {...(title && { "aria-label": title })}
    >
      {name}
    </span>
  );
};
