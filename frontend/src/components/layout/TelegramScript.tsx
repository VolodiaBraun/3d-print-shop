"use client";

import { useEffect } from "react";

/**
 * Only loads the Telegram WebApp SDK when running inside Telegram.
 * In a regular browser this script is unnecessary and causes a
 * "search local network devices" permission prompt.
 */
export function TelegramScript() {
  useEffect(() => {
    const hash = window.location.hash || "";
    const search = window.location.search || "";
    const ua = navigator.userAgent || "";

    const isTelegram =
      hash.includes("tgWebAppData") ||
      search.includes("tgWebAppStartParam") ||
      ua.includes("Telegram");

    if (!isTelegram) return;

    // Check if already loaded
    if (document.querySelector('script[src*="telegram-web-app.js"]')) return;

    const script = document.createElement("script");
    script.src = "https://telegram.org/js/telegram-web-app.js";
    script.async = false;
    document.head.appendChild(script);
  }, []);

  return null;
}
