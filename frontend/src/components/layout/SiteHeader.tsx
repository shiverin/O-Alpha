import { BrandMark } from "../BrandMark";
import { PillButton } from "../PillButton";

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
            <BrandMark />
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
          <PillButton variant="outline" size="sm">
            Login
          </PillButton>
          <PillButton className="scale-95 active:scale-90" size="sm">
            Launch App
          </PillButton>
        </div>
      </div>
    </nav>
  );
}
