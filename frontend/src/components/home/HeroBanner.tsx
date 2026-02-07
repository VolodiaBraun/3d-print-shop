import Link from "next/link";
import { Button } from "@/components/ui/button";
import { ArrowRight } from "lucide-react";

export function HeroBanner() {
  return (
    <section className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-primary/10 via-background to-primary/5 border border-border">
      <div className="px-6 py-16 sm:px-12 sm:py-24 lg:py-32">
        <div className="max-w-2xl">
          <h1 className="text-4xl font-bold tracking-tight sm:text-5xl lg:text-6xl">
            3D-печатные изделия{" "}
            <span className="text-primary">премиум качества</span>
          </h1>
          <p className="mt-4 text-lg text-muted-foreground sm:text-xl">
            Фигурки, модели и аксессуары для коллекционеров и геймеров. Каждое
            изделие создано с вниманием к деталям.
          </p>
          <div className="mt-8 flex flex-wrap gap-3">
            <Button size="lg" asChild>
              <Link href="/catalog">
                Смотреть каталог
                <ArrowRight className="ml-2 h-4 w-4" />
              </Link>
            </Button>
            <Button size="lg" variant="outline" asChild>
              <Link href="/catalog?sort=newest">Новинки</Link>
            </Button>
          </div>
        </div>
      </div>
    </section>
  );
}
