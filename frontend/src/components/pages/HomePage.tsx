import { LandingShell } from "../layout/LandingShell";
import { HomeContent } from "./HomeContent";

export function HomePage() {
  return (
    <LandingShell activePath="/">
      <HomeContent />
    </LandingShell>
  );
}
