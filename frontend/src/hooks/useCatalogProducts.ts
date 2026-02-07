"use client";

import { useInfiniteQuery } from "@tanstack/react-query";
import { getProducts } from "@/lib/api";
import type { CatalogFilters } from "./useCatalogFilters";

const LIMIT = 12;

export function useCatalogProducts(filters: CatalogFilters) {
  const queryParams = {
    category: filters.category || undefined,
    search: filters.search || undefined,
    sort: filters.sort,
    min_price: filters.minPrice ?? undefined,
    max_price: filters.maxPrice ?? undefined,
    material: filters.materials.length > 0 ? filters.materials.join(",") : undefined,
  };

  const query = useInfiniteQuery({
    queryKey: ["catalog-products", queryParams],
    queryFn: async ({ pageParam = 1 }) => {
      return getProducts({
        ...queryParams,
        page: pageParam,
        limit: LIMIT,
      });
    },
    initialPageParam: 1,
    getNextPageParam: (lastPage) => {
      if (lastPage.meta.page < lastPage.meta.totalPages) {
        return lastPage.meta.page + 1;
      }
      return undefined;
    },
  });

  const products = query.data?.pages.flatMap((page) => page.data) ?? [];
  const meta = query.data?.pages[query.data.pages.length - 1]?.meta ?? null;

  return {
    products,
    meta,
    isLoading: query.isLoading,
    isLoadingMore: query.isFetchingNextPage,
    hasMore: query.hasNextPage ?? false,
    loadMore: query.fetchNextPage,
    error: query.error,
  };
}
