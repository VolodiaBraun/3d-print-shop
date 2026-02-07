import Image from "next/image";
import Link from "next/link";
import { Star } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import type { Product } from "@/lib/types";

function getMainImage(product: Product): string | null {
  if (!product.images || product.images.length === 0) return null;
  const main = product.images.find((img) => img.isMain);
  const img = main || product.images[0];
  return img.urlThumbnail || img.url;
}

function formatPrice(price: number): string {
  return new Intl.NumberFormat("ru-RU").format(price) + " \u20BD";
}

function RatingStars({ rating }: { rating: number }) {
  return (
    <div className="flex items-center gap-1">
      {[1, 2, 3, 4, 5].map((star) => (
        <Star
          key={star}
          className={`h-3.5 w-3.5 ${
            star <= Math.round(rating)
              ? "fill-yellow-400 text-yellow-400"
              : "fill-muted text-muted"
          }`}
        />
      ))}
      {rating > 0 && (
        <span className="ml-1 text-xs text-muted-foreground">
          {rating.toFixed(1)}
        </span>
      )}
    </div>
  );
}

export function ProductCard({ product }: { product: Product }) {
  const imageUrl = getMainImage(product);

  return (
    <Link href={`/product/${product.slug}`}>
      <Card className="group overflow-hidden transition-all hover:shadow-lg hover:shadow-primary/5 hover:-translate-y-1 duration-200 h-full">
        {/* Image */}
        <div className="relative aspect-square overflow-hidden bg-muted">
          {imageUrl ? (
            <Image
              src={imageUrl}
              alt={product.name}
              fill
              className="object-cover transition-transform duration-300 group-hover:scale-105"
              sizes="(max-width: 640px) 50vw, (max-width: 1024px) 33vw, 25vw"
            />
          ) : (
            <div className="flex h-full items-center justify-center text-muted-foreground">
              <span className="text-4xl">📦</span>
            </div>
          )}
          {/* Badges */}
          <div className="absolute top-2 left-2 flex flex-col gap-1">
            {product.oldPrice && product.oldPrice > product.price && (
              <Badge variant="destructive" className="text-xs">
                -
                {Math.round(
                  ((product.oldPrice - product.price) / product.oldPrice) * 100
                )}
                %
              </Badge>
            )}
            {product.isFeatured && (
              <Badge className="bg-orange-500 text-xs text-white hover:bg-orange-600">
                HIT
              </Badge>
            )}
          </div>
          {product.material && (
            <Badge
              variant="secondary"
              className="absolute bottom-2 right-2 text-xs"
            >
              {product.material}
            </Badge>
          )}
        </div>

        {/* Content */}
        <CardContent className="p-3">
          <h3 className="mb-1 line-clamp-2 text-sm font-medium leading-tight">
            {product.name}
          </h3>
          <RatingStars rating={product.rating} />
          <div className="mt-2 flex items-baseline gap-2">
            <span className="text-lg font-bold">
              {formatPrice(product.price)}
            </span>
            {product.oldPrice && product.oldPrice > product.price && (
              <span className="text-sm text-muted-foreground line-through">
                {formatPrice(product.oldPrice)}
              </span>
            )}
          </div>
          {product.stockQuantity <= 0 && (
            <p className="mt-1 text-xs text-destructive">Нет в наличии</p>
          )}
        </CardContent>
      </Card>
    </Link>
  );
}
