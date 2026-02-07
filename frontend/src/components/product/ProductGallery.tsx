"use client";

import { useState, useRef, useCallback } from "react";
import Image from "next/image";
import { cn } from "@/lib/utils";
import type { ProductImage } from "@/lib/types";
import { FullscreenViewer } from "./FullscreenViewer";
import { Package } from "lucide-react";

interface ProductGalleryProps {
  images: ProductImage[];
  productName: string;
}

export function ProductGallery({ images, productName }: ProductGalleryProps) {
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [fullscreenOpen, setFullscreenOpen] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);

  // Sort images: main first, then by displayOrder
  const sorted = [...images].sort((a, b) => {
    if (a.isMain && !b.isMain) return -1;
    if (!a.isMain && b.isMain) return 1;
    return a.displayOrder - b.displayOrder;
  });

  // Track scroll position for mobile dot indicators
  const handleScroll = useCallback(() => {
    const el = scrollRef.current;
    if (!el || el.clientWidth === 0) return;
    const index = Math.round(el.scrollLeft / el.clientWidth);
    setSelectedIndex(index);
  }, []);

  // Scroll to selected image on mobile when thumbnail/dot is clicked
  const scrollToIndex = useCallback((index: number) => {
    setSelectedIndex(index);
    const el = scrollRef.current;
    if (el) {
      el.scrollTo({ left: index * el.clientWidth, behavior: "smooth" });
    }
  }, []);

  if (sorted.length === 0) {
    return (
      <div className="aspect-square rounded-xl bg-muted flex flex-col items-center justify-center gap-2">
        <Package className="h-16 w-16 text-muted-foreground/40" />
        <span className="text-muted-foreground text-sm">Нет изображений</span>
      </div>
    );
  }

  const selectedImage = sorted[selectedIndex] || sorted[0];

  return (
    <>
      <div className="space-y-3">
        {/* Mobile: scroll-snap carousel */}
        <div
          ref={scrollRef}
          onScroll={handleScroll}
          className="flex snap-x snap-mandatory overflow-x-auto scrollbar-hide lg:hidden"
        >
          {sorted.map((img, i) => (
            <div
              key={img.id}
              className="w-full flex-none snap-start"
              onClick={() => {
                setSelectedIndex(i);
                setFullscreenOpen(true);
              }}
            >
              <div className="relative aspect-square rounded-xl overflow-hidden bg-muted cursor-zoom-in">
                <Image
                  src={img.urlLarge || img.urlMedium || img.url}
                  alt={`${productName} — фото ${i + 1}`}
                  fill
                  className="object-cover"
                  sizes="100vw"
                  priority={i === 0}
                />
              </div>
            </div>
          ))}
        </div>

        {/* Mobile: dot indicators */}
        {sorted.length > 1 && (
          <div className="flex justify-center gap-1.5 lg:hidden">
            {sorted.map((_, i) => (
              <button
                key={i}
                onClick={() => scrollToIndex(i)}
                className={cn(
                  "h-2 w-2 rounded-full transition-colors",
                  i === selectedIndex
                    ? "bg-primary"
                    : "bg-muted-foreground/30"
                )}
                aria-label={`Фото ${i + 1}`}
              />
            ))}
          </div>
        )}

        {/* Desktop: main image */}
        <div
          className="hidden lg:block relative aspect-square rounded-xl overflow-hidden bg-muted cursor-zoom-in"
          onClick={() => setFullscreenOpen(true)}
        >
          <Image
            src={
              selectedImage.urlLarge ||
              selectedImage.urlMedium ||
              selectedImage.url
            }
            alt={`${productName} — фото ${selectedIndex + 1}`}
            fill
            className="object-cover"
            sizes="60vw"
            priority
          />
        </div>

        {/* Desktop: thumbnail strip */}
        {sorted.length > 1 && (
          <div className="hidden lg:flex gap-2 overflow-x-auto scrollbar-hide">
            {sorted.map((img, i) => (
              <button
                key={img.id}
                onClick={() => setSelectedIndex(i)}
                className={cn(
                  "relative h-16 w-16 flex-none rounded-lg overflow-hidden border-2 transition-colors",
                  i === selectedIndex
                    ? "border-primary"
                    : "border-transparent hover:border-muted-foreground/50"
                )}
              >
                <Image
                  src={img.urlThumbnail || img.url}
                  alt={`Миниатюра ${i + 1}`}
                  fill
                  className="object-cover"
                  sizes="64px"
                />
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Fullscreen viewer */}
      <FullscreenViewer
        images={sorted}
        productName={productName}
        initialIndex={selectedIndex}
        open={fullscreenOpen}
        onClose={() => setFullscreenOpen(false)}
      />
    </>
  );
}
