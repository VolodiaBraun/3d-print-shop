import { useState, useEffect, useCallback } from "react";
import { Table, Tag, Select, Card, Space, Button, App } from "antd";
import {
  ShoppingCartOutlined,
  EyeOutlined,
} from "@ant-design/icons";
import type { ColumnsType, TablePaginationConfig } from "antd/es/table";
import { useNavigate } from "react-router-dom";
import api from "../../lib/api";

interface OrderItem {
  id: number;
  productId: number;
  quantity: number;
  unitPrice: number;
  totalPrice: number;
}

interface Order {
  id: number;
  orderNumber: string;
  status: string;
  subtotal: number;
  discountAmount: number;
  totalPrice: number;
  promoCode?: string;
  deliveryMethod: string;
  paymentMethod: string;
  customerName: string;
  customerPhone: string;
  customerEmail?: string;
  trackingNumber?: string;
  items: OrderItem[];
  createdAt: string;
}

const STATUS_CONFIG: Record<string, { color: string; label: string }> = {
  new: { color: "gold", label: "Новый" },
  confirmed: { color: "blue", label: "Подтверждён" },
  processing: { color: "cyan", label: "В обработке" },
  shipped: { color: "purple", label: "Отправлен" },
  delivered: { color: "green", label: "Доставлен" },
  cancelled: { color: "red", label: "Отменён" },
};

const STATUS_OPTIONS = [
  { value: "", label: "Все статусы" },
  { value: "new", label: "Новый" },
  { value: "confirmed", label: "Подтверждён" },
  { value: "processing", label: "В обработке" },
  { value: "shipped", label: "Отправлен" },
  { value: "delivered", label: "Доставлен" },
  { value: "cancelled", label: "Отменён" },
];

export const OrderList = () => {
  const { message } = App.useApp();
  const navigate = useNavigate();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [statusFilter, setStatusFilter] = useState("");

  const fetchOrders = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, string | number> = {
        page,
        limit: pageSize,
      };
      if (statusFilter) params.status = statusFilter;

      const { data: resp } = await api.get("/admin/orders", { params });
      setOrders(resp.data || []);
      setTotal(resp.meta?.total || 0);
    } catch {
      message.error("Не удалось загрузить заказы");
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, statusFilter, message]);

  useEffect(() => {
    fetchOrders();
  }, [fetchOrders]);

  const formatPrice = (price: number) =>
    new Intl.NumberFormat("ru-RU", {
      style: "currency",
      currency: "RUB",
      maximumFractionDigits: 0,
    }).format(price);

  const formatDate = (dateStr: string) => {
    const d = new Date(dateStr);
    return d.toLocaleDateString("ru-RU", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const columns: ColumnsType<Order> = [
    {
      title: "Номер",
      dataIndex: "orderNumber",
      key: "orderNumber",
      width: 180,
      render: (num: string) => (
        <span style={{ fontWeight: 500, fontFamily: "monospace" }}>{num}</span>
      ),
    },
    {
      title: "Клиент",
      key: "customer",
      width: 200,
      render: (_: unknown, record: Order) => (
        <div>
          <div style={{ fontWeight: 500 }}>{record.customerName}</div>
          <div style={{ fontSize: 12, color: "#999" }}>
            {record.customerPhone}
          </div>
        </div>
      ),
    },
    {
      title: "Товаров",
      key: "itemsCount",
      width: 90,
      align: "center",
      render: (_: unknown, record: Order) => record.items?.length || 0,
    },
    {
      title: "Сумма",
      dataIndex: "totalPrice",
      key: "totalPrice",
      width: 130,
      render: (price: number, record: Order) => (
        <div>
          <div style={{ fontWeight: 500 }}>{formatPrice(price)}</div>
          {record.discountAmount > 0 && (
            <div style={{ fontSize: 12, color: "#52c41a" }}>
              -{formatPrice(record.discountAmount)}
            </div>
          )}
        </div>
      ),
    },
    {
      title: "Статус",
      dataIndex: "status",
      key: "status",
      width: 130,
      render: (status: string) => {
        const cfg = STATUS_CONFIG[status] || {
          color: "default",
          label: status,
        };
        return <Tag color={cfg.color}>{cfg.label}</Tag>;
      },
    },
    {
      title: "Доставка",
      dataIndex: "deliveryMethod",
      key: "delivery",
      width: 110,
      render: (method: string) =>
        method === "courier" ? (
          <Tag>Курьер</Tag>
        ) : (
          <Tag>Самовывоз</Tag>
        ),
    },
    {
      title: "Дата",
      dataIndex: "createdAt",
      key: "createdAt",
      width: 150,
      render: (date: string) => (
        <span style={{ fontSize: 13, color: "#666" }}>{formatDate(date)}</span>
      ),
    },
    {
      title: "",
      key: "actions",
      width: 60,
      render: (_: unknown, record: Order) => (
        <Button
          type="text"
          size="small"
          icon={<EyeOutlined />}
          onClick={() => navigate(`/orders/${record.id}`)}
        />
      ),
    },
  ];

  const handleTableChange = (pagination: TablePaginationConfig) => {
    setPage(pagination.current || 1);
    setPageSize(pagination.pageSize || 20);
  };

  return (
    <Card
      title={
        <Space>
          <ShoppingCartOutlined />
          <span>Заказы</span>
        </Space>
      }
    >
      <Space style={{ marginBottom: 16 }} wrap>
        <Select
          style={{ width: 200 }}
          options={STATUS_OPTIONS}
          value={statusFilter}
          onChange={(val) => {
            setStatusFilter(val);
            setPage(1);
          }}
        />
      </Space>

      <Table
        columns={columns}
        dataSource={orders}
        rowKey="id"
        loading={loading}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showTotal: (t) => `Всего: ${t}`,
        }}
        onChange={handleTableChange}
        onRow={(record) => ({
          onClick: () => navigate(`/orders/${record.id}`),
          style: { cursor: "pointer" },
        })}
        scroll={{ x: 900 }}
        size="middle"
      />
    </Card>
  );
};
