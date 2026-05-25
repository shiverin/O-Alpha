"use client";

import { LandingShell } from "@/components/layout/LandingShell";
import { HomeContent } from "@/components/pages/HomeContent";
import { useSearchParams } from "next/navigation";
import { useState, useEffect } from "react";

export function HomePageClient() {
  const searchParams = useSearchParams();
  const [showLoginModal, setShowLoginModal] = useState(false);

  useEffect(() => {
    if (searchParams.get("auth") === "required") {
      setShowLoginModal(true);
    }
  }, [searchParams]);

  return (
    <LandingShell activePath="/" loginModalOpen={showLoginModal} onLoginModalOpenChange={setShowLoginModal}>
      <HomeContent />
    </LandingShell>
  );
}
