"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { getCategories } from "@/lib/api";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

const categoryIcons: Record<string, string> = {
  figurki: "🎭",
  modeli: "🏗️",
  aksessuary: "🎮",
  dekor: "🏠",
  podarki: "🎁",
};

export function CategoriesSection() {
  const { data: categories, isLoading } = useQuery({
    queryKey: ["categories"],
    queryFn: getCategories,
  });

  return (
    <section>
      <h2 className="mb-6 text-2xl font-bold">Категории</h2>
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
        {isLoading
          ? Array.from({ length: 4 }).map((_, i) => (
              <Card key={i} className="overflow-hidden">
                <CardContent className="flex flex-col items-center gap-3 p-6">
                  <Skeleton className="h-12 w-12 rounded-full" />
                  <Skeleton className="h-4 w-20" />
                </CardContent>
              </Card>
            ))
          : categories?.map((category) => (
              <Link key={category.id} href={`/catalog?category=${category.slug}`}>
                <Card className="group overflow-hidden transition-all hover:shadow-md hover:-translate-y-0.5 duration-200 h-full">
                  <CardContent className="flex flex-col items-center gap-3 p-6">
                    <span className="text-4xl">
                      {categoryIcons[category.slug] || "📦"}
                    </span>
                    <span className="text-sm font-medium text-center group-hover:text-primary transition-colors">
                      {category.name}
                    </span>
                    {category.children && category.children.length > 0 && (
                      <span className="text-xs text-muted-foreground">
                        {category.children.length} подкатегорий
                      </span>
                    )}
                  </CardContent>
                </Card>
              </Link>
            ))}
        {!isLoading && categories?.length === 0 && (
          <p className="col-span-full text-center text-muted-foreground py-8">
            Категории скоро появятся
          </p>
        )}
      </div>
    </section>
  );
}
