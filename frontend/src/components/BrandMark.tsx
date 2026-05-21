type BrandMarkProps = {
  className?: string;
  logoSize?: "sm" | "md";
};

const logoSizeClasses: Record<NonNullable<BrandMarkProps["logoSize"]>, string> = {
  sm: "h-9 w-9",
  md: "h-11 w-11",
};

export function BrandMark({ className = "", logoSize = "md" }: BrandMarkProps) {
  return (
    <span
      className={`inline-flex items-center gap-1 ${className}`}
      aria-label="O(Alpha)"
    >
      <img
        alt="O(Alpha) Brand Mark"
        className={`${logoSizeClasses[logoSize]} shrink-0`}
        src="/brand-mark.png"
      />
      <span className="font-headline-lg text-headline-lg font-semibold tracking-tight leading-none">
        <span className="text-secondary-container">O</span>
        <span className="text-primary-container">(Alpha)</span>
      </span>
    </span>
  );
}
