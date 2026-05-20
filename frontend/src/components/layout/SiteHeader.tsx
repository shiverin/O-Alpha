type NavLink = {
  label: string;
  href: string;
  active?: boolean;
};

type SiteHeaderProps = {
  activePath?: string;
};

const navLinks: NavLink[] = [
  { label: "Product", href: "/" },
  { label: "Performance", href: "/performance" },
  { label: "Pricing", href: "/pricing" },
  { label: "Mission", href: "/mission" },
];

export function SiteHeader({ activePath }: SiteHeaderProps) {
  const links = navLinks.map((link) => ({
    ...link,
    active: activePath ? link.href === activePath : link.active,
  }));

  return (
    <nav className="fixed top-0 w-full z-50 bg-background/80 backdrop-blur-xl border-b border-outline-variant/30">
      <div className="flex justify-between items-center px-margin-desktop py-4 max-w-[1440px] mx-auto">
        <div className="flex items-center gap-12">
          <a className="flex items-center gap-3" href="/">
            <img
              alt="O(Alpha) Brand Mark"
              className="h-10 w-auto"
              src="https://lh3.googleusercontent.com/aida/ADBb0uiLcVGEmFIldOXt-b8Aw8wcEKHbR2dL9OpZyqR_nRl52WinVpztI50GXzZVQeuO4-zboFLhA-5Ho6PMnvokDCTFMsjixlVK4YCDHFjdSvGVCir5NZKVGiusSGvv9NcX8Cu97Sno2kjvdwxIFvJnNmO4_UAfkgts9MACekYch70mSHYxLMTB9yVJJlgjox1LDGfkljNIkP8INxZq0VDXBSZ8tV2Mx5geHlBMe7bbMfM3Q0uff9feqwcHNK4"
            />
            <span className="font-headline-lg text-headline-lg font-bold text-secondary-container tracking-tight">
              O(Alpha)
            </span>
          </a>
          <div className="hidden md:flex gap-8 items-center pt-1">
            {links.map((link) => (
              <a
                key={link.label}
                className={
                  link.active
                    ? "font-body-md text-body-md text-primary-container border-b-2 border-primary-container pb-1 transition-colors duration-300"
                    : "font-body-md text-body-md text-on-surface hover:text-primary-container transition-colors duration-300"
                }
                href={link.href}
              >
                {link.label}
              </a>
            ))}
          </div>
        </div>
        <div className="hidden md:flex items-center gap-3">
          <button className="border border-outline-variant/60 text-on-surface px-5 py-2 rounded font-body-md text-body-md font-medium transition-colors hover:text-primary-container hover:border-primary-container/70">
            Login
          </button>
          <button className="bg-primary-container text-on-primary-container px-6 py-2 rounded font-body-md text-body-md font-medium scale-95 active:scale-90 transition-transform hover:bg-primary-fixed">
            Launch App
          </button>
        </div>
      </div>
    </nav>
  );
}
