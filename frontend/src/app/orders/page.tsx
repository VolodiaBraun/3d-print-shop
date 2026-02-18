"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  ClipboardList,
  Search,
  Loader2,
  Package,
  FlaskConical,
  ChevronRight,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { getOrder, getMyOrders, type OrderResponse } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";

const STATUS_MAP: Record<string, { label: string; color: string }> = {
  // Regular order statuses
  new: { label: "Новый", color: "bg-yellow-500/15 text-yellow-600" },
  confirmed: { label: "Подтверждён", color: "bg-blue-500/15 text-blue-600" },
  processing: { label: "В обработке", color: "bg-blue-500/15 text-blue-600" },
  shipped: { label: "Отправлен", color: "bg-violet-500/15 text-violet-600" },
  delivered: { label: "Выдан", color: "bg-green-500/15 text-green-600" },
  cancelled: { label: "Отменён", color: "bg-red-500/15 text-red-600" },
  // Custom order statuses
  in_progress: { label: "В работе", color: "bg-cyan-500/15 text-cyan-600" },
  ready: { label: "Готов", color: "bg-purple-500/15 text-purple-600" },
};

function formatPrice(price: number): string {
  return new Intl.NumberFormat("ru-RU").format(price) + " \u20BD";
}

function formatDate(dateStr: string): string {
  return new Intl.DateTimeFormat("ru-RU", {
    day: "numeric",
    month: "short",
    year: "numeric",
  }).format(new Date(dateStr));
}

export default function OrdersPage() {
  const router = useRouter();
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  const [orderNumber, setOrderNumber] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const [orders, setOrders] = useState<OrderResponse[]>([]);
  const [ordersLoading, setOrdersLoading] = useState(false);

  // Load user's orders when authenticated
  useEffect(() => {
    if (authLoading || !isAuthenticated) return;

    setOrdersLoading(true);
    getMyOrders()
      .then(setOrders)
      .catch(() => {
        // silently fail — will show search fallback
      })
      .finally(() => setOrdersLoading(false));
  }, [isAuthenticated, authLoading]);

  async function handleSearch() {
    const num = orderNumber.trim();
    if (!num) return;

    setLoading(true);
    setError("");
    try {
      await getOrder(num);
      router.push(`/order/${num}`);
    } catch {
      setError("Заказ не найден. Проверьте номер и попробуйте снова.");
    } finally {
      setLoading(false);
    }
  }

  if (authLoading || ordersLoading) {
    return (
      <div className="mx-auto max-w-lg px-4 py-12 text-center">
        <Loader2 className="mx-auto h-8 w-8 animate-spin text-muted-foreground" />
        <p className="mt-4 text-muted-foreground">Загрузка заказов...</p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-lg px-4 py-8">
      <h1 className="text-2xl font-bold text-center">Мои заказы</h1>

      {/* Orders list for authenticated users */}
      {isAuthenticated && orders.length > 0 && (
        <div className="mt-6 space-y-3">
          {orders.map((order) => {
            const status = STATUS_MAP[order.status] || {
              label: order.status,
              color: "bg-muted text-muted-foreground",
            };
            return (
              <Link
                key={order.id}
                href={`/order/${order.orderNumber}`}
                className="flex items-center gap-3 rounded-lg border border-border bg-card p-4 transition-colors hover:bg-accent/50"
              >
                {order.orderType === "custom" ? (
                  <FlaskConical className="h-5 w-5 shrink-0 text-violet-500" />
                ) : (
                  <Package className="h-5 w-5 text-muted-foreground shrink-0" />
                )}
                <div className="flex-1 min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="font-medium font-mono text-sm">
                      {order.orderNumber}
                    </span>
                    {order.orderType === "custom" && (
                      <span className="inline-flex items-center rounded-full bg-violet-500/10 px-2 py-0.5 text-xs font-medium text-violet-600">
                        Инд. заказ
                      </span>
                    )}
                    <span
                      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${status.color}`}
                    >
                      {status.label}
                    </span>
                  </div>
                  {order.orderType === "custom" &&
                    order.customDetails?.clientDescription && (
                      <p className="mt-0.5 truncate text-xs text-muted-foreground">
                        {order.customDetails.clientDescription}
                      </p>
                    )}
                  <div className="mt-1 flex items-center gap-2 text-sm text-muted-foreground">
                    {order.totalPrice > 0 ? (
                      <>
                        <span>{formatPrice(order.totalPrice)}</span>
                        <span>&middot;</span>
                      </>
                    ) : order.orderType === "custom" ? (
                      <>
                        <span className="text-xs">Цена уточняется</span>
                        <span>&middot;</span>
                      </>
                    ) : null}
                    <span>{formatDate(order.createdAt)}</span>
                  </div>
                </div>
                <ChevronRight className="h-4 w-4 text-muted-foreground shrink-0" />
              </Link>
            );
          })}
        </div>
      )}

      {/* Empty state for authenticated users */}
      {isAuthenticated && orders.length === 0 && (
        <div className="mt-8 text-center">
          <ClipboardList className="mx-auto h-16 w-16 text-muted-foreground/50" />
          <p className="mt-4 text-muted-foreground">
            У вас пока нет заказов
          </p>
          <Button asChild className="mt-4">
            <Link href="/catalog">Перейти в каталог</Link>
          </Button>
        </div>
      )}

      {/* Search by order number — always visible */}
      <div className="mt-8">
        <p className="text-sm text-muted-foreground text-center">
          {isAuthenticated
            ? "Или найдите заказ по номеру"
            : "Введите номер заказа, чтобы посмотреть его статус"}
        </p>
        <div className="mt-3 flex gap-2">
          <Input
            placeholder="ORD-..."
            value={orderNumber}
            onChange={(e) => {
              setOrderNumber(e.target.value);
              if (error) setError("");
            }}
            onKeyDown={(e) => {
              if (e.key === "Enter") handleSearch();
            }}
          />
          <Button
            onClick={handleSearch}
            disabled={loading || !orderNumber.trim()}
            className="shrink-0"
          >
            {loading ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Search className="h-4 w-4" />
            )}
          </Button>
        </div>
        {error && (
          <p className="mt-3 text-sm text-destructive text-center">{error}</p>
        )}
      </div>
    </div>
  );
}
