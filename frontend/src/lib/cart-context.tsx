"use client";

import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  useRef,
  type ReactNode,
} from "react";
import type { CartItem, Product } from "./types";
import { useAuth } from "./auth-context";
import {
  getServerCart,
  addServerCartItem,
  updateServerCartItem,
  removeServerCartItem,
  clearServerCart,
  type ServerCart,
} from "./api";

const STORAGE_KEY = "avangard_cart";

interface CartContextType {
  items: CartItem[];
  addItem: (product: Product, quantity?: number) => void;
  updateQuantity: (productId: number, quantity: number) => void;
  removeItem: (productId: number) => void;
  clearCart: () => void;
  getItemQuantity: (productId: number) => number;
  totalItems: number;
  totalPrice: number;
  loaded: boolean;
}

const CartContext = createContext<CartContextType | null>(null);

// Convert server cart to local CartItem format
function serverCartToItems(cart: ServerCart): CartItem[] {
  return (cart.items || []).map((si) => {
    const mainImage = si.product.images?.find((img) => img.isMain);
    const image =
      mainImage?.urlThumbnail ||
      mainImage?.url ||
      si.product.images?.[0]?.urlThumbnail ||
      si.product.images?.[0]?.url;
    return {
      id: si.id,
      productId: si.productId,
      name: si.product.name,
      slug: si.product.slug,
      price: si.product.price,
      oldPrice: si.product.oldPrice,
      image,
      quantity: si.quantity,
      stockQuantity: si.product.stockQuantity,
    };
  });
}

function loadLocalCart(): CartItem[] {
  if (typeof window === "undefined") return [];
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
}

function saveLocalCart(items: CartItem[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
}

export function CartProvider({ children }: { children: ReactNode }) {
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  const [items, setItems] = useState<CartItem[]>([]);
  const [loaded, setLoaded] = useState(false);
  const mergedRef = useRef(false);

  // Load cart: server for authenticated, localStorage for guests
  useEffect(() => {
    if (authLoading) return;

    if (isAuthenticated) {
      // Load server cart
      getServerCart()
        .then(async (serverCart) => {
          const serverItems = serverCartToItems(serverCart);

          // Merge local cart into server (only once per session)
          if (!mergedRef.current) {
            mergedRef.current = true;
            const localItems = loadLocalCart();
            if (localItems.length > 0) {
              const serverProductIds = new Set(
                serverItems.map((i) => i.productId)
              );
              for (const li of localItems) {
                if (!serverProductIds.has(li.productId)) {
                  try {
                    await addServerCartItem(li.productId, li.quantity);
                  } catch {
                    // product may no longer exist
                  }
                }
              }
              localStorage.removeItem(STORAGE_KEY);
              // Reload server cart after merge
              const merged = await getServerCart();
              setItems(serverCartToItems(merged));
              setLoaded(true);
              return;
            }
          }

          setItems(serverItems);
          setLoaded(true);
        })
        .catch(() => {
          // Fallback to localStorage
          setItems(loadLocalCart());
          setLoaded(true);
        });
    } else {
      mergedRef.current = false;
      setItems(loadLocalCart());
      setLoaded(true);
    }
  }, [isAuthenticated, authLoading]);

  // Save to localStorage for guests only
  useEffect(() => {
    if (loaded && !isAuthenticated) {
      saveLocalCart(items);
    }
  }, [items, loaded, isAuthenticated]);

  const addItem = useCallback(
    (product: Product, quantity = 1) => {
      if (isAuthenticated) {
        addServerCartItem(product.id, quantity)
          .then((cart) => setItems(serverCartToItems(cart)))
          .catch(console.warn);
      } else {
        setItems((prev) => {
          const existing = prev.find((i) => i.productId === product.id);
          if (existing) {
            const newQty = Math.min(
              existing.quantity + quantity,
              product.stockQuantity
            );
            return prev.map((i) =>
              i.productId === product.id ? { ...i, quantity: newQty } : i
            );
          }
          const mainImage = product.images?.find((img) => img.isMain);
          const image =
            mainImage?.urlThumbnail ||
            mainImage?.url ||
            product.images?.[0]?.urlThumbnail ||
            product.images?.[0]?.url;
          return [
            ...prev,
            {
              productId: product.id,
              name: product.name,
              slug: product.slug,
              price: product.price,
              oldPrice: product.oldPrice,
              image,
              quantity: Math.min(quantity, product.stockQuantity),
              stockQuantity: product.stockQuantity,
            },
          ];
        });
      }
    },
    [isAuthenticated]
  );

  const updateQuantity = useCallback(
    (productId: number, quantity: number) => {
      if (isAuthenticated) {
        const item = items.find((i) => i.productId === productId);
        if (item?.id) {
          updateServerCartItem(item.id, quantity)
            .then((cart) => setItems(serverCartToItems(cart)))
            .catch(console.warn);
        }
      } else {
        setItems((prev) =>
          prev.map((i) =>
            i.productId === productId
              ? {
                  ...i,
                  quantity: Math.max(1, Math.min(quantity, i.stockQuantity)),
                }
              : i
          )
        );
      }
    },
    [isAuthenticated, items]
  );

  const removeItem = useCallback(
    (productId: number) => {
      if (isAuthenticated) {
        const item = items.find((i) => i.productId === productId);
        if (item?.id) {
          removeServerCartItem(item.id)
            .then((cart) => setItems(serverCartToItems(cart)))
            .catch(console.warn);
        }
      } else {
        setItems((prev) => prev.filter((i) => i.productId !== productId));
      }
    },
    [isAuthenticated, items]
  );

  const clearCart = useCallback(() => {
    if (isAuthenticated) {
      clearServerCart()
        .then(() => setItems([]))
        .catch(console.warn);
    } else {
      setItems([]);
    }
  }, [isAuthenticated]);

  const getItemQuantity = useCallback(
    (productId: number) =>
      items.find((i) => i.productId === productId)?.quantity ?? 0,
    [items]
  );

  const totalItems = items.reduce((sum, i) => sum + i.quantity, 0);
  const totalPrice =
    Math.round(items.reduce((sum, i) => sum + i.price * i.quantity, 0) * 100) /
    100;

  return (
    <CartContext.Provider
      value={{
        items,
        addItem,
        updateQuantity,
        removeItem,
        clearCart,
        getItemQuantity,
        totalItems,
        totalPrice,
        loaded,
      }}
    >
      {children}
    </CartContext.Provider>
  );
}

export function useCart() {
  const ctx = useContext(CartContext);
  if (!ctx) {
    throw new Error("useCart must be used within CartProvider");
  }
  return ctx;
}
