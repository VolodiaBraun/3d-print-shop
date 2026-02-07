"use client";

import { useQuery } from "@tanstack/react-query";
import { getProducts } from "@/lib/api";
import { ProductCard } from "./ProductCard";
import { ProductCardSkeleton } from "./ProductCardSkeleton";

interface ProductGridProps {
  title: string;
  sort?: string;
  limit?: number;
}

export function ProductGrid({ title, sort, limit = 8 }: ProductGridProps) {
  const { data, isLoading } = useQuery({
    queryKey: ["products", sort, limit],
    queryFn: () => getProducts({ sort, limit, page: 1 }),
  });

  return (
    <section>
      <h2 className="mb-6 text-2xl font-bold">{title}</h2>
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
        {isLoading
          ? Array.from({ length: limit }).map((_, i) => (
              <ProductCardSkeleton key={i} />
            ))
          : data?.data.map((product) => (
              <ProductCard key={product.id} product={product} />
            ))}
        {!isLoading && data?.data.length === 0 && (
          <p className="col-span-full text-center text-muted-foreground py-8">
            Товары скоро появятся
          </p>
        )}
      </div>
    </section>
  );
}
