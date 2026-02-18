import { Refine, Authenticated } from "@refinedev/core";
import { ThemedLayoutV2, useNotificationProvider } from "@refinedev/antd";
import routerProvider, {
  NavigateToResource,
  CatchAllNavigate,
} from "@refinedev/react-router-v6";
import { BrowserRouter, Routes, Route, Outlet } from "react-router-dom";
import { App as AntdApp, ConfigProvider } from "antd";
import ruRU from "antd/locale/ru_RU";
import {
  DashboardOutlined,
  FolderOutlined,
  ShoppingOutlined,
  ShoppingCartOutlined,
  GiftOutlined,
  MessageOutlined,
  HeartOutlined,
  EditOutlined,
  ExperimentOutlined,
} from "@ant-design/icons";

import { dataProvider } from "./providers/dataProvider";
import { authProvider } from "./providers/authProvider";
import { LoginPage } from "./pages/login";
import { DashboardPage } from "./pages/dashboard";
import { CategoryList } from "./pages/categories/list";
import { ProductList } from "./pages/products/list";
import { ProductForm } from "./pages/products/form";
import { OrderList } from "./pages/orders/list";
import { OrderDetail } from "./pages/orders/detail";
import { CustomOrderList } from "./pages/custom-orders/list";
import { CustomOrderDetail } from "./pages/custom-orders/detail";
import { CustomOrderCreate } from "./pages/custom-orders/create";
import { PromoList } from "./pages/promos/list";
import { PromoForm } from "./pages/promos/form";
import { ReviewList } from "./pages/reviews/list";
import { LoyaltySettings } from "./pages/loyalty/settings";
import { ContentEditor } from "./pages/content/editor";

import "@refinedev/antd/dist/reset.css";

const AdminApp = () => {
  return (
    <BrowserRouter basename="/admin">
      <ConfigProvider locale={ruRU}>
        <AntdApp>
          <Refine
            dataProvider={dataProvider}
            authProvider={authProvider}
            routerProvider={routerProvider}
            notificationProvider={useNotificationProvider}
            resources={[
              {
                name: "dashboard",
                list: "/",
                meta: { label: "Главная", icon: <DashboardOutlined /> },
              },
              {
                name: "categories",
                list: "/categories",
                meta: { label: "Категории", icon: <FolderOutlined /> },
              },
              {
                name: "products",
                list: "/products",
                create: "/products/create",
                edit: "/products/:id/edit",
                meta: { label: "Товары", icon: <ShoppingOutlined /> },
              },
              {
                name: "orders",
                list: "/orders",
                show: "/orders/:id",
                meta: { label: "Заказы", icon: <ShoppingCartOutlined /> },
              },
              {
                name: "custom-orders",
                list: "/custom-orders",
                create: "/custom-orders/create",
                show: "/custom-orders/:id",
                meta: { label: "Инд. заказы", icon: <ExperimentOutlined /> },
              },
              {
                name: "promos",
                list: "/promos",
                create: "/promos/create",
                edit: "/promos/:id/edit",
                meta: { label: "Промокоды", icon: <GiftOutlined /> },
              },
              {
                name: "reviews",
                list: "/reviews",
                meta: { label: "Отзывы", icon: <MessageOutlined /> },
              },
              {
                name: "loyalty",
                list: "/loyalty",
                meta: { label: "Лояльность", icon: <HeartOutlined /> },
              },
              {
                name: "content",
                list: "/content",
                meta: { label: "Контент", icon: <EditOutlined /> },
              },
            ]}
            options={{
              syncWithLocation: true,
              warnWhenUnsavedChanges: true,
            }}
          >
            <Routes>
              <Route
                element={
                  <Authenticated
                    key="auth-routes"
                    fallback={<CatchAllNavigate to="/login" />}
                  >
                    <ThemedLayoutV2
                      Title={() => (
                        <div style={{ padding: "0 12px", fontWeight: 600 }}>
                          АВАНГАРД
                        </div>
                      )}
                    >
                      <Outlet />
                    </ThemedLayoutV2>
                  </Authenticated>
                }
              >
                <Route index element={<DashboardPage />} />
                <Route path="/categories" element={<CategoryList />} />
                <Route path="/products" element={<ProductList />} />
                <Route path="/products/create" element={<ProductForm />} />
                <Route path="/products/:id/edit" element={<ProductForm />} />
                <Route path="/orders" element={<OrderList />} />
                <Route path="/orders/:id" element={<OrderDetail />} />
                <Route path="/custom-orders" element={<CustomOrderList />} />
                <Route path="/custom-orders/create" element={<CustomOrderCreate />} />
                <Route path="/custom-orders/:id" element={<CustomOrderDetail />} />
                <Route path="/promos" element={<PromoList />} />
                <Route path="/promos/create" element={<PromoForm />} />
                <Route path="/promos/:id/edit" element={<PromoForm />} />
                <Route path="/reviews" element={<ReviewList />} />
                <Route path="/loyalty" element={<LoyaltySettings />} />
                <Route path="/content" element={<ContentEditor />} />
              </Route>

              <Route
                element={
                  <Authenticated key="auth-login" fallback={<Outlet />}>
                    <NavigateToResource resource="dashboard" />
                  </Authenticated>
                }
              >
                <Route path="/login" element={<LoginPage />} />
              </Route>
            </Routes>
          </Refine>
        </AntdApp>
      </ConfigProvider>
    </BrowserRouter>
  );
};

export default AdminApp;
