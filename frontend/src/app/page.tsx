import { HeroBanner } from "@/components/home/HeroBanner";
import { CustomOrderBanner } from "@/components/home/CustomOrderBanner";
import { ProductGrid } from "@/components/product/ProductGrid";
import { CategoriesSection } from "@/components/home/CategoriesSection";

export default function HomePage() {
  return (
    <div className="mx-auto max-w-7xl space-y-12 px-4 py-8">
      <HeroBanner />
      <CustomOrderBanner />
      <ProductGrid title="Популярные товары" sort="popular" limit={8} />
      <ProductGrid title="Новинки" sort="newest" limit={8} />
      <CategoriesSection />
    </div>
  );
}
