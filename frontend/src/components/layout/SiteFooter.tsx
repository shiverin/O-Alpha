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
    <footer className="w-full py-16 bg-surface-container-lowest border-t border-outline-variant/40 relative z-10">
      <div className="grid grid-cols-1 md:grid-cols-4 gap-gutter px-margin-desktop max-w-[1440px] mx-auto">
        <div className="col-span-1 md:col-span-1">
          <a className="flex items-center gap-4 mb-4" href="#">
            <img
              alt="O(Alpha) Brand Mark"
              className="h-12 w-auto"
              src="https://lh3.googleusercontent.com/aida/ADBb0uiLcVGEmFIldOXt-b8Aw8wcEKHbR2dL9OpZyqR_nRl52WinVpztI50GXzZVQeuO4-zboFLhA-5Ho6PMnvokDCTFMsjixlVK4YCDHFjdSvGVCir5NZKVGiusSGvv9NcX8Cu97Sno2kjvdwxIFvJnNmO4_UAfkgts9MACekYch70mSHYxLMTB9yVJJlgjox1LDGfkljNIkP8INxZq0VDXBSZ8tV2Mx5geHlBMe7bbMfM3Q0uff9feqwcHNK4"
            />
            <span className="font-headline-lg text-headline-lg text-secondary-fixed font-bold tracking-tight">
              O(Alpha)
            </span>
          </a>
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
