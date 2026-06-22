"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { useAuth } from "@/context/AuthContext";

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const { user, loading } = useAuth();

  useEffect(() => {
    if (loading) return;
    if (!user) {
      router.replace("/?auth=required");
    }
  }, [loading, router, user]);

  if (loading || !user) {
    return (
      <div className="min-h-screen bg-background text-on-background flex items-center justify-center">
        <span className="font-body-md text-body-md text-on-surface-variant">
          Loading...
        </span>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background text-on-background">
      {children}
    </div>
  );
}
