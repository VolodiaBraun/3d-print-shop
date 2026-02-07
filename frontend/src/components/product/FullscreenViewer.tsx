"use client";

import { useState, useCallback, useEffect } from "react";
import Image from "next/image";
import { Dialog, VisuallyHidden } from "radix-ui";
import { X, ChevronLeft, ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";
import type { ProductImage } from "@/lib/types";

interface FullscreenViewerProps {
  images: ProductImage[];
  productName: string;
  initialIndex: number;
  open: boolean;
  onClose: () => void;
}

export function FullscreenViewer({
  images,
  productName,
  initialIndex,
  open,
  onClose,
}: FullscreenViewerProps) {
  const [currentIndex, setCurrentIndex] = useState(initialIndex);
  const [zoomed, setZoomed] = useState(false);
  const [transformOrigin, setTransformOrigin] = useState("center center");

  // Sync initialIndex when dialog opens
  useEffect(() => {
    if (open) {
      setCurrentIndex(initialIndex);
      setZoomed(false);
    }
  }, [open, initialIndex]);

  const goNext = useCallback(() => {
    setZoomed(false);
    setCurrentIndex((i) => (i < images.length - 1 ? i + 1 : i));
  }, [images.length]);

  const goPrev = useCallback(() => {
    setZoomed(false);
    setCurrentIndex((i) => (i > 0 ? i - 1 : i));
  }, []);

  const handleImageClick = useCallback(
    (e: React.MouseEvent<HTMLDivElement>) => {
      if (zoomed) {
        setZoomed(false);
      } else {
        const rect = e.currentTarget.getBoundingClientRect();
        const x = ((e.clientX - rect.left) / rect.width) * 100;
        const y = ((e.clientY - rect.top) / rect.height) * 100;
        setTransformOrigin(`${x}% ${y}%`);
        setZoomed(true);
      }
    },
    [zoomed]
  );

  // Keyboard navigation
  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === "ArrowLeft") goPrev();
      if (e.key === "ArrowRight") goNext();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [open, goPrev, goNext]);

  const currentImage = images[currentIndex];
  if (!currentImage) return null;

  return (
    <Dialog.Root open={open} onOpenChange={(v) => !v && onClose()}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-50 bg-black/95" />
        <Dialog.Content className="fixed inset-0 z-50 flex items-center justify-center outline-none">
          <VisuallyHidden.Root>
            <Dialog.Title>
              {productName} — фото {currentIndex + 1}
            </Dialog.Title>
          </VisuallyHidden.Root>

          {/* Close button */}
          <button
            onClick={onClose}
            className="absolute top-4 right-4 z-10 p-2 rounded-full text-white/70 hover:text-white hover:bg-white/10 transition-colors"
          >
            <X className="h-6 w-6" />
          </button>

          {/* Counter */}
          <div className="absolute top-4 left-4 z-10 text-white/60 text-sm">
            {currentIndex + 1} / {images.length}
          </div>

          {/* Previous */}
          {currentIndex > 0 && (
            <button
              onClick={goPrev}
              className="absolute left-4 z-10 p-2 rounded-full text-white/70 hover:text-white hover:bg-white/10 transition-colors"
            >
              <ChevronLeft className="h-8 w-8" />
            </button>
          )}

          {/* Next */}
          {currentIndex < images.length - 1 && (
            <button
              onClick={goNext}
              className="absolute right-4 z-10 p-2 rounded-full text-white/70 hover:text-white hover:bg-white/10 transition-colors"
            >
              <ChevronRight className="h-8 w-8" />
            </button>
          )}

          {/* Image with zoom */}
          <div
            className={cn(
              "relative w-full h-full max-w-[90vw] max-h-[85vh] transition-transform duration-300 overflow-hidden",
              zoomed ? "cursor-zoom-out" : "cursor-zoom-in"
            )}
            style={{
              transform: zoomed ? "scale(2)" : "scale(1)",
              transformOrigin,
            }}
            onClick={handleImageClick}
          >
            <Image
              src={currentImage.urlLarge || currentImage.url}
              alt={`${productName} — фото ${currentIndex + 1}`}
              fill
              className="object-contain"
              sizes="90vw"
              priority
            />
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
