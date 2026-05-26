"use client";

import { useEffect, useMemo, useState } from "react";
import { createPortal } from "react-dom";
import { useRouter } from "next/navigation";
import { Panel } from "@/components/ui/Panel";
import { Icon } from "@/components/ui/Icon";
import { api } from "@/lib/api";
import { setToken } from "@/lib/auth";

type LoginModalProps = {
  isOpen: boolean;
  onClose: () => void;
  redirectPath?: string;
};

export function LoginModal({ isOpen, onClose, redirectPath }: LoginModalProps) {
  const router = useRouter();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [mounted, setMounted] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const canSubmit = useMemo(
    () => username.length > 0 && password.length > 0,
    [username, password],
  );

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    window.addEventListener("keydown", handleKeyDown);

    return () => {
      document.body.style.overflow = previousOverflow;
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [isOpen, onClose]);

  if (!isOpen || !mounted) return null;

  const createLocalDemoToken = (): string => {
    const encode = (value: object) =>
      btoa(JSON.stringify(value))
        .replace(/\+/g, "-")
        .replace(/\//g, "_")
        .replace(/=+$/g, "");

    const now = Math.floor(Date.now() / 1000);
    const header = encode({ alg: "HS256", typ: "JWT" });
    const payload = encode({
      user_id: 999,
      username: "demouser",
      exp: now + 60 * 60 * 24,
      iat: now,
    });

    return `${header}.${payload}.offline-demo-signature`;
  };

  const isBackendUnavailable = (err: unknown): boolean => {
    if (err instanceof TypeError) {
      return true;
    }
    return (
      err instanceof Error &&
      /Failed to fetch|NetworkError|Request failed \(5\d\d\)/i.test(err.message)
    );
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!canSubmit) return;

    setLoading(true);
    setError(null);

    try {
      const response = await api.post<{
        token: string;
        user: { id: number; username: string };
      }>("/auth/login", {
        username,
        password,
      });

      setToken(response.token);
      router.push(redirectPath || "/app/dashboard");
      onClose();
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Login failed. Please check your credentials.");
      }
    } finally {
      setLoading(false);
    }
  };

  const handleBypass = async () => {
    setLoading(true);
    setError(null);

    const demoTarget = "/app/dashboard";

    try {
      setToken(createLocalDemoToken());
      router.push(demoTarget);
      onClose();
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Demo login failed.");
      }
    } finally {
      setLoading(false);
    }
  };

  return createPortal(
    <div className="fixed inset-0 z-[100] flex items-center justify-center px-margin-mobile">
      <button
        className="absolute inset-0 h-full w-full bg-background/70 backdrop-blur-sm"
        type="button"
        aria-label="Close login"
        onClick={onClose}
      />
      <Panel className="z-10 w-full max-w-sm overflow-hidden rounded-[24px] border border-outline-variant/40 bg-surface-container-high/90 shadow-[0_20px_50px_rgba(0,0,0,0.55)]">
        <div className="pointer-events-none absolute left-0 right-0 top-0 h-28 bg-gradient-to-b from-white/5 to-transparent" />

        <button
          className="absolute right-4 top-4 z-20 inline-flex h-9 w-9 items-center justify-center rounded-full border border-outline-variant/40 text-on-surface-variant transition-colors hover:bg-surface-container-high hover:text-on-background"
          type="button"
          onClick={onClose}
          aria-label="Cancel login"
        >
          <Icon name="close" size="small" color="text-on-surface-variant" />
        </button>

        <div className="relative z-10 flex flex-col items-center p-6 sm:p-8">
          <div className="mb-6 flex flex-col items-center text-center">
            <h1 className="mb-1 text-2xl font-bold text-on-background">
              Log In
            </h1>
          </div>

          <form className="w-full space-y-5" onSubmit={handleSubmit}>
            <div className="space-y-3">
              <div className="group relative">
                <Icon
                  name="badge"
                  className="absolute left-4 top-1/2 -translate-y-1/2 text-on-surface-variant transition-colors group-focus-within:text-primary-container"
                />
                <input
                  className="w-full rounded-t-lg border-x-0 border-b border-t-0 border-outline-variant/60 bg-surface-container-low py-3 pl-12 pr-4 font-body-md text-on-background transition-colors focus:border-primary-container focus:bg-surface-container-highest focus:ring-0"
                  placeholder="Username"
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  required
                />
              </div>
              <div className="group relative">
                <Icon
                  name="key"
                  className="absolute left-4 top-1/2 -translate-y-1/2 text-on-surface-variant transition-colors group-focus-within:text-primary-container"
                />
                <input
                  className="w-full rounded-t-lg border-x-0 border-b border-t-0 border-outline-variant/60 bg-surface-container-low py-3 pl-12 pr-12 font-body-md text-on-background transition-colors focus:border-primary-container focus:bg-surface-container-highest focus:ring-0"
                  placeholder="Password"
                  type={showPassword ? "text" : "password"}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-4 top-1/2 -translate-y-1/2 text-on-surface-variant transition-colors hover:text-on-background"
                  aria-label={showPassword ? "Hide password" : "Show password"}
                >
                  <Icon name={showPassword ? "visibility" : "visibility_off"} />
                </button>
              </div>
            </div>

            {error && (
              <div className="mb-4 p-3 bg-red-50 border-l-4 border-red-500 text-red-800 text-sm rounded">
                {error}
              </div>
            )}

            <div className="flex items-center justify-between gap-4 text-data-sm text-on-surface-variant">
              <label className="inline-flex items-center gap-2 cursor-pointer">
                <input
                  className="h-4 w-4 rounded border-outline-variant/60 bg-transparent text-primary-container focus:ring-primary-container"
                  type="checkbox"
                />
                Remember Me
              </label>
              <button
                className="text-primary-container transition-colors hover:text-white"
                type="button"
              >
                Reset Password
              </button>
            </div>

            <button
              className={`w-full rounded-full bg-primary-container px-8 py-3 text-base font-semibold text-background transition-transform duration-200 hover:scale-[1.02] ${loading ? "opacity-50 cursor-not-allowed" : ""}`}
              type="submit"
              disabled={loading}
            >
              {loading ? "Logging in..." : "Start"}
            </button>

            <div className="flex items-center gap-4 text-data-sm text-on-surface-variant">
              <div className="h-px flex-1 bg-outline-variant/50" />
              <span>OR</span>
              <div className="h-px flex-1 bg-outline-variant/50" />
            </div>

            <button
              className={`flex w-full items-center justify-center gap-2 rounded-full border border-outline-variant/60 px-8 py-3 text-base font-medium text-on-background transition-colors hover:bg-surface-container-high ${loading ? "opacity-50 cursor-not-allowed" : ""}`}
              type="button"
              onClick={handleBypass}
              disabled={loading}
            >
              <Icon name="login" size="small" color="text-primary-container" />
              {loading ? "Logging in..." : "Demo login"}
            </button>
          </form>
        </div>
      </Panel>
    </div>,
    document.body,
  );
}
