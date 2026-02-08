"use client";

import { useState, useEffect } from "react";
import { Star, ShoppingCart, Check, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useCart } from "@/lib/cart-context";
import type { Product } from "@/lib/types";

function formatPrice(price: number): string {
  return new Intl.NumberFormat("ru-RU").format(price) + " \u20BD";
}

function RatingStars({ rating }: { rating: number }) {
  return (
    <div className="flex items-center gap-1">
      {[1, 2, 3, 4, 5].map((star) => (
        <Star
          key={star}
          className={`h-4 w-4 ${
            star <= Math.round(rating)
              ? "fill-yellow-400 text-yellow-400"
              : "fill-muted text-muted"
          }`}
        />
      ))}
      {rating > 0 && (
        <span className="ml-1 text-sm text-muted-foreground">
          {rating.toFixed(1)}
        </span>
      )}
    </div>
  );
}

export function ProductInfo({ product }: { product: Product }) {
  const { addItem, getItemQuantity } = useCart();
  const [justAdded, setJustAdded] = useState(false);

  const inStock = product.stockQuantity > 0;
  const hasDiscount = product.oldPrice && product.oldPrice > product.price;
  const discountPercent = hasDiscount
    ? Math.round(
        ((product.oldPrice! - product.price) / product.oldPrice!) * 100
      )
    : 0;

  const inCartQty = getItemQuantity(product.id);

  useEffect(() => {
    if (justAdded) {
      const timer = setTimeout(() => setJustAdded(false), 2000);
      return () => clearTimeout(timer);
    }
  }, [justAdded]);

  const handleAddToCart = () => {
    addItem(product);
    setJustAdded(true);
  };

  return (
    <div className="space-y-6">
      {/* Name */}
      <h1 className="text-2xl font-bold lg:text-3xl">{product.name}</h1>

      {/* Rating + SKU */}
      <div className="flex items-center gap-4">
        <RatingStars rating={product.rating} />
        {product.sku && (
          <span className="text-sm text-muted-foreground">
            Арт: {product.sku}
          </span>
        )}
      </div>

      {/* Price block */}
      <div className="flex items-baseline gap-3">
        <span className="text-3xl font-bold">{formatPrice(product.price)}</span>
        {hasDiscount && (
          <>
            <span className="text-lg text-muted-foreground line-through">
              {formatPrice(product.oldPrice!)}
            </span>
            <Badge variant="destructive">-{discountPercent}%</Badge>
          </>
        )}
      </div>

      {/* Material badge */}
      {product.material && (
        <Badge variant="secondary">{product.material}</Badge>
      )}

      {/* Short description */}
      {product.shortDescription && (
        <p className="text-muted-foreground leading-relaxed">
          {product.shortDescription}
        </p>
      )}

      {/* Add to cart / Out of stock */}
      <div className="pt-2">
        {inStock ? (
          <div className="flex flex-wrap items-center gap-3">
            {justAdded ? (
              <Button
                size="lg"
                className="w-full sm:w-auto gap-2 bg-green-600 hover:bg-green-700"
              >
                <Check className="h-5 w-5" />
                Добавлено
              </Button>
            ) : inCartQty > 0 ? (
              <>
                <Button
                  size="lg"
                  variant="outline"
                  className="gap-2"
                  onClick={handleAddToCart}
                  disabled={inCartQty >= product.stockQuantity}
                >
                  <Plus className="h-5 w-5" />
                  Ещё
                </Button>
                <span className="text-sm text-muted-foreground">
                  В корзине: {inCartQty} шт.
                </span>
              </>
            ) : (
              <Button
                size="lg"
                className="w-full sm:w-auto gap-2"
                onClick={handleAddToCart}
              >
                <ShoppingCart className="h-5 w-5" />
                В корзину
              </Button>
            )}
          </div>
        ) : (
          <Button
            size="lg"
            variant="secondary"
            disabled
            className="w-full sm:w-auto"
          >
            Нет в наличии
          </Button>
        )}
        {inStock && product.stockQuantity <= 5 && (
          <p className="mt-2 text-sm text-orange-400">
            Осталось мало: {product.stockQuantity} шт.
          </p>
        )}
      </div>

      {/* Full description */}
      {product.description && (
        <div className="pt-4 border-t border-border">
          <h2 className="text-lg font-semibold mb-2">Описание</h2>
          <div className="text-muted-foreground leading-relaxed whitespace-pre-line">
            {product.description}
          </div>
        </div>
      )}
    </div>
  );
}
