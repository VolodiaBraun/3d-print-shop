"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import Image from "next/image";
import {
  ShoppingCart,
  Package,
  Truck,
  MapPin,
  CreditCard,
  Banknote,
  Tag,
  X,
  Loader2,
  CheckCircle2,
  ArrowLeft,
  Clock,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useCart } from "@/lib/cart-context";
import { useTelegram } from "@/lib/telegram";
import {
  createOrder,
  validatePromoCode,
  calculateDelivery,
  type PromoValidationResult,
  type OrderResponse,
  type DeliveryCalculationResult,
  type PickupPointData,
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
        // requestContact not supported
      }
    }
  }, [isTelegram, firstName, lastName]); // eslint-disable-line react-hooks/exhaustive-deps

  // Delivery
  const [deliveryMethod, setDeliveryMethod] = useState("pickup");
  const [city, setCity] = useState("");
  const [address, setAddress] = useState("");
  const [deliveryLoading, setDeliveryLoading] = useState(false);
  const [deliveryResult, setDeliveryResult] = useState<DeliveryCalculationResult | null>(null);
  const [selectedPickupPoint, setSelectedPickupPoint] = useState<PickupPointData | null>(null);
  const [deliveryCost, setDeliveryCost] = useState(0);

  // Calculate delivery when city changes
  const doCalculateDelivery = useCallback(async (cityValue: string) => {
    if (!cityValue.trim() || totalPrice <= 0) return;
    setDeliveryLoading(true);
    try {
      const result = await calculateDelivery(cityValue.trim(), totalPrice);
      setDeliveryResult(result);
    } catch {
      setDeliveryResult(null);
    } finally {
      setDeliveryLoading(false);
    }
  }, [totalPrice]);

  // Debounced city input
  useEffect(() => {
    if (!city.trim()) {
      setDeliveryResult(null);
      return;
    }
    const timer = setTimeout(() => {
      doCalculateDelivery(city);
    }, 600);
    return () => clearTimeout(timer);
  }, [city, doCalculateDelivery]);

  // Update delivery cost when method or result changes
  useEffect(() => {
    if (deliveryMethod === "pickup") {
      setDeliveryCost(0);
    } else if (deliveryMethod === "courier" && deliveryResult?.courierOptions?.length) {
      setDeliveryCost(deliveryResult.courierOptions[0].cost);
    } else if (deliveryMethod === "pickup_point") {
      setDeliveryCost(0);
    } else {
      setDeliveryCost(0);
    }
  }, [deliveryMethod, deliveryResult]);

  // Reset pickup point when changing method
  useEffect(() => {
    if (deliveryMethod !== "pickup_point") {
      setSelectedPickupPoint(null);
    }
  }, [deliveryMethod]);

  // Payment
  const [paymentMethod, setPaymentMethod] = useState("card");

  // Notes
  const [notes, setNotes] = useState("");

  // Promo
  const [promoCode, setPromoCode] = useState("");
  const [promoResult, setPromoResult] = useState<PromoValidationResult | null>(null);
  const [promoError, setPromoError] = useState("");
  const [promoLoading, setPromoLoading] = useState(false);

  // Form state
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState("");
  const [order, setOrder] = useState<OrderResponse | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const discountAmount = promoResult?.discountAmount ?? 0;
  const finalPrice = Math.max(0, totalPrice - discountAmount + deliveryCost);

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
      setPromoError(axiosErr?.response?.data?.message || "Промокод не найден");
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
    if (deliveryMethod === "courier") {
      if (!city.trim()) errors.city = "Введите город";
      if (!address.trim()) errors.address = "Введите адрес доставки";
    }
    if (deliveryMethod === "pickup_point") {
      if (!city.trim()) errors.city = "Введите город";
      if (!selectedPickupPoint) errors.pickupPoint = "Выберите пункт выдачи";
    }
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
        pickupPointId: selectedPickupPoint?.id || undefined,
        city: city.trim() || undefined,
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
            {order.deliveryCost > 0 && (
              <div className="flex justify-between">
                <span className="text-muted-foreground">Доставка</span>
                <span>{formatPrice(order.deliveryCost)}</span>
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

  // Empty cart
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

  const courierOption = deliveryResult?.courierOptions?.[0];
  const pickupPoints = deliveryResult?.pickupPoints ?? [];

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
                    if (fieldErrors.name) setFieldErrors((p) => ({ ...p, name: "" }));
                  }}
                  aria-invalid={!!fieldErrors.name}
                />
                {fieldErrors.name && (
                  <p className="mt-1 text-sm text-destructive">{fieldErrors.name}</p>
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
                    if (fieldErrors.phone) setFieldErrors((p) => ({ ...p, phone: "" }));
                  }}
                  aria-invalid={!!fieldErrors.phone}
                />
                {fieldErrors.phone && (
                  <p className="mt-1 text-sm text-destructive">{fieldErrors.phone}</p>
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
              {/* City input */}
              <div>
                <label className="mb-1 block text-sm font-medium">
                  Город {deliveryMethod !== "pickup" && <span className="text-destructive">*</span>}
                </label>
                <div className="relative">
                  <Input
                    placeholder="Начните вводить город..."
                    value={city}
                    onChange={(e) => {
                      setCity(e.target.value);
                      if (fieldErrors.city) setFieldErrors((p) => ({ ...p, city: "" }));
                    }}
                    aria-invalid={!!fieldErrors.city}
                  />
                  {deliveryLoading && (
                    <Loader2 className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 animate-spin text-muted-foreground" />
                  )}
                </div>
                {fieldErrors.city && (
                  <p className="mt-1 text-sm text-destructive">{fieldErrors.city}</p>
                )}
              </div>

              {/* Option: Самовывоз со склада */}
              <RadioOption
                name="delivery"
                value="pickup"
                checked={deliveryMethod === "pickup"}
                onChange={() => setDeliveryMethod("pickup")}
                icon={<Package className="h-5 w-5 text-muted-foreground" />}
                title="Самовывоз со склада"
                subtitle="Бесплатно"
              />

              {/* Option: Курьер */}
              <RadioOption
                name="delivery"
                value="courier"
                checked={deliveryMethod === "courier"}
                onChange={() => setDeliveryMethod("courier")}
                icon={<Truck className="h-5 w-5 text-muted-foreground" />}
                title="Курьерская доставка"
                subtitle={
                  !city.trim()
                    ? "Введите город для расчёта"
                    : deliveryLoading
                    ? "Рассчитываем..."
                    : courierOption
                    ? courierOption.isFreeDelivery
                      ? `Бесплатно (${courierOption.estimatedDaysMin}-${courierOption.estimatedDaysMax} дн.)`
                      : `${formatPrice(courierOption.cost)} (${courierOption.estimatedDaysMin}-${courierOption.estimatedDaysMax} дн.)`
                    : "Доставка в ваш город недоступна"
                }
                disabled={!!city.trim() && !deliveryLoading && !courierOption}
              />

              {/* Address for courier */}
              {deliveryMethod === "courier" && (
                <div className="ml-8">
                  <label className="mb-1 block text-sm font-medium">
                    Адрес доставки <span className="text-destructive">*</span>
                  </label>
                  <Input
                    placeholder="Улица, дом, квартира"
                    value={address}
                    onChange={(e) => {
                      setAddress(e.target.value);
                      if (fieldErrors.address) setFieldErrors((p) => ({ ...p, address: "" }));
                    }}
                    aria-invalid={!!fieldErrors.address}
                  />
                  {fieldErrors.address && (
                    <p className="mt-1 text-sm text-destructive">{fieldErrors.address}</p>
                  )}
                </div>
              )}

              {/* Option: Пункт выдачи */}
              {(pickupPoints.length > 0 || (city.trim() && !deliveryLoading)) && (
                <RadioOption
                  name="delivery"
                  value="pickup_point"
                  checked={deliveryMethod === "pickup_point"}
                  onChange={() => setDeliveryMethod("pickup_point")}
                  icon={<MapPin className="h-5 w-5 text-muted-foreground" />}
                  title="Пункт выдачи"
                  subtitle={
                    pickupPoints.length > 0
                      ? `Бесплатно (${pickupPoints.length} ${pickupPoints.length === 1 ? "пункт" : pickupPoints.length < 5 ? "пункта" : "пунктов"})`
                      : "Нет пунктов выдачи в вашем городе"
                  }
                  disabled={pickupPoints.length === 0}
                />
              )}

              {/* Pickup point list */}
              {deliveryMethod === "pickup_point" && pickupPoints.length > 0 && (
                <div className="ml-8 space-y-2">
                  {fieldErrors.pickupPoint && (
                    <p className="text-sm text-destructive">{fieldErrors.pickupPoint}</p>
                  )}
                  {pickupPoints.map((point) => (
                    <label
                      key={point.id}
                      className={`flex cursor-pointer items-start gap-3 rounded-lg border p-3 transition-colors ${
                        selectedPickupPoint?.id === point.id
                          ? "border-primary bg-primary/5"
                          : "border-border hover:border-muted-foreground/30"
                      }`}
                    >
                      <input
                        type="radio"
                        name="pickup_point"
                        value={point.id}
                        checked={selectedPickupPoint?.id === point.id}
                        onChange={() => {
                          setSelectedPickupPoint(point);
                          if (fieldErrors.pickupPoint) setFieldErrors((p) => ({ ...p, pickupPoint: "" }));
                        }}
                        className="sr-only"
                      />
                      <div
                        className={`mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded-full border-2 ${
                          selectedPickupPoint?.id === point.id
                            ? "border-primary"
                            : "border-muted-foreground/40"
                        }`}
                      >
                        {selectedPickupPoint?.id === point.id && (
                          <div className="h-2 w-2 rounded-full bg-primary" />
                        )}
                      </div>
                      <div className="min-w-0">
                        <div className="text-sm font-medium">{point.name}</div>
                        <div className="text-xs text-muted-foreground">{point.address}</div>
                        <div className="mt-1 flex items-center gap-1 text-xs text-muted-foreground">
                          <Clock className="h-3 w-3" />
                          {point.workingHours}
                        </div>
                        {point.phone && (
                          <div className="text-xs text-muted-foreground">{point.phone}</div>
                        )}
                      </div>
                    </label>
                  ))}
                </div>
              )}
            </div>
          </section>

          {/* Payment */}
          <section>
            <h2 className="text-lg font-semibold">Способ оплаты</h2>
            <div className="mt-4 space-y-3">
              <RadioOption
                name="payment"
                value="card"
                checked={paymentMethod === "card"}
                onChange={() => setPaymentMethod("card")}
                icon={<CreditCard className="h-5 w-5 text-muted-foreground" />}
                title="Банковская карта"
              />
              <RadioOption
                name="payment"
                value="cash"
                checked={paymentMethod === "cash"}
                onChange={() => setPaymentMethod("cash")}
                icon={<Banknote className="h-5 w-5 text-muted-foreground" />}
                title="Наличные при получении"
              />
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
                <span className="text-muted-foreground">Товары ({totalItems})</span>
                <span>{formatPrice(totalPrice)}</span>
              </div>
              {deliveryCost > 0 && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Доставка</span>
                  <span>{formatPrice(deliveryCost)}</span>
                </div>
              )}
              {deliveryMethod !== "pickup" && deliveryCost === 0 && city.trim() && courierOption?.isFreeDelivery && (
                <div className="flex justify-between text-green-600">
                  <span>Доставка</span>
                  <span>Бесплатно</span>
                </div>
              )}
            </div>

            {/* Promo code */}
            <div className="mt-4 border-t border-border pt-4">
              {promoResult ? (
                <div className="flex items-center justify-between rounded-md bg-green-500/10 px-3 py-2">
                  <div className="flex items-center gap-2 text-sm">
                    <Tag className="h-4 w-4 text-green-600" />
                    <span className="font-medium text-green-700">{promoResult.code}</span>
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
                    <p className="mt-2 text-sm text-destructive">{promoError}</p>
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

// Reusable radio option component
function RadioOption({
  name,
  value,
  checked,
  onChange,
  icon,
  title,
  subtitle,
  disabled,
}: {
  name: string;
  value: string;
  checked: boolean;
  onChange: () => void;
  icon: React.ReactNode;
  title: string;
  subtitle?: string;
  disabled?: boolean;
}) {
  return (
    <label
      className={`flex cursor-pointer items-center gap-3 rounded-lg border p-4 transition-colors ${
        disabled
          ? "cursor-not-allowed opacity-50 border-border"
          : checked
          ? "border-primary bg-primary/5"
          : "border-border hover:border-muted-foreground/30"
      }`}
    >
      <input
        type="radio"
        name={name}
        value={value}
        checked={checked}
        onChange={onChange}
        disabled={disabled}
        className="sr-only"
      />
      <div
        className={`flex h-5 w-5 items-center justify-center rounded-full border-2 ${
          checked ? "border-primary" : "border-muted-foreground/40"
        }`}
      >
        {checked && <div className="h-2.5 w-2.5 rounded-full bg-primary" />}
      </div>
      {icon}
      <div>
        <div className="font-medium">{title}</div>
        {subtitle && (
          <div className="text-sm text-muted-foreground">{subtitle}</div>
        )}
      </div>
    </label>
  );
}
