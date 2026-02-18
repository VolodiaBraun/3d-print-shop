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
  bonusAmount?: number;
  notes?: string;
  telegramId?: number;
  pickupPointId?: number;
  city?: string;
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
  orderType: string; // "regular" | "custom"
  status: string;
  subtotal: number;
  discountAmount: number;
  bonusDiscount: number;
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

// --- Delivery API ---

export interface DeliveryOption {
  type: string;
  name: string;
  cost: number;
  originalCost: number;
  estimatedDaysMin: number;
  estimatedDaysMax: number;
  isFreeDelivery: boolean;
  providerName: string;
}

export interface PickupPointData {
  id: number;
  name: string;
  address: string;
  city: string;
  latitude?: number;
  longitude?: number;
  phone?: string;
  workingHours: string;
}

export interface DeliveryCalculationResult {
  courierOptions: DeliveryOption[];
  pickupPoints: PickupPointData[];
  hasPickupPoints: boolean;
}

export async function calculateDelivery(
  city: string,
  orderTotal: number,
  totalWeight?: number
): Promise<DeliveryCalculationResult> {
  const { data } = await api.post<ApiResponse<DeliveryCalculationResult>>(
    "/delivery/calculate",
    { city, orderTotal, totalWeight }
  );
  return data.data;
}

export async function getPickupPoints(
  city: string
): Promise<PickupPointData[]> {
  const { data } = await api.get<ApiResponse<PickupPointData[]>>(
    "/delivery/pickup-points",
    { params: { city } }
  );
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

// --- Profile API ---

export interface ProfileData {
  id: number;
  email?: string;
  emailVerified?: boolean;
  phone?: string;
  firstName?: string;
  lastName?: string;
  username?: string;
  telegramId?: number;
  role: string;
  createdAt: string;
}

export async function getProfile(): Promise<ProfileData> {
  const { data } = await api.get<ApiResponse<ProfileData>>("/users/me");
  return data.data;
}

export async function updateProfile(
  input: Partial<Pick<ProfileData, "firstName" | "lastName" | "phone" | "email">>
): Promise<ProfileData> {
  const { data } = await api.put<ApiResponse<ProfileData>>("/users/me", input);
  return data.data;
}

// --- Reviews API ---

export interface ReviewData {
  id: number;
  userId?: number;
  productId: number;
  orderId: number;
  rating: number;
  comment?: string;
  status: string;
  user?: { id: number; firstName?: string; lastName?: string };
  product?: { id: number; name: string; slug: string };
  createdAt: string;
}

export async function getProductReviews(
  productId: number
): Promise<ReviewData[]> {
  const { data } = await api.get<ApiResponse<ReviewData[]>>(
    `/products/${productId}/reviews`
  );
  return data.data;
}

export interface CreateReviewInput {
  orderId: number;
  rating: number;
  comment?: string;
}

export async function createReview(
  productId: number,
  input: CreateReviewInput
): Promise<ReviewData> {
  const { data } = await api.post<ApiResponse<ReviewData>>(
    `/products/${productId}/reviews`,
    input
  );
  return data.data;
}

export async function getMyReviews(): Promise<ReviewData[]> {
  const { data } = await api.get<ApiResponse<ReviewData[]>>("/reviews/my");
  return data.data;
}

// --- Referral / Loyalty API ---

export interface ReferralInfo {
  referralCode: string;
  referralLink: string;
  referralsCount: number;
  bonusBalance: number;
}

export async function getReferralInfo(): Promise<ReferralInfo> {
  const { data } = await api.get<ApiResponse<ReferralInfo>>("/users/me/referral");
  return data.data;
}

export async function applyReferralCode(code: string): Promise<void> {
  await api.post("/users/me/referral/apply", { code });
}

export interface BonusHistoryItem {
  id: number;
  userId: number;
  amount: number;
  type: string;
  referenceId?: number;
  description?: string;
  createdAt: string;
}

export async function getBonusHistory(): Promise<BonusHistoryItem[]> {
  const { data } = await api.get<ApiResponse<BonusHistoryItem[]>>("/users/me/bonuses");
  return data.data;
}

// --- Email Verification API ---

export async function sendVerificationCode(): Promise<void> {
  await api.post("/users/me/email/verify");
}

export async function confirmVerificationCode(code: string): Promise<void> {
  await api.post("/users/me/email/confirm", { code });
}

// --- Content API ---

export async function getContentBlock<T = Record<string, string>>(slug: string): Promise<T> {
  const { data } = await api.get<T>(`/content/${slug}`);
  return data;
}

// --- Custom Orders (public) ---

export interface SubmitCustomOrderInput {
  customerName: string;
  customerPhone: string;
  customerEmail?: string;
  clientDescription?: string;
  paymentMethod: string;
  deliveryMethod: string;
  deliveryAddress?: string;
  notes?: string;
}

export interface CustomOrderSubmitResponse {
  id: number;
  orderNumber: string;
  status: string;
  customerName: string;
  customerPhone: string;
  createdAt: string;
}

export async function submitCustomOrder(
  input: SubmitCustomOrderInput
): Promise<CustomOrderSubmitResponse> {
  const { data } = await api.post<ApiResponse<CustomOrderSubmitResponse>>(
    "/custom-orders",
    input
  );
  return data.data;
}

export async function uploadCustomOrderFile(
  orderId: number,
  file: File
): Promise<string> {
  const form = new FormData();
  form.append("file", file);
  const { data } = await api.post<ApiResponse<{ url: string }>>(
    `/custom-orders/${orderId}/files`,
    form,
    { headers: { "Content-Type": "multipart/form-data" }, timeout: 60000 }
  );
  return data.data.url;
}

export default api;
