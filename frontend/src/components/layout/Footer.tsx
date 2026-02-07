import Link from "next/link";

export function Footer() {
  return (
    <footer className="border-t border-border bg-card mt-auto">
      <div className="mx-auto max-w-7xl px-4 py-8">
        <div className="grid grid-cols-1 gap-8 sm:grid-cols-3">
          <div>
            <h3 className="mb-3 font-bold text-lg">АВАНГАРД</h3>
            <p className="text-sm text-muted-foreground">
              3D-печатные изделия премиального качества. Фигурки, модели и
              аксессуары для коллекционеров и геймеров.
            </p>
          </div>
          <div>
            <h4 className="mb-3 font-semibold text-sm">Навигация</h4>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li>
                <Link href="/catalog" className="hover:text-foreground">
                  Каталог
                </Link>
              </li>
              <li>
                <Link
                  href="/catalog?sort=newest"
                  className="hover:text-foreground"
                >
                  Новинки
                </Link>
              </li>
              <li>
                <Link
                  href="/catalog?sort=popular"
                  className="hover:text-foreground"
                >
                  Популярное
                </Link>
              </li>
            </ul>
          </div>
          <div>
            <h4 className="mb-3 font-semibold text-sm">Контакты</h4>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li>Telegram: @avangard3d</li>
              <li>Email: info@avangard3d.ru</li>
            </ul>
          </div>
        </div>
        <div className="mt-8 border-t border-border pt-4 text-center text-xs text-muted-foreground">
          &copy; {new Date().getFullYear()} АВАНГАРД. Все права защищены.
        </div>
      </div>
    </footer>
  );
}
