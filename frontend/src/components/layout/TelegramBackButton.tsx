"use client";

import { useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { getTelegramWebApp } from "@/lib/telegram";

export function TelegramBackButton() {
  const pathname = usePathname();
  const router = useRouter();

  useEffect(() => {
    const tg = getTelegramWebApp();
    if (!tg) return;

    const isHome = pathname === "/" || pathname === "";

    if (isHome) {
      tg.BackButton.hide();
    } else {
      tg.BackButton.show();
    }

    const handleBack = () => {
      router.back();
    };

    tg.BackButton.onClick(handleBack);
    return () => {
      tg.BackButton.offClick(handleBack);
    };
  }, [pathname, router]);

  return null;
}
