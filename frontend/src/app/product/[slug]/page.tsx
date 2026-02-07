import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getProductBySlug } from "@/lib/api";
import { Breadcrumbs } from "@/components/product/Breadcrumbs";
import { ProductDetailContent } from "@/components/product/ProductDetailContent";

interface Props {
  params: Promise<{ slug: string }>;
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  try {
    const product = await getProductBySlug(slug);
    const mainImage =
      product.images?.find((img) => img.isMain) || product.images?.[0];
    return {
      title: `${product.name} — АВАНГАРД`,
      description:
        product.shortDescription ||
        product.description?.slice(0, 160) ||
        `Купить ${product.name} в интернет-магазине АВАНГАРД`,
      openGraph: {
        title: product.name,
        description:
          product.shortDescription || product.description?.slice(0, 160),
        images: mainImage
          ? [{ url: mainImage.urlLarge || mainImage.url }]
          : [],
      },
    };
  } catch {
    return { title: "Товар не найден — АВАНГАРД" };
  }
}

export default async function ProductPage({ params }: Props) {
  const { slug } = await params;

  let product;
  try {
    product = await getProductBySlug(slug);
  } catch {
    notFound();
  }

  return (
    <div className="mx-auto max-w-7xl px-4 py-8">
      <Breadcrumbs product={product} />
      <ProductDetailContent product={product} />
    </div>
  );
}
