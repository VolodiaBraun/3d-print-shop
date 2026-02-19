import Link from "next/link";
import { FlaskConical, ArrowRight } from "lucide-react";

export function CustomOrderBanner() {
  return (
    <Link
      href="/custom-order"
      className="group block rounded-2xl bg-gradient-to-r from-violet-600/20 to-violet-500/10 border border-violet-500/20 p-5 transition-colors hover:bg-violet-600/25"
    >
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-4">
          <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-violet-500/20">
            <FlaskConical className="h-6 w-6 text-violet-400" />
          </div>
          <div>
            <p className="font-semibold text-foreground">
              Индивидуальная 3D-печать
            </p>
            <p className="mt-0.5 text-sm text-muted-foreground">
              Загрузите свою модель или опишите идею — рассчитаем стоимость
            </p>
          </div>
        </div>
        <ArrowRight className="h-5 w-5 shrink-0 text-violet-400 transition-transform group-hover:translate-x-1" />
      </div>
    </Link>
  );
}
