# ТЕХНИЧЕСКОЕ ЗАДАНИЕ
## Интернет-магазин 3D-моделей с Telegram-ботом

**Версия:** 2.0
**Дата:** 05.02.2026

---

## 1. ОБЩЕЕ ОПИСАНИЕ ПРОЕКТА

### 1.1 Цель проекта
Создание интернет-магазина для продажи 3D-печатных изделий с возможностью размещения заказов через веб-интерфейс и Telegram-бот.

### 1.2 Целевая аудитория
- Покупатели 3D-моделей (физические лица)
- Администраторы магазина
- Менеджеры по обработке заказов

### 1.3 Основные возможности
- Просмотр каталога товаров
- Оформление заказов
- Интеграция с Telegram
- Административная панель для управления товарами
- Система обработки заказов

---

## 2. ТЕХНИЧЕСКИЙ СТЕК

### 2.1 Backend
- **Язык:** Go 1.21+
- **Фреймворк:** Gin / Echo / Fiber
- **База данных:** PostgreSQL 15+
- **Миграции:** golang-migrate / goose
- **ORM:** GORM / sqlx (на выбор разработчика)

### 2.2 Хранилище файлов
- **S3-совместимое хранилище:**
  - AWS S3 / Yandex Object Storage / MinIO
- **Библиотека:** aws-sdk-go-v2

### 2.3 Frontend (веб)
- **Фреймворк:** React 18+ / Next.js 14+
- **UI-библиотека:** Tailwind CSS / Material UI
- **Запросы:** Axios / React Query
- **Роутинг:** React Router / Next.js App Router

### 2.4 Telegram Bot
- **Библиотека:** telebot / go-telegram-bot-api
- **Интеграция:** Telegram Web App для магазина

### 2.5 DevOps
- **Контейнеризация:** Docker + Docker Compose
- **CI/CD:** GitHub Actions / GitLab CI
- **Мониторинг:** Prometheus + Grafana (опционально)

---

## 3. АРХИТЕКТУРА СИСТЕМЫ

### 3.1 Общая архитектура

```
┌─────────────────┐         ┌──────────────────┐
│   Web Frontend  │────────▶│                  │
│   (React/Next)  │         │                  │
└─────────────────┘         │                  │
                            │   Go Backend     │
┌─────────────────┐         │   (REST API)     │
│  Telegram Bot   │────────▶│                  │
│   (Web App)     │         │                  │
└─────────────────┘         └────────┬─────────┘
                                     │
                            ┌────────▼─────────┐
                            │   PostgreSQL     │
                            └──────────────────┘
                                     │
                            ┌────────▼─────────┐
                            │  S3 Storage      │
                            │  (Images/Files)  │
                            └──────────────────┘
```

### 3.2 Структура Backend (Clean Architecture)

```
project/
├── cmd/
│   └── api/
│       └── main.go              # Точка входа
├── internal/
│   ├── config/                  # Конфигурация
│   ├── domain/                  # Бизнес-логика и модели
│   │   ├── entity/              # Сущности
│   │   ├── repository/          # Интерфейсы репозиториев
│   │   └── service/             # Интерфейсы сервисов
│   ├── repository/              # Реализация репозиториев
│   │   └── postgres/
│   ├── service/                 # Реализация бизнес-логики
│   ├── handler/                 # HTTP handlers
│   │   ├── rest/                # REST API handlers
│   │   └── telegram/            # Telegram bot handlers
│   ├── middleware/              # Middleware (auth, logging)
│   ├── storage/                 # S3 storage logic
│   └── dto/                     # Data Transfer Objects
├── pkg/                         # Общие утилиты
├── migrations/                  # SQL миграции
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── go.sum
```

---

## 4. БАЗЫ ДАННЫХ

### 4.1 Схема PostgreSQL

```sql
-- Таблица категорий
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    parent_id INTEGER REFERENCES categories(id),
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица товаров (3D-модели)
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    short_description VARCHAR(500),
    price DECIMAL(10, 2) NOT NULL,
    old_price DECIMAL(10, 2),
    stock_quantity INTEGER DEFAULT 0,
    sku VARCHAR(100) UNIQUE,
    weight DECIMAL(10, 2),  -- вес в граммах
    dimensions JSONB,  -- {"length": 10, "width": 5, "height": 3}
    material VARCHAR(100),  -- PLA, ABS, PETG и т.д.
    print_time INTEGER,  -- время печати в минутах
    category_id INTEGER REFERENCES categories(id),
    is_active BOOLEAN DEFAULT true,
    is_featured BOOLEAN DEFAULT false,
    views_count INTEGER DEFAULT 0,
    sales_count INTEGER DEFAULT 0,
    rating DECIMAL(3, 2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для products
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_slug ON products(slug);
CREATE INDEX idx_products_active ON products(is_active);

-- Таблица изображений товаров
CREATE TABLE product_images (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    image_url VARCHAR(500) NOT NULL,
    s3_key VARCHAR(500) NOT NULL,
    display_order INTEGER DEFAULT 0,
    is_main BOOLEAN DEFAULT false,
    alt_text VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_product_images_product ON product_images(product_id);

-- Таблица тегов
CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL
);

-- Связь товаров и тегов
CREATE TABLE product_tags (
    product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
    tag_id INTEGER REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, tag_id)
);

-- Таблица пользователей
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE,  -- ID пользователя в Telegram
    phone VARCHAR(20),
    email VARCHAR(255),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    username VARCHAR(100),  -- username в Telegram
    role VARCHAR(20) DEFAULT 'customer',  -- customer, admin, manager
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_telegram ON users(telegram_id);
CREATE INDEX idx_users_email ON users(email);

-- Таблица адресов доставки
CREATE TABLE addresses (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    full_name VARCHAR(200),
    phone VARCHAR(20),
    city VARCHAR(100),
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    postal_code VARCHAR(20),
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица заказов
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    order_number VARCHAR(50) UNIQUE NOT NULL,
    user_id INTEGER REFERENCES users(id),
    
    -- Контактные данные (копия на момент заказа)
    customer_name VARCHAR(200),
    customer_phone VARCHAR(20),
    customer_email VARCHAR(255),
    
    -- Адрес доставки
    delivery_address TEXT,
    delivery_city VARCHAR(100),
    delivery_postal_code VARCHAR(20),
    
    -- Финансы
    subtotal DECIMAL(10, 2) NOT NULL,
    delivery_cost DECIMAL(10, 2) DEFAULT 0,
    discount DECIMAL(10, 2) DEFAULT 0,
    total_amount DECIMAL(10, 2) NOT NULL,
    
    -- Статусы
    status VARCHAR(50) DEFAULT 'new',  -- new, confirmed, processing, shipped, delivered, cancelled
    payment_status VARCHAR(50) DEFAULT 'pending',  -- pending, paid, failed, refunded
    payment_method VARCHAR(50),  -- card, cash, transfer
    
    -- Доставка
    delivery_method VARCHAR(50),  -- courier, pickup, post, cdek
    tracking_number VARCHAR(100),
    
    -- Дополнительно
    comment TEXT,
    source VARCHAR(20) DEFAULT 'web',  -- web, telegram
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_number ON orders(order_number);

-- Таблица товаров в заказе
CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER REFERENCES products(id),
    product_name VARCHAR(255) NOT NULL,  -- копия на момент заказа
    product_sku VARCHAR(100),
    quantity INTEGER NOT NULL,
    price DECIMAL(10, 2) NOT NULL,  -- цена на момент заказа
    subtotal DECIMAL(10, 2) NOT NULL
);

CREATE INDEX idx_order_items_order ON order_items(order_id);

-- Таблица корзины
CREATE TABLE cart_items (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    telegram_id BIGINT,  -- для неавторизованных пользователей из Telegram
    product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, product_id),
    UNIQUE(telegram_id, product_id)
);

-- Таблица отзывов
CREATE TABLE reviews (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id),
    order_id INTEGER REFERENCES orders(id),
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    is_approved BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_reviews_product ON reviews(product_id);

-- Таблица промокодов
CREATE TABLE promo_codes (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    discount_type VARCHAR(20),  -- percentage, fixed
    discount_value DECIMAL(10, 2) NOT NULL,
    min_order_amount DECIMAL(10, 2),
    max_uses INTEGER,
    used_count INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    valid_from TIMESTAMP,
    valid_until TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица настроек
CREATE TABLE settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT,
    description TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- МОДУЛЬ ДОСТАВКИ
-- =====================================================

-- Таблица служб доставки
CREATE TABLE delivery_services (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,  -- cdek, boxberry, russian_post, pickup
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    api_url VARCHAR(255),
    api_key_encrypted TEXT,
    settings JSONB,  -- дополнительные настройки провайдера
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица тарифов доставки
CREATE TABLE delivery_tariffs (
    id SERIAL PRIMARY KEY,
    service_id INTEGER REFERENCES delivery_services(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50),  -- код тарифа у провайдера
    delivery_type VARCHAR(50),  -- door_to_door, door_to_pickup, pickup_to_door, pickup_to_pickup
    min_weight DECIMAL(10, 2) DEFAULT 0,
    max_weight DECIMAL(10, 2),
    min_days INTEGER,
    max_days INTEGER,
    base_cost DECIMAL(10, 2),
    is_active BOOLEAN DEFAULT true
);

-- Таблица зон доставки
CREATE TABLE delivery_zones (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    regions TEXT[],  -- список регионов/городов
    delivery_cost DECIMAL(10, 2),
    free_delivery_from DECIMAL(10, 2),
    min_days INTEGER,
    max_days INTEGER,
    is_active BOOLEAN DEFAULT true
);

-- Таблица пунктов выдачи
CREATE TABLE pickup_points (
    id SERIAL PRIMARY KEY,
    service_id INTEGER REFERENCES delivery_services(id),
    external_id VARCHAR(100),  -- ID у провайдера
    code VARCHAR(100),
    name VARCHAR(255),
    city VARCHAR(100),
    address TEXT,
    work_hours VARCHAR(255),
    phone VARCHAR(50),
    latitude DECIMAL(10, 7),
    longitude DECIMAL(10, 7),
    is_active BOOLEAN DEFAULT true,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_pickup_points_city ON pickup_points(city);
CREATE INDEX idx_pickup_points_coords ON pickup_points(latitude, longitude);

-- Таблица отслеживания доставки
CREATE TABLE delivery_tracking (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
    tracking_number VARCHAR(100),
    service_code VARCHAR(50),
    status VARCHAR(100),
    status_description TEXT,
    location VARCHAR(255),
    event_time TIMESTAMP,
    raw_data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_delivery_tracking_order ON delivery_tracking(order_id);

-- =====================================================
-- МОДУЛЬ ИМПОРТА ТОВАРОВ
-- =====================================================

-- Таблица импортов
CREATE TABLE product_imports (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    filename VARCHAR(255) NOT NULL,
    file_path VARCHAR(500),
    file_type VARCHAR(20),  -- csv, xlsx
    status VARCHAR(50) DEFAULT 'pending',  -- pending, processing, completed, failed
    total_rows INTEGER DEFAULT 0,
    processed_rows INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    field_mapping JSONB,  -- маппинг колонок файла на поля БД
    options JSONB,  -- опции импорта (обновлять существующие, пропускать ошибки и т.д.)
    error_log TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица ошибок импорта
CREATE TABLE import_errors (
    id SERIAL PRIMARY KEY,
    import_id INTEGER REFERENCES product_imports(id) ON DELETE CASCADE,
    row_number INTEGER,
    field_name VARCHAR(100),
    value TEXT,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_import_errors_import ON import_errors(import_id);

-- Шаблоны импорта
CREATE TABLE import_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    field_mapping JSONB NOT NULL,
    options JSONB,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- МОДУЛЬ АНАЛИТИКИ И ЭКСПОРТА
-- =====================================================

-- Таблица экспортов
CREATE TABLE data_exports (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    export_type VARCHAR(50) NOT NULL,  -- sales, products, orders, customers
    format VARCHAR(20) NOT NULL,  -- csv, xlsx, pdf
    status VARCHAR(50) DEFAULT 'pending',  -- pending, processing, completed, failed
    filters JSONB,  -- применённые фильтры
    file_path VARCHAR(500),
    file_url VARCHAR(500),
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- Агрегированная статистика продаж (для быстрых отчётов)
CREATE TABLE sales_daily_stats (
    id SERIAL PRIMARY KEY,
    date DATE UNIQUE NOT NULL,
    orders_count INTEGER DEFAULT 0,
    items_sold INTEGER DEFAULT 0,
    revenue DECIMAL(12, 2) DEFAULT 0,
    average_order_value DECIMAL(10, 2) DEFAULT 0,
    new_customers INTEGER DEFAULT 0,
    returning_customers INTEGER DEFAULT 0,
    cancelled_orders INTEGER DEFAULT 0,
    refunded_amount DECIMAL(12, 2) DEFAULT 0,
    top_products JSONB,  -- [{product_id, name, quantity, revenue}]
    top_categories JSONB,
    traffic_sources JSONB,  -- {web: 10, telegram: 5}
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sales_stats_date ON sales_daily_stats(date);

-- Статистика по товарам
CREATE TABLE product_stats (
    id SERIAL PRIMARY KEY,
    product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    views INTEGER DEFAULT 0,
    cart_adds INTEGER DEFAULT 0,
    purchases INTEGER DEFAULT 0,
    revenue DECIMAL(10, 2) DEFAULT 0,
    UNIQUE(product_id, date)
);

CREATE INDEX idx_product_stats_product_date ON product_stats(product_id, date);

-- =====================================================
-- РЕФЕРАЛЬНАЯ ПРОГРАММА
-- =====================================================

-- Реферальные настройки (расширение таблицы settings или отдельная)
-- Используем JSONB в settings для гибкости

-- Таблица рефералов
CREATE TABLE referrals (
    id SERIAL PRIMARY KEY,
    referrer_id INTEGER NOT NULL REFERENCES users(id),  -- кто пригласил
    referred_id INTEGER NOT NULL REFERENCES users(id),  -- кого пригласили
    referral_code VARCHAR(50) NOT NULL,
    status VARCHAR(50) DEFAULT 'registered',  -- registered, first_order, rewarded
    reward_amount DECIMAL(10, 2),
    reward_paid BOOLEAN DEFAULT false,
    first_order_id INTEGER REFERENCES orders(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    rewarded_at TIMESTAMP,
    UNIQUE(referred_id)  -- один пользователь может быть приглашён только один раз
);

CREATE INDEX idx_referrals_referrer ON referrals(referrer_id);
CREATE INDEX idx_referrals_code ON referrals(referral_code);

-- Добавляем реферальный код в таблицу users
ALTER TABLE users ADD COLUMN IF NOT EXISTS referral_code VARCHAR(50) UNIQUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS referral_balance DECIMAL(10, 2) DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS referred_by INTEGER REFERENCES users(id);

-- Бонусный баланс пользователей
CREATE TABLE bonus_transactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    amount DECIMAL(10, 2) NOT NULL,  -- положительное - начисление, отрицательное - списание
    type VARCHAR(50) NOT NULL,  -- referral_reward, referral_bonus, order_cashback, promo, manual, order_payment
    description TEXT,
    order_id INTEGER REFERENCES orders(id),
    referral_id INTEGER REFERENCES referrals(id),
    balance_after DECIMAL(10, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bonus_transactions_user ON bonus_transactions(user_id);

-- Расширение таблицы промокодов для реферальной программы
ALTER TABLE promo_codes ADD COLUMN IF NOT EXISTS owner_id INTEGER REFERENCES users(id);
ALTER TABLE promo_codes ADD COLUMN IF NOT EXISTS is_referral BOOLEAN DEFAULT false;
ALTER TABLE promo_codes ADD COLUMN IF NOT EXISTS cashback_percent DECIMAL(5, 2);

-- История использования промокодов
CREATE TABLE promo_code_usages (
    id SERIAL PRIMARY KEY,
    promo_code_id INTEGER REFERENCES promo_codes(id),
    user_id INTEGER REFERENCES users(id),
    order_id INTEGER REFERENCES orders(id),
    discount_applied DECIMAL(10, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_promo_usages_code ON promo_code_usages(promo_code_id);
CREATE INDEX idx_promo_usages_user ON promo_code_usages(user_id);

-- Вставка базовых настроек
INSERT INTO settings (key, value, description) VALUES
    ('site_name', '3D Print Store', 'Название магазина'),
    ('site_description', 'Магазин 3D-печатных изделий', 'Описание магазина'),
    ('delivery_cost', '300', 'Стоимость доставки по умолчанию'),
    ('free_delivery_from', '3000', 'Бесплатная доставка от суммы'),
    ('admin_email', 'admin@example.com', 'Email администратора'),
    ('telegram_bot_token', '', 'Токен Telegram бота');

-- Таблица для логов (опционально)
CREATE TABLE activity_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(100),
    entity_type VARCHAR(50),
    entity_id INTEGER,
    details JSONB,
    ip_address VARCHAR(45),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 4.2 Индексы для оптимизации

```sql
-- Полнотекстовый поиск по товарам
CREATE INDEX idx_products_search ON products 
    USING gin(to_tsvector('russian', name || ' ' || COALESCE(description, '')));

-- Индексы для сортировки и фильтрации
CREATE INDEX idx_products_price ON products(price);
CREATE INDEX idx_products_created ON products(created_at DESC);
CREATE INDEX idx_products_rating ON products(rating DESC);
CREATE INDEX idx_products_featured ON products(is_featured) WHERE is_featured = true;
```

---

## 5. REST API СПЕЦИФИКАЦИЯ

### 5.1 Базовый URL
```
https://api.example.com/v1
```

### 5.2 Аутентификация
- **JWT токены** для веб-приложения
- **Telegram ID** для Telegram-бота

### 5.3 Endpoints

#### 5.3.1 Категории

```
GET    /categories              # Список всех категорий
GET    /categories/{slug}       # Категория по slug
POST   /admin/categories        # Создать категорию (admin)
PUT    /admin/categories/{id}   # Обновить категорию (admin)
DELETE /admin/categories/{id}   # Удалить категорию (admin)
```

#### 5.3.2 Товары

```
GET    /products                # Список товаров (с пагинацией, фильтрами)
GET    /products/{slug}         # Товар по slug
GET    /products/{id}           # Товар по ID
POST   /admin/products          # Создать товар (admin)
PUT    /admin/products/{id}     # Обновить товар (admin)
DELETE /admin/products/{id}     # Удалить товар (admin)
POST   /admin/products/{id}/images  # Загрузить изображения (admin)
DELETE /admin/products/images/{id}  # Удалить изображение (admin)
```

**GET /products - параметры запроса:**
```
?page=1
&limit=20
&category={slug}
&min_price=100
&max_price=5000
&sort=price_asc|price_desc|rating|newest|popular
&search=текст поиска
&tags=tag1,tag2
```

**Пример ответа GET /products/{slug}:**
```json
{
  "id": 1,
  "name": "Фигурка Дракона",
  "slug": "figurka-drakona",
  "description": "Детализированная 3D-модель дракона...",
  "short_description": "Красивая фигурка дракона",
  "price": 1200.00,
  "old_price": 1500.00,
  "stock_quantity": 15,
  "sku": "DRAG-001",
  "weight": 150.5,
  "dimensions": {
    "length": 15,
    "width": 10,
    "height": 20
  },
  "material": "PLA",
  "print_time": 420,
  "category": {
    "id": 2,
    "name": "Фигурки",
    "slug": "figurki"
  },
  "images": [
    {
      "id": 1,
      "url": "https://s3.example.com/products/dragon-1.jpg",
      "is_main": true,
      "alt_text": "Дракон вид спереди"
    }
  ],
  "tags": ["фэнтези", "дракон", "коллекция"],
  "rating": 4.8,
  "reviews_count": 23,
  "is_featured": true,
  "created_at": "2026-01-15T10:30:00Z"
}
```

#### 5.3.3 Корзина

```
GET    /cart                    # Получить корзину
POST   /cart/items              # Добавить товар в корзину
PUT    /cart/items/{id}         # Обновить количество
DELETE /cart/items/{id}         # Удалить товар из корзины
DELETE /cart                    # Очистить корзину
```

**POST /cart/items:**
```json
{
  "product_id": 1,
  "quantity": 2
}
```

#### 5.3.4 Заказы

```
GET    /orders                  # История заказов пользователя
GET    /orders/{order_number}   # Детали заказа
POST   /orders                  # Создать заказ
POST   /orders/{id}/cancel      # Отменить заказ

GET    /admin/orders            # Список всех заказов (admin)
PUT    /admin/orders/{id}       # Обновить статус заказа (admin)
```

**POST /orders - создание заказа:**
```json
{
  "customer_name": "Иван Петров",
  "customer_phone": "+79991234567",
  "customer_email": "ivan@example.com",
  "delivery_address": "ул. Ленина, д. 10, кв. 5",
  "delivery_city": "Москва",
  "delivery_postal_code": "101000",
  "delivery_method": "cdek",
  "payment_method": "card",
  "promo_code": "SALE10",
  "comment": "Позвоните за час до доставки",
  "items": [
    {
      "product_id": 1,
      "quantity": 2
    }
  ]
}
```

**Ответ:**
```json
{
  "order_number": "ORD-20260204-0001",
  "total_amount": 2700.00,
  "status": "new",
  "created_at": "2026-02-04T14:30:00Z"
}
```

#### 5.3.5 Пользователи

```
POST   /auth/telegram           # Авторизация через Telegram
POST   /auth/register           # Регистрация
POST   /auth/login              # Вход
GET    /users/me                # Профиль текущего пользователя
PUT    /users/me                # Обновить профиль
```

#### 5.3.6 Отзывы

```
GET    /products/{id}/reviews   # Отзывы о товаре
POST   /reviews                 # Добавить отзыв
GET    /admin/reviews           # Все отзывы (admin)
PUT    /admin/reviews/{id}/approve  # Одобрить отзыв (admin)
```

#### 5.3.7 Поиск

```
GET    /search?q=текст          # Поиск по товарам
GET    /search/suggestions?q=текст  # Подсказки для поиска
```

#### 5.3.8 Загрузка файлов

```
POST   /upload/image            # Загрузка изображения в S3
```

#### 5.3.9 Модуль доставки

```
GET    /delivery/calculate      # Расчёт стоимости доставки
GET    /delivery/services       # Список доступных служб доставки
GET    /delivery/pickup-points  # Пункты выдачи (с фильтром по городу)
GET    /delivery/tracking/{order_id}  # Отслеживание доставки

GET    /admin/delivery/services         # Список служб доставки (admin)
POST   /admin/delivery/services         # Добавить службу доставки (admin)
PUT    /admin/delivery/services/{id}    # Обновить службу (admin)
DELETE /admin/delivery/services/{id}    # Удалить службу (admin)

GET    /admin/delivery/zones            # Зоны доставки (admin)
POST   /admin/delivery/zones            # Создать зону (admin)
PUT    /admin/delivery/zones/{id}       # Обновить зону (admin)
DELETE /admin/delivery/zones/{id}       # Удалить зону (admin)

POST   /admin/delivery/sync-pickup-points  # Синхронизировать ПВЗ с провайдерами
```

**GET /delivery/calculate - расчёт доставки:**
```json
// Запрос
{
  "city": "Москва",
  "postal_code": "101000",
  "weight": 500,
  "dimensions": {"length": 20, "width": 15, "height": 10},
  "items": [{"product_id": 1, "quantity": 2}]
}

// Ответ
{
  "options": [
    {
      "service": "cdek",
      "service_name": "СДЭК",
      "tariff": "door_to_door",
      "tariff_name": "Курьер до двери",
      "cost": 350.00,
      "min_days": 2,
      "max_days": 4
    },
    {
      "service": "cdek",
      "service_name": "СДЭК",
      "tariff": "door_to_pickup",
      "tariff_name": "До пункта выдачи",
      "cost": 250.00,
      "min_days": 2,
      "max_days": 5,
      "pickup_points_count": 45
    },
    {
      "service": "russian_post",
      "service_name": "Почта России",
      "tariff": "parcel",
      "tariff_name": "Посылка",
      "cost": 180.00,
      "min_days": 5,
      "max_days": 14
    }
  ],
  "free_delivery_available": true,
  "free_delivery_from": 3000.00,
  "current_cart_total": 2400.00
}
```

#### 5.3.10 Импорт товаров

```
POST   /admin/import/products          # Загрузить файл для импорта
GET    /admin/import/products/preview  # Предпросмотр данных из файла
POST   /admin/import/products/start    # Запустить импорт
GET    /admin/import/products/{id}     # Статус импорта
GET    /admin/import/products/{id}/errors  # Ошибки импорта
GET    /admin/import/history           # История импортов
DELETE /admin/import/products/{id}     # Отменить/удалить импорт

GET    /admin/import/templates         # Шаблоны импорта
POST   /admin/import/templates         # Создать шаблон
GET    /admin/import/sample            # Скачать образец файла
```

**POST /admin/import/products - загрузка файла:**
```
Content-Type: multipart/form-data
file: products.csv
```

**POST /admin/import/products/start - запуск импорта:**
```json
{
  "file_id": "temp-upload-123",
  "mapping": {
    "name": "Название",
    "price": "Цена",
    "sku": "Артикул",
    "category": "Категория",
    "description": "Описание",
    "stock_quantity": "Остаток",
    "weight": "Вес",
    "material": "Материал"
  },
  "options": {
    "update_existing": true,
    "skip_errors": false,
    "match_by": "sku"
  }
}
```

#### 5.3.11 Аналитика и экспорт

```
GET    /admin/analytics/dashboard      # Сводная аналитика
GET    /admin/analytics/sales          # Статистика продаж
GET    /admin/analytics/products       # Аналитика по товарам
GET    /admin/analytics/customers      # Аналитика по клиентам
GET    /admin/analytics/traffic        # Источники трафика

POST   /admin/export/sales             # Экспорт продаж
POST   /admin/export/products          # Экспорт товаров
POST   /admin/export/orders            # Экспорт заказов
POST   /admin/export/customers         # Экспорт клиентов
GET    /admin/export/{id}              # Статус/скачивание экспорта
GET    /admin/export/history           # История экспортов
```

**GET /admin/analytics/dashboard:**
```json
{
  "period": "month",
  "summary": {
    "revenue": 156000.00,
    "revenue_change": 12.5,
    "orders_count": 89,
    "orders_change": 8.2,
    "average_order": 1752.81,
    "items_sold": 234,
    "new_customers": 45,
    "returning_rate": 23.5
  },
  "charts": {
    "revenue_by_day": [...],
    "orders_by_day": [...]
  },
  "top_products": [
    {"id": 1, "name": "Фигурка Дракона", "sales": 23, "revenue": 27600.00}
  ],
  "top_categories": [...],
  "recent_orders": [...],
  "low_stock_alerts": [
    {"id": 5, "name": "Брелок Кот", "stock": 3}
  ]
}
```

**POST /admin/export/sales:**
```json
{
  "format": "xlsx",
  "date_from": "2026-01-01",
  "date_to": "2026-01-31",
  "include_items": true,
  "group_by": "day",
  "columns": ["date", "orders_count", "revenue", "items_sold", "avg_order"]
}
```

#### 5.3.12 Реферальная программа

```
GET    /referral/info                  # Информация о реф. программе
GET    /referral/my                    # Моя реф. статистика и код
GET    /referral/my/history            # История начислений
POST   /referral/apply                 # Применить реферальный код при регистрации

GET    /admin/referral/stats           # Статистика реф. программы
GET    /admin/referral/list            # Список рефералов
PUT    /admin/referral/settings        # Настройки программы
POST   /admin/referral/reward/{id}     # Начислить награду вручную
```

**GET /referral/my:**
```json
{
  "referral_code": "VOVA2026",
  "referral_link": "https://shop.example.com/?ref=VOVA2026",
  "balance": 1500.00,
  "total_earned": 4500.00,
  "referrals_count": 12,
  "active_referrals": 8,
  "pending_rewards": 500.00,
  "program_rules": {
    "reward_for_referrer": 500,
    "reward_type": "fixed",
    "bonus_for_referred": 10,
    "bonus_type": "percentage",
    "min_order_amount": 1000
  }
}
```

#### 5.3.13 Промокоды (расширенный)

```
POST   /promo/validate                 # Проверить промокод
POST   /promo/apply                    # Применить к корзине

GET    /admin/promo                    # Список промокодов
POST   /admin/promo                    # Создать промокод
PUT    /admin/promo/{id}               # Обновить промокод
DELETE /admin/promo/{id}               # Удалить промокод
GET    /admin/promo/{id}/stats         # Статистика использования
POST   /admin/promo/generate-bulk      # Массовая генерация промокодов
```

**POST /admin/promo - создание промокода:**
```json
{
  "code": "SUMMER2026",
  "discount_type": "percentage",
  "discount_value": 15,
  "min_order_amount": 2000,
  "max_discount_amount": 1000,
  "max_uses": 100,
  "max_uses_per_user": 1,
  "valid_from": "2026-06-01T00:00:00Z",
  "valid_until": "2026-08-31T23:59:59Z",
  "categories": [1, 2, 3],
  "products": null,
  "exclude_sale_items": true,
  "first_order_only": false,
  "is_active": true
}
```

#### 5.3.14 Бонусный баланс

```
GET    /bonus/balance                  # Мой бонусный баланс
GET    /bonus/history                  # История транзакций
POST   /bonus/apply                    # Применить бонусы к заказу

GET    /admin/bonus/transactions       # Все транзакции
POST   /admin/bonus/adjust             # Ручная корректировка баланса
```

### 5.4 Коды ответов

```
200 OK                  # Успешный запрос
201 Created             # Ресурс создан
204 No Content          # Успешное удаление
400 Bad Request         # Ошибка валидации
401 Unauthorized        # Не авторизован
403 Forbidden           # Нет прав доступа
404 Not Found           # Ресурс не найден
500 Internal Server Error  # Ошибка сервера
```

### 5.5 Формат ошибок

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Некорректные данные",
    "details": [
      {
        "field": "email",
        "message": "Неверный формат email"
      }
    ]
  }
}
```

---

## 6. TELEGRAM BOT

### 6.1 Функционал бота

#### Команды:
- `/start` - Приветствие и главное меню
- `/catalog` - Открыть каталог
- `/cart` - Корзина
- `/orders` - Мои заказы
- `/help` - Помощь

#### Inline-меню:
```
🏠 Главная
📦 Каталог
🛒 Корзина (3)
📋 Мои заказы
ℹ️ Помощь
```

### 6.2 Telegram Web App

**URL:** `https://yourdomain.com/telegram-shop`

**Интеграция:**
- Открывается через кнопку в боте
- Авторизация через `window.Telegram.WebApp.initDataUnsafe`
- Полноценный магазин в Web App
- Данные корзины синхронизируются с ботом

### 6.3 Уведомления

**Пользователю:**
- Подтверждение заказа
- Изменение статуса заказа
- Получение трек-номера
- Напоминание об оплате

**Администратору:**
- Новый заказ
- Новый отзыв
- Товар заканчивается (stock < 5)

### 6.4 Структура бота

```go
// internal/handler/telegram/bot.go
type TelegramBot struct {
    bot     *telebot.Bot
    service *service.Service
}

// Handlers
func (tb *TelegramBot) HandleStart(c telebot.Context)
func (tb *TelegramBot) HandleCatalog(c telebot.Context)
func (tb *TelegramBot) HandleCart(c telebot.Context)
func (tb *TelegramBot) HandleOrders(c telebot.Context)
func (tb *TelegramBot) HandleCallback(c telebot.Context)
```

---

## 7. WEB FRONTEND

### 7.1 Страницы

#### Публичные:
- `/` - Главная (популярные товары, акции)
- `/catalog` - Каталог с фильтрами
- `/catalog/{category}` - Категория
- `/product/{slug}` - Страница товара
- `/cart` - Корзина
- `/checkout` - Оформление заказа
- `/order/{order_number}` - Детали заказа
- `/about` - О магазине
- `/contacts` - Контакты

#### Личный кабинет:
- `/profile` - Профиль
- `/orders` - История заказов
- `/addresses` - Адреса доставки

### 7.2 Основные компоненты

```
components/
├── Layout/
│   ├── Header.tsx
│   ├── Footer.tsx
│   └── Navigation.tsx
├── Product/
│   ├── ProductCard.tsx
│   ├── ProductGrid.tsx
│   ├── ProductDetail.tsx
│   ├── ProductGallery.tsx
│   └── ProductFilter.tsx
├── Cart/
│   ├── CartItem.tsx
│   ├── CartSummary.tsx
│   └── CartDrawer.tsx
├── Order/
│   ├── CheckoutForm.tsx
│   ├── OrderSummary.tsx
│   └── OrderStatus.tsx
└── Common/
    ├── Button.tsx
    ├── Input.tsx
    ├── Modal.tsx
    └── Loader.tsx
```

### 7.3 Адаптивность
- Mobile First подход
- Breakpoints: 320px, 768px, 1024px, 1440px
- Адаптивная галерея изображений
- Мобильное меню

---

## 8. АДМИНИСТРАТИВНАЯ ПАНЕЛЬ

### 8.1 Доступ
- URL: `/admin` или отдельный поддомен `admin.example.com`
- Авторизация: JWT с ролью `admin` или `manager`

### 8.2 Разделы

#### 8.2.1 Dashboard (главная)
- Статистика продаж (сегодня, неделя, месяц)
- Количество заказов по статусам
- Популярные товары
- Последние заказы
- Товары с низким остатком

#### 8.2.2 Товары
**Список товаров:**
- Таблица с колонками: фото, название, цена, остаток, статус
- Фильтры: категория, статус, наличие
- Поиск по названию, SKU
- Сортировка
- Массовые действия (удалить, изменить статус)

**Создание/редактирование товара:**
```
Форма:
- Название *
- Slug (автогенерация из названия)
- Краткое описание
- Полное описание (WYSIWYG редактор)
- Цена *
- Старая цена
- SKU
- Количество на складе *
- Вес
- Размеры (Д x Ш x В)
- Материал (выбор из списка)
- Время печати
- Категория *
- Теги (мультивыбор)
- Статус (активен/неактивен)
- Рекомендуемый товар
- Изображения (drag-n-drop, до 10 фото)
  - Главное изображение
  - Альтернативный текст для каждого
```

#### 8.2.3 Категории
- Древовидная структура
- CRUD операции
- Изменение порядка отображения

#### 8.2.4 Заказы
**Список заказов:**
- Таблица: номер, клиент, сумма, статус, дата
- Фильтры по статусу, дате, способу оплаты
- Поиск по номеру, телефону, email

**Детали заказа:**
- Информация о клиенте
- Список товаров
- Адрес доставки
- Изменение статуса
- Добавление трек-номера
- История изменений статуса
- Кнопка "Распечатать накладную"

#### 8.2.5 Пользователи
- Список пользователей
- Просмотр профиля
- История заказов пользователя
- Блокировка/разблокировка

#### 8.2.6 Отзывы
- Список всех отзывов
- Модерация (одобрить/отклонить)
- Фильтр по товарам, рейтингу, статусу

#### 8.2.7 Промокоды
- Создание промокодов
- Просмотр статистики использования
- Деактивация

#### 8.2.8 Настройки
- Общие настройки магазина
- Настройки доставки
- Способы оплаты
- Email/Telegram уведомления
- Токен Telegram-бота
- Настройки S3

### 8.3 Технологии админки
- React Admin / Refine / AdminBro
- Или кастомное решение на React + Tailwind

---

## 9. МОДУЛЬ ДОСТАВКИ

### 9.1 Обзор функционала

Модуль обеспечивает полный цикл работы с доставкой:
- Расчёт стоимости и сроков доставки
- Интеграция с транспортными компаниями
- Отслеживание статуса
- Управление пунктами выдачи

### 9.2 Поддерживаемые службы доставки

| Служба | Код | API | Функционал |
|--------|-----|-----|------------|
| СДЭК | `cdek` | REST API v2 | Расчёт, создание заказа, отслеживание, ПВЗ |
| Boxberry | `boxberry` | REST API | Расчёт, отслеживание, ПВЗ |
| Почта России | `russian_post` | API отправки | Расчёт, создание отправления |
| Самовывоз | `pickup` | - | Собственные точки выдачи |
| Курьер (своя служба) | `local_courier` | - | Ручное управление |

### 9.3 Расчёт стоимости доставки

**Алгоритм расчёта:**
```
1. Получить параметры заказа (вес, габариты, город)
2. Для каждой активной службы доставки:
   a. Если есть API интеграция → запрос к API
   b. Если нет → расчёт по тарифным зонам
3. Применить бесплатную доставку (если сумма заказа >= порога)
4. Отсортировать по цене/срокам
5. Вернуть список вариантов
```

**Формула ручного расчёта:**
```
Стоимость = Базовая ставка зоны + (Вес × Тариф за кг) + Надбавка за габариты
```

### 9.4 Интеграция с СДЭК

```go
// internal/delivery/cdek/client.go
type CDEKClient struct {
    baseURL     string
    clientID    string
    clientSecret string
    httpClient  *http.Client
}

// Методы
func (c *CDEKClient) Authenticate() error
func (c *CDEKClient) CalculateDelivery(req CalculateRequest) ([]Tariff, error)
func (c *CDEKClient) CreateOrder(order Order) (*OrderResponse, error)
func (c *CDEKClient) GetOrderStatus(uuid string) (*StatusResponse, error)
func (c *CDEKClient) GetPickupPoints(cityCode int) ([]PickupPoint, error)
```

### 9.5 Отслеживание доставки

**Webhook от служб доставки:**
```
POST /webhook/delivery/cdek
POST /webhook/delivery/boxberry
```

**Автоматическое обновление:**
- Cron задача каждые 2 часа для активных заказов
- Обновление статуса в БД
- Уведомление клиента при изменении статуса

### 9.6 Карта пунктов выдачи

**Frontend компонент:**
- Интеграция с Яндекс.Картами / 2GIS
- Фильтрация по городу
- Поиск ближайших ПВЗ по геолокации
- Информация: адрес, часы работы, как добраться

**Синхронизация ПВЗ:**
- Ежедневное обновление из API провайдеров
- Кеширование в PostgreSQL
- Поиск по городу с индексом

---

## 10. МОДУЛЬ ИМПОРТА ТОВАРОВ

### 10.1 Обзор

Массовая загрузка товаров из файлов CSV/Excel с валидацией, предпросмотром и обработкой ошибок.

### 10.2 Поддерживаемые форматы

| Формат | Расширения | Библиотека |
|--------|------------|------------|
| CSV | .csv | encoding/csv |
| Excel | .xlsx, .xls | excelize |

### 10.3 Процесс импорта

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Загрузка   │───▶│  Preview    │───▶│  Маппинг   │───▶│   Импорт    │
│   файла     │    │  (первые    │    │   полей    │    │  (фоновая   │
│             │    │   10 строк) │    │            │    │   задача)   │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

### 10.4 Маппинг полей

**Обязательные поля:**
- `name` - Название товара
- `price` - Цена
- `category` - Категория (название или ID)

**Опциональные поля:**
- `sku` - Артикул (для обновления существующих)
- `description` - Описание
- `short_description` - Краткое описание
- `old_price` - Старая цена
- `stock_quantity` - Остаток
- `weight` - Вес (граммы)
- `material` - Материал
- `tags` - Теги (через запятую)
- `images` - URL изображений (через запятую)
- `is_active` - Активен (1/0, true/false)

### 10.5 Валидация

```go
type ImportValidator struct {
    rules map[string][]ValidationRule
}

// Правила валидации
- name: required, min:2, max:255
- price: required, numeric, min:0
- sku: unique (если update_existing = false)
- category: exists_in_db
- stock_quantity: integer, min:0
- weight: numeric, min:0
- images: valid_url
```

### 10.6 Обработка ошибок

**Режимы:**
1. **Строгий** - остановка при первой ошибке
2. **Пропуск ошибок** - импорт валидных строк, логирование ошибок

**Лог ошибок:**
```json
{
  "row": 15,
  "field": "price",
  "value": "abc",
  "error": "Значение должно быть числом"
}
```

### 10.7 Интерфейс в админке

**Страница импорта:**
1. Drag-n-drop зона для файла
2. Выбор шаблона маппинга (если есть сохранённые)
3. Таблица предпросмотра с маппингом колонок
4. Настройки: обновлять существующие, пропускать ошибки
5. Кнопка "Начать импорт"
6. Прогресс-бар и статистика
7. Скачать отчёт об ошибках

### 10.8 Образец файла

```csv
name;price;sku;category;stock_quantity;weight;material;description
"Фигурка Дракона";1200;DRAG-001;Фигурки;15;150;PLA;"Детализированная модель дракона"
"Брелок Кот";350;KEY-CAT-01;Брелоки;50;20;PETG;"Милый брелок в виде кота"
```

---

## 11. МОДУЛЬ АНАЛИТИКИ И ЭКСПОРТА

### 11.1 Dashboard аналитики

**Виджеты:**
- 📈 Выручка (день/неделя/месяц/год) с графиком
- 📦 Количество заказов с разбивкой по статусам
- 🛒 Средний чек
- 👥 Новые vs возвращающиеся клиенты
- 🏆 Топ-10 товаров
- 📊 Продажи по категориям (pie chart)
- ⚠️ Товары с низким остатком
- 🔔 Последние заказы

### 11.2 Отчёты

| Отчёт | Описание | Фильтры |
|-------|----------|---------|
| Продажи | Выручка, заказы, товары по периодам | Период, категория, источник |
| Товары | Продажи, просмотры, конверсия | Период, категория |
| Клиенты | LTV, частота покупок, сегменты | Период, источник регистрации |
| Заказы | Детализация всех заказов | Период, статус, способ оплаты |
| Доставка | Статистика по службам доставки | Период, служба |

### 11.3 Расчёт метрик

```go
// Ключевые метрики
type AnalyticsMetrics struct {
    Revenue           decimal.Decimal  // Сумма всех оплаченных заказов
    OrdersCount       int              // Количество заказов
    AverageOrderValue decimal.Decimal  // Выручка / Кол-во заказов
    ItemsSold         int              // Сумма quantity из order_items
    ConversionRate    float64          // Заказы / Уникальные посетители
    CustomerLTV       decimal.Decimal  // Средняя выручка с клиента
    RepeatPurchaseRate float64         // % клиентов с >1 заказа
}
```

### 11.4 Экспорт данных

**Форматы:**
- **CSV** - для импорта в другие системы
- **XLSX** - для работы в Excel с форматированием
- **PDF** - для печати отчётов (опционально)

**Процесс экспорта:**
```
1. Пользователь выбирает тип, период, фильтры, формат
2. Создаётся задача экспорта (status: pending)
3. Background worker формирует файл
4. Файл загружается в S3 (temp папка)
5. Пользователь получает ссылку на скачивание
6. Файл удаляется через 24 часа
```

### 11.5 Агрегация данных

**Ежедневный cron (00:05):**
```sql
INSERT INTO sales_daily_stats (date, orders_count, revenue, ...)
SELECT
    DATE(created_at) as date,
    COUNT(*) as orders_count,
    SUM(total_amount) as revenue,
    ...
FROM orders
WHERE DATE(created_at) = CURRENT_DATE - 1
  AND status NOT IN ('cancelled')
GROUP BY DATE(created_at);
```

---

## 12. РЕФЕРАЛЬНАЯ ПРОГРАММА И ПРОМОКОДЫ

### 12.1 Обзор реферальной программы

Программа мотивирует клиентов приглашать друзей, награждая обе стороны.

### 12.2 Механика программы

```
┌─────────────────┐                    ┌─────────────────┐
│   Реферер       │                    │   Приглашённый  │
│   (Вова)        │                    │   (Друг)        │
└────────┬────────┘                    └────────┬────────┘
         │                                      │
         │  1. Делится ссылкой/кодом           │
         │────────────────────────────────────▶│
         │                                      │
         │  2. Друг регистрируется             │
         │◀────────────────────────────────────│
         │                                      │
         │  3. Друг делает заказ ≥ мин. суммы  │
         │◀────────────────────────────────────│
         │                                      │
    ┌────▼────┐                          ┌─────▼─────┐
    │  +500₽  │                          │  Скидка   │
    │ бонусов │                          │   10%     │
    └─────────┘                          └───────────┘
```

### 12.3 Настройки программы

```json
{
  "referral_program": {
    "enabled": true,
    "reward_for_referrer": {
      "type": "fixed",         // fixed | percentage
      "value": 500,            // 500₽ или 5%
      "trigger": "first_order" // registration | first_order | order_delivered
    },
    "bonus_for_referred": {
      "type": "percentage",
      "value": 10,
      "apply_to": "first_order"
    },
    "min_order_amount": 1000,
    "max_referrals_per_user": null,
    "code_format": "USERNAME_RANDOM4"  // VOVA_A1B2
  }
}
```

### 12.4 Генерация реферального кода

```go
func GenerateReferralCode(user *User) string {
    // Варианты:
    // 1. Username + случайные символы: VOVA_A1B2
    // 2. Только случайные: REF-8X4K2M
    // 3. На основе ID: REF-{base36(userId)}

    if user.Username != "" {
        return strings.ToUpper(user.Username) + "_" + randomString(4)
    }
    return "REF-" + randomString(6)
}
```

### 12.5 Бонусный баланс

**Начисление:**
- Реферальная награда
- Кешбэк с заказа (если настроен)
- Ручное начисление администратором
- Компенсация

**Списание:**
- Оплата части заказа бонусами
- Истечение срока (если настроено)

**Ограничения:**
- Максимум X% от суммы заказа можно оплатить бонусами
- Минимальная сумма заказа для использования бонусов
- Бонусы не применяются к товарам со скидкой (опционально)

### 12.6 Система промокодов (расширенная)

**Типы промокодов:**

| Тип | Описание | Пример |
|-----|----------|--------|
| `percentage` | Процент от суммы | 15% скидка |
| `fixed` | Фиксированная сумма | 500₽ скидка |
| `free_delivery` | Бесплатная доставка | - |
| `gift` | Подарок к заказу | + товар ID:123 |

**Условия применения:**
- Минимальная сумма заказа
- Максимальная сумма скидки
- Ограничение по категориям/товарам
- Только для первого заказа
- Ограничение по количеству использований
- Период действия
- Исключить товары со скидкой

### 12.7 Интерфейс пользователя

**Страница "Моя реферальная программа" (/profile/referral):**
```
┌─────────────────────────────────────────────────────────┐
│  🎁 Пригласи друга — получи 500₽                        │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Твой код: VOVA_A1B2     [📋 Копировать]               │
│                                                         │
│  Ссылка: https://shop.com/?ref=VOVA_A1B2  [📋] [📤]    │
│                                                         │
├─────────────────────────────────────────────────────────┤
│  💰 Баланс бонусов: 1 500 ₽        [Как использовать?] │
├─────────────────────────────────────────────────────────┤
│  📊 Статистика                                          │
│  ├─ Приглашено друзей: 12                              │
│  ├─ Сделали заказ: 8                                   │
│  └─ Всего заработано: 4 000 ₽                          │
├─────────────────────────────────────────────────────────┤
│  📜 История начислений                                  │
│  ┌────────────┬───────────────┬──────────┐             │
│  │ 03.02.2026 │ Реферал Иван  │ +500 ₽   │             │
│  │ 28.01.2026 │ Реферал Мария │ +500 ₽   │             │
│  └────────────┴───────────────┴──────────┘             │
└─────────────────────────────────────────────────────────┘
```

### 12.8 Админка реферальной программы

**Разделы:**
1. **Настройки** - параметры программы
2. **Статистика** - общая эффективность
3. **Рефералы** - список всех связей
4. **Транзакции** - история начислений/списаний
5. **Ручные операции** - корректировка балансов

---

## 13. ФАЙЛОВОЕ ХРАНИЛИЩЕ (S3)

### 9.1 Структура bucket'а

```
bucket-name/
├── products/
│   ├── {product_id}/
│   │   ├── main.jpg
│   │   ├── gallery-1.jpg
│   │   ├── gallery-2.jpg
│   │   └── ...
├── categories/
│   └── {category_id}.jpg
└── temp/
    └── {uuid}.jpg  # Временные файлы для удаления
```

### 9.2 Обработка изображений

**При загрузке:**
1. Валидация (формат, размер до 10MB)
2. Генерация уникального имени
3. Создание нескольких версий:
   - Original (до 2000px)
   - Large (1200px)
   - Medium (800px)
   - Thumbnail (300px)
4. Загрузка в S3
5. Сохранение URL в БД

**Библиотеки:**
- `github.com/disintegration/imaging` - ресайз
- `github.com/aws/aws-sdk-go-v2` - работа с S3

### 9.3 CDN (опционально)
- CloudFront / Cloudflare
- Кеширование изображений
- Сжатие

---

## 14. БЕЗОПАСНОСТЬ

### 10.1 Аутентификация и авторизация
- JWT токены с коротким временем жизни (15 мин access, 7 дней refresh)
- HTTPS обязательно
- Rate limiting для API
- CORS настройки

### 10.2 Валидация данных
- Валидация на уровне Go (validator библиотека)
- Sanitization пользовательского ввода
- Защита от SQL injection (prepared statements)

### 10.3 Защита файлов
- Проверка типа файлов (не только по расширению)
- Ограничение размера
- Антивирус сканирование (опционально)

### 10.4 Защита API
- Rate limiting (100 req/min для анонимных, 1000 для авторизованных)
- API ключи для Telegram бота
- Логирование всех действий

---

## 15. ПРОИЗВОДИТЕЛЬНОСТЬ

### 11.1 Кеширование
- **Redis** для:
  - Сессии пользователей
  - Кеш каталога товаров (5 мин)
  - Кеш категорий (30 мин)
  - Rate limiting counters

### 11.2 Оптимизация запросов
- Индексы в PostgreSQL
- Eager loading связанных данных
- Пагинация всех списков
- Ленивая загрузка изображений на фронтенде

### 11.3 Масштабирование
- Горизонтальное масштабирование Go backend
- Connection pooling для PostgreSQL
- S3 для статики

---

## 16. МОНИТОРИНГ И ЛОГИРОВАНИЕ

### 12.1 Логи
- Структурированное логирование (JSON)
- Уровни: DEBUG, INFO, WARN, ERROR
- Ротация логов

**Что логировать:**
- Все API запросы
- Ошибки
- Изменения заказов
- Действия администраторов

### 12.2 Метрики (опционально)
- Prometheus для сбора метрик
- Grafana для визуализации
- Отслеживание:
  - Время ответа API
  - Количество запросов
  - Ошибки
  - Использование ресурсов

---

## 17. РАЗВЕРТЫВАНИЕ

### 13.1 Docker Compose для разработки

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: shop_db
      POSTGRES_USER: shop_user
      POSTGRES_PASSWORD: shop_pass
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  minio:  # S3-совместимое хранилище для разработки
    image: minio/minio
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - minio_data:/data

  backend:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
      - minio
    environment:
      DATABASE_URL: postgresql://shop_user:shop_pass@postgres:5432/shop_db
      REDIS_URL: redis://redis:6379
      S3_ENDPOINT: http://minio:9000
    volumes:
      - ./:/app

volumes:
  postgres_data:
  minio_data:
```

### 13.2 Production
- VPS / Dedicated Server / Cloud (AWS, Yandex Cloud)
- Nginx как reverse proxy
- SSL сертификат (Let's Encrypt)
- Автоматические бэкапы БД
- CI/CD pipeline

---

## 18. ТЕСТИРОВАНИЕ

### 14.1 Backend
- Unit тесты для бизнес-логики
- Integration тесты для API
- Покрытие кода >70%

### 14.2 Frontend
- Unit тесты компонентов (Jest, React Testing Library)
- E2E тесты критических флоу (Playwright, Cypress)

### 14.3 Telegram Bot
- Тесты команд и callback'ов
- Мок Telegram API

---

## 19. ДОКУМЕНТАЦИЯ

### 15.1 API Documentation
- Swagger/OpenAPI спецификация
- Автоматическая генерация из кода (swag для Go)
- Доступна по адресу `/api/docs`

### 15.2 README
- Инструкции по запуску проекта
- Переменные окружения
- Команды для миграций
- Примеры использования API

---

## 20. ЭТАПЫ РАЗРАБОТКИ

### MVP (Минимально жизнеспособный продукт) - 4-6 недель

**Этап 1: Backend основа (1-2 недели)**
- [ ] Настройка проекта, структура
- [ ] PostgreSQL схема и миграции
- [ ] CRUD API для товаров
- [ ] CRUD API для категорий
- [ ] Загрузка изображений в S3
- [ ] Базовая авторизация

**Этап 2: Core функционал (2 недели)**
- [ ] Корзина
- [ ] Создание заказов
- [ ] Управление заказами
- [ ] Интеграция с Telegram (базовый бот)

**Этап 3: Frontend (2 недели)**
- [ ] Главная страница
- [ ] Каталог с фильтрами
- [ ] Страница товара
- [ ] Корзина и оформление заказа
- [ ] Базовая админка (управление товарами и заказами)

**Этап 4: Telegram Web App (1 неделя)**
- [ ] Адаптация frontend для Telegram
- [ ] Интеграция с ботом
- [ ] Синхронизация корзины

### Развитие — Этап 5: Модуль доставки (1-2 недели)
- [ ] Интеграция с СДЭК API
- [ ] Интеграция с Boxberry API
- [ ] Расчёт стоимости доставки
- [ ] Карта пунктов выдачи (Яндекс.Карты)
- [ ] Webhook для отслеживания статуса
- [ ] Синхронизация ПВЗ
- [ ] Интерфейс выбора доставки в checkout

### Развитие — Этап 6: Импорт товаров (1 неделя)
- [ ] Загрузка и парсинг CSV/XLSX
- [ ] Интерфейс маппинга полей
- [ ] Предпросмотр данных
- [ ] Валидация с детальными ошибками
- [ ] Фоновый импорт с прогрессом
- [ ] Шаблоны маппинга
- [ ] Образец файла для скачивания

### Развитие — Этап 7: Аналитика и экспорт (1-2 недели)
- [ ] Dashboard с виджетами
- [ ] Графики продаж (Chart.js / Recharts)
- [ ] Отчёты по товарам, клиентам, заказам
- [ ] Экспорт в CSV/XLSX
- [ ] Агрегация ежедневной статистики (cron)
- [ ] Фильтры и выбор периода

### Развитие — Этап 8: Реферальная программа (1 неделя)
- [ ] Генерация реферальных кодов
- [ ] Регистрация по реф. ссылке
- [ ] Начисление бонусов за рефералов
- [ ] Личный кабинет: страница рефералов
- [ ] Админка: настройки программы
- [ ] Статистика реферальной программы

### Развитие — Этап 9: Расширенные промокоды (1 неделя)
- [ ] Расширенные условия (категории, товары)
- [ ] Лимит использований на пользователя
- [ ] Массовая генерация промокодов
- [ ] Статистика использования
- [ ] Персональные промокоды
- [ ] Бонусный баланс и оплата бонусами

### Дополнительно (постепенно)
- [ ] Отзывы и рейтинги
- [ ] Email/Telegram уведомления
- [ ] Улучшенная админка
- [ ] Интеграция с платежными системами (ЮKassa, Тинькофф)
- [ ] Интеграция с Почтой России API
- [ ] Мобильное приложение

---

## 21. ПРИМЕРНАЯ СТОИМОСТЬ РАЗРАБОТКИ

### Команда фрилансеров:
- Backend разработчик (Go): 150-250k ₽
- Frontend разработчик (React): 120-200k ₽
- **Итого MVP:** 270-450k ₽

### Студия разработки:
- MVP: 400-800k ₽
- Полная версия: 1-2M ₽

### Самостоятельная разработка:
- Время: 3-6 месяцев
- Инфраструктура: 5-10k ₽/месяц

---

## 22. ПЕРЕМЕННЫЕ ОКРУЖЕНИЯ

```env
# Server
PORT=8080
ENV=development  # development, production

# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/shop_db
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE=5

# Redis
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=

# S3 Storage
S3_ENDPOINT=https://s3.amazonaws.com
S3_REGION=us-east-1
S3_BUCKET=my-shop-bucket
S3_ACCESS_KEY=your-access-key
S3_SECRET_KEY=your-secret-key
S3_PUBLIC_URL=https://cdn.example.com

# JWT
JWT_SECRET=your-super-secret-key
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

# Telegram
TELEGRAM_BOT_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
TELEGRAM_WEBHOOK_URL=https://yourdomain.com/webhook/telegram

# CORS
ALLOWED_ORIGINS=http://localhost:3000,https://yourdomain.com

# Email (опционально)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-password

# Payment (для будущего)
PAYMENT_PROVIDER=yookassa
PAYMENT_API_KEY=
```

---

## 23. ПОЛЕЗНЫЕ БИБЛИОТЕКИ GO

```go
// Backend framework
"github.com/gin-gonic/gin"

// Database
"gorm.io/gorm"
"gorm.io/driver/postgres"

// Migrations
"github.com/golang-migrate/migrate/v4"

// S3
"github.com/aws/aws-sdk-go-v2/service/s3"

// Image processing
"github.com/disintegration/imaging"

// Telegram
"gopkg.in/telebot.v3"

// JWT
"github.com/golang-jwt/jwt/v5"

// Validation
"github.com/go-playground/validator/v10"

// Config
"github.com/spf13/viper"

// Logging
"github.com/sirupsen/logrus"
"go.uber.org/zap"

// Redis
"github.com/go-redis/redis/v8"

// UUID
"github.com/google/uuid"

// CORS
"github.com/gin-contrib/cors"

// Excel для импорта/экспорта
"github.com/xuri/excelize/v2"

// CSV
"encoding/csv"  // стандартная библиотека

// Background jobs
"github.com/hibiken/asynq"  // или
"github.com/robfig/cron/v3"

// Decimal для денежных операций
"github.com/shopspring/decimal"
```

---

## 24. КОНТАКТЫ И ПОДДЕРЖКА

**Проектная документация:**
- Этот файл - основное ТЗ
- API документация: `/api/docs` после запуска
- README проекта: инструкции по разработке

**Для вопросов:**
- Создавайте issues в репозитории
- Документируйте все изменения

---

## 25. ЧЕКЛИСТ ПЕРЕД ЗАПУСКОМ

### Development:
- [ ] Все зависимости установлены
- [ ] PostgreSQL запущен, миграции применены
- [ ] Redis запущен (если используется)
- [ ] MinIO/S3 настроен и доступен
- [ ] .env файл заполнен
- [ ] Backend запускается без ошибок
- [ ] Frontend запускается без ошибок
- [ ] Telegram бот отвечает на команды

### Production:
- [ ] SSL сертификат настроен
- [ ] Все секреты в безопасности
- [ ] Бэкапы настроены
- [ ] Мониторинг настроен
- [ ] Логи пишутся корректно
- [ ] Rate limiting включен
- [ ] Домен настроен
- [ ] Email/Telegram уведомления работают
- [ ] Проведено нагрузочное тестирование

---

**Версия документа:** 2.0
**Дата последнего обновления:** 05.02.2026
**Статус:** Утверждено к разработке

---

## ИСТОРИЯ ИЗМЕНЕНИЙ

| Версия | Дата | Описание |
|--------|------|----------|
| 1.0 | 04.02.2026 | Первоначальная версия ТЗ |
| 2.0 | 05.02.2026 | Добавлены модули: доставка (интеграция СДЭК/Boxberry, расчёт стоимости, ПВЗ), импорт товаров (CSV/XLSX), аналитика и экспорт (dashboard, отчёты), реферальная программа и расширенные промокоды |

