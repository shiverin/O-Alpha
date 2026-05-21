type SectionHeaderProps = {
  label: string;
  tone?: "primary" | "secondary";
};

export function SectionHeader({ label, tone = "secondary" }: SectionHeaderProps) {
  const colorClass =
    tone === "primary" ? "text-primary-container" : "text-secondary-container";

  return (
    <div className="mb-8 border-b border-outline-variant/40 pb-2">
      <h2 className={`font-label-caps text-label-caps ${colorClass} tracking-widest`}>
        {label}
      </h2>
    </div>
  );
}
