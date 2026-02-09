"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import Image from "next/image";
import {
  Package,
  Truck,
  CreditCard,
  Banknote,
  ShoppingCart,
  ArrowLeft,
  Loader2,
  AlertCircle,
  User,
  Phone,
  Mail,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { getOrder, type OrderResponse } from "@/lib/api";

function formatPrice(price: number): string {
  return new Intl.NumberFormat("ru-RU").format(price) + " \u20BD";
}

function formatDate(dateStr: string): string {
  return new Intl.DateTimeFormat("ru-RU", {
    day: "numeric",
    month: "long",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(dateStr));
}

const STATUS_MAP: Record<string, { label: string; color: string }> = {
  new: { label: "Новый", color: "bg-yellow-500/15 text-yellow-600" },
  confirmed: { label: "Подтверждён", color: "bg-blue-500/15 text-blue-600" },
  processing: {
    label: "В обработке",
    color: "bg-blue-500/15 text-blue-600",
  },
  shipped: {
    label: "Отправлен",
    color: "bg-violet-500/15 text-violet-600",
  },
  delivered: {
    label: "Доставлен",
    color: "bg-green-500/15 text-green-600",
  },
  cancelled: { label: "Отменён", color: "bg-red-500/15 text-red-600" },
};

const DELIVERY_MAP: Record<string, string> = {
  pickup: "Самовывоз",
  courier: "Курьер",
};

const PAYMENT_MAP: Record<string, string> = {
  card: "Банковская карта",
  cash: "Наличные при получении",
};

export default function OrderPage() {
  const params = useParams();
  const orderNumber = params.orderNumber as string;

  const [order, setOrder] = useState<OrderResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!orderNumber) return;
    setLoading(true);
    getOrder(orderNumber)
      .then(setOrder)
      .catch((err) => {
        const axiosErr = err as { response?: { status?: number } };
        if (axiosErr?.response?.status === 404) {
          setError("Заказ не найден");
        } else {
          setError("Ошибка загрузки заказа");
        }
      })
      .finally(() => setLoading(false));
  }, [orderNumber]);

  if (loading) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-16 text-center">
        <Loader2 className="mx-auto h-8 w-8 animate-spin text-muted-foreground" />
        <p className="mt-4 text-muted-foreground">Загрузка заказа...</p>
      </div>
    );
  }

  if (error || !order) {
    return (
      <div className="mx-auto max-w-4xl px-4 py-16 text-center">
        <AlertCircle className="mx-auto h-16 w-16 text-muted-foreground/50" />
        <h1 className="mt-4 text-2xl font-bold">
          {error || "Заказ не найден"}
        </h1>
        <p className="mt-2 text-muted-foreground">
          Проверьте правильность номера заказа
        </p>
        <Button asChild className="mt-6">
          <Link href="/">На главную</Link>
        </Button>
      </div>
    );
  }

  const status = STATUS_MAP[order.status] || {
    label: order.status,
    color: "bg-muted text-muted-foreground",
  };

  return (
    <div className="mx-auto max-w-4xl px-4 py-8">
      <Button variant="ghost" size="sm" asChild className="mb-6">
        <Link href="/">
          <ArrowLeft className="mr-1 h-4 w-4" />
          На главную
        </Link>
      </Button>

      {/* Header */}
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold">
            Заказ{" "}
            <span className="font-mono">{order.orderNumber}</span>
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            от {formatDate(order.createdAt)}
          </p>
        </div>
        <span
          className={`inline-flex w-fit items-center rounded-full px-3 py-1 text-sm font-medium ${status.color}`}
        >
          {status.label}
        </span>
      </div>

      <div className="mt-8 grid gap-8 lg:grid-cols-[1fr_340px]">
        {/* Left: Items + Info */}
        <div className="space-y-8">
          {/* Order items */}
          <section>
            <h2 className="text-lg font-semibold">Товары</h2>
            <div className="mt-4 space-y-3">
              {order.items.map((item) => {
                const mainImage = item.product?.images?.find(
                  (img) => img.isMain
                );
                const imageUrl =
                  mainImage?.urlThumbnail || mainImage?.url || null;

                return (
                  <div
                    key={item.id}
                    className="flex gap-4 rounded-lg border border-border bg-card p-4"
                  >
                    <div className="relative h-16 w-16 shrink-0 overflow-hidden rounded-md bg-muted">
                      {imageUrl ? (
                        <Image
                          src={imageUrl}
                          alt={item.product?.name || ""}
                          fill
                          className="object-cover"
                          sizes="64px"
                        />
                      ) : (
                        <div className="flex h-full items-center justify-center text-muted-foreground">
                          <ShoppingCart className="h-5 w-5" />
                        </div>
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      {item.product ? (
                        <Link
                          href={`/product/${item.product.slug}`}
                          className="font-medium hover:text-primary transition-colors"
                        >
                          {item.product.name}
                        </Link>
                      ) : (
                        <span className="font-medium">
                          Товар #{item.productId}
                        </span>
                      )}
                      <p className="text-sm text-muted-foreground">
                        {item.quantity} x {formatPrice(item.unitPrice)}
                      </p>
                    </div>
                    <span className="font-semibold shrink-0">
                      {formatPrice(item.totalPrice)}
                    </span>
                  </div>
                );
              })}
            </div>
          </section>

          {/* Contact info */}
          <section>
            <h2 className="text-lg font-semibold">Информация</h2>
            <div className="mt-4 rounded-lg border border-border bg-card p-4 space-y-3">
              <div className="flex items-center gap-3">
                <User className="h-4 w-4 text-muted-foreground shrink-0" />
                <span className="text-sm">{order.customerName}</span>
              </div>
              <div className="flex items-center gap-3">
                <Phone className="h-4 w-4 text-muted-foreground shrink-0" />
                <span className="text-sm">{order.customerPhone}</span>
              </div>
              {order.customerEmail && (
                <div className="flex items-center gap-3">
                  <Mail className="h-4 w-4 text-muted-foreground shrink-0" />
                  <span className="text-sm">{order.customerEmail}</span>
                </div>
              )}
              <div className="border-t border-border pt-3 flex items-center gap-3">
                {order.deliveryMethod === "courier" ? (
                  <Truck className="h-4 w-4 text-muted-foreground shrink-0" />
                ) : (
                  <Package className="h-4 w-4 text-muted-foreground shrink-0" />
                )}
                <div>
                  <span className="text-sm font-medium">
                    {DELIVERY_MAP[order.deliveryMethod] ||
                      order.deliveryMethod}
                  </span>
                  {order.deliveryAddress && (
                    <p className="text-sm text-muted-foreground">
                      {order.deliveryAddress}
                    </p>
                  )}
                </div>
              </div>
              <div className="border-t border-border pt-3 flex items-center gap-3">
                {order.paymentMethod === "card" ? (
                  <CreditCard className="h-4 w-4 text-muted-foreground shrink-0" />
                ) : (
                  <Banknote className="h-4 w-4 text-muted-foreground shrink-0" />
                )}
                <span className="text-sm">
                  {PAYMENT_MAP[order.paymentMethod] || order.paymentMethod}
                </span>
              </div>
              {order.notes && (
                <div className="border-t border-border pt-3">
                  <p className="text-sm text-muted-foreground">
                    {order.notes}
                  </p>
                </div>
              )}
            </div>
          </section>
        </div>

        {/* Right: Summary */}
        <div className="lg:sticky lg:top-24 self-start">
          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="text-lg font-bold">Итого</h2>
            <div className="mt-4 space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Товары</span>
                <span>{formatPrice(order.subtotal)}</span>
              </div>
              {order.discountAmount > 0 && (
                <div className="flex justify-between text-green-600">
                  <span>
                    Скидка{order.promoCode ? ` (${order.promoCode})` : ""}
                  </span>
                  <span>&minus;{formatPrice(order.discountAmount)}</span>
                </div>
              )}
              {order.deliveryCost > 0 && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Доставка</span>
                  <span>{formatPrice(order.deliveryCost)}</span>
                </div>
              )}
            </div>
            <div className="mt-4 border-t border-border pt-4">
              <div className="flex justify-between text-lg font-bold">
                <span>К оплате</span>
                <span>{formatPrice(order.totalPrice)}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
