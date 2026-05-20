import { SectionHeader } from "./SectionHeader";

type FlowItem = {
  label: string;
  icon: string;
};

type FooterNote = {
  text: string;
};

const configureItems: FlowItem[] = [
  { label: "Risk Profile", icon: "tune" },
  { label: "Return Target", icon: "track_changes" },
  { label: "Asset Universe", icon: "layers" },
];

const executeItems: FlowItem[] = [
  { label: "Portfolio Rebalancing", icon: "sync" },
  { label: "Risk-Adjusted Orders", icon: "gavel" },
  { label: "Live P&L Reporting", icon: "monitoring" },
];

const footerNote: FooterNote = {
  text: "No trading experience required. Just preferences — converted into alpha-seeking execution.",
};

export function ExecutionFlow() {
  return (
    <section className="px-margin-desktop max-w-[1440px] mx-auto w-full relative">
      <SectionHeader label="SYSTEMATIC EXECUTION FLOW" tone="primary" />
      <div className="flex flex-col lg:flex-row items-center justify-between bg-surface-container/40 border border-outline-variant/30 rounded-xl p-8 relative overflow-hidden backdrop-blur-sm">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-primary-container/10 via-background to-background pointer-events-none"></div>

        <div className="flex flex-col items-start w-full lg:w-1/4 z-10">
          <h4 className="font-label-caps text-label-caps text-secondary-container mb-6">
            YOU CONFIGURE
          </h4>
          <ul className="flex flex-col gap-4 w-full font-body-md text-body-md">
            {configureItems.map((item) => (
              <li
                key={item.label}
                className="flex items-center gap-3 bg-surface-container-high border border-outline-variant/40 p-3 rounded"
              >
                <span className="material-symbols-outlined text-primary-container text-sm">
                  {item.icon}
                </span>
                {item.label}
              </li>
            ))}
          </ul>
        </div>

        <div className="hidden lg:flex w-1/6 items-center justify-center relative h-32 z-10">
          <div className="w-full h-px bg-outline-variant/60 relative">
            <div className="absolute top-1/2 left-0 w-2 h-2 bg-primary-container rounded-full -translate-y-1/2 -translate-x-1/2 glow-cyan"></div>
            <div className="absolute top-1/2 right-0 w-2 h-2 bg-primary-container rounded-full -translate-y-1/2 translate-x-1/2 glow-cyan"></div>
            <div className="h-px bg-primary-container data-stream w-full absolute top-0 left-0"></div>
          </div>
        </div>

        <div className="flex flex-col items-center justify-center w-full lg:w-1/3 py-12 lg:py-0 z-10 relative">
          <div className="absolute w-64 h-64 border border-outline-variant/30 rounded-full flex items-center justify-center bg-surface-container-low/20">
            <div className="w-48 h-48 border border-primary-container/30 rounded-full flex items-center justify-center bg-primary-container/5">
              <div className="w-32 h-32 border border-secondary-container/40 rounded-full absolute bg-secondary-container/5"></div>
            </div>
          </div>
          <div className="w-24 h-24 bg-surface-container-highest border-2 border-primary-container rounded-full flex items-center justify-center z-20 glow-cyan shadow-xl">
            <span className="font-headline-xl text-headline-xl text-secondary-container text-glow-gold">
              α
            </span>
          </div>
          <div className="text-center mt-6 z-20">
            <h3 className="font-headline-lg text-headline-lg text-on-background">
              O(Alpha)
            </h3>
            <p className="font-data-sm text-data-sm text-on-surface-variant mt-2 max-w-[200px] text-center">
              Continuous Logic Processing
            </p>
          </div>
        </div>

        <div className="hidden lg:flex w-1/6 items-center justify-center relative h-32 z-10">
          <div className="w-full h-px bg-outline-variant/60 relative">
            <div className="absolute top-1/2 left-0 w-2 h-2 bg-primary-container rounded-full -translate-y-1/2 -translate-x-1/2 glow-cyan"></div>
            <div className="absolute top-1/2 right-0 w-2 h-2 bg-primary-container rounded-full -translate-y-1/2 translate-x-1/2 glow-cyan"></div>
            <div
              className="h-px bg-primary-container data-stream w-full absolute top-0 left-0"
              style={{ animationDelay: "1.5s" }}
            ></div>
          </div>
        </div>

        <div className="flex flex-col items-end w-full lg:w-1/4 z-10">
          <h4 className="font-label-caps text-label-caps text-secondary-container mb-6">
            AGENT EXECUTES
          </h4>
          <ul className="flex flex-col gap-4 w-full font-body-md text-body-md">
            {executeItems.map((item) => (
              <li
                key={item.label}
                className="flex items-center justify-end gap-3 bg-surface-container-high border border-outline-variant/40 p-3 rounded"
              >
                {item.label}
                <span className="material-symbols-outlined text-primary-container text-sm">
                  {item.icon}
                </span>
              </li>
            ))}
          </ul>
        </div>
      </div>
      <div className="text-center mt-8">
        <p className="font-data-sm text-data-sm text-secondary-container">
          {footerNote.text}
        </p>
      </div>
    </section>
  );
}
