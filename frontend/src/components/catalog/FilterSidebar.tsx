"use client";

import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { SearchFilter } from "./SearchFilter";
import { CategoryFilter } from "./CategoryFilter";
import { PriceFilter } from "./PriceFilter";
import { MaterialFilter } from "./MaterialFilter";
import type { CatalogFilters } from "@/hooks/useCatalogFilters";

interface FilterSidebarProps {
  filters: CatalogFilters;
  setFilter: (key: keyof CatalogFilters, value: unknown) => void;
  setFilters: (updates: Partial<CatalogFilters>) => void;
  resetFilters: () => void;
  hasActiveFilters: boolean;
}

export function FilterSidebar({
  filters,
  setFilter,
  setFilters,
  resetFilters,
  hasActiveFilters,
}: FilterSidebarProps) {
  return (
    <div className="space-y-4">
      <SearchFilter
        search={filters.search}
        onSearchChange={(v) => setFilter("search", v)}
      />
      <Separator />
      <CategoryFilter
        selectedCategory={filters.category}
        onCategoryChange={(v) => setFilter("category", v)}
      />
      <Separator />
      <PriceFilter
        minPrice={filters.minPrice}
        maxPrice={filters.maxPrice}
        onPriceChange={(min, max) => setFilters({ minPrice: min, maxPrice: max })}
      />
      <Separator />
      <MaterialFilter
        selectedMaterials={filters.materials}
        onMaterialsChange={(v) => setFilter("materials", v)}
      />
      {hasActiveFilters && (
        <>
          <Separator />
          <Button
            variant="outline"
            size="sm"
            onClick={resetFilters}
            className="w-full"
          >
            Сбросить фильтры
          </Button>
        </>
      )}
    </div>
  );
}
