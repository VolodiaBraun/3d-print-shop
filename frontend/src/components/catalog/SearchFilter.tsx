"use client";

import { useState, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Search, X } from "lucide-react";
import { useDebounce } from "@/hooks/useDebounce";

interface SearchFilterProps {
  search: string | null;
  onSearchChange: (search: string | null) => void;
}

export function SearchFilter({ search, onSearchChange }: SearchFilterProps) {
  const [localValue, setLocalValue] = useState(search ?? "");
  const debouncedValue = useDebounce(localValue, 400);

  // Sync from URL
  useEffect(() => {
    setLocalValue(search ?? "");
  }, [search]);

  // Apply debounced
  useEffect(() => {
    const value = debouncedValue.trim() || null;
    if (value !== search) {
      onSearchChange(value);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debouncedValue]);

  return (
    <div>
      <h3 className="text-sm font-semibold mb-3">Поиск</h3>
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
        <Input
          placeholder="Найти товар..."
          value={localValue}
          onChange={(e) => setLocalValue(e.target.value)}
          className="pl-9 pr-9 h-9"
        />
        {localValue && (
          <button
            onClick={() => {
              setLocalValue("");
              onSearchChange(null);
            }}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
          >
            <X className="h-4 w-4" />
          </button>
        )}
      </div>
    </div>
  );
}
