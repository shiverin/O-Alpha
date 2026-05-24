import { forwardRef } from "react";

import type { InputHTMLAttributes } from "react";

export type InputProps = InputHTMLAttributes<HTMLInputElement> & {
  size?: "sm" | "md";
};

const sizeClasses: Record<NonNullable<InputProps["size"]>, string> = {
  sm: "h-10 px-3",
  md: "h-12 px-4",
};

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className = "", size = "md", ...props }, ref) => (
    <input
      ref={ref}
      className={[
        "flex h-[calc(1.5rem+1px)] items-center justify-between rounded-md border border-input-bg bg-transparent px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:focus-visible:ring-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 [&::-webkit-search-button]:appearance-none",
        sizeClasses[size as keyof typeof sizeClasses],
        className,
      ].join(" ")}
      {...props}
    />
  )
);

Input.displayName = "Input";