import { LandingShell } from "../layout/LandingShell";

const configureItems = [
  { icon: "tune", text: "Risk profile" },
  { icon: "my_location", text: "Return target" },
  { icon: "layers", text: "Asset universe" },
] as const;

const executeItems = [
  { icon: "sync", text: "Portfolio rebalancing" },
  { icon: "monitoring", text: "Regime detection" },
  { icon: "data_exploration", text: "Risk-adjusted orders" },
] as const;

const teamMembers = [
  {
    name: "Shizhen",
    role: "QUANTITATIVE RESEARCH",
    roleTone: "text-secondary-container",
    icon: "terminal",
    accent: "border-l-primary-container",
    summary:
      "Focusing on statistical arbitrage models, regime detection algorithms, and translating complex market dynamics into executable, risk-managed logic streams.",
    tags: ["Python", "Data Science"],
  },
  {
    name: "Jia Jun",
    role: "SYSTEMS ENGINEERING",
    roleTone: "text-primary-container",
    icon: "dns",
    accent: "border-l-secondary-container",
    summary:
      "Architecting the low-latency execution environment, ensuring robust API integrations, and building the seamless interface that connects user intent to market action.",
    tags: ["Infrastructure", "Frontend"],
  },
] as const;

export function MissionPage() {
  return (
    <LandingShell activePath="/mission" className="bg-mission-grid">
      <div className="ambient-glow"></div>
      <div className="ambient-glow-secondary"></div>
      <main className="pt-32 pb-24">
        <section className="max-w-[1440px] mx-auto px-margin-mobile md:px-margin-desktop mb-32 relative">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-gutter items-center min-h-[614px]">
            <div className="flex flex-col gap-6 z-10">
              <div className="font-label-caps text-label-caps text-primary-container uppercase tracking-widest">
                Orbital 2026
              </div>
              <h1 className="font-headline-xl text-headline-xl text-on-background leading-tight">
                Eliminating Emotion <br />
                <span className="text-primary-container">from Execution.</span>
              </h1>
              <p className="font-body-md text-body-md text-on-surface-variant max-w-xl mt-4">
                O(Alpha) was born from a simple thesis: human emotion is the
                primary drag on retail trading performance. We are building a
                future where sophisticated, quant-level systematic trading is
                accessible to anyone capable of defining their preferences.
              </p>
              <div className="flex items-center gap-4 mt-8 font-data-sm text-data-sm text-outline">
                <span>VERSION: 1.0.0-ALPHA</span>
                <span className="w-1 h-1 rounded-full bg-outline-variant"></span>
                <span>
                  STATUS: <span className="text-secondary-container">LIVE DEVELOPMENT</span>
                </span>
              </div>
            </div>
            <div className="relative h-[400px] flex justify-center items-center lg:justify-end opacity-80 z-0">
              <div className="absolute w-64 h-64 border border-outline-variant rounded-full flex items-center justify-center">
                <div className="absolute w-48 h-48 border border-primary-container/30 rounded-full"></div>
                <div className="w-32 h-32 bg-surface-container-lowest border border-primary-container rounded-full flex items-center justify-center shadow-[0_0_30px_rgba(0,229,255,0.2)]">
                  <span className="font-headline-xl text-headline-xl text-secondary-container">
                    α
                  </span>
                </div>
                <div className="absolute top-10 right-0 h-[1px] w-32 bg-outline-variant/50">
                  <div className="absolute right-0 -top-4 font-data-sm text-data-sm text-on-surface-variant text-[10px]">
                    SYSTEM ACTIVE
                  </div>
                  <div className="absolute left-0 -top-[2px] w-8 h-[2px] bg-primary-container"></div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="max-w-[1440px] mx-auto px-margin-mobile md:px-margin-desktop mb-32">
          <div className="mb-16">
            <h2 className="font-headline-lg-mobile md:font-headline-lg text-headline-lg-mobile md:text-headline-lg text-on-background">
              Our Vision 
            </h2>
            <p className="font-body-md text-body-md text-on-surface-variant max-w-2xl mt-4">
              A seamless translation layer between human intent and algorithmic
              precision. No trading experience required. Just preferences -
              converted into alpha-seeking execution.
            </p>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-0 relative">
            <div className="hidden md:block absolute top-1/2 left-0 w-full h-[1px] bg-outline-variant/30 -z-10"></div>
            <div className="glass-panel p-8 md:mr-gutter relative">
              <div className="font-label-caps text-label-caps text-secondary-container mb-8">
                YOU CONFIGURE
              </div>
              <ul className="flex flex-col gap-6 font-data-lg text-data-lg text-on-background">
                {configureItems.map((item) => (
                  <li key={item.text} className="flex items-center gap-4">
                    <span className="material-symbols-outlined text-primary-container text-xl">
                      {item.icon}
                    </span>
                    {item.text}
                  </li>
                ))}
              </ul>
            </div>
            <div className="flex flex-col justify-center items-center py-12 md:py-0 relative">
              <div className="absolute h-full w-[1px] bg-outline-variant/30 hidden md:block"></div>
              <div className="bg-surface-container-lowest border border-outline-variant/50 rounded px-6 py-4 flex flex-col items-center gap-2 z-10 shadow-[0_0_20px_rgba(0,0,0,0.5)]">
                <span className="font-headline-lg text-headline-lg text-secondary-container">
                  α
                </span>
                <span className="font-data-sm text-data-sm text-on-surface-variant tracking-widest uppercase">
                  Translation Layer
                </span>
              </div>
            </div>
            <div className="glass-panel p-8 md:ml-gutter relative">
              <div className="font-label-caps text-label-caps text-secondary-container mb-8">
                AGENT EXECUTES
              </div>
              <ul className="flex flex-col gap-6 font-data-lg text-data-lg text-on-background">
                {executeItems.map((item) => (
                  <li key={item.text} className="flex items-center gap-4">
                    <span className="material-symbols-outlined text-outline text-xl">
                      {item.icon}
                    </span>
                    {item.text}
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </section>

        <section className="max-w-[1440px] mx-auto px-margin-mobile md:px-margin-desktop">
          <div className="mb-12">
            <div className="font-label-caps text-label-caps text-primary-container mb-2">
              SYSTEM ARCHITECTS
            </div>
            <h2 className="font-headline-lg-mobile md:font-headline-lg text-headline-lg-mobile md:text-headline-lg text-on-background">
              The Team
            </h2>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-gutter">
            {teamMembers.map((member) => (
              <div
                key={member.name}
                className={`glass-panel p-8 border-l-2 ${member.accent} hover:bg-surface-container-high/50 transition-colors duration-300`}
              >
                <div className="flex justify-between items-start mb-6">
                  <div>
                    <h3 className="font-headline-lg-mobile text-headline-lg-mobile text-on-background mb-1">
                      {member.name}
                    </h3>
                    <div className={`font-data-sm text-data-sm ${member.roleTone}`}>
                      {member.role}
                    </div>
                  </div>
                  <span className="material-symbols-outlined text-outline-variant">
                    {member.icon}
                  </span>
                </div>
                <p className="font-body-md text-body-md text-on-surface-variant">
                  {member.summary}
                </p>
                <div className="mt-8 flex gap-2">
                  {member.tags.map((tag) => (
                    <span
                      key={tag}
                      className="px-2 py-1 border border-outline-variant/50 rounded font-data-sm text-data-sm text-on-surface-variant"
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </section>
      </main>
    </LandingShell>
  );
}
