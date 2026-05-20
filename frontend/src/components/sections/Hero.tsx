export function Hero() {
  return (
    <section className="px-margin-desktop max-w-[1440px] mx-auto w-full relative">
      <div className="absolute top-0 right-1/4 w-96 h-96 bg-primary-container/10 rounded-full blur-[120px] -z-10"></div>
      <div className="max-w-3xl">
        <div className="flex items-center gap-2 mb-4">
          <span className="font-data-sm text-data-sm text-primary-container border border-primary-container/40 bg-primary-container/15 px-2 py-0.5 rounded">
            BUILD YOUR AGENT
          </span>
          <span className="font-data-sm text-data-sm text-outline-variant">
            /// PERSONALIZED EXECUTION
          </span>
        </div>
        <h1 className="font-headline-xl text-headline-xl text-on-background mb-6">
          Build Your Own Trading Agent.
        </h1>
        <p className="font-body-md text-body-md text-on-surface-variant text-lg max-w-2xl">
          Design a strategy that trades the way you think. Tune risk, define
          targets, and let O(Alpha) turn your preferences into a live, automated
          agent that executes with quant-grade precision around the clock.
        </p>
      </div>
    </section>
  );
}
