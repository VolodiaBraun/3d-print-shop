"use client";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

const SORT_OPTIONS = [
  { value: "newest", label: "Новинки" },
  { value: "popular", label: "Популярные" },
  { value: "price_asc", label: "Сначала дешёвые" },
  { value: "price_desc", label: "Сначала дорогие" },
  { value: "rating", label: "По рейтингу" },
];

interface SortSelectProps {
  sort: string;
  onSortChange: (sort: string) => void;
}

export function SortSelect({ sort, onSortChange }: SortSelectProps) {
  return (
    <Select value={sort} onValueChange={onSortChange}>
      <SelectTrigger className="w-[200px] h-9">
        <SelectValue placeholder="Сортировка" />
      </SelectTrigger>
      <SelectContent>
        {SORT_OPTIONS.map((option) => (
          <SelectItem key={option.value} value={option.value}>
            {option.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
