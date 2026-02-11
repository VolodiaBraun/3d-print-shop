"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAuth, type TelegramWidgetData } from "@/lib/auth-context";
import { Loader2 } from "lucide-react";

export default function LoginPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { loginWithEmail, loginWithTelegramWidget, isAuthenticated } = useAuth();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const redirectTo = searchParams.get("redirect") || "/";

  useEffect(() => {
    if (isAuthenticated) {
      router.replace(redirectTo);
    }
  }, [isAuthenticated, router, redirectTo]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await loginWithEmail(email, password);
      router.push(redirectTo);
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message || "Ошибка входа";
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  // Telegram Login Widget callback
  const handleTelegramAuth = useCallback(
    async (user: TelegramWidgetData) => {
      setError("");
      setLoading(true);
      try {
        await loginWithTelegramWidget(user);
        router.push(redirectTo);
      } catch {
        setError("Ошибка входа через Telegram");
      } finally {
        setLoading(false);
      }
    },
    [loginWithTelegramWidget, router, redirectTo]
  );

  // Mount Telegram Login Widget
  useEffect(() => {
    // Make callback accessible globally
    (window as unknown as Record<string, unknown>).onTelegramAuth = handleTelegramAuth;

    const container = document.getElementById("telegram-login-widget");
    if (!container || container.childNodes.length > 0) return;

    const script = document.createElement("script");
    script.src = "https://telegram.org/js/telegram-widget.js?22";
    script.async = true;
    script.setAttribute("data-telegram-login", "avangard_print_bot");
    script.setAttribute("data-size", "large");
    script.setAttribute("data-radius", "8");
    script.setAttribute("data-onauth", "onTelegramAuth(user)");
    script.setAttribute("data-request-access", "write");
    container.appendChild(script);

    return () => {
      delete (window as unknown as Record<string, unknown>).onTelegramAuth;
    };
  }, [handleTelegramAuth]);

  return (
    <div className="mx-auto max-w-md px-4 py-12">
      <div className="space-y-6">
        <div className="text-center">
          <h1 className="text-2xl font-bold">Вход в аккаунт</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Войдите, чтобы отслеживать заказы и оставлять отзывы
          </p>
        </div>

        {/* Telegram Login Widget */}
        <div className="flex justify-center">
          <div id="telegram-login-widget" />
        </div>

        <div className="relative">
          <div className="absolute inset-0 flex items-center">
            <span className="w-full border-t" />
          </div>
          <div className="relative flex justify-center text-xs uppercase">
            <span className="bg-background px-2 text-muted-foreground">
              или по email
            </span>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
              {error}
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              type="email"
              placeholder="your@email.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="password">Пароль</Label>
            <Input
              id="password"
              type="password"
              placeholder="Минимум 6 символов"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={6}
            />
          </div>

          <Button type="submit" className="w-full" disabled={loading}>
            {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Войти
          </Button>
        </form>

        <p className="text-center text-sm text-muted-foreground">
          Нет аккаунта?{" "}
          <Link
            href={`/register${redirectTo !== "/" ? `?redirect=${encodeURIComponent(redirectTo)}` : ""}`}
            className="font-medium text-primary underline-offset-4 hover:underline"
          >
            Зарегистрироваться
          </Link>
        </p>
      </div>
    </div>
  );
}
