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
} from "@ant-design/icons";

import { dataProvider } from "./providers/dataProvider";
import { authProvider } from "./providers/authProvider";
import { LoginPage } from "./pages/login";
import { DashboardPage } from "./pages/dashboard";
import { CategoryList } from "./pages/categories/list";

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
