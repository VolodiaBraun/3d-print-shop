import { cn } from "@/lib/utils";
import type { Product, Dimensions } from "@/lib/types";

function formatDimensions(d: Dimensions): string {
  const parts: string[] = [];
  if (d.length) parts.push(`${d.length}`);
  if (d.width) parts.push(`${d.width}`);
  if (d.height) parts.push(`${d.height}`);
  return parts.join(" × ") + " см";
}

function formatPrintTime(minutes: number): string {
  if (minutes < 60) return `${minutes} мин`;
  const hours = Math.floor(minutes / 60);
  const mins = minutes % 60;
  return mins > 0 ? `${hours} ч ${mins} мин` : `${hours} ч`;
}

export function ProductCharacteristics({ product }: { product: Product }) {
  const rows: { label: string; value: string }[] = [];

  if (product.material)
    rows.push({ label: "Материал", value: product.material });
  if (product.weight)
    rows.push({ label: "Вес", value: `${product.weight} г` });
  if (product.dimensions)
    rows.push({ label: "Размеры", value: formatDimensions(product.dimensions) });
  if (product.printTime)
    rows.push({ label: "Время печати", value: formatPrintTime(product.printTime) });
  if (product.sku)
    rows.push({ label: "Артикул", value: product.sku });

  if (rows.length === 0) return null;

  return (
    <div className="mt-12">
      <h2 className="text-xl font-semibold mb-4">Характеристики</h2>
      <div className="rounded-lg border border-border overflow-hidden">
        {rows.map((row, i) => (
          <div
            key={row.label}
            className={cn(
              "flex justify-between px-4 py-3 text-sm",
              i % 2 === 0 ? "bg-muted/30" : "bg-transparent"
            )}
          >
            <span className="text-muted-foreground">{row.label}</span>
            <span className="font-medium">{row.value}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
