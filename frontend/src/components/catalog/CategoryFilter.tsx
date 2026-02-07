"use client";

import { useQuery } from "@tanstack/react-query";
import { getCategories } from "@/lib/api";
import type { Category } from "@/lib/types";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

interface CategoryFilterProps {
  selectedCategory: string | null;
  onCategoryChange: (slug: string | null) => void;
}

export function CategoryFilter({
  selectedCategory,
  onCategoryChange,
}: CategoryFilterProps) {
  const { data: categories, isLoading } = useQuery({
    queryKey: ["categories"],
    queryFn: getCategories,
  });

  if (isLoading) {
    return (
      <div className="space-y-2">
        <h3 className="text-sm font-semibold mb-3">Категории</h3>
        {Array.from({ length: 5 }).map((_, i) => (
          <Skeleton key={i} className="h-8 w-full" />
        ))}
      </div>
    );
  }

  return (
    <div>
      <h3 className="text-sm font-semibold mb-3">Категории</h3>
      <div className="space-y-0.5">
        <button
          onClick={() => onCategoryChange(null)}
          className={cn(
            "w-full text-left px-3 py-2 rounded-md text-sm transition-colors",
            !selectedCategory
              ? "bg-primary/10 text-primary font-medium"
              : "text-muted-foreground hover:text-foreground hover:bg-muted"
          )}
        >
          Все категории
        </button>
        {categories?.map((category) => (
          <CategoryItem
            key={category.id}
            category={category}
            selectedCategory={selectedCategory}
            onCategoryChange={onCategoryChange}
            depth={0}
          />
        ))}
      </div>
    </div>
  );
}

function CategoryItem({
  category,
  selectedCategory,
  onCategoryChange,
  depth,
}: {
  category: Category;
  selectedCategory: string | null;
  onCategoryChange: (slug: string | null) => void;
  depth: number;
}) {
  const isSelected = selectedCategory === category.slug;

  return (
    <>
      <button
        onClick={() => onCategoryChange(isSelected ? null : category.slug)}
        className={cn(
          "w-full text-left px-3 py-2 rounded-md text-sm transition-colors",
          isSelected
            ? "bg-primary/10 text-primary font-medium"
            : "text-muted-foreground hover:text-foreground hover:bg-muted"
        )}
        style={{ paddingLeft: `${12 + depth * 16}px` }}
      >
        {category.name}
      </button>
      {category.children?.map((child) => (
        <CategoryItem
          key={child.id}
          category={child}
          selectedCategory={selectedCategory}
          onCategoryChange={onCategoryChange}
          depth={depth + 1}
        />
      ))}
    </>
  );
}
