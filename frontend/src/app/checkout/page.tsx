"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import Image from "next/image";
import {
  ShoppingCart,
  Package,
  Truck,
  CreditCard,
  Banknote,
  Tag,
  X,
  Loader2,
  CheckCircle2,
  ArrowLeft,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useCart } from "@/lib/cart-context";
import { useTelegram } from "@/lib/telegram";
import {
  createOrder,
  validatePromoCode,
  type PromoValidationResult,
  type OrderResponse,
} from "@/lib/api";

function formatPrice(price: number): string {
  return new Intl.NumberFormat("ru-RU").format(price) + " \u20BD";
}

export default function CheckoutPage() {
  const { items, totalItems, totalPrice, clearCart, loaded } = useCart();
  const { isTelegram, firstName, lastName, userId: telegramId } = useTelegram();

  // Contact form
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");
  const [email, setEmail] = useState("");

  // Pre-fill from Telegram
  useEffect(() => {
    if (!isTelegram) return;

    if (firstName && !name) {
      const tgName = [firstName, lastName].filter(Boolean).join(" ");
      setName(tgName);
    }

    // Request phone number via Telegram
    const tg = window.Telegram?.WebApp;
    if (tg && !phone) {
      try {
        tg.requestContact?.((ok: boolean, response?: { responseUnsafe?: { contact?: { phone_number?: string } } }) => {
          if (ok && response?.responseUnsafe?.contact?.phone_number) {
            let num = response.responseUnsafe.contact.phone_number;
            if (!num.startsWith("+")) num = "+" + num;
            setPhone(num);
          }
        });
      } catch {
        // requestContact not supported in this TG version, ignore
      }
    }
  }, [isTelegram, firstName, lastName]); // eslint-disable-line react-hooks/exhaustive-deps

  // Delivery
  const [deliveryMethod, setDeliveryMethod] = useState("pickup");
  const [address, setAddress] = useState("");

  // Payment
  const [paymentMethod, setPaymentMethod] = useState("card");

  // Notes
  const [notes, setNotes] = useState("");

  // Promo
  const [promoCode, setPromoCode] = useState("");
  const [promoResult, setPromoResult] = useState<PromoValidationResult | null>(
    null
  );
  const [promoError, setPromoError] = useState("");
  const [promoLoading, setPromoLoading] = useState(false);

  // Form state
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState("");
  const [order, setOrder] = useState<OrderResponse | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const discountAmount = promoResult?.discountAmount ?? 0;
  const finalPrice = Math.max(0, totalPrice - discountAmount);

  async function handleApplyPromo() {
    const code = promoCode.trim();
    if (!code) return;
    setPromoLoading(true);
    setPromoError("");
    try {
      const result = await validatePromoCode(code, totalPrice);
      setPromoResult(result);
    } catch (err: unknown) {
      setPromoResult(null);
      const axiosErr = err as { response?: { data?: { message?: string } } };
      setPromoError(
        axiosErr?.response?.data?.message || "Промокод не найден"
      );
    } finally {
      setPromoLoading(false);
    }
  }

  function handleRemovePromo() {
    setPromoResult(null);
    setPromoCode("");
    setPromoError("");
  }

  function validate(): boolean {
    const errors: Record<string, string> = {};
    if (!name.trim()) errors.name = "Введите имя";
    if (!phone.trim()) errors.phone = "Введите телефон";
    else if (phone.replace(/\D/g, "").length < 10)
      errors.phone = "Некорректный номер телефона";
    if (deliveryMethod === "courier" && !address.trim())
      errors.address = "Введите адрес доставки";
    setFieldErrors(errors);
    return Object.keys(errors).length === 0;
  }

  async function handleSubmit() {
    if (!validate()) return;
    setSubmitting(true);
    setSubmitError("");
    try {
      const result = await createOrder({
        items: items.map((i) => ({
          productId: i.productId,
          quantity: i.quantity,
        })),
        customerName: name.trim(),
        customerPhone: phone.trim(),
        customerEmail: email.trim() || undefined,
        deliveryMethod,
        telegramId: telegramId || undefined,
        deliveryAddress:
          deliveryMethod === "courier" ? address.trim() : undefined,
        paymentMethod,
        promoCode: promoResult?.code || undefined,
        notes: notes.trim() || undefined,
      });
      setOrder(result);
      clearCart();
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { message?: string } } };
      setSubmitError(
        axiosErr?.response?.data?.message || "Ошибка при оформлении заказа"
      );
    } finally {
      setSubmitting(false);
    }
  }

  // Success state
  if (order) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-16 text-center">
        <CheckCircle2 className="mx-auto h-16 w-16 text-green-500" />
        <h1 className="mt-4 text-2xl font-bold">Заказ оформлен!</h1>
        <p className="mt-2 text-lg text-muted-foreground">
          Номер заказа:{" "}
          <span className="font-mono font-bold text-foreground">
            {order.orderNumber}
          </span>
        </p>
        <div className="mt-6 rounded-lg border border-border bg-card p-6 text-left">
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-muted-foreground">Сумма товаров</span>
              <span>{formatPrice(order.subtotal)}</span>
            </div>
            {order.discountAmount > 0 && (
              <div className="flex justify-between text-green-600">
                <span>Скидка ({order.promoCode})</span>
                <span>&minus;{formatPrice(order.discountAmount)}</span>
              </div>
            )}
            <div className="flex justify-between border-t border-border pt-2 text-base font-bold">
              <span>Итого</span>
              <span>{formatPrice(order.totalPrice)}</span>
            </div>
          </div>
        </div>
        <p className="mt-4 text-sm text-muted-foreground">
          Мы свяжемся с вами для подтверждения заказа
        </p>
        <div className="mt-6 flex flex-col items-center gap-3 sm:flex-row sm:justify-center">
          <Button asChild>
            <Link href={`/order/${order.orderNumber}`}>Посмотреть заказ</Link>
          </Button>
          <Button variant="outline" asChild>
            <Link href="/catalog">Продолжить покупки</Link>
          </Button>
        </div>
      </div>
    );
  }

  // Empty cart — redirect hint
  if (loaded && items.length === 0) {
    return (
      <div className="mx-auto max-w-7xl px-4 py-16 text-center">
        <ShoppingCart className="mx-auto h-16 w-16 text-muted-foreground/50" />
        <h1 className="mt-4 text-2xl font-bold">Корзина пуста</h1>
        <p className="mt-2 text-muted-foreground">
          Добавьте товары, чтобы оформить заказ
        </p>
        <Button asChild className="mt-6">
          <Link href="/catalog">Перейти в каталог</Link>
        </Button>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-7xl px-4 py-8">
      <div className="mb-6">
        <Button variant="ghost" size="sm" asChild>
          <Link href="/cart">
            <ArrowLeft className="mr-1 h-4 w-4" />
            Вернуться в корзину
          </Link>
        </Button>
      </div>

      <h1 className="text-2xl font-bold lg:text-3xl">Оформление заказа</h1>

      <div className="mt-8 grid gap-8 lg:grid-cols-[1fr_400px]">
        {/* Left: Form */}
        <div className="space-y-8">
          {/* Contacts */}
          <section>
            <h2 className="text-lg font-semibold">Контактные данные</h2>
            <div className="mt-4 space-y-4">
              <div>
                <label className="mb-1 block text-sm font-medium">
                  Имя <span className="text-destructive">*</span>
                </label>
                <Input
                  placeholder="Ваше имя"
                  value={name}
                  onChange={(e) => {
                    setName(e.target.value);
                    if (fieldErrors.name)
                      setFieldErrors((p) => ({ ...p, name: "" }));
                  }}
                  aria-invalid={!!fieldErrors.name}
                />
                {fieldErrors.name && (
                  <p className="mt-1 text-sm text-destructive">
                    {fieldErrors.name}
                  </p>
                )}
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">
                  Телефон <span className="text-destructive">*</span>
                </label>
                <Input
                  type="tel"
                  placeholder="+7 (900) 123-45-67"
                  value={phone}
                  onChange={(e) => {
                    setPhone(e.target.value);
                    if (fieldErrors.phone)
                      setFieldErrors((p) => ({ ...p, phone: "" }));
                  }}
                  aria-invalid={!!fieldErrors.phone}
                />
                {fieldErrors.phone && (
                  <p className="mt-1 text-sm text-destructive">
                    {fieldErrors.phone}
                  </p>
                )}
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium">Email</label>
                <Input
                  type="email"
                  placeholder="email@example.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
              </div>
            </div>
          </section>

          {/* Delivery */}
          <section>
            <h2 className="text-lg font-semibold">Способ доставки</h2>
            <div className="mt-4 space-y-3">
              <label
                className={`flex cursor-pointer items-center gap-3 rounded-lg border p-4 transition-colors ${
                  deliveryMethod === "pickup"
                    ? "border-primary bg-primary/5"
                    : "border-border hover:border-muted-foreground/30"
                }`}
              >
                <input
                  type="radio"
                  name="delivery"
                  value="pickup"
                  checked={deliveryMethod === "pickup"}
                  onChange={() => setDeliveryMethod("pickup")}
                  className="sr-only"
                />
                <div
                  className={`flex h-5 w-5 items-center justify-center rounded-full border-2 ${
                    deliveryMethod === "pickup"
                      ? "border-primary"
                      : "border-muted-foreground/40"
                  }`}
                >
                  {deliveryMethod === "pickup" && (
                    <div className="h-2.5 w-2.5 rounded-full bg-primary" />
                  )}
                </div>
                <Package className="h-5 w-5 text-muted-foreground" />
                <div>
                  <div className="font-medium">Самовывоз</div>
                  <div className="text-sm text-muted-foreground">
                    Бесплатно
                  </div>
                </div>
              </label>

              <label
                className={`flex cursor-pointer items-center gap-3 rounded-lg border p-4 transition-colors ${
                  deliveryMethod === "courier"
                    ? "border-primary bg-primary/5"
                    : "border-border hover:border-muted-foreground/30"
                }`}
              >
                <input
                  type="radio"
                  name="delivery"
                  value="courier"
                  checked={deliveryMethod === "courier"}
                  onChange={() => setDeliveryMethod("courier")}
                  className="sr-only"
                />
                <div
                  className={`flex h-5 w-5 items-center justify-center rounded-full border-2 ${
                    deliveryMethod === "courier"
                      ? "border-primary"
                      : "border-muted-foreground/40"
                  }`}
                >
                  {deliveryMethod === "courier" && (
                    <div className="h-2.5 w-2.5 rounded-full bg-primary" />
                  )}
                </div>
                <Truck className="h-5 w-5 text-muted-foreground" />
                <div>
                  <div className="font-medium">Курьер</div>
                  <div className="text-sm text-muted-foreground">
                    Стоимость рассчитывается отдельно
                  </div>
                </div>
              </label>

              {deliveryMethod === "courier" && (
                <div className="ml-8">
                  <label className="mb-1 block text-sm font-medium">
                    Адрес доставки{" "}
                    <span className="text-destructive">*</span>
                  </label>
                  <Input
                    placeholder="Город, улица, дом, квартира"
                    value={address}
                    onChange={(e) => {
                      setAddress(e.target.value);
                      if (fieldErrors.address)
                        setFieldErrors((p) => ({ ...p, address: "" }));
                    }}
                    aria-invalid={!!fieldErrors.address}
                  />
                  {fieldErrors.address && (
                    <p className="mt-1 text-sm text-destructive">
                      {fieldErrors.address}
                    </p>
                  )}
                </div>
              )}
            </div>
          </section>

          {/* Payment */}
          <section>
            <h2 className="text-lg font-semibold">Способ оплаты</h2>
            <div className="mt-4 space-y-3">
              <label
                className={`flex cursor-pointer items-center gap-3 rounded-lg border p-4 transition-colors ${
                  paymentMethod === "card"
                    ? "border-primary bg-primary/5"
                    : "border-border hover:border-muted-foreground/30"
                }`}
              >
                <input
                  type="radio"
                  name="payment"
                  value="card"
                  checked={paymentMethod === "card"}
                  onChange={() => setPaymentMethod("card")}
                  className="sr-only"
                />
                <div
                  className={`flex h-5 w-5 items-center justify-center rounded-full border-2 ${
                    paymentMethod === "card"
                      ? "border-primary"
                      : "border-muted-foreground/40"
                  }`}
                >
                  {paymentMethod === "card" && (
                    <div className="h-2.5 w-2.5 rounded-full bg-primary" />
                  )}
                </div>
                <CreditCard className="h-5 w-5 text-muted-foreground" />
                <div>
                  <div className="font-medium">Банковская карта</div>
                </div>
              </label>

              <label
                className={`flex cursor-pointer items-center gap-3 rounded-lg border p-4 transition-colors ${
                  paymentMethod === "cash"
                    ? "border-primary bg-primary/5"
                    : "border-border hover:border-muted-foreground/30"
                }`}
              >
                <input
                  type="radio"
                  name="payment"
                  value="cash"
                  checked={paymentMethod === "cash"}
                  onChange={() => setPaymentMethod("cash")}
                  className="sr-only"
                />
                <div
                  className={`flex h-5 w-5 items-center justify-center rounded-full border-2 ${
                    paymentMethod === "cash"
                      ? "border-primary"
                      : "border-muted-foreground/40"
                  }`}
                >
                  {paymentMethod === "cash" && (
                    <div className="h-2.5 w-2.5 rounded-full bg-primary" />
                  )}
                </div>
                <Banknote className="h-5 w-5 text-muted-foreground" />
                <div>
                  <div className="font-medium">Наличные при получении</div>
                </div>
              </label>
            </div>
          </section>

          {/* Notes */}
          <section>
            <h2 className="text-lg font-semibold">Комментарий к заказу</h2>
            <div className="mt-4">
              <textarea
                className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary min-h-[80px] resize-y"
                placeholder="Пожелания к заказу..."
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
              />
            </div>
          </section>
        </div>

        {/* Right: Summary */}
        <div className="lg:sticky lg:top-24 self-start">
          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="text-lg font-bold">Ваш заказ</h2>

            {/* Items */}
            <div className="mt-4 space-y-3">
              {items.map((item) => (
                <div key={item.productId} className="flex gap-3">
                  <div className="relative h-12 w-12 shrink-0 overflow-hidden rounded bg-muted">
                    {item.image ? (
                      <Image
                        src={item.image}
                        alt={item.name}
                        fill
                        className="object-cover"
                        sizes="48px"
                      />
                    ) : (
                      <div className="flex h-full items-center justify-center text-muted-foreground">
                        <ShoppingCart className="h-4 w-4" />
                      </div>
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium truncate">{item.name}</p>
                    <p className="text-xs text-muted-foreground">
                      {item.quantity} x {formatPrice(item.price)}
                    </p>
                  </div>
                  <span className="text-sm font-medium shrink-0">
                    {formatPrice(item.price * item.quantity)}
                  </span>
                </div>
              ))}
            </div>

            <div className="mt-4 border-t border-border pt-4 space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">
                  Товары ({totalItems})
                </span>
                <span>{formatPrice(totalPrice)}</span>
              </div>
            </div>

            {/* Promo code */}
            <div className="mt-4 border-t border-border pt-4">
              {promoResult ? (
                <div className="flex items-center justify-between rounded-md bg-green-500/10 px-3 py-2">
                  <div className="flex items-center gap-2 text-sm">
                    <Tag className="h-4 w-4 text-green-600" />
                    <span className="font-medium text-green-700">
                      {promoResult.code}
                    </span>
                    <span className="text-green-600">
                      &minus;{formatPrice(promoResult.discountAmount)}
                    </span>
                  </div>
                  <button
                    onClick={handleRemovePromo}
                    className="text-green-600 hover:text-green-800 transition-colors"
                  >
                    <X className="h-4 w-4" />
                  </button>
                </div>
              ) : (
                <div>
                  <div className="flex gap-2">
                    <Input
                      placeholder="Промокод"
                      value={promoCode}
                      onChange={(e) => {
                        setPromoCode(e.target.value);
                        if (promoError) setPromoError("");
                      }}
                      onKeyDown={(e) => {
                        if (e.key === "Enter") handleApplyPromo();
                      }}
                      className="text-sm"
                    />
                    <Button
                      variant="outline"
                      size="sm"
                      className="shrink-0 px-4"
                      disabled={promoLoading || !promoCode.trim()}
                      onClick={handleApplyPromo}
                    >
                      {promoLoading ? (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      ) : (
                        "Применить"
                      )}
                    </Button>
                  </div>
                  {promoError && (
                    <p className="mt-2 text-sm text-destructive">
                      {promoError}
                    </p>
                  )}
                </div>
              )}
            </div>

            {/* Total */}
            <div className="mt-4 border-t border-border pt-4">
              {discountAmount > 0 && (
                <div className="flex justify-between text-sm text-green-600 mb-2">
                  <span>Скидка</span>
                  <span>&minus;{formatPrice(discountAmount)}</span>
                </div>
              )}
              <div className="flex justify-between text-lg font-bold">
                <span>К оплате</span>
                <span>{formatPrice(finalPrice)}</span>
              </div>
            </div>

            {submitError && (
              <p className="mt-4 text-sm text-destructive rounded-md bg-destructive/10 p-3">
                {submitError}
              </p>
            )}

            <Button
              className="mt-6 w-full gap-2"
              size="lg"
              disabled={submitting || !loaded || items.length === 0}
              onClick={handleSubmit}
            >
              {submitting ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Оформляем...
                </>
              ) : (
                `Оформить заказ — ${formatPrice(finalPrice)}`
              )}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
