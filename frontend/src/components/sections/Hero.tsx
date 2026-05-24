import { Container } from '@/components/ui/Container';

export function Hero() {
  return (
    <section className="relative min-h-screen flex items-center">
      <Container>
        <div className="absolute top-8 right-1/4 w-80 h-80 bg-primary-container/10 rounded-full blur-3xl -z-10"></div>
        <div className="max-w-3xl">
          <div className="flex items-center gap-2 mb-4">
          </div>
          <h1 className="font-headline-xl text-headline-xl text-on-background mb-6">
            <span className="text-primary-container">Build</span> Your Own Trading Agent.
          </h1>
          <p className="font-body-md text-body-md text-on-surface-variant text-lg max-w-2xl">
              O(Alpha) translates your strategic intent into autonomous execution. High-frequency capability meets institutional-grade intelligence, now in your control.
          </p>
        </div>
      </Container>
    </section>
  );
}
