export default async function ProductPage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;

  return (
    <div className="mx-auto max-w-7xl px-4 py-8">
      <h1 className="text-3xl font-bold">Товар: {slug}</h1>
      <p className="mt-4 text-muted-foreground">
        Скоро здесь будет детальная карточка товара с галереей.
      </p>
    </div>
  );
}
