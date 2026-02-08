"use client";

import {
  createContext,
  useContext,
  useReducer,
  useEffect,
  useCallback,
  type ReactNode,
} from "react";
import type { CartItem, Product } from "./types";

const STORAGE_KEY = "avangard_cart";

interface CartState {
  items: CartItem[];
  loaded: boolean;
}

type CartAction =
  | { type: "LOAD"; items: CartItem[] }
  | { type: "ADD"; product: Product; quantity: number }
  | { type: "UPDATE_QTY"; productId: number; quantity: number }
  | { type: "REMOVE"; productId: number }
  | { type: "CLEAR" };

function cartReducer(state: CartState, action: CartAction): CartState {
  switch (action.type) {
    case "LOAD":
      return { items: action.items, loaded: true };

    case "ADD": {
      const { product, quantity } = action;
      const existing = state.items.find((i) => i.productId === product.id);
      if (existing) {
        const newQty = Math.min(
          existing.quantity + quantity,
          product.stockQuantity
        );
        return {
          ...state,
          items: state.items.map((i) =>
            i.productId === product.id ? { ...i, quantity: newQty } : i
          ),
        };
      }
      const mainImage = product.images?.find((img) => img.isMain);
      const image =
        mainImage?.urlThumbnail || mainImage?.url || product.images?.[0]?.urlThumbnail || product.images?.[0]?.url;
      return {
        ...state,
        items: [
          ...state.items,
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
        ],
      };
    }

    case "UPDATE_QTY": {
      return {
        ...state,
        items: state.items.map((i) =>
          i.productId === action.productId
            ? { ...i, quantity: Math.max(1, Math.min(action.quantity, i.stockQuantity)) }
            : i
        ),
      };
    }

    case "REMOVE":
      return {
        ...state,
        items: state.items.filter((i) => i.productId !== action.productId),
      };

    case "CLEAR":
      return { ...state, items: [] };

    default:
      return state;
  }
}

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

export function CartProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(cartReducer, {
    items: [],
    loaded: false,
  });

  // Load from localStorage on mount
  useEffect(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        const items = JSON.parse(stored) as CartItem[];
        dispatch({ type: "LOAD", items });
      } else {
        dispatch({ type: "LOAD", items: [] });
      }
    } catch {
      dispatch({ type: "LOAD", items: [] });
    }
  }, []);

  // Save to localStorage on change
  useEffect(() => {
    if (state.loaded) {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(state.items));
    }
  }, [state.items, state.loaded]);

  const addItem = useCallback(
    (product: Product, quantity = 1) => {
      dispatch({ type: "ADD", product, quantity });
    },
    []
  );

  const updateQuantity = useCallback(
    (productId: number, quantity: number) => {
      dispatch({ type: "UPDATE_QTY", productId, quantity });
    },
    []
  );

  const removeItem = useCallback(
    (productId: number) => {
      dispatch({ type: "REMOVE", productId });
    },
    []
  );

  const clearCart = useCallback(() => {
    dispatch({ type: "CLEAR" });
  }, []);

  const getItemQuantity = useCallback(
    (productId: number) => {
      return state.items.find((i) => i.productId === productId)?.quantity ?? 0;
    },
    [state.items]
  );

  const totalItems = state.items.reduce((sum, i) => sum + i.quantity, 0);
  const totalPrice =
    Math.round(
      state.items.reduce((sum, i) => sum + i.price * i.quantity, 0) * 100
    ) / 100;

  return (
    <CartContext.Provider
      value={{
        items: state.items,
        addItem,
        updateQuantity,
        removeItem,
        clearCart,
        getItemQuantity,
        totalItems,
        totalPrice,
        loaded: state.loaded,
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
