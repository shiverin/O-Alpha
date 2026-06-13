import Link from "next/link";
import { BrandMark } from "../BrandMark";

export function SiteFooter() {
  return (
    <footer className="w-full py-16 bg-surface-container border-t border-outline-variant/40 relative z-10">
      <div className="px-margin-desktop max-w-[1440px] mx-auto">
        <Link href="/" className="flex items-center gap-4 mb-4">
          <BrandMark />
        </Link>
        <span className="font-data-sm text-data-sm text-on-surface-variant block mt-8">
          © 2026 Orbital. All rights reserved.
        </span>
      </div>
    </footer>
  );
}
