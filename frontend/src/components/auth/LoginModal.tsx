"use client";

import { useEffect, useMemo, useState } from "react";
import { createPortal } from "react-dom";
import { useRouter } from "next/navigation";
import { Panel } from '@/components/ui/Panel';
import { Icon } from '@/components/ui/Icon';

type LoginModalProps = {
  isOpen: boolean;
  onClose: () => void;
  // FIXED: Added redirectPath as an optional string
  redirectPath?: string; 
};

export function LoginModal({ isOpen, onClose, redirectPath = "/app/dashboard" }: LoginModalProps) {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [mounted, setMounted] = useState(false);

  const canSubmit = useMemo(() => email.length > 0 && password.length > 0, [email, password]);

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (!isOpen) {
      return;
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      }
    };

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    window.addEventListener("keydown", handleKeyDown);

    return () => {
      document.body.style.overflow = previousOverflow;
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [isOpen, onClose]);

  if (!isOpen || !mounted) {
    return null;
  }

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!canSubmit) {
      return;
    }

    localStorage.setItem("oa-auth", "true");
    onClose();
    router.push(redirectPath);
  };

  const handleBypass = () => {
    localStorage.setItem("oa-auth", "true");
    onClose();
    router.push(redirectPath);
  };

  return createPortal(
    <div className="fixed inset-0 z-[100] flex items-center justify-center px-margin-mobile">
      <button
        className="absolute inset-0 h-full w-full bg-background/70 backdrop-blur-sm"
        type="button"
        aria-label="Close login"
        onClick={onClose}
      />
      {/* Changed max-w-md to max-w-sm to scale the entire box down */}
      <Panel className="z-10 w-full max-w-sm overflow-hidden rounded-[24px] border border-outline-variant/40 bg-surface-container-high/90 shadow-[0_20px_50px_rgba(0,0,0,0.55)]">
        <div className="pointer-events-none absolute left-0 right-0 top-0 h-28 bg-gradient-to-b from-white/5 to-transparent" />

        {/* Added z-20 to ensure the close button is clickable above the inner content */}
        <button
          className="absolute right-4 top-4 z-20 inline-flex h-9 w-9 items-center justify-center rounded-full border border-outline-variant/40 text-on-surface-variant transition-colors hover:bg-surface-container-high hover:text-on-background"
          type="button"
          onClick={onClose}
          aria-label="Cancel login"
        >
          <Icon name="close" size="small" color="text-on-surface-variant" />
        </button>

        {/* Reduced padding from p-8 sm:p-10 to p-6 sm:p-8 */}
        <div className="relative z-10 flex flex-col items-center p-6 sm:p-8">
          <div className="mb-6 flex flex-col items-center text-center">
            {/* Reduced icon container size slightly */}
            {/* <div className="mb-4 flex h-14 w-14 items-center justify-center rounded-xl border border-outline-variant/40 bg-surface-container-highest shadow-[0_0_20px_rgba(0,213,255,0.15)]">
              <span
                className="material-symbols-outlined text-3xl text-primary-container"
                style={{ fontVariationSettings: "'FILL' 1" }}
              >
                security
              </span>
            </div> */}
            {/* Tweaked text sizing/margins for the smaller container */}
            <h1 className="mb-1 text-2xl font-bold text-on-background">
              Log In
            </h1>
            <p className="font-data-sm text-data-sm text-on-surface-variant">

            </p>
          </div>
          <form className="w-full space-y-5" onSubmit={handleSubmit}>
            <div className="space-y-3">
              <div className="group relative">
                <Icon name="badge" className="absolute left-4 top-1/2 -translate-y-1/2 text-on-surface-variant transition-colors group-focus-within:text-primary-container" />
                {/* Reduced vertical padding (py-4 to py-3) */}
                <input
                  className="w-full rounded-t-lg border-x-0 border-b border-t-0 border-outline-variant/60 bg-surface-container-low py-3 pl-12 pr-4 font-body-md text-on-background transition-colors focus:border-primary-container focus:bg-surface-container-highest focus:ring-0"
                  placeholder="Username"
                  type="text"
                  value={email}
                  onChange={(event) => setEmail(event.target.value)}
                  required
                />
              </div>
              <div className="group relative">
                <Icon name="key" className="absolute left-4 top-1/2 -translate-y-1/2 text-on-surface-variant transition-colors group-focus-within:text-primary-container" />
                {/* Reduced vertical padding (py-4 to py-3) */}
                <input
                  className="w-full rounded-t-lg border-x-0 border-b border-t-0 border-outline-variant/60 bg-surface-container-low py-3 pl-12 pr-12 font-body-md text-on-background transition-colors focus:border-primary-container focus:bg-surface-container-highest focus:ring-0"
                  placeholder="Password"
                  type="password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  required
                />
                <Icon name="visibility_off" className="absolute right-4 top-1/2 -translate-y-1/2 text-on-surface-variant" />
              </div>
            </div>

            <div className="flex items-center justify-between gap-4 text-data-sm text-on-surface-variant">
              <label className="inline-flex items-center gap-2 cursor-pointer">
                <input
                  className="h-4 w-4 rounded border-outline-variant/60 bg-transparent text-primary-container focus:ring-primary-container"
                  type="checkbox"
                />
                Remember Me
              </label>
              <button className="text-primary-container transition-colors hover:text-white" type="button">
                Reset Password
              </button>
            </div>

            {/* Reduced button vertical padding (py-4 to py-3) */}
            <button
              className="w-full rounded-full bg-primary-container px-8 py-3 text-base font-semibold text-background transition-transform duration-200 hover:scale-[1.02]"
              type="submit"
            >
              Start
            </button>

            <div className="flex items-center gap-4 text-data-sm text-on-surface-variant">
              <div className="h-px flex-1 bg-outline-variant/50" />
              <span>OR</span>
              <div className="h-px flex-1 bg-outline-variant/50" />
            </div>

            {/* Reduced button vertical padding (py-4 to py-3) */}
            <button
              className="flex w-full items-center justify-center gap-2 rounded-full border border-outline-variant/60 px-8 py-3 text-base font-medium text-on-background transition-colors hover:bg-surface-container-high"
              type="button"
              onClick={handleBypass}
            >
              <Icon name="login" size="small" color="text-primary-container" />
              Skip login
            </button>
          </form>
        </div>
      </Panel>
    </div>,
    document.body,
  );
}