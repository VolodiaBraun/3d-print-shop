import Link from "next/link";
import { Button } from "@/components/ui/button";
import { PackageX } from "lucide-react";

export default function ProductNotFound() {
  return (
    <div className="mx-auto max-w-7xl px-4 py-24 text-center">
      <PackageX className="mx-auto h-16 w-16 text-muted-foreground/50 mb-4" />
      <h1 className="text-3xl font-bold mb-4">Товар не найден</h1>
      <p className="text-muted-foreground mb-8">
        Возможно, товар был удалён или ссылка неверна.
      </p>
      <Button asChild>
        <Link href="/catalog">Перейти в каталог</Link>
      </Button>
    </div>
  );
}
