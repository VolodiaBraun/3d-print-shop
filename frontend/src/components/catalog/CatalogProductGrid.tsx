"use client";

import { ProductCard } from "@/components/product/ProductCard";
import { ProductCardSkeleton } from "@/components/product/ProductCardSkeleton";
import { Button } from "@/components/ui/button";
import { Loader2, PackageOpen } from "lucide-react";
import type { Product, PaginationMeta } from "@/lib/types";

interface CatalogProductGridProps {
  products: Product[];
  meta: PaginationMeta | null;
  isLoading: boolean;
  isLoadingMore: boolean;
  hasMore: boolean;
  onLoadMore: () => void;
  onResetFilters: () => void;
}

function pluralProducts(count: number): string {
  const mod10 = count % 10;
  const mod100 = count % 100;
  if (mod10 === 1 && mod100 !== 11) return `${count} товар`;
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20))
    return `${count} товара`;
  return `${count} товаров`;
}

export function CatalogProductGrid({
  products,
  meta,
  isLoading,
  isLoadingMore,
  hasMore,
  onLoadMore,
  onResetFilters,
}: CatalogProductGridProps) {
  if (isLoading) {
    return (
      <div>
        <div className="h-6 mb-4" />
        <div className="grid grid-cols-2 gap-4 lg:grid-cols-3">
          {Array.from({ length: 12 }).map((_, i) => (
            <ProductCardSkeleton key={i} />
          ))}
        </div>
      </div>
    );
  }

  if (products.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <PackageOpen className="h-16 w-16 text-muted-foreground/50 mb-4" />
        <h3 className="text-lg font-medium mb-2">Ничего не найдено</h3>
        <p className="text-sm text-muted-foreground mb-4">
          По вашему запросу товары не найдены. Попробуйте изменить параметры
          поиска.
        </p>
        <Button variant="outline" onClick={onResetFilters}>
          Сбросить фильтры
        </Button>
      </div>
    );
  }

  const remaining = meta ? meta.total - products.length : 0;

  return (
    <div>
      {meta && (
        <p className="text-sm text-muted-foreground mb-4">
          Найдено {pluralProducts(meta.total)}
        </p>
      )}

      <div className="grid grid-cols-2 gap-4 lg:grid-cols-3">
        {products.map((product) => (
          <ProductCard key={product.id} product={product} />
        ))}
      </div>

      {hasMore && (
        <div className="flex justify-center mt-8">
          <Button
            variant="outline"
            size="lg"
            onClick={onLoadMore}
            disabled={isLoadingMore}
          >
            {isLoadingMore ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Загрузка...
              </>
            ) : (
              `Показать ещё${remaining > 0 ? ` (${remaining})` : ""}`
            )}
          </Button>
        </div>
      )}
    </div>
  );
}
