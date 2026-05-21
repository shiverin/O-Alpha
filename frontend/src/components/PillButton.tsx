import type { ButtonHTMLAttributes } from "react";

export type PillButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "solid" | "outline";
  size?: "sm" | "md";
};

const sizeClasses: Record<NonNullable<PillButtonProps["size"]>, string> = {
  sm: "px-4 py-1.5 text-body-sm",
  md: "px-6 py-2 text-body-md",
};

const variantClasses: Record<NonNullable<PillButtonProps["variant"]>, string> = {
  solid:
    "bg-primary-container text-on-primary-container shadow-[0_10px_24px_-14px_rgba(0,229,255,0.9)] hover:bg-primary-fixed hover:shadow-[0_12px_28px_-14px_rgba(0,229,255,0.95)]",
  outline:
    "border border-outline-variant/60 text-on-surface bg-transparent hover:border-primary-container/70 hover:text-primary-container hover:bg-primary-container/10",
};

export function PillButton({
  className = "",
  variant = "solid",
  size = "md",
  ...props
}: PillButtonProps) {
  return (
    <button
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
  );
}
