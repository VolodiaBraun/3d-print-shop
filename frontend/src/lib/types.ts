export interface Category {
  id: number;
  name: string;
  slug: string;
  description?: string;
  parentId?: number;
  displayOrder: number;
  imageUrl?: string;
  isActive: boolean;
  children?: Category[];
}

export interface ProductImage {
  id: number;
  productId: number;
  url: string;
  urlLarge?: string;
  urlMedium?: string;
  urlThumbnail?: string;
  s3Key: string;
  isMain: boolean;
  displayOrder: number;
}

export interface Dimensions {
  length?: number;
  width?: number;
  height?: number;
}

export interface Product {
  id: number;
  name: string;
  slug: string;
  description?: string;
  shortDescription?: string;
  price: number;
  oldPrice?: number;
  stockQuantity: number;
  sku?: string;
  weight?: number;
  dimensions?: Dimensions;
  material?: string;
  printTime?: number;
  categoryId?: number;
  category?: Category;
  images?: ProductImage[];
  isActive: boolean;
  isFeatured: boolean;
  viewsCount: number;
  salesCount: number;
  rating: number;
  createdAt: string;
  updatedAt: string;
}

export interface CartItem {
  id?: number; // server-side cart item ID
  productId: number;
  name: string;
  slug: string;
  price: number;
  oldPrice?: number;
  image?: string;
  quantity: number;
  stockQuantity: number;
}

export interface PaginationMeta {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  meta: PaginationMeta;
}

export interface ApiResponse<T> {
  data: T;
}
