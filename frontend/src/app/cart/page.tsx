"use client";

import Link from "next/link";
import Image from "next/image";
import { ShoppingCart, Trash2, Minus, Plus, ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useCart } from "@/lib/cart-context";

function formatPrice(price: number): string {
  return new Intl.NumberFormat("ru-RU").format(price) + " \u20BD";
}

export default function CartPage() {
  const { items, updateQuantity, removeItem, clearCart, totalItems, totalPrice } =
    useCart();

  if (items.length === 0) {
    return (
      <div className="mx-auto max-w-7xl px-4 py-16 text-center">
        <ShoppingCart className="mx-auto h-16 w-16 text-muted-foreground/50" />
        <h1 className="mt-4 text-2xl font-bold">Корзина пуста</h1>
        <p className="mt-2 text-muted-foreground">
          Добавьте товары из каталога, чтобы оформить заказ
        </p>
        <Button asChild className="mt-6">
          <Link href="/catalog">Перейти в каталог</Link>
        </Button>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-7xl px-4 py-8">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold lg:text-3xl">
          Корзина{" "}
          <span className="text-muted-foreground font-normal text-lg">
            ({totalItems}{" "}
            {totalItems === 1
              ? "товар"
              : totalItems < 5
                ? "товара"
                : "товаров"}
            )
          </span>
        </h1>
        <Button
          variant="ghost"
          size="sm"
          className="text-muted-foreground hover:text-destructive"
          onClick={clearCart}
        >
          <Trash2 className="mr-1 h-4 w-4" />
          Очистить
        </Button>
      </div>

      <div className="mt-8 grid gap-8 lg:grid-cols-[1fr_360px]">
        {/* Cart items */}
        <div className="space-y-4">
          {items.map((item) => (
            <div
              key={item.productId}
              className="flex gap-4 rounded-lg border border-border bg-card p-4"
            >
              {/* Image */}
              <Link
                href={`/product/${item.slug}`}
                className="shrink-0"
              >
                <div className="relative h-24 w-24 overflow-hidden rounded-md bg-muted sm:h-28 sm:w-28">
                  {item.image ? (
                    <Image
                      src={item.image}
                      alt={item.name}
                      fill
                      className="object-cover"
                      sizes="112px"
                    />
                  ) : (
                    <div className="flex h-full items-center justify-center text-muted-foreground">
                      <ShoppingCart className="h-8 w-8" />
                    </div>
                  )}
                </div>
              </Link>

              {/* Info */}
              <div className="flex flex-1 flex-col justify-between min-w-0">
                <div>
                  <Link
                    href={`/product/${item.slug}`}
                    className="font-medium hover:text-primary transition-colors line-clamp-2"
                  >
                    {item.name}
                  </Link>
                  <div className="mt-1 flex items-baseline gap-2">
                    <span className="font-semibold">
                      {formatPrice(item.price)}
                    </span>
                    {item.oldPrice && item.oldPrice > item.price && (
                      <span className="text-sm text-muted-foreground line-through">
                        {formatPrice(item.oldPrice)}
                      </span>
                    )}
                  </div>
                </div>

                {/* Quantity + remove */}
                <div className="mt-3 flex items-center justify-between">
                  <div className="flex items-center gap-1">
                    <Button
                      variant="outline"
                      size="icon"
                      className="h-8 w-8"
                      disabled={item.quantity <= 1}
                      onClick={() =>
                        updateQuantity(item.productId, item.quantity - 1)
                      }
                    >
                      <Minus className="h-3 w-3" />
                    </Button>
                    <span className="w-10 text-center text-sm font-medium">
                      {item.quantity}
                    </span>
                    <Button
                      variant="outline"
                      size="icon"
                      className="h-8 w-8"
                      disabled={item.quantity >= item.stockQuantity}
                      onClick={() =>
                        updateQuantity(item.productId, item.quantity + 1)
                      }
                    >
                      <Plus className="h-3 w-3" />
                    </Button>
                  </div>

                  <div className="flex items-center gap-3">
                    <span className="font-bold">
                      {formatPrice(item.price * item.quantity)}
                    </span>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8 text-muted-foreground hover:text-destructive"
                      onClick={() => removeItem(item.productId)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* Summary */}
        <div className="lg:sticky lg:top-24">
          <div className="rounded-lg border border-border bg-card p-6">
            <h2 className="text-lg font-bold">Итого</h2>

            <div className="mt-4 space-y-3">
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">
                  Товары ({totalItems})
                </span>
                <span>{formatPrice(totalPrice)}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">Доставка</span>
                <span className="text-muted-foreground">Рассчитывается при оформлении</span>
              </div>
            </div>

            <div className="mt-4 border-t border-border pt-4">
              <div className="flex justify-between text-lg font-bold">
                <span>К оплате</span>
                <span>{formatPrice(totalPrice)}</span>
              </div>
            </div>

            <Button className="mt-6 w-full gap-2" size="lg" asChild>
              <Link href="/checkout">
                Оформить заказ
                <ArrowRight className="h-4 w-4" />
              </Link>
            </Button>

            <Button
              variant="ghost"
              className="mt-2 w-full text-muted-foreground"
              asChild
            >
              <Link href="/catalog">Продолжить покупки</Link>
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
