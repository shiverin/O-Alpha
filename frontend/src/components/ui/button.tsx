import { forwardRef } from "react";

import type { ButtonHTMLAttributes } from "react";

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "solid" | "outline";
  size?: "sm" | "md";
};

const sizeClasses: Record<NonNullable<ButtonProps["size"]>, string> = {
  sm: "px-4 py-1.5 text-body-sm",
  md: "px-6 py-2 text-body-md",
};

const variantClasses: Record<NonNullable<ButtonProps["variant"]>, string> = {
  solid:
    "bg-primary-container text-on-primary-container shadow-[0_8px_20px_-12px_rgba(0,213,255,0.6)] hover:bg-primary-fixed hover:shadow-[0_10px_24px_-12px_rgba(0,213,255,0.7)]",
  outline:
    "border border-outline-variant/70 text-on-surface bg-transparent hover:border-primary-container/70 hover:text-primary-container hover:bg-primary-container/10",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className = "", variant = "solid", size = "md", ...props }, ref) => (
    <button
      ref={ref}
      className={[
        "relative inline-flex items-center justify-center gap-2 rounded-full font-body-md font-medium transition-all duration-200",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-container/60 focus-visible:ring-offset-2 focus-visible:ring-offset-background",
        "active:translate-y-[1px]",
        sizeClasses[size],
        variantClasses[variant],
        className,
      ].join(" ")}
      {...props}
    />
  )
);

Button.displayName = "Button";