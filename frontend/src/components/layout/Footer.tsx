"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { getContentBlock } from "@/lib/api";

interface FooterContent {
  description: string;
  telegram: string;
  email: string;
  phone: string;
  address: string;
}

const defaults: FooterContent = {
  description:
    "3D-печатные изделия премиального качества. Фигурки, модели и аксессуары для коллекционеров и геймеров.",
  telegram: "@avangard3d",
  email: "info@avangard-print.ru",
  phone: "",
  address: "",
};

export function Footer() {
  const [content, setContent] = useState<FooterContent>(defaults);

  useEffect(() => {
    getContentBlock<FooterContent>("footer")
      .then((data) => setContent({ ...defaults, ...data }))
      .catch(() => {});
  }, []);

  return (
    <footer className="border-t border-border bg-card mt-auto">
      <div className="mx-auto max-w-7xl px-4 py-8">
        <div className="grid grid-cols-1 gap-8 sm:grid-cols-3">
          <div>
            <h3 className="mb-3 font-bold text-lg">АВАНГАРД</h3>
            <p className="text-sm text-muted-foreground">
              {content.description}
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
              <li>
                <Link href="/custom-order" className="hover:text-foreground">
                  Индивидуальный заказ
                </Link>
              </li>
            </ul>
          </div>
          <div>
            <h4 className="mb-3 font-semibold text-sm">Контакты</h4>
            <ul className="space-y-2 text-sm text-muted-foreground">
              {content.telegram && <li>Telegram: {content.telegram}</li>}
              {content.email && <li>Email: {content.email}</li>}
              {content.phone && <li>Телефон: {content.phone}</li>}
              {content.address && <li>Адрес: {content.address}</li>}
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
