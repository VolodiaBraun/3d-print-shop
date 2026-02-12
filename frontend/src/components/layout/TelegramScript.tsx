"use client";

import Script from "next/script";
import { useEffect, useState } from "react";

/**
 * Only loads the Telegram WebApp SDK when running inside Telegram.
 * In a regular browser this script is unnecessary and causes a
 * "search local network devices" permission prompt.
 */
export function TelegramScript() {
  const [isTelegram, setIsTelegram] = useState(false);

  useEffect(() => {
    const hash = window.location.hash || "";
    const search = window.location.search || "";
    const ua = navigator.userAgent || "";

    if (
      hash.includes("tgWebAppData") ||
      search.includes("tgWebAppStartParam") ||
      ua.includes("Telegram")
    ) {
      setIsTelegram(true);
    }
  }, []);

  if (!isTelegram) return null;

  return (
    <Script
      src="https://telegram.org/js/telegram-web-app.js"
      strategy="afterInteractive"
    />
  );
}
