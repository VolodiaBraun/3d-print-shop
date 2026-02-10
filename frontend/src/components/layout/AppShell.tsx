"use client";

import { useTelegram } from "@/lib/telegram";
import { Header } from "./Header";
import { Footer } from "./Footer";
import { TelegramBackButton } from "./TelegramBackButton";
import { TelegramNavBar } from "./TelegramNavBar";

export function AppShell({ children }: { children: React.ReactNode }) {
  const { isTelegram, ready } = useTelegram();

  if (!ready) {
    return null;
  }

  if (isTelegram) {
    return (
      <>
        <TelegramBackButton />
        <main className="flex-1 pb-16">{children}</main>
        <TelegramNavBar />
      </>
    );
  }

  return (
    <>
      <Header />
      <main className="flex-1">{children}</main>
      <Footer />
    </>
  );
}
