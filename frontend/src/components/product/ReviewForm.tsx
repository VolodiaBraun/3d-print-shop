"use client";

import { useState } from "react";
import { Star, Loader2, Check } from "lucide-react";
import { Button } from "@/components/ui/button";
import { createReview } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import Link from "next/link";

interface ReviewFormProps {
  productId: number;
  /** IDs of delivered orders containing this product */
  deliveredOrderIds: number[];
}

export function ReviewForm({ productId, deliveredOrderIds }: ReviewFormProps) {
  const { isAuthenticated } = useAuth();
  const [rating, setRating] = useState(0);
  const [hoverRating, setHoverRating] = useState(0);
  const [comment, setComment] = useState("");
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState("");

  if (!isAuthenticated) {
    return (
      <div className="rounded-lg border border-dashed p-4 text-center">
        <p className="text-sm text-muted-foreground">
          <Link href="/login" className="text-primary hover:underline">
            Войдите
          </Link>
          , чтобы оставить отзыв
        </p>
      </div>
    );
  }

  if (deliveredOrderIds.length === 0) {
    return null;
  }

  if (success) {
    return (
      <div className="flex items-center gap-2 rounded-lg bg-green-500/10 p-4 text-sm text-green-600 dark:text-green-400">
        <Check className="h-4 w-4" />
        Спасибо! Отзыв отправлен на модерацию
      </div>
    );
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (rating === 0) {
      setError("Выберите рейтинг");
      return;
    }
    setError("");
    setLoading(true);
    try {
      await createReview(productId, {
        orderId: deliveredOrderIds[0],
        rating,
        comment: comment.trim() || undefined,
      });
      setSuccess(true);
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message || "Не удалось отправить отзыв";
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-3 rounded-lg border p-4">
      <h3 className="text-sm font-semibold">Оставить отзыв</h3>

      {error && (
        <p className="text-sm text-destructive">{error}</p>
      )}

      <div className="flex gap-1">
        {[1, 2, 3, 4, 5].map((i) => (
          <button
            key={i}
            type="button"
            className="p-0.5"
            onMouseEnter={() => setHoverRating(i)}
            onMouseLeave={() => setHoverRating(0)}
            onClick={() => setRating(i)}
          >
            <Star
              className={`h-6 w-6 transition-colors ${
                i <= (hoverRating || rating)
                  ? "fill-yellow-400 text-yellow-400"
                  : "text-muted-foreground/30"
              }`}
            />
          </button>
        ))}
        {rating > 0 && (
          <span className="ml-2 text-sm text-muted-foreground">
            {rating === 1 && "Ужасно"}
            {rating === 2 && "Плохо"}
            {rating === 3 && "Нормально"}
            {rating === 4 && "Хорошо"}
            {rating === 5 && "Отлично"}
          </span>
        )}
      </div>

      <textarea
        className="w-full resize-none rounded-md border bg-transparent px-3 py-2 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring"
        rows={3}
        placeholder="Ваш комментарий (необязательно)"
        value={comment}
        onChange={(e) => setComment(e.target.value)}
        maxLength={1000}
      />

      <Button type="submit" size="sm" disabled={loading || rating === 0}>
        {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
        Отправить отзыв
      </Button>
    </form>
  );
}
