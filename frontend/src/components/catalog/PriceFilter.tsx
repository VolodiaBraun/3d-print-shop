"use client";

import { useState, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { useDebounce } from "@/hooks/useDebounce";

interface PriceFilterProps {
  minPrice: number | null;
  maxPrice: number | null;
  onPriceChange: (min: number | null, max: number | null) => void;
}

export function PriceFilter({
  minPrice,
  maxPrice,
  onPriceChange,
}: PriceFilterProps) {
  const [localMin, setLocalMin] = useState(minPrice?.toString() ?? "");
  const [localMax, setLocalMax] = useState(maxPrice?.toString() ?? "");

  const debouncedMin = useDebounce(localMin, 500);
  const debouncedMax = useDebounce(localMax, 500);

  // Sync from URL to local state
  useEffect(() => {
    setLocalMin(minPrice?.toString() ?? "");
  }, [minPrice]);

  useEffect(() => {
    setLocalMax(maxPrice?.toString() ?? "");
  }, [maxPrice]);

  // Apply debounced values
  useEffect(() => {
    const min = debouncedMin ? Number(debouncedMin) : null;
    const max = debouncedMax ? Number(debouncedMax) : null;
    if (min !== minPrice || max !== maxPrice) {
      onPriceChange(min, max);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debouncedMin, debouncedMax]);

  return (
    <div>
      <h3 className="text-sm font-semibold mb-3">Цена, ₽</h3>
      <div className="flex items-center gap-2">
        <Input
          type="number"
          placeholder="От"
          min={0}
          value={localMin}
          onChange={(e) => setLocalMin(e.target.value)}
          className="h-9"
        />
        <span className="text-muted-foreground shrink-0">—</span>
        <Input
          type="number"
          placeholder="До"
          min={0}
          value={localMax}
          onChange={(e) => setLocalMax(e.target.value)}
          className="h-9"
        />
      </div>
    </div>
  );
}
