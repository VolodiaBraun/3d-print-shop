"use client";

import { useEffect, useState } from "react";
import type { Product } from "@/lib/types";
import { ProductGallery } from "./ProductGallery";
import { ProductInfo } from "./ProductInfo";
import { ProductCharacteristics } from "./ProductCharacteristics";
import { ProductReviews } from "./ProductReviews";
import { ReviewForm } from "./ReviewForm";
import { useAuth } from "@/lib/auth-context";
import { getMyOrders } from "@/lib/api";

export function ProductDetailContent({ product }: { product: Product }) {
  const { isAuthenticated } = useAuth();
  const [deliveredOrderIds, setDeliveredOrderIds] = useState<number[]>([]);

  // Find delivered orders containing this product
  useEffect(() => {
    if (!isAuthenticated) return;
    getMyOrders()
      .then((orders) => {
        const ids = orders
          .filter(
            (o) =>
              o.status === "delivered" &&
              o.items?.some((item) => item.productId === product.id)
          )
          .map((o) => o.id);
        setDeliveredOrderIds(ids);
      })
      .catch(() => {});
  }, [isAuthenticated, product.id]);

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

      {/* Reviews */}
      <div className="mt-10 space-y-6">
        <ProductReviews
          productId={product.id}
          rating={product.rating || 0}
          reviewsCount={(product as Product & { reviewsCount?: number }).reviewsCount || 0}
        />
        <ReviewForm
          productId={product.id}
          deliveredOrderIds={deliveredOrderIds}
        />
      </div>
    </>
  );
}
