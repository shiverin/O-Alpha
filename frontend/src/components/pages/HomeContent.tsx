import { ExecutionFlow } from "../sections/ExecutionFlow";
import { FeatureGrid } from "../sections/FeatureGrid";
import { Hero } from "../sections/Hero";

export function HomeContent() {
  return (
    <main className="flex-grow pb-24 flex flex-col gap-24 relative z-10">
      <Hero />
      <FeatureGrid />
      <ExecutionFlow />
    </main>
  );
}
