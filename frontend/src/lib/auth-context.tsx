"use client";

import {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
  type ReactNode,
} from "react";
import api from "./api";

interface AuthUser {
  id: number;
  firstName: string;
  telegramId: number;
  role: string;
}

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  user: AuthUser | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

interface AuthContextType extends AuthState {
  loginWithTelegram: (initData: string) => Promise<void>;
  logout: () => void;
}

const AUTH_KEY = "avangard_auth";

const AuthContext = createContext<AuthContextType>({
  accessToken: null,
  refreshToken: null,
  user: null,
  isAuthenticated: false,
  isLoading: true,
  loginWithTelegram: async () => {},
  logout: () => {},
});

function loadStoredAuth(): Omit<AuthState, "isLoading"> {
  if (typeof window === "undefined") {
    return { accessToken: null, refreshToken: null, user: null, isAuthenticated: false };
  }
  try {
    const raw = localStorage.getItem(AUTH_KEY);
    if (!raw) return { accessToken: null, refreshToken: null, user: null, isAuthenticated: false };
    const parsed = JSON.parse(raw);
    if (parsed.accessToken && parsed.user) {
      return {
        accessToken: parsed.accessToken,
        refreshToken: parsed.refreshToken,
        user: parsed.user,
        isAuthenticated: true,
      };
    }
  } catch {
    // ignore
  }
  return { accessToken: null, refreshToken: null, user: null, isAuthenticated: false };
}

function saveAuth(accessToken: string, refreshToken: string, user: AuthUser) {
  localStorage.setItem(AUTH_KEY, JSON.stringify({ accessToken, refreshToken, user }));
}

function clearAuth() {
  localStorage.removeItem(AUTH_KEY);
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({
    ...loadStoredAuth(),
    isLoading: true,
  });

  const logout = useCallback(() => {
    clearAuth();
    setState({
      accessToken: null,
      refreshToken: null,
      user: null,
      isAuthenticated: false,
      isLoading: false,
    });
  }, []);

  const loginWithTelegram = useCallback(async (initData: string) => {
    const { data } = await api.post<{
      data: {
        accessToken: string;
        refreshToken: string;
        user: AuthUser;
      };
    }>("/auth/telegram", { initData });

    const { accessToken, refreshToken, user } = data.data;
    saveAuth(accessToken, refreshToken, user);
    setState({
      accessToken,
      refreshToken,
      user,
      isAuthenticated: true,
      isLoading: false,
    });
  }, []);

  // Auto-login when inside Telegram
  useEffect(() => {
    const tg = window.Telegram?.WebApp;
    if (!tg?.initData) {
      setState((s) => ({ ...s, isLoading: false }));
      return;
    }

    // Already authenticated — skip
    const stored = loadStoredAuth();
    if (stored.isAuthenticated) {
      setState({ ...stored, isLoading: false });
      return;
    }

    // Attempt TG login
    loginWithTelegram(tg.initData).catch((err) => {
      console.warn("Telegram auto-login failed:", err);
      setState((s) => ({ ...s, isLoading: false }));
    });
  }, [loginWithTelegram]);

  return (
    <AuthContext.Provider value={{ ...state, loginWithTelegram, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
