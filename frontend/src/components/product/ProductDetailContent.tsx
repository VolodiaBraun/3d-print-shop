"use client";

import type { Product } from "@/lib/types";
import { ProductGallery } from "./ProductGallery";
import { ProductInfo } from "./ProductInfo";
import { ProductCharacteristics } from "./ProductCharacteristics";

export function ProductDetailContent({ product }: { product: Product }) {
  return (
    <>
      <div className="grid grid-cols-1 gap-8 lg:grid-cols-5">
        {/* Gallery: 3/5 = 60% */}
        <div className="lg:col-span-3">
          <ProductGallery
            images={product.images || []}
            productName={product.name}
          />
        </div>
        {/* Info: 2/5 = 40% */}
        <div className="lg:col-span-2">
          <ProductInfo product={product} />
        </div>
      </div>
      {/* Characteristics below */}
      <ProductCharacteristics product={product} />
    </>
  );
}
