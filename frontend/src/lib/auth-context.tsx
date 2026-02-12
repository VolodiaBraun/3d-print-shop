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

export interface AuthUser {
  id: number;
  firstName: string;
  email?: string;
  telegramId?: number;
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
  loginWithEmail: (email: string, password: string) => Promise<void>;
  loginWithTelegramWidget: (data: TelegramWidgetData) => Promise<void>;
  register: (name: string, email: string, password: string, referralCode?: string) => Promise<void>;
  logout: () => void;
}

export interface TelegramWidgetData {
  id: number;
  first_name?: string;
  last_name?: string;
  username?: string;
  photo_url?: string;
  auth_date: number;
  hash: string;
}

const AUTH_KEY = "avangard_auth";

const AuthContext = createContext<AuthContextType>({
  accessToken: null,
  refreshToken: null,
  user: null,
  isAuthenticated: false,
  isLoading: true,
  loginWithTelegram: async () => {},
  loginWithEmail: async () => {},
  loginWithTelegramWidget: async () => {},
  register: async () => {},
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

  const setAuthFromResponse = useCallback(
    (data: { accessToken: string; refreshToken: string; user: AuthUser }) => {
      saveAuth(data.accessToken, data.refreshToken, data.user);
      setState({
        accessToken: data.accessToken,
        refreshToken: data.refreshToken,
        user: data.user,
        isAuthenticated: true,
        isLoading: false,
      });
    },
    []
  );

  const loginWithTelegram = useCallback(
    async (initData: string) => {
      const { data } = await api.post<{
        data: { accessToken: string; refreshToken: string; user: AuthUser };
      }>("/auth/telegram", { initData });
      setAuthFromResponse(data.data);
    },
    [setAuthFromResponse]
  );

  const loginWithEmail = useCallback(
    async (email: string, password: string) => {
      const { data } = await api.post<{
        data: { accessToken: string; refreshToken: string };
      }>("/auth/login", { email, password });
      // Login doesn't return user, fetch profile
      saveAuth(data.data.accessToken, data.data.refreshToken, { id: 0, firstName: "", role: "customer" });
      const profileResp = await api.get<{ data: AuthUser }>("/users/me", {
        headers: { Authorization: `Bearer ${data.data.accessToken}` },
      });
      const user: AuthUser = {
        id: profileResp.data.data.id,
        firstName: profileResp.data.data.firstName || "",
        email: profileResp.data.data.email,
        role: profileResp.data.data.role,
      };
      setAuthFromResponse({ ...data.data, user });
    },
    [setAuthFromResponse]
  );

  const register = useCallback(
    async (name: string, email: string, password: string, referralCode?: string) => {
      const payload: Record<string, string> = { name, email, password };
      if (referralCode) payload.referralCode = referralCode;
      const { data } = await api.post<{
        data: { accessToken: string; refreshToken: string; user: AuthUser };
      }>("/auth/register", payload);
      setAuthFromResponse(data.data);
    },
    [setAuthFromResponse]
  );

  const loginWithTelegramWidget = useCallback(
    async (widgetData: TelegramWidgetData) => {
      const { data } = await api.post<{
        data: { accessToken: string; refreshToken: string; user: AuthUser };
      }>("/auth/telegram-widget", {
        id: widgetData.id,
        firstName: widgetData.first_name || "",
        lastName: widgetData.last_name || "",
        username: widgetData.username || "",
        photoUrl: widgetData.photo_url || "",
        authDate: widgetData.auth_date,
        hash: widgetData.hash,
      });
      setAuthFromResponse(data.data);
    },
    [setAuthFromResponse]
  );

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
    <AuthContext.Provider
      value={{
        ...state,
        loginWithTelegram,
        loginWithEmail,
        loginWithTelegramWidget,
        register,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
