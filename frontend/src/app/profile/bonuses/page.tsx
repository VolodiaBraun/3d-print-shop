"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useAuth } from "@/lib/auth-context";
import { getBonusHistory, type BonusHistoryItem } from "@/lib/api";
import { ArrowLeft, Loader2, TrendingUp, TrendingDown, Gift, ShoppingCart } from "lucide-react";

const typeLabels: Record<string, string> = {
  referral_welcome: "Приветственный бонус",
  referral_reward: "Бонус за реферала",
  order_deduction: "Списание при заказе",
  admin_adjustment: "Корректировка",
};

const typeIcons: Record<string, typeof Gift> = {
  referral_welcome: Gift,
  referral_reward: TrendingUp,
  order_deduction: ShoppingCart,
  admin_adjustment: TrendingDown,
};

export default function BonusHistoryPage() {
  const router = useRouter();
  const { isAuthenticated, isLoading: authLoading } = useAuth();

  const [history, setHistory] = useState<BonusHistoryItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (authLoading) return;
    if (!isAuthenticated) {
      router.replace("/login?redirect=/profile/bonuses");
      return;
    }

    getBonusHistory()
      .then((data) => setHistory(data || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [isAuthenticated, authLoading, router]);

  if (authLoading || loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-2xl px-4 py-8">
      <Link
        href="/profile"
        className="mb-6 inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" />
        Назад к профилю
      </Link>

      <h1 className="mb-6 text-2xl font-bold">История бонусов</h1>

      {history.length === 0 ? (
        <div className="py-12 text-center">
          <Gift className="mx-auto h-12 w-12 text-muted-foreground/30" />
          <p className="mt-4 text-muted-foreground">Пока нет операций</p>
          <Link
            href="/profile/referral"
            className="mt-2 inline-block text-sm text-primary underline-offset-4 hover:underline"
          >
            Пригласите друга и получите бонус
          </Link>
        </div>
      ) : (
        <div className="space-y-3">
          {history.map((item) => {
            const Icon = typeIcons[item.type] || Gift;
            const isPositive = item.amount > 0;

            return (
              <div
                key={item.id}
                className="flex items-center gap-4 rounded-lg border p-4"
              >
                <div
                  className={`flex h-10 w-10 items-center justify-center rounded-full ${
                    isPositive
                      ? "bg-green-500/10 text-green-600 dark:text-green-400"
                      : "bg-red-500/10 text-red-600 dark:text-red-400"
                  }`}
                >
                  <Icon className="h-5 w-5" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium">
                    {typeLabels[item.type] || item.type}
                  </p>
                  {item.description && (
                    <p className="text-xs text-muted-foreground truncate">
                      {item.description}
                    </p>
                  )}
                  <p className="text-xs text-muted-foreground">
                    {new Date(item.createdAt).toLocaleDateString("ru-RU", {
                      day: "numeric",
                      month: "long",
                      year: "numeric",
                      hour: "2-digit",
                      minute: "2-digit",
                    })}
                  </p>
                </div>
                <span
                  className={`text-sm font-bold whitespace-nowrap ${
                    isPositive
                      ? "text-green-600 dark:text-green-400"
                      : "text-red-600 dark:text-red-400"
                  }`}
                >
                  {isPositive ? "+" : ""}
                  {item.amount.toFixed(0)} &#8381;
                </span>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
