"use client";

import { useQuery } from "@tanstack/react-query";
import { getCategories } from "@/lib/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { X } from "lucide-react";
import type { CatalogFilters } from "@/hooks/useCatalogFilters";
import type { Category } from "@/lib/types";

interface ActiveFiltersProps {
  filters: CatalogFilters;
  onRemoveFilter: (key: keyof CatalogFilters, value?: string) => void;
  onResetAll: () => void;
}

function findCategoryName(
  categories: Category[] | undefined,
  slug: string
): string {
  if (!categories) return slug;
  for (const cat of categories) {
    if (cat.slug === slug) return cat.name;
    if (cat.children) {
      const found = findCategoryName(cat.children, slug);
      if (found !== slug) return found;
    }
  }
  return slug;
}

export function ActiveFilters({
  filters,
  onRemoveFilter,
  onResetAll,
}: ActiveFiltersProps) {
  const { data: categories } = useQuery({
    queryKey: ["categories"],
    queryFn: getCategories,
  });

  const hasFilters =
    filters.category ||
    filters.search ||
    filters.minPrice ||
    filters.maxPrice ||
    filters.materials.length > 0;

  if (!hasFilters) return null;

  return (
    <div className="flex flex-wrap items-center gap-2 mb-4">
      {filters.category && (
        <FilterBadge
          label={findCategoryName(categories, filters.category)}
          onRemove={() => onRemoveFilter("category")}
        />
      )}
      {filters.search && (
        <FilterBadge
          label={`«${filters.search}»`}
          onRemove={() => onRemoveFilter("search")}
        />
      )}
      {filters.minPrice != null && (
        <FilterBadge
          label={`от ${filters.minPrice} ₽`}
          onRemove={() => onRemoveFilter("minPrice")}
        />
      )}
      {filters.maxPrice != null && (
        <FilterBadge
          label={`до ${filters.maxPrice} ₽`}
          onRemove={() => onRemoveFilter("maxPrice")}
        />
      )}
      {filters.materials.map((m) => (
        <FilterBadge
          key={m}
          label={m}
          onRemove={() => onRemoveFilter("materials", m)}
        />
      ))}
      <Button
        variant="ghost"
        size="sm"
        onClick={onResetAll}
        className="text-xs text-muted-foreground h-7"
      >
        Сбросить все
      </Button>
    </div>
  );
}

function FilterBadge({
  label,
  onRemove,
}: {
  label: string;
  onRemove: () => void;
}) {
  return (
    <Badge variant="secondary" className="gap-1 pr-1">
      {label}
      <button
        onClick={onRemove}
        className="ml-1 rounded-full p-0.5 hover:bg-muted-foreground/20 transition-colors"
      >
        <X className="h-3 w-3" />
      </button>
    </Badge>
  );
}
