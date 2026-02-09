import { useState, useEffect, useCallback } from "react";
import {
  Card,
  Descriptions,
  Table,
  Tag,
  Button,
  Space,
  Input,
  App,
  Spin,
  Row,
  Col,
  Popconfirm,
  Image,
} from "antd";
import {
  ArrowLeftOutlined,
  CheckOutlined,
  SendOutlined,
  CarOutlined,
  GiftOutlined,
  CloseOutlined,
} from "@ant-design/icons";
import type { ColumnsType } from "antd/es/table";
import { useParams, useNavigate } from "react-router-dom";
import api from "../../lib/api";

interface ProductImage {
  id: number;
  url: string;
  urlThumbnail?: string;
  isMain: boolean;
}

interface Product {
  id: number;
  name: string;
  slug: string;
  images?: ProductImage[];
}

interface OrderItem {
  id: number;
  productId: number;
  quantity: number;
  unitPrice: number;
  totalPrice: number;
  product?: Product;
}

interface Order {
  id: number;
  orderNumber: string;
  status: string;
  subtotal: number;
  discountAmount: number;
  deliveryCost: number;
  totalPrice: number;
  promoCode?: string;
  deliveryMethod: string;
  deliveryAddress?: string;
  paymentMethod: string;
  isPaid: boolean;
  customerName: string;
  customerPhone: string;
  customerEmail?: string;
  trackingNumber?: string;
  notes?: string;
  items: OrderItem[];
  createdAt: string;
  updatedAt: string;
}

const STATUS_CONFIG: Record<string, { color: string; label: string }> = {
  new: { color: "gold", label: "Новый" },
  confirmed: { color: "blue", label: "Подтверждён" },
  processing: { color: "cyan", label: "В обработке" },
  shipped: { color: "purple", label: "Отправлен" },
  delivered: { color: "green", label: "Доставлен" },
  cancelled: { color: "red", label: "Отменён" },
};

const TRANSITIONS: Record<
  string,
  { status: string; label: string; icon: React.ReactNode; color?: string }[]
> = {
  new: [
    { status: "confirmed", label: "Подтвердить", icon: <CheckOutlined /> },
    {
      status: "cancelled",
      label: "Отменить",
      icon: <CloseOutlined />,
      color: "danger",
    },
  ],
  confirmed: [
    { status: "processing", label: "В обработку", icon: <SendOutlined /> },
    {
      status: "cancelled",
      label: "Отменить",
      icon: <CloseOutlined />,
      color: "danger",
    },
  ],
  processing: [
    { status: "shipped", label: "Отправлен", icon: <CarOutlined /> },
  ],
  shipped: [
    { status: "delivered", label: "Доставлен", icon: <GiftOutlined /> },
  ],
};

export const OrderDetail = () => {
  const { message } = App.useApp();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [order, setOrder] = useState<Order | null>(null);
  const [loading, setLoading] = useState(true);
  const [statusLoading, setStatusLoading] = useState(false);
  const [trackingNumber, setTrackingNumber] = useState("");
  const [trackingLoading, setTrackingLoading] = useState(false);

  const fetchOrder = useCallback(async () => {
    setLoading(true);
    try {
      const { data: resp } = await api.get(`/admin/orders/${id}`);
      setOrder(resp.data);
      setTrackingNumber(resp.data.trackingNumber || "");
    } catch {
      message.error("Не удалось загрузить заказ");
    } finally {
      setLoading(false);
    }
  }, [id, message]);

  useEffect(() => {
    fetchOrder();
  }, [fetchOrder]);

  const handleStatusChange = async (newStatus: string) => {
    setStatusLoading(true);
    try {
      await api.put(`/admin/orders/${id}/status`, { status: newStatus });
      message.success("Статус обновлён");
      fetchOrder();
    } catch (err: unknown) {
      const axiosErr = err as {
        response?: { data?: { error?: { message?: string } } };
      };
      message.error(
        axiosErr.response?.data?.error?.message || "Ошибка обновления статуса"
      );
    } finally {
      setStatusLoading(false);
    }
  };

  const handleTrackingSave = async () => {
    if (!trackingNumber.trim()) return;
    setTrackingLoading(true);
    try {
      await api.put(`/admin/orders/${id}/tracking`, {
        trackingNumber: trackingNumber.trim(),
      });
      message.success("Трек-номер сохранён");
      fetchOrder();
    } catch {
      message.error("Не удалось сохранить трек-номер");
    } finally {
      setTrackingLoading(false);
    }
  };

  const formatPrice = (price: number) =>
    new Intl.NumberFormat("ru-RU", {
      style: "currency",
      currency: "RUB",
      maximumFractionDigits: 0,
    }).format(price);

  const formatDate = (dateStr: string) =>
    new Date(dateStr).toLocaleDateString("ru-RU", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });

  if (loading) {
    return (
      <div style={{ textAlign: "center", padding: 64 }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!order) {
    return (
      <Card>
        <div style={{ textAlign: "center", padding: 32 }}>
          Заказ не найден
        </div>
      </Card>
    );
  }

  const statusCfg = STATUS_CONFIG[order.status] || {
    color: "default",
    label: order.status,
  };
  const transitions = TRANSITIONS[order.status] || [];

  const itemColumns: ColumnsType<OrderItem> = [
    {
      title: "Фото",
      key: "image",
      width: 60,
      render: (_: unknown, record: OrderItem) => {
        const img =
          record.product?.images?.find((i) => i.isMain) ||
          record.product?.images?.[0];
        return img ? (
          <Image
            src={img.urlThumbnail || img.url}
            alt=""
            width={40}
            height={40}
            style={{ objectFit: "cover", borderRadius: 4 }}
            preview={false}
          />
        ) : (
          <div
            style={{
              width: 40,
              height: 40,
              background: "#f0f0f0",
              borderRadius: 4,
            }}
          />
        );
      },
    },
    {
      title: "Товар",
      key: "product",
      render: (_: unknown, record: OrderItem) => (
        <span>{record.product?.name || `ID: ${record.productId}`}</span>
      ),
    },
    {
      title: "Цена",
      dataIndex: "unitPrice",
      key: "unitPrice",
      width: 120,
      render: (p: number) => formatPrice(p),
    },
    {
      title: "Кол-во",
      dataIndex: "quantity",
      key: "quantity",
      width: 80,
      align: "center",
    },
    {
      title: "Итого",
      dataIndex: "totalPrice",
      key: "totalPrice",
      width: 120,
      render: (p: number) => <strong>{formatPrice(p)}</strong>,
    },
  ];

  return (
    <div>
      <Button
        type="text"
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate("/orders")}
        style={{ marginBottom: 16 }}
      >
        Назад к заказам
      </Button>

      {/* Header */}
      <Card style={{ marginBottom: 16 }}>
        <Row justify="space-between" align="middle">
          <Col>
            <Space size="middle">
              <span
                style={{
                  fontSize: 20,
                  fontWeight: 600,
                  fontFamily: "monospace",
                }}
              >
                {order.orderNumber}
              </span>
              <Tag color={statusCfg.color} style={{ fontSize: 14 }}>
                {statusCfg.label}
              </Tag>
            </Space>
            <div style={{ color: "#999", marginTop: 4, fontSize: 13 }}>
              Создан: {formatDate(order.createdAt)}
              {order.updatedAt !== order.createdAt &&
                ` · Обновлён: ${formatDate(order.updatedAt)}`}
            </div>
          </Col>
          <Col>
            <Space>
              {transitions.map((t) =>
                t.color === "danger" ? (
                  <Popconfirm
                    key={t.status}
                    title="Отменить заказ?"
                    onConfirm={() => handleStatusChange(t.status)}
                    okText="Да"
                    cancelText="Нет"
                    okButtonProps={{ danger: true }}
                  >
                    <Button
                      danger
                      icon={t.icon}
                      loading={statusLoading}
                    >
                      {t.label}
                    </Button>
                  </Popconfirm>
                ) : (
                  <Button
                    key={t.status}
                    type="primary"
                    icon={t.icon}
                    loading={statusLoading}
                    onClick={() => handleStatusChange(t.status)}
                  >
                    {t.label}
                  </Button>
                )
              )}
            </Space>
          </Col>
        </Row>
      </Card>

      <Row gutter={16}>
        {/* Left column */}
        <Col xs={24} lg={16}>
          {/* Items */}
          <Card title="Товары" style={{ marginBottom: 16 }}>
            <Table
              columns={itemColumns}
              dataSource={order.items}
              rowKey="id"
              pagination={false}
              size="small"
            />
            <div
              style={{
                marginTop: 16,
                textAlign: "right",
                borderTop: "1px solid #f0f0f0",
                paddingTop: 12,
              }}
            >
              <div style={{ color: "#666" }}>
                Подитог: {formatPrice(order.subtotal)}
              </div>
              {order.discountAmount > 0 && (
                <div style={{ color: "#52c41a" }}>
                  Скидка {order.promoCode && `(${order.promoCode})`}:{" "}
                  -{formatPrice(order.discountAmount)}
                </div>
              )}
              {order.deliveryCost > 0 && (
                <div style={{ color: "#666" }}>
                  Доставка: {formatPrice(order.deliveryCost)}
                </div>
              )}
              <div
                style={{ fontSize: 18, fontWeight: 600, marginTop: 4 }}
              >
                Итого: {formatPrice(order.totalPrice)}
              </div>
            </div>
          </Card>

          {/* Tracking */}
          {(order.status === "processing" ||
            order.status === "shipped" ||
            order.status === "delivered") && (
            <Card title="Трек-номер" style={{ marginBottom: 16 }}>
              <Space>
                <Input
                  placeholder="Введите трек-номер"
                  value={trackingNumber}
                  onChange={(e) => setTrackingNumber(e.target.value)}
                  style={{ width: 300 }}
                />
                <Button
                  type="primary"
                  loading={trackingLoading}
                  onClick={handleTrackingSave}
                  disabled={!trackingNumber.trim()}
                >
                  Сохранить
                </Button>
              </Space>
            </Card>
          )}
        </Col>

        {/* Right column */}
        <Col xs={24} lg={8}>
          {/* Customer */}
          <Card title="Клиент" style={{ marginBottom: 16 }}>
            <Descriptions column={1} size="small">
              <Descriptions.Item label="Имя">
                {order.customerName}
              </Descriptions.Item>
              <Descriptions.Item label="Телефон">
                <a href={`tel:${order.customerPhone}`}>
                  {order.customerPhone}
                </a>
              </Descriptions.Item>
              {order.customerEmail && (
                <Descriptions.Item label="Email">
                  <a href={`mailto:${order.customerEmail}`}>
                    {order.customerEmail}
                  </a>
                </Descriptions.Item>
              )}
            </Descriptions>
          </Card>

          {/* Delivery */}
          <Card title="Доставка" style={{ marginBottom: 16 }}>
            <Descriptions column={1} size="small">
              <Descriptions.Item label="Способ">
                {order.deliveryMethod === "courier"
                  ? "Курьерская доставка"
                  : "Самовывоз"}
              </Descriptions.Item>
              {order.deliveryAddress && (
                <Descriptions.Item label="Адрес">
                  {order.deliveryAddress}
                </Descriptions.Item>
              )}
              {order.trackingNumber && (
                <Descriptions.Item label="Трек-номер">
                  <Tag color="blue">{order.trackingNumber}</Tag>
                </Descriptions.Item>
              )}
            </Descriptions>
          </Card>

          {/* Payment */}
          <Card title="Оплата" style={{ marginBottom: 16 }}>
            <Descriptions column={1} size="small">
              <Descriptions.Item label="Способ">
                {order.paymentMethod === "card"
                  ? "Банковская карта"
                  : "Наличные"}
              </Descriptions.Item>
              <Descriptions.Item label="Статус">
                <Tag color={order.isPaid ? "green" : "orange"}>
                  {order.isPaid ? "Оплачен" : "Не оплачен"}
                </Tag>
              </Descriptions.Item>
            </Descriptions>
          </Card>

          {/* Notes */}
          {order.notes && (
            <Card title="Примечания">
              <div style={{ whiteSpace: "pre-wrap" }}>{order.notes}</div>
            </Card>
          )}
        </Col>
      </Row>
    </div>
  );
};
