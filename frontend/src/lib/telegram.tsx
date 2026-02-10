"use client";

import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";

interface TelegramContextType {
  isTelegram: boolean;
  ready: boolean;
  userId?: number;
  firstName?: string;
  lastName?: string;
  username?: string;
  colorScheme: "light" | "dark";
}

const TelegramContext = createContext<TelegramContextType>({
  isTelegram: false,
  ready: false,
  colorScheme: "dark",
});

export function TelegramProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<TelegramContextType>({
    isTelegram: false,
    ready: false,
    colorScheme: "dark",
  });

  useEffect(() => {
    // Check if running inside Telegram Web App
    const tg = typeof window !== "undefined" ? window.Telegram?.WebApp : null;
    if (!tg || !tg.initData) {
      setState((s) => ({ ...s, ready: true }));
      return;
    }

    // We're inside Telegram
    tg.ready();
    tg.expand();

    // Disable vertical swipe to close (better UX for scrollable content)
    if (tg.disableVerticalSwipes) {
      tg.disableVerticalSwipes();
    }

    const user = tg.initDataUnsafe?.user;
    const colorScheme = tg.colorScheme || "dark";

    setState({
      isTelegram: true,
      ready: true,
      userId: user?.id,
      firstName: user?.first_name,
      lastName: user?.last_name,
      username: user?.username,
      colorScheme,
    });

    // Apply Telegram theme colors as CSS variables
    const themeParams = tg.themeParams;
    if (themeParams) {
      const root = document.documentElement;
      if (themeParams.bg_color) {
        root.style.setProperty("--tg-bg-color", themeParams.bg_color);
      }
      if (themeParams.text_color) {
        root.style.setProperty("--tg-text-color", themeParams.text_color);
      }
      if (themeParams.button_color) {
        root.style.setProperty("--tg-button-color", themeParams.button_color);
      }
      if (themeParams.button_text_color) {
        root.style.setProperty(
          "--tg-button-text-color",
          themeParams.button_text_color
        );
      }
      if (themeParams.secondary_bg_color) {
        root.style.setProperty(
          "--tg-secondary-bg-color",
          themeParams.secondary_bg_color
        );
      }
    }

    // Toggle light/dark mode based on Telegram theme
    if (colorScheme === "light") {
      document.documentElement.classList.remove("dark");
    } else {
      document.documentElement.classList.add("dark");
    }
  }, []);

  return (
    <TelegramContext.Provider value={state}>
      {children}
    </TelegramContext.Provider>
  );
}

export function useTelegram() {
  return useContext(TelegramContext);
}

// Get raw Telegram WebApp object (for BackButton, MainButton, etc.)
export function getTelegramWebApp() {
  if (typeof window === "undefined") return null;
  return window.Telegram?.WebApp ?? null;
}

// Get raw initData string for authentication
export function getInitData(): string | null {
  if (typeof window === "undefined") return null;
  return window.Telegram?.WebApp?.initData || null;
}
