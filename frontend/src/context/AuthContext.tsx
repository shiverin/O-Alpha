"use client";

import { createContext, useContext, useEffect, useState } from "react";
import { api } from "@/lib/api";
import {
  setToken,
  decodeToken,
  getToken,
  removeToken,
  type User,
} from "@/lib/auth";

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  markOnboarded: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState<boolean>(true);

  useEffect(() => {
    const checkAuth = async () => {
      const token = getToken();
      if (!token) {
        setLoading(false);
        return;
      }

      const decoded = decodeToken(token);
      if (decoded) {
        // Optimistically hydrate user state from token while backend validation runs.
        setUser(decoded);
      }

      const isLocalDemoToken =
        token.endsWith(".offline-demo-signature") && decoded?.id === 999;
      if (decoded && isLocalDemoToken) {
        setLoading(false);
        return;
      }

      try {
        const response = await api.get<{
          id: number;
          username: string;
          is_onboarded: boolean;
        }>("/auth/me");
        setUser({
          id: response.id,
          username: response.username,
          is_onboarded: response.is_onboarded,
        });
      } catch (err) {
        const isNetworkError =
          err instanceof TypeError ||
          (err instanceof Error &&
            /Failed to fetch|NetworkError/i.test(err.message));

        if (decoded && decoded.id === 999 && isNetworkError) {
          // Keep token-backed session for local/demo mode when backend is down.
          setUser(decoded);
        } else {
          removeToken();
          setUser(null);
        }
      } finally {
        setLoading(false);
      }
    };

    checkAuth();
  }, []);

  const login = async (username: string, password: string) => {
    try {
      const response = await api.post<{
        token: string;
        user: { id: number; username: string; is_onboarded: boolean };
      }>("/auth/login", { username, password });

      setToken(response.token);

      setUser({
        id: response.user.id,
        username: response.user.username,
        is_onboarded: response.user.is_onboarded,
      });
    } catch (error) {
      throw error;
    }
  };

  const logout = async () => {
    try {
      await api.post("/auth/logout", {});
    } catch {
      // Logout stays local if backend session cleanup is unavailable.
    } finally {
      removeToken();
      setUser(null);
    }
  };

  const markOnboarded = () => {
    setUser((currentUser) =>
      currentUser ? { ...currentUser, is_onboarded: true } : currentUser,
    );
  };

  const value = {
    user,
    loading,
    login,
    logout,
    markOnboarded,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
