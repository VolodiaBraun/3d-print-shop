import { Skeleton } from "@/components/ui/skeleton";

export default function ProductPageLoading() {
  return (
    <div className="mx-auto max-w-7xl px-4 py-8">
      {/* Breadcrumb skeleton */}
      <div className="mb-6 flex items-center gap-2">
        <Skeleton className="h-4 w-16" />
        <Skeleton className="h-4 w-4" />
        <Skeleton className="h-4 w-24" />
        <Skeleton className="h-4 w-4" />
        <Skeleton className="h-4 w-32" />
      </div>

      <div className="grid grid-cols-1 gap-8 lg:grid-cols-5">
        {/* Gallery skeleton */}
        <div className="lg:col-span-3">
          <Skeleton className="aspect-square w-full rounded-xl" />
          <div className="mt-3 flex gap-2">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="h-16 w-16 rounded-lg" />
            ))}
          </div>
        </div>
        {/* Info skeleton */}
        <div className="lg:col-span-2 space-y-4">
          <Skeleton className="h-8 w-3/4" />
          <Skeleton className="h-5 w-32" />
          <Skeleton className="h-10 w-40" />
          <Skeleton className="h-20 w-full" />
          <Skeleton className="h-12 w-full sm:w-40" />
        </div>
      </div>
    </div>
  );
}
