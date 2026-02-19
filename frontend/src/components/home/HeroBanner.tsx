"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { ArrowRight } from "lucide-react";
import { getContentBlock } from "@/lib/api";

interface HeroContent {
  title: string;
  titleAccent: string;
  subtitle: string;
  ctaPrimary: string;
  ctaPrimaryLink: string;
  ctaSecondary: string;
  ctaSecondaryLink: string;
}

const defaults: HeroContent = {
  title: "3D-печатные изделия",
  titleAccent: "премиум качества",
  subtitle:
    "Фигурки, модели и аксессуары для коллекционеров и геймеров. Каждое изделие создано с вниманием к деталям.",
  ctaPrimary: "Смотреть каталог",
  ctaPrimaryLink: "/catalog",
  ctaSecondary: "Новинки",
  ctaSecondaryLink: "/catalog?sort=newest",
};

export function HeroBanner() {
  const [content, setContent] = useState<HeroContent>(defaults);

  useEffect(() => {
    getContentBlock<HeroContent>("hero")
      .then((data) => setContent({ ...defaults, ...data }))
      .catch(() => {});
  }, []);

  return (
    <section className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-primary/10 via-background to-primary/5 border border-border">
      <div className="px-6 py-16 sm:px-12 sm:py-24 lg:py-32">
        <div className="max-w-2xl">
          <h1 className="text-4xl font-bold tracking-tight sm:text-5xl lg:text-6xl">
            {content.title}{" "}
            <span className="text-primary">{content.titleAccent}</span>
          </h1>
          <p className="mt-4 text-lg text-muted-foreground sm:text-xl">
            {content.subtitle}
          </p>
          <div className="mt-8 flex flex-wrap gap-3">
            <Button size="lg" asChild>
              <Link href={content.ctaPrimaryLink}>
                {content.ctaPrimary}
                <ArrowRight className="ml-2 h-4 w-4" />
              </Link>
            </Button>
            <Button size="lg" variant="outline" asChild>
              <Link href={content.ctaSecondaryLink}>
                {content.ctaSecondary}
              </Link>
            </Button>
          </div>
        </div>
      </div>
    </section>
  );
}
