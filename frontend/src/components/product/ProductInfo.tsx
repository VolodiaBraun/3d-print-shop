"use client";

import { Star, ShoppingCart } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import type { Product } from "@/lib/types";

function formatPrice(price: number): string {
  return new Intl.NumberFormat("ru-RU").format(price) + " ₽";
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
  const inStock = product.stockQuantity > 0;
  const hasDiscount = product.oldPrice && product.oldPrice > product.price;
  const discountPercent = hasDiscount
    ? Math.round(
        ((product.oldPrice! - product.price) / product.oldPrice!) * 100
      )
    : 0;

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
          <Button size="lg" className="w-full sm:w-auto gap-2">
            <ShoppingCart className="h-5 w-5" />
            В корзину
          </Button>
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
