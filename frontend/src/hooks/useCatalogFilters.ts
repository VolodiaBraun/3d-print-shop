"use client";

import { useSearchParams, useRouter, usePathname } from "next/navigation";
import { useCallback, useMemo } from "react";

export interface CatalogFilters {
  category: string | null;
  search: string | null;
  sort: string;
  minPrice: number | null;
  maxPrice: number | null;
  materials: string[];
  page: number;
}

const DEFAULT_SORT = "newest";

export function useCatalogFilters() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();

  const filters: CatalogFilters = useMemo(() => {
    const materialParam = searchParams.get("material");
    return {
      category: searchParams.get("category"),
      search: searchParams.get("search"),
      sort: searchParams.get("sort") || DEFAULT_SORT,
      minPrice: searchParams.get("min_price")
        ? Number(searchParams.get("min_price"))
        : null,
      maxPrice: searchParams.get("max_price")
        ? Number(searchParams.get("max_price"))
        : null,
      materials: materialParam ? materialParam.split(",").filter(Boolean) : [],
      page: searchParams.get("page") ? Number(searchParams.get("page")) : 1,
    };
  }, [searchParams]);

  const buildUrl = useCallback(
    (updates: Partial<CatalogFilters>) => {
      const params = new URLSearchParams(searchParams.toString());

      const merged = { ...filters, ...updates };

      // Reset page when filters change (not when page itself changes)
      if (!("page" in updates)) {
        params.delete("page");
      }

      // category
      if (merged.category) {
        params.set("category", merged.category);
      } else {
        params.delete("category");
      }

      // search
      if (merged.search) {
        params.set("search", merged.search);
      } else {
        params.delete("search");
      }

      // sort
      if (merged.sort && merged.sort !== DEFAULT_SORT) {
        params.set("sort", merged.sort);
      } else {
        params.delete("sort");
      }

      // min_price
      if (merged.minPrice != null && merged.minPrice > 0) {
        params.set("min_price", String(merged.minPrice));
      } else {
        params.delete("min_price");
      }

      // max_price
      if (merged.maxPrice != null && merged.maxPrice > 0) {
        params.set("max_price", String(merged.maxPrice));
      } else {
        params.delete("max_price");
      }

      // materials
      if (merged.materials.length > 0) {
        params.set("material", merged.materials.join(","));
      } else {
        params.delete("material");
      }

      // page
      if ("page" in updates && merged.page > 1) {
        params.set("page", String(merged.page));
      }

      const qs = params.toString();
      return qs ? `${pathname}?${qs}` : pathname;
    },
    [searchParams, filters, pathname]
  );

  const setFilter = useCallback(
    (key: keyof CatalogFilters, value: unknown) => {
      router.replace(buildUrl({ [key]: value } as Partial<CatalogFilters>), {
        scroll: false,
      });
    },
    [router, buildUrl]
  );

  const setFilters = useCallback(
    (updates: Partial<CatalogFilters>) => {
      router.replace(buildUrl(updates), { scroll: false });
    },
    [router, buildUrl]
  );

  const removeFilter = useCallback(
    (key: keyof CatalogFilters, value?: string) => {
      if (key === "materials" && value) {
        const next = filters.materials.filter((m) => m !== value);
        setFilter("materials", next);
      } else if (key === "sort") {
        setFilter("sort", DEFAULT_SORT);
      } else {
        setFilter(key, key === "materials" ? [] : null);
      }
    },
    [filters.materials, setFilter]
  );

  const resetFilters = useCallback(() => {
    router.replace(pathname, { scroll: false });
  }, [router, pathname]);

  const hasActiveFilters = useMemo(() => {
    return !!(
      filters.category ||
      filters.search ||
      filters.minPrice ||
      filters.maxPrice ||
      filters.materials.length > 0
    );
  }, [filters]);

  const activeFilterCount = useMemo(() => {
    let count = 0;
    if (filters.category) count++;
    if (filters.search) count++;
    if (filters.minPrice || filters.maxPrice) count++;
    count += filters.materials.length;
    return count;
  }, [filters]);

  return {
    filters,
    setFilter,
    setFilters,
    removeFilter,
    resetFilters,
    hasActiveFilters,
    activeFilterCount,
  };
}
