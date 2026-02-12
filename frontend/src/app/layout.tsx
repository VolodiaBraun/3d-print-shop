import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { Providers } from "@/lib/providers";
import { AppShell } from "@/components/layout/AppShell";
import { TelegramScript } from "@/components/layout/TelegramScript";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "АВАНГАРД — 3D-печатные изделия",
  description:
    "Интернет-магазин 3D-печатных изделий премиум качества. Фигурки, модели и аксессуары для коллекционеров и геймеров.",
  keywords: [
    "3D печать",
    "фигурки",
    "модели",
    "коллекционные",
    "3D изделия",
    "купить фигурку",
  ],
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ru" className="dark">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased min-h-screen flex flex-col`}
      >
        <TelegramScript />
        <Providers>
          <AppShell>{children}</AppShell>
        </Providers>
      </body>
    </html>
  );
}
