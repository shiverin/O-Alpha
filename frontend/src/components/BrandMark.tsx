import Image from "next/image";

type BrandMarkProps = {
  className?: string;
  logoSize?: "sm" | "md";
  showText?: boolean;
};

const logoSizeClasses: Record<
  NonNullable<BrandMarkProps["logoSize"]>,
  string
> = {
  sm: "h-9 w-9",
  md: "h-11 w-11",
};

const logoSizePixels: Record<
  NonNullable<BrandMarkProps["logoSize"]>,
  number
> = {
  sm: 36,
  md: 44,
};

export function BrandMark({
  className = "",
  logoSize = "md",
  showText = true,
}: BrandMarkProps) {
  return (
    <span
      className={`inline-flex items-center gap-1 ${className}`}
      aria-label="O(Alpha)"
    >
      <Image
        alt="O(Alpha) Brand Mark"
        className={`${logoSizeClasses[logoSize]} shrink-0`}
        height={logoSizePixels[logoSize]}
        src="/brand-mark.png"
        width={logoSizePixels[logoSize]}
      />
      {showText && (
        <span className="font-headline-lg text-headline-lg font-semibold tracking-tight leading-none">
          <span className="text-secondary-container">O</span>
          <span className="text-primary-container">(Alpha)</span>
        </span>
      )}
    </span>
  );
}
