export function Hero() {
  return (
    <section className="px-margin-desktop max-w-[1440px] mx-auto w-full relative">
      <div className="absolute top-8 right-1/4 w-80 h-80 bg-primary-container/10 rounded-full blur-3xl -z-10"></div>
      <div className="max-w-3xl">
        <div className="flex items-center gap-2 mb-4">
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
