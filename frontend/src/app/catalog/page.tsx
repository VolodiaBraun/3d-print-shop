import { Suspense } from "react";
import { CatalogContent } from "@/components/catalog/CatalogContent";
import { ProductCardSkeleton } from "@/components/product/ProductCardSkeleton";
import { Skeleton } from "@/components/ui/skeleton";

export const metadata = {
  title: "Каталог — АВАНГАРД",
  description: "Каталог 3D-печатных изделий с фильтрами и сортировкой",
};

function CatalogSkeleton() {
  return (
    <div className="mx-auto max-w-7xl px-4 py-8">
      <Skeleton className="h-9 w-48 mb-6" />
      <div className="flex gap-8">
        <aside className="hidden md:block w-[260px] shrink-0 space-y-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="space-y-2">
              <Skeleton className="h-5 w-24" />
              <Skeleton className="h-9 w-full" />
              <Skeleton className="h-9 w-full" />
            </div>
          ))}
        </aside>
        <div className="flex-1 min-w-0">
          <div className="grid grid-cols-2 gap-4 lg:grid-cols-3">
            {Array.from({ length: 12 }).map((_, i) => (
              <ProductCardSkeleton key={i} />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

export default function CatalogPage() {
  return (
    <Suspense fallback={<CatalogSkeleton />}>
      <CatalogContent />
    </Suspense>
  );
}
