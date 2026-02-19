CREATE TABLE content_blocks (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(100) UNIQUE NOT NULL,
    data JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO content_blocks (slug, data) VALUES
('hero', '{"title":"3D-печатные изделия","titleAccent":"премиум качества","subtitle":"Фигурки, модели и аксессуары для коллекционеров и геймеров. Каждое изделие создано с вниманием к деталям.","ctaPrimary":"Смотреть каталог","ctaPrimaryLink":"/catalog","ctaSecondary":"Новинки","ctaSecondaryLink":"/catalog?sort=newest"}'),
('footer', '{"description":"3D-печатные изделия премиального качества. Фигурки, модели и аксессуары для коллекционеров и геймеров.","telegram":"@avangard3d","email":"info@avangard-print.ru","phone":"","address":""}');
