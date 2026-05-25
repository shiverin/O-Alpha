"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { isAuthenticated } from "@/lib/auth";

export default function AppLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const authed = isAuthenticated();
    if (!authed) {
      router.replace("/login");
      return;
    }
    setReady(true);
  }, [router]);

  if (!ready) {
    return (
      <div className="min-h-screen bg-background text-on-background flex items-center justify-center">
        <span className="font-body-md text-body-md text-on-surface-variant">
          Loading...
        </span>
      </div>
    );
  }

  return <div className="min-h-screen bg-background text-on-background">{children}</div>;
}
