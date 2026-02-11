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

const AUTH_KEY = "avangard_auth";

// Request interceptor: attach access token
api.interceptors.request.use((config) => {
  if (typeof window !== "undefined") {
    try {
      const raw = localStorage.getItem(AUTH_KEY);
      if (raw) {
        const { accessToken } = JSON.parse(raw);
        if (accessToken) {
          config.headers.Authorization = `Bearer ${accessToken}`;
        }
      }
    } catch {
      // ignore
    }
  }
  return config;
});

// Response interceptor: refresh token on 401
let refreshPromise: Promise<string> | null = null;

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const original = error.config;
    if (
      error.response?.status === 401 &&
      !original._retry &&
      !original.url?.includes("/auth/")
    ) {
      original._retry = true;

      try {
        if (!refreshPromise) {
          refreshPromise = (async () => {
            const raw = localStorage.getItem(AUTH_KEY);
            if (!raw) throw new Error("no auth");
            const stored = JSON.parse(raw);
            if (!stored.refreshToken) throw new Error("no refresh token");

            const { data } = await axios.post(
              `${api.defaults.baseURL}/auth/refresh`,
              { refreshToken: stored.refreshToken }
            );
            const newAccess = data.data.accessToken;
            const newRefresh = data.data.refreshToken;
            stored.accessToken = newAccess;
            stored.refreshToken = newRefresh;
            localStorage.setItem(AUTH_KEY, JSON.stringify(stored));
            return newAccess;
          })();
        }

        const newToken = await refreshPromise;
        refreshPromise = null;
        original.headers.Authorization = `Bearer ${newToken}`;
        return api(original);
      } catch {
        refreshPromise = null;
        localStorage.removeItem(AUTH_KEY);
        return Promise.reject(error);
      }
    }
    return Promise.reject(error);
  }
);

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
  telegramId?: number;
}

export interface OrderItemProduct {
  id: number;
  name: string;
  slug: string;
  images?: { url: string; urlThumbnail?: string; isMain: boolean }[];
}

export interface OrderResponseItem {
  id: number;
  productId: number;
  quantity: number;
  unitPrice: number;
  totalPrice: number;
  product?: OrderItemProduct;
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
  deliveryAddress?: string;
  paymentMethod: string;
  isPaid: boolean;
  notes?: string;
  items: OrderResponseItem[];
  createdAt: string;
  updatedAt: string;
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

export async function getOrder(
  orderNumber: string
): Promise<OrderResponse> {
  const { data } = await api.get<ApiResponse<OrderResponse>>(
    `/orders/${orderNumber}`
  );
  return data.data;
}

export async function getMyOrders(): Promise<OrderResponse[]> {
  const { data } = await api.get<ApiResponse<OrderResponse[]>>("/orders/my");
  return data.data;
}

// --- Server Cart API ---

export interface ServerCartItem {
  id: number;
  productId: number;
  quantity: number;
  product: Product;
}

export interface ServerCart {
  items: ServerCartItem[];
  totalItems: number;
  totalPrice: number;
}

export async function getServerCart(): Promise<ServerCart> {
  const { data } = await api.get<ApiResponse<ServerCart>>("/cart");
  return data.data;
}

export async function addServerCartItem(
  productId: number,
  quantity: number
): Promise<ServerCart> {
  const { data } = await api.post<ApiResponse<ServerCart>>("/cart/items", {
    productId,
    quantity,
  });
  return data.data;
}

export async function updateServerCartItem(
  itemId: number,
  quantity: number
): Promise<ServerCart> {
  const { data } = await api.put<ApiResponse<ServerCart>>(
    `/cart/items/${itemId}`,
    { quantity }
  );
  return data.data;
}

export async function removeServerCartItem(
  itemId: number
): Promise<ServerCart> {
  const { data } = await api.delete<ApiResponse<ServerCart>>(
    `/cart/items/${itemId}`
  );
  return data.data;
}

export async function clearServerCart(): Promise<void> {
  await api.delete("/cart");
}

export default api;
