"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Home, Search, ShoppingCart, ClipboardList, User } from "lucide-react";
import { useCart } from "@/lib/cart-context";

export function TelegramNavBar() {
  const pathname = usePathname();
  const { totalItems, loaded } = useCart();

  const tabs = [
    { href: "/", icon: Home, label: "Главная" },
    { href: "/catalog", icon: Search, label: "Каталог" },
    { href: "/cart", icon: ShoppingCart, label: "Корзина", badge: totalItems },
    { href: "/orders", icon: ClipboardList, label: "Заказы" },
    { href: "/profile", icon: User, label: "Профиль" },
  ];

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 border-t border-border bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/80 safe-area-bottom">
      <div className="flex items-center justify-around h-14">
        {tabs.map((tab) => {
          const isActive =
            tab.href === "/"
              ? pathname === "/"
              : pathname.startsWith(tab.href);

          return (
            <Link
              key={tab.href}
              href={tab.href}
              className={`flex flex-col items-center justify-center gap-0.5 flex-1 h-full transition-colors ${
                isActive
                  ? "text-primary"
                  : "text-muted-foreground"
              }`}
            >
              <div className="relative">
                <tab.icon className="h-5 w-5" />
                {loaded && tab.badge != null && tab.badge > 0 && (
                  <span className="absolute -top-1.5 -right-2.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-primary px-0.5 text-[10px] font-bold text-primary-foreground">
                    {tab.badge > 99 ? "99+" : tab.badge}
                  </span>
                )}
              </div>
              <span className="text-[10px] font-medium">{tab.label}</span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
