import type { AuthProvider } from "@refinedev/core";
import api from "../lib/api";

export const authProvider: AuthProvider = {
  login: async ({ email, password }) => {
    try {
      const { data } = await api.post("/auth/login", { email, password });
      const tokens = data.data;

      localStorage.setItem("accessToken", tokens.accessToken);
      localStorage.setItem("refreshToken", tokens.refreshToken);

      return {
        success: true,
        redirectTo: "/",
      };
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: { code?: string; message?: string } } } };
      const code = err.response?.data?.error?.code;
      const message = err.response?.data?.error?.message;

      if (code === "ACCOUNT_DISABLED") {
        return {
          success: false,
          error: {
            name: "Аккаунт деактивирован",
            message: message || "Обратитесь к администратору",
          },
        };
      }

      return {
        success: false,
        error: {
          name: "Ошибка входа",
          message: message || "Неверный email или пароль",
        },
      };
    }
  },

  logout: async () => {
    localStorage.removeItem("accessToken");
    localStorage.removeItem("refreshToken");

    return {
      success: true,
      redirectTo: "/login",
    };
  },

  check: async () => {
    const token = localStorage.getItem("accessToken");
    if (!token) {
      return {
        authenticated: false,
        redirectTo: "/login",
      };
    }

    return {
      authenticated: true,
    };
  },

  getIdentity: async () => {
    const token = localStorage.getItem("accessToken");
    if (!token) return null;

    try {
      // Decode JWT payload (no verification — server validates)
      const payload = JSON.parse(atob(token.split(".")[1]));
      return {
        id: payload.sub,
        name: payload.role === "admin" ? "Администратор" : "Пользователь",
        role: payload.role,
      };
    } catch {
      return null;
    }
  },

  onError: async (error) => {
    const status = (error as { statusCode?: number }).statusCode;
    if (status === 401) {
      return {
        logout: true,
        redirectTo: "/login",
      };
    }
    return { error };
  },
};
