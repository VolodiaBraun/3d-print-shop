import axios from "axios";
import type {
  Product,
  Category,
  PaginatedResponse,
  ApiResponse,
} from "./types";

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1",
  timeout: 10000,
});

export async function getProducts(params?: {
  page?: number;
  limit?: number;
  category?: string;
  sort?: string;
  search?: string;
  min_price?: number;
  max_price?: number;
  material?: string;
}): Promise<PaginatedResponse<Product>> {
  const { data } = await api.get<PaginatedResponse<Product>>("/products", {
    params,
  });
  return data;
}

export async function getProductBySlug(slug: string): Promise<Product> {
  const { data } = await api.get<ApiResponse<Product>>(`/products/${slug}`);
  return data.data;
}

export async function getCategories(): Promise<Category[]> {
  const { data } = await api.get<ApiResponse<Category[]>>("/categories");
  return data.data;
}

export async function getSearchSuggestions(q: string): Promise<string[]> {
  const { data } = await api.get<ApiResponse<string[]>>(
    "/search/suggestions",
    { params: { q } }
  );
  return data.data;
}

export interface PromoValidationResult {
  valid: boolean;
  code: string;
  discountType: "percent" | "fixed";
  discountValue: number;
  discountAmount: number;
  message?: string;
}

export async function validatePromoCode(
  code: string,
  orderTotal: number
): Promise<PromoValidationResult> {
  const { data } = await api.post<ApiResponse<PromoValidationResult>>(
    "/promo/validate",
    { code, orderTotal }
  );
  return data.data;
}

export interface CreateOrderInput {
  items: { productId: number; quantity: number }[];
  customerName: string;
  customerPhone: string;
  customerEmail?: string;
  deliveryMethod: string;
  deliveryAddress?: string;
  paymentMethod: string;
  promoCode?: string;
  notes?: string;
}

export interface OrderResponse {
  id: number;
  orderNumber: string;
  status: string;
  subtotal: number;
  discountAmount: number;
  deliveryCost: number;
  totalPrice: number;
  promoCode?: string;
  customerName: string;
  customerPhone: string;
  customerEmail?: string;
  deliveryMethod: string;
  paymentMethod: string;
  items: {
    productId: number;
    quantity: number;
    unitPrice: number;
    totalPrice: number;
  }[];
  createdAt: string;
}

export async function createOrder(
  input: CreateOrderInput
): Promise<OrderResponse> {
  const { data } = await api.post<ApiResponse<OrderResponse>>(
    "/orders",
    input
  );
  return data.data;
}

export default api;
