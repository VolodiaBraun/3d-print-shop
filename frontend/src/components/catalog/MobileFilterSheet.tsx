"use client";

import { useState } from "react";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { SlidersHorizontal } from "lucide-react";
import { FilterSidebar } from "./FilterSidebar";
import type { CatalogFilters } from "@/hooks/useCatalogFilters";

interface MobileFilterSheetProps {
  filters: CatalogFilters;
  setFilter: (key: keyof CatalogFilters, value: unknown) => void;
  setFilters: (updates: Partial<CatalogFilters>) => void;
  resetFilters: () => void;
  hasActiveFilters: boolean;
  activeFilterCount: number;
}

export function MobileFilterSheet({
  filters,
  setFilter,
  setFilters,
  resetFilters,
  hasActiveFilters,
  activeFilterCount,
}: MobileFilterSheetProps) {
  const [open, setOpen] = useState(false);

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button variant="outline" size="sm" className="gap-2">
          <SlidersHorizontal className="h-4 w-4" />
          Фильтры
          {activeFilterCount > 0 && (
            <Badge variant="default" className="ml-1 h-5 w-5 p-0 justify-center text-xs">
              {activeFilterCount}
            </Badge>
          )}
        </Button>
      </SheetTrigger>
      <SheetContent side="left" className="w-[300px] p-0">
        <SheetHeader className="px-4 pt-4 pb-2">
          <SheetTitle>Фильтры</SheetTitle>
        </SheetHeader>
        <ScrollArea className="h-[calc(100vh-120px)] px-4">
          <FilterSidebar
            filters={filters}
            setFilter={setFilter}
            setFilters={setFilters}
            resetFilters={resetFilters}
            hasActiveFilters={hasActiveFilters}
          />
        </ScrollArea>
        <div className="p-4 border-t">
          <Button className="w-full" onClick={() => setOpen(false)}>
            Показать результаты
          </Button>
        </div>
      </SheetContent>
    </Sheet>
  );
}
