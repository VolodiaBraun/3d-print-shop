"use client";

import { useCatalogFilters } from "@/hooks/useCatalogFilters";
import { useCatalogProducts } from "@/hooks/useCatalogProducts";
import { FilterSidebar } from "./FilterSidebar";
import { MobileFilterSheet } from "./MobileFilterSheet";
import { SortSelect } from "./SortSelect";
import { ActiveFilters } from "./ActiveFilters";
import { CatalogProductGrid } from "./CatalogProductGrid";

export function CatalogContent() {
  const {
    filters,
    setFilter,
    setFilters,
    removeFilter,
    resetFilters,
    hasActiveFilters,
    activeFilterCount,
  } = useCatalogFilters();

  const { products, meta, isLoading, isLoadingMore, hasMore, loadMore } =
    useCatalogProducts(filters);

  return (
    <div className="mx-auto max-w-7xl px-4 py-8">
      {/* Page header */}
      <h1 className="text-3xl font-bold mb-6">Каталог</h1>

      {/* Mobile: filter trigger + sort */}
      <div className="flex items-center gap-3 mb-4 md:hidden">
        <MobileFilterSheet
          filters={filters}
          setFilter={setFilter}
          setFilters={setFilters}
          resetFilters={resetFilters}
          hasActiveFilters={hasActiveFilters}
          activeFilterCount={activeFilterCount}
        />
        <SortSelect
          sort={filters.sort}
          onSortChange={(v) => setFilter("sort", v)}
        />
      </div>

      {/* Active filter chips */}
      <ActiveFilters
        filters={filters}
        onRemoveFilter={removeFilter}
        onResetAll={resetFilters}
      />

      {/* Main layout: sidebar + grid */}
      <div className="flex gap-8">
        {/* Desktop sidebar */}
        <aside className="hidden md:block w-[260px] shrink-0">
          <div className="sticky top-20">
            <FilterSidebar
              filters={filters}
              setFilter={setFilter}
              setFilters={setFilters}
              resetFilters={resetFilters}
              hasActiveFilters={hasActiveFilters}
            />
          </div>
        </aside>

        {/* Product grid area */}
        <div className="flex-1 min-w-0">
          {/* Desktop sort row */}
          <div className="hidden md:flex items-center justify-between mb-4">
            {meta && !isLoading ? (
              <span className="text-sm text-muted-foreground">
                {/* Count shown in CatalogProductGrid */}
              </span>
            ) : (
              <span />
            )}
            <SortSelect
              sort={filters.sort}
              onSortChange={(v) => setFilter("sort", v)}
            />
          </div>

          <CatalogProductGrid
            products={products}
            meta={meta}
            isLoading={isLoading}
            isLoadingMore={isLoadingMore}
            hasMore={hasMore}
            onLoadMore={loadMore}
            onResetFilters={resetFilters}
          />
        </div>
      </div>
    </div>
  );
}
