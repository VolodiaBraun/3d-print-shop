"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { ClipboardList, Search, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { getOrder } from "@/lib/api";

export default function OrdersPage() {
  const router = useRouter();
  const [orderNumber, setOrderNumber] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSearch() {
    const num = orderNumber.trim();
    if (!num) return;

    setLoading(true);
    setError("");
    try {
      await getOrder(num);
      router.push(`/order/${num}`);
    } catch {
      setError("Заказ не найден. Проверьте номер и попробуйте снова.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="mx-auto max-w-lg px-4 py-12 text-center">
      <ClipboardList className="mx-auto h-16 w-16 text-muted-foreground/50" />
      <h1 className="mt-4 text-2xl font-bold">Мои заказы</h1>
      <p className="mt-2 text-muted-foreground">
        Введите номер заказа, чтобы посмотреть его статус
      </p>

      <div className="mt-6 flex gap-2">
        <Input
          placeholder="Номер заказа (например, ORD-...)"
          value={orderNumber}
          onChange={(e) => {
            setOrderNumber(e.target.value);
            if (error) setError("");
          }}
          onKeyDown={(e) => {
            if (e.key === "Enter") handleSearch();
          }}
        />
        <Button
          onClick={handleSearch}
          disabled={loading || !orderNumber.trim()}
          className="shrink-0"
        >
          {loading ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <Search className="h-4 w-4" />
          )}
        </Button>
      </div>

      {error && (
        <p className="mt-3 text-sm text-destructive">{error}</p>
      )}
    </div>
  );
}
