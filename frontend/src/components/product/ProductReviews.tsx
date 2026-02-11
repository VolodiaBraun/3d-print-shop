"use client";

import { useEffect, useState } from "react";
import { Star } from "lucide-react";
import { getProductReviews, type ReviewData } from "@/lib/api";

function StarRating({ rating }: { rating: number }) {
  return (
    <div className="flex gap-0.5">
      {[1, 2, 3, 4, 5].map((i) => (
        <Star
          key={i}
          className={`h-4 w-4 ${
            i <= rating
              ? "fill-yellow-400 text-yellow-400"
              : "text-muted-foreground/30"
          }`}
        />
      ))}
    </div>
  );
}

interface ProductReviewsProps {
  productId: number;
  rating: number;
  reviewsCount: number;
}

export function ProductReviews({
  productId,
  rating,
  reviewsCount,
}: ProductReviewsProps) {
  const [reviews, setReviews] = useState<ReviewData[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getProductReviews(productId)
      .then(setReviews)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [productId]);

  if (loading) {
    return (
      <div className="space-y-4">
        <h2 className="text-xl font-bold">Отзывы</h2>
        <div className="animate-pulse space-y-3">
          {[1, 2].map((i) => (
            <div key={i} className="h-24 rounded-lg bg-muted" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-bold">
          Отзывы{" "}
          {reviewsCount > 0 && (
            <span className="text-base font-normal text-muted-foreground">
              ({reviewsCount})
            </span>
          )}
        </h2>
        {reviewsCount > 0 && (
          <div className="flex items-center gap-2">
            <StarRating rating={Math.round(rating)} />
            <span className="text-sm font-medium">{rating.toFixed(1)}</span>
          </div>
        )}
      </div>

      {reviews.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          Пока нет отзывов. Будьте первым!
        </p>
      ) : (
        <div className="space-y-3">
          {reviews.map((review) => (
            <div key={review.id} className="rounded-lg border p-4">
              <div className="mb-2 flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium">
                    {review.user?.firstName || "Покупатель"}
                    {review.user?.lastName ? ` ${review.user.lastName.charAt(0)}.` : ""}
                  </span>
                  <StarRating rating={review.rating} />
                </div>
                <span className="text-xs text-muted-foreground">
                  {new Date(review.createdAt).toLocaleDateString("ru-RU")}
                </span>
              </div>
              {review.comment && (
                <p className="text-sm text-muted-foreground">{review.comment}</p>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
