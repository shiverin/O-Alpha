"use client";

import { createContext, useContext, useEffect, useState } from "react";
import { api } from "@/lib/api";
import { decodeToken, getToken, removeToken, type User } from "@/lib/auth";

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
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
        const response = await api.get<{ id: number; email: string }>(
          "/auth/me",
        );
        setUser({ id: response.id, email: response.email });
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

  const login = async (email: string, password: string) => {
    try {
      const response = await api.post<{
        token: string;
        user: { id: number; email: string };
      }>("/auth/login", { email, password });
      // Token will be stored by the API interceptor in headers
      setUser({ id: response.user.id, email: response.user.email });
    } catch (error) {
      throw error;
    }
  };

  const logout = async () => {
    try {
      await api.post("/auth/logout", {});
    } catch {
    } finally {
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
