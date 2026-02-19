"use client";

import Link from "next/link";
import { ShoppingCart, Search, Menu, X, User, LogOut, FlaskConical } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useState } from "react";
import { useCart } from "@/lib/cart-context";
import { useAuth } from "@/lib/auth-context";

export function Header() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const { totalItems, loaded } = useCart();
  const { isAuthenticated, user, logout } = useAuth();

  return (
    <header className="sticky top-0 z-50 border-b border-border bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4">
        {/* Logo */}
        <Link href="/" className="flex items-center gap-2">
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-primary-foreground font-bold text-lg">
            A
          </div>
          <span className="hidden font-bold text-xl sm:block">АВАНГАРД</span>
        </Link>

        {/* Desktop nav */}
        <nav className="hidden items-center gap-6 md:flex">
          <Link
            href="/catalog"
            className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
          >
            Каталог
          </Link>
          <Link
            href="/catalog?sort=newest"
            className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
          >
            Новинки
          </Link>
          <Link
            href="/catalog?sort=popular"
            className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
          >
            Популярное
          </Link>
          <Link
            href="/custom-order"
            className="flex items-center gap-1.5 text-sm font-medium text-violet-500 transition-colors hover:text-violet-400"
          >
            <FlaskConical className="h-4 w-4" />
            Заказать печать
          </Link>
        </nav>

        {/* Actions */}
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="icon" aria-label="Поиск">
            <Search className="h-5 w-5" />
          </Button>
          <Button variant="ghost" size="icon" asChild aria-label="Корзина" className="relative">
            <Link href="/cart">
              <ShoppingCart className="h-5 w-5" />
              {loaded && totalItems > 0 && (
                <span className="absolute -top-1 -right-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-primary px-1 text-[10px] font-bold text-primary-foreground">
                  {totalItems > 99 ? "99+" : totalItems}
                </span>
              )}
            </Link>
          </Button>

          {/* Auth */}
          {isAuthenticated ? (
            <div className="hidden items-center gap-1 md:flex">
              <Button variant="ghost" size="icon" asChild aria-label="Профиль">
                <Link href="/profile">
                  <User className="h-5 w-5" />
                </Link>
              </Button>
              <Button
                variant="ghost"
                size="icon"
                aria-label="Выйти"
                onClick={logout}
              >
                <LogOut className="h-4 w-4" />
              </Button>
            </div>
          ) : (
            <Button
              variant="ghost"
              size="sm"
              asChild
              className="hidden md:inline-flex"
            >
              <Link href="/login">Войти</Link>
            </Button>
          )}

          <Button
            variant="ghost"
            size="icon"
            className="md:hidden"
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            aria-label="Меню"
          >
            {mobileMenuOpen ? (
              <X className="h-5 w-5" />
            ) : (
              <Menu className="h-5 w-5" />
            )}
          </Button>
        </div>
      </div>

      {/* Mobile menu */}
      {mobileMenuOpen && (
        <nav className="border-t border-border px-4 py-4 md:hidden">
          <div className="flex flex-col gap-3">
            <Link
              href="/catalog"
              className="text-sm font-medium"
              onClick={() => setMobileMenuOpen(false)}
            >
              Каталог
            </Link>
            <Link
              href="/catalog?sort=newest"
              className="text-sm font-medium"
              onClick={() => setMobileMenuOpen(false)}
            >
              Новинки
            </Link>
            <Link
              href="/catalog?sort=popular"
              className="text-sm font-medium"
              onClick={() => setMobileMenuOpen(false)}
            >
              Популярное
            </Link>
            <Link
              href="/custom-order"
              className="flex items-center gap-2 text-sm font-medium text-violet-500"
              onClick={() => setMobileMenuOpen(false)}
            >
              <FlaskConical className="h-4 w-4" />
              Заказать печать
            </Link>
            <hr className="border-border" />
            {isAuthenticated ? (
              <>
                <Link
                  href="/profile"
                  className="text-sm font-medium"
                  onClick={() => setMobileMenuOpen(false)}
                >
                  Профиль ({user?.firstName || user?.email || "Аккаунт"})
                </Link>
                <Link
                  href="/orders"
                  className="text-sm font-medium"
                  onClick={() => setMobileMenuOpen(false)}
                >
                  Мои заказы
                </Link>
                <button
                  className="text-sm font-medium text-left text-muted-foreground"
                  onClick={() => {
                    logout();
                    setMobileMenuOpen(false);
                  }}
                >
                  Выйти
                </button>
              </>
            ) : (
              <Link
                href="/login"
                className="text-sm font-medium"
                onClick={() => setMobileMenuOpen(false)}
              >
                Войти / Регистрация
              </Link>
            )}
          </div>
        </nav>
      )}
    </header>
  );
}
