import Link from "next/link";
import { ChevronRight, Home } from "lucide-react";
import type { Product } from "@/lib/types";

export function Breadcrumbs({ product }: { product: Product }) {
  return (
    <nav
      aria-label="Breadcrumb"
      className="mb-6 flex items-center gap-1.5 text-sm text-muted-foreground"
    >
      <Link
        href="/"
        className="flex items-center gap-1 hover:text-foreground transition-colors"
      >
        <Home className="h-3.5 w-3.5" />
        <span>Главная</span>
      </Link>
      <ChevronRight className="h-3.5 w-3.5" />
      <Link
        href="/catalog"
        className="hover:text-foreground transition-colors"
      >
        Каталог
      </Link>
      {product.category && (
        <>
          <ChevronRight className="h-3.5 w-3.5" />
          <Link
            href={`/catalog?category=${product.category.slug}`}
            className="hover:text-foreground transition-colors"
          >
            {product.category.name}
          </Link>
        </>
      )}
      <ChevronRight className="h-3.5 w-3.5" />
      <span className="text-foreground truncate max-w-[200px]">
        {product.name}
      </span>
    </nav>
  );
}
