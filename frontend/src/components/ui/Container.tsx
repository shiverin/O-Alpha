import React from "react";

interface ContainerProps {
  children: React.ReactNode;
  className?: string;
  fluid?: boolean;
  px?: "mobile" | "desktop" | "none";
  maxWidth?: string | number;
}

/**
 * Layout container component for consistent horizontal spacing and max-width
 *
 * Defaults: px-margin-desktop max-w-[1440px] mx-auto w-full
 *
 * Props:
 *   fluid: if true, removes max-width constraint (w-full only)
 *   px: controls horizontal padding ('mobile', 'desktop', or 'none')
 *   maxWidth: custom max-width value (overrides default when not fluid)
 */
export const Container = ({
  children,
  className = "",
  fluid = false,
  px = "desktop",
  maxWidth = fluid ? undefined : "[1440px]",
}: ContainerProps) => {
  // Padding classes
  let pxClass = "";
  switch (px) {
    case "mobile":
      pxClass = "px-margin-mobile";
      break;
    case "desktop":
      pxClass = "px-margin-desktop";
      break;
    case "none":
      pxClass = "px-0";
      break;
  }

  // Width/max-width classes
  let widthClass = "w-full";
  if (!fluid && maxWidth) {
    widthClass += ` max-w-${typeof maxWidth === "number" ? `[${maxWidth}px]` : maxWidth}`;
  }

  // Centering (only when not fluid and has max-width)
  const mxClass = !fluid ? "mx-auto" : "";

  return (
    <div className={`${pxClass} ${widthClass} ${mxClass} ${className}`}>
      {children}
    </div>
  );
};
