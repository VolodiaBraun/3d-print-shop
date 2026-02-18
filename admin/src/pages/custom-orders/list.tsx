import { useState, useEffect, useCallback } from "react";
import { Table, Tag, Select, Card, Space, Button, Badge, App } from "antd";
import {
  ExperimentOutlined,
  EyeOutlined,
  PlusOutlined,
} from "@ant-design/icons";
import type { ColumnsType, TablePaginationConfig } from "antd/es/table";
import { useNavigate } from "react-router-dom";
import api from "../../lib/api";

interface CustomOrder {
  id: number;
  orderNumber: string;
  status: string;
  totalPrice: number;
  isPaid: boolean;
  paymentMethod: string;
  deliveryMethod: string;
  customerName: string;
  customerPhone: string;
  customerEmail?: string;
  customDetails?: {
    clientDescription?: string;
    bitrixDealId?: string;
    fileUrls: string[];
  };
  createdAt: string;
}

const STATUS_CONFIG: Record<string, { color: string; label: string }> = {
  new:         { color: "gold",    label: "Новая заявка" },
  confirmed:   { color: "blue",    label: "Подтверждён" },
  in_progress: { color: "cyan",    label: "В работе" },
  ready:       { color: "purple",  label: "Готов" },
  delivered:   { color: "green",   label: "Выдан" },
  cancelled:   { color: "red",     label: "Отменён" },
};

const STATUS_OPTIONS = [
  { value: "",            label: "Все статусы" },
  { value: "new",         label: "Новая заявка" },
  { value: "confirmed",   label: "Подтверждён" },
  { value: "in_progress", label: "В работе" },
  { value: "ready",       label: "Готов" },
  { value: "delivered",   label: "Выдан" },
  { value: "cancelled",   label: "Отменён" },
];

const PAYMENT_METHOD_LABEL: Record<string, string> = {
  card: "Карта",
  cash: "Наличные",
};

export const CustomOrderList = () => {
  const { message } = App.useApp();
  const navigate = useNavigate();
  const [orders, setOrders] = useState<CustomOrder[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [statusFilter, setStatusFilter] = useState("");

  const fetchOrders = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, string | number> = { page, limit: pageSize };
      if (statusFilter) params.status = statusFilter;
      const { data: resp } = await api.get("/admin/custom-orders", { params });
      setOrders(resp.data?.items || []);
      setTotal(resp.data?.total || 0);
    } catch {
      message.error("Не удалось загрузить индивидуальные заказы");
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, statusFilter, message]);

  useEffect(() => {
    fetchOrders();
  }, [fetchOrders]);

  const formatPrice = (price: number) =>
    price > 0
      ? new Intl.NumberFormat("ru-RU", {
          style: "currency",
          currency: "RUB",
          maximumFractionDigits: 0,
        }).format(price)
      : "—";

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

  const columns: ColumnsType<CustomOrder> = [
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
      render: (_: unknown, r: CustomOrder) => (
        <div>
          <div style={{ fontWeight: 500 }}>{r.customerName}</div>
          <div style={{ fontSize: 12, color: "#999" }}>{r.customerPhone}</div>
        </div>
      ),
    },
    {
      title: "Описание",
      key: "description",
      ellipsis: true,
      render: (_: unknown, r: CustomOrder) =>
        r.customDetails?.clientDescription ? (
          <span style={{ fontSize: 13, color: "#555" }}>
            {r.customDetails.clientDescription.slice(0, 60)}
            {r.customDetails.clientDescription.length > 60 ? "…" : ""}
          </span>
        ) : (
          <span style={{ color: "#bbb" }}>—</span>
        ),
    },
    {
      title: "Сумма",
      dataIndex: "totalPrice",
      key: "totalPrice",
      width: 130,
      render: (price: number) => (
        <span style={{ fontWeight: 500 }}>{formatPrice(price)}</span>
      ),
    },
    {
      title: "Статус",
      dataIndex: "status",
      key: "status",
      width: 140,
      render: (status: string) => {
        const cfg = STATUS_CONFIG[status] || { color: "default", label: status };
        return <Tag color={cfg.color}>{cfg.label}</Tag>;
      },
    },
    {
      title: "Оплата",
      key: "payment",
      width: 120,
      render: (_: unknown, r: CustomOrder) => (
        <Space direction="vertical" size={2}>
          <span style={{ fontSize: 12 }}>
            {PAYMENT_METHOD_LABEL[r.paymentMethod] || r.paymentMethod}
          </span>
          {r.isPaid ? (
            <Badge status="success" text="Оплачен" />
          ) : (
            <Badge status="warning" text="Не оплачен" />
          )}
        </Space>
      ),
    },
    {
      title: "Файлы",
      key: "files",
      width: 70,
      align: "center",
      render: (_: unknown, r: CustomOrder) => {
        const count = r.customDetails?.fileUrls?.length || 0;
        return count > 0 ? (
          <Tag>{count}</Tag>
        ) : (
          <span style={{ color: "#bbb" }}>—</span>
        );
      },
    },
    {
      title: "Bitrix",
      key: "bitrix",
      width: 90,
      render: (_: unknown, r: CustomOrder) =>
        r.customDetails?.bitrixDealId ? (
          <Tag color="geekblue">#{r.customDetails.bitrixDealId}</Tag>
        ) : (
          <span style={{ color: "#bbb" }}>—</span>
        ),
    },
    {
      title: "Дата",
      dataIndex: "createdAt",
      key: "createdAt",
      width: 140,
      render: (date: string) => (
        <span style={{ fontSize: 12, color: "#666" }}>{formatDate(date)}</span>
      ),
    },
    {
      title: "",
      key: "actions",
      width: 50,
      render: (_: unknown, r: CustomOrder) => (
        <Button
          type="text"
          size="small"
          icon={<EyeOutlined />}
          onClick={(e) => {
            e.stopPropagation();
            navigate(`/custom-orders/${r.id}`);
          }}
        />
      ),
    },
  ];

  return (
    <Card
      title={
        <Space>
          <ExperimentOutlined />
          <span>Индивидуальные заказы</span>
        </Space>
      }
      extra={
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => navigate("/custom-orders/create")}
        >
          Создать заказ
        </Button>
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
        onChange={(p: TablePaginationConfig) => {
          setPage(p.current || 1);
          setPageSize(p.pageSize || 20);
        }}
        onRow={(record) => ({
          onClick: () => navigate(`/custom-orders/${record.id}`),
          style: { cursor: "pointer" },
        })}
        scroll={{ x: 1100 }}
        size="middle"
      />
    </Card>
  );
};
