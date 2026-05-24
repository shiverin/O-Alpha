import Link from "next/link";
import { BrandMark } from "../BrandMark";

type FooterLink = {
  label: string;
  href: string;
};

const footerLinks: FooterLink[] = [
  { label: "Privacy Policy", href: "#" },
  { label: "Terms of Service", href: "#" },
  { label: "Security", href: "#" },
  { label: "API Documentation", href: "#" },
];

export function SiteFooter() {
  return (
    <footer className="w-full py-16 bg-surface-container border-t border-outline-variant/40 relative z-10">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-gutter px-margin-desktop max-w-[1440px] mx-auto">
        <div className="col-span-1 md:col-span-1">
          <Link href="/" className="flex items-center gap-4 mb-4">
            <BrandMark />
          </Link>
          <span className="font-data-sm text-data-sm text-on-surface-variant block mt-8">
            © 2026 Orbital. All rights reserved.
          </span>
        </div>
        <div className="col-span-1 md:col-span-3 flex flex-col md:flex-row justify-end gap-8 md:gap-16 pt-2">
          {footerLinks.map((link) => (
            <a
              key={link.label}
              className="font-data-sm text-data-sm text-on-surface-variant hover:text-secondary-fixed transition-colors opacity-80 hover:opacity-100"
              href={link.href}
            >
              {link.label}
            </a>
          ))}
        </div>
      </div>
    </footer>
  );
}