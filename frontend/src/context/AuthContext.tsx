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

      try {
        // 🚀 UPDATED: Added is_onboarded type tracking to matching API layout response
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

        if (decoded && isNetworkError) {
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
      // 🚀 UPDATED: Extended the layout schema specification to preserve onboarding metrics
      const response = await api.post<{
        token: string;
        user: { id: number; username: string; is_onboarded: boolean };
      }>("/auth/login", { username, password });

      // Explicitly persist the token securely
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
      // Absorb tracking network logging variances silently
    } finally {
      removeToken(); // Guard clean tracking clear states
      setUser(null);
    }
  };

  const value = {
    user,
    loading,
    login,
    logout,
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
